// Package api provides the HTTP server for gpuctl -- both public API and internal endpoints.
package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gpuai/gpuctl/internal/availability"
	"github.com/gpuai/gpuctl/internal/config"
	"github.com/gpuai/gpuctl/internal/db"
	"github.com/gpuai/gpuctl/internal/provision"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

// Server holds HTTP routes and all injected dependencies.
type Server struct {
	mux            *http.ServeMux
	db             *db.Pool
	redis          *redis.Client
	config         *config.Config
	engine         *provision.Engine
	statusBroker   *StatusBroker
	orgEventBroker *OrgEventBroker
	availCache     *availability.Cache
}

// ServerDeps contains the dependencies injected into the Server.
type ServerDeps struct {
	DB         *db.Pool
	Redis      *redis.Client
	Config     *config.Config
	Engine     *provision.Engine
	AvailCache *availability.Cache
}

// NewServer creates a Server, registers routes, and returns it.
func NewServer(deps ServerDeps) *Server {
	s := &Server{
		mux:            http.NewServeMux(),
		db:             deps.DB,
		redis:          deps.Redis,
		config:         deps.Config,
		engine:         deps.Engine,
		statusBroker:   NewStatusBroker(),
		orgEventBroker: NewOrgEventBroker(),
		availCache:     deps.AvailCache,
	}

	// Health endpoint behind localhost restriction + internal token auth
	s.mux.Handle("GET /health", LocalhostOnly(InternalAuthMiddleware(deps.Config.InternalAPIToken, http.HandlerFunc(s.handleHealth))))

	// Auth + rate limiting middleware chain.
	// 10 req/s sustained with burst of 20 per org.
	clerkAuth := ClerkAuthMiddleware(deps.Config.ClerkSecretKey)
	requireOrg := RequireOrg
	rateLimiter := NewOrgRateLimiter(rate.Every(100*time.Millisecond), 20)

	// Middleware chain helper: Clerk auth -> org required -> rate limiter.
	authChain := func(h http.Handler) http.Handler {
		return clerkAuth(requireOrg(rateLimiter.Middleware(h)))
	}

	// Start rate limiter cleanup goroutine.
	rateLimiter.StartCleanup(context.Background(), 5*time.Minute)

	// Instance CRUD routes with auth chain.
	idempotency := IdempotencyMiddleware(deps.DB)
	s.mux.Handle("POST /api/v1/instances",
		authChain(idempotency(http.HandlerFunc(s.handleCreateInstance))))
	s.mux.Handle("GET /api/v1/instances",
		authChain(http.HandlerFunc(s.handleListInstances)))
	s.mux.Handle("GET /api/v1/instances/{id}",
		authChain(http.HandlerFunc(s.handleGetInstance)))
	s.mux.Handle("DELETE /api/v1/instances/{id}",
		authChain(http.HandlerFunc(s.handleDeleteInstance)))

	// SSH key CRUD routes with auth chain.
	s.mux.Handle("POST /api/v1/ssh-keys",
		authChain(http.HandlerFunc(s.handleCreateSSHKey)))
	s.mux.Handle("GET /api/v1/ssh-keys",
		authChain(http.HandlerFunc(s.handleListSSHKeys)))
	s.mux.Handle("DELETE /api/v1/ssh-keys/{id}",
		authChain(http.HandlerFunc(s.handleDeleteSSHKey)))

	// Billing endpoints.
	s.mux.Handle("GET /api/v1/billing/usage",
		authChain(http.HandlerFunc(s.handleGetUsage)))
	s.mux.Handle("PUT /api/v1/billing/spending-limit",
		authChain(http.HandlerFunc(s.handleSetSpendingLimit)))
	s.mux.Handle("GET /api/v1/billing/spending-limit",
		authChain(http.HandlerFunc(s.handleGetSpendingLimit)))
	s.mux.Handle("DELETE /api/v1/billing/spending-limit",
		authChain(http.HandlerFunc(s.handleDeleteSpendingLimit)))

	// GPU availability endpoint.
	s.mux.Handle("GET /api/v1/gpu/available",
		authChain(http.HandlerFunc(s.handleListGPUAvailability)))

	// SSE status streaming.
	s.mux.Handle("GET /api/v1/instances/{id}/events",
		authChain(http.HandlerFunc(s.handleInstanceSSE)))

	// Per-org events: SSE streaming + REST catch-up.
	s.mux.Handle("GET /api/v1/events",
		authChain(http.HandlerFunc(s.handleEvents)))

	// Internal callbacks (per-instance token auth, reachable from GPU instances).
	s.mux.Handle("POST /internal/instances/{id}/ready",
		InstanceTokenAuth(deps.DB, http.HandlerFunc(s.handleInstanceReady)))
	s.mux.Handle("POST /internal/instances/{id}/health",
		InstanceTokenAuth(deps.DB, http.HandlerFunc(s.handleInstanceHealth)))

	return s
}

// Handler returns the root HTTP handler for the server.
func (s *Server) Handler() http.Handler {
	return s.mux
}

// PublishOrgEvent publishes an instance event to per-org SSE subscribers.
// Called by the health monitor via OnEvent callback.
func (s *Server) PublishOrgEvent(event db.InstanceEvent) {
	payload := OrgEventPayload{
		EventID:    event.EventID,
		EventType:  event.EventType,
		InstanceID: event.InstanceID,
		Metadata:   event.Metadata,
		CreatedAt:  event.CreatedAt.Format(time.RFC3339),
	}
	s.orgEventBroker.Publish(event.OrgID, payload)
}

// PublishStatusChange publishes an instance status change event to SSE subscribers.
// Called by the provisioning engine via OnStatusChange callback.
func (s *Server) PublishStatusChange(instanceID, internalStatus string) {
	s.statusBroker.Publish(instanceID, StatusEvent{
		InstanceID:     instanceID,
		Status:         provision.ExternalState(internalStatus),
		InternalStatus: internalStatus,
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
	})
}

// healthResponse is the JSON body returned by the health endpoint.
type healthResponse struct {
	Status string `json:"status"`
	DB     string `json:"db"`
	Redis  string `json:"redis"`
}

// handleHealth checks database and Redis connectivity and returns a JSON status.
// Returns 200 if both are connected, 503 if either is degraded.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	dbStatus := "connected"
	if err := s.db.Ping(ctx); err != nil {
		dbStatus = "disconnected"
	}

	redisStatus := "connected"
	if err := s.redis.Ping(ctx).Err(); err != nil {
		redisStatus = "disconnected"
	}

	status := "ok"
	httpCode := http.StatusOK
	if dbStatus == "disconnected" || redisStatus == "disconnected" {
		status = "degraded"
		httpCode = http.StatusServiceUnavailable
	}

	resp := healthResponse{
		Status: status,
		DB:     dbStatus,
		Redis:  redisStatus,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	json.NewEncoder(w).Encode(resp)
}

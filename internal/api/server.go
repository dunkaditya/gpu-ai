// Package api provides the HTTP server for gpuctl -- both public API and internal endpoints.
package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gpuai/gpuctl/internal/config"
	"github.com/gpuai/gpuctl/internal/db"
	"github.com/gpuai/gpuctl/internal/provision"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

// Server holds HTTP routes and all injected dependencies.
type Server struct {
	mux          *http.ServeMux
	db           *db.Pool
	redis        *redis.Client
	config       *config.Config
	engine       *provision.Engine
	statusBroker *StatusBroker
}

// ServerDeps contains the dependencies injected into the Server.
type ServerDeps struct {
	DB     *db.Pool
	Redis  *redis.Client
	Config *config.Config
	Engine *provision.Engine
}

// NewServer creates a Server, registers routes, and returns it.
func NewServer(deps ServerDeps) *Server {
	s := &Server{
		mux:          http.NewServeMux(),
		db:           deps.DB,
		redis:        deps.Redis,
		config:       deps.Config,
		engine:       deps.Engine,
		statusBroker: NewStatusBroker(),
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

	// SSE status streaming.
	s.mux.Handle("GET /api/v1/instances/{id}/events",
		authChain(http.HandlerFunc(s.handleInstanceSSE)))

	// Internal callback (localhost + internal token, NOT Clerk auth).
	s.mux.Handle("POST /internal/instances/{id}/ready",
		LocalhostOnly(InternalAuthMiddleware(deps.Config.InternalAPIToken,
			http.HandlerFunc(s.handleInstanceReady))))
	s.mux.Handle("POST /internal/instances/{id}/health",
		LocalhostOnly(InternalAuthMiddleware(deps.Config.InternalAPIToken,
			http.HandlerFunc(s.handleInstanceHealth))))

	return s
}

// Handler returns the root HTTP handler for the server.
func (s *Server) Handler() http.Handler {
	return s.mux
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

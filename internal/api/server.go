// Package api provides the HTTP server for gpuctl -- both public API and internal endpoints.
package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gpuai/gpuctl/internal/config"
	"github.com/gpuai/gpuctl/internal/db"
	"github.com/redis/go-redis/v9"
)

// Server holds HTTP routes and Phase 1 dependencies.
type Server struct {
	mux    *http.ServeMux
	db     *db.Pool
	redis  *redis.Client
	config *config.Config
}

// ServerDeps contains the dependencies injected into the Server.
type ServerDeps struct {
	DB     *db.Pool
	Redis  *redis.Client
	Config *config.Config
}

// NewServer creates a Server, registers routes, and returns it.
func NewServer(deps ServerDeps) *Server {
	s := &Server{
		mux:    http.NewServeMux(),
		db:     deps.DB,
		redis:  deps.Redis,
		config: deps.Config,
	}

	// Health endpoint behind internal token auth
	s.mux.Handle("GET /health", InternalAuthMiddleware(deps.Config.InternalAPIToken, http.HandlerFunc(s.handleHealth)))

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

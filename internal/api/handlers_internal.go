package api

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gpuai/gpuctl/internal/db"
	"github.com/gpuai/gpuctl/internal/provision"
)

// handleInstanceReady handles POST /internal/instances/{id}/ready.
// Called by cloud-init when an instance boots successfully.
// Transitions booting -> running and publishes SSE event.
func (s *Server) handleInstanceReady(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Extract instance ID from path.
	instanceID := r.PathValue("id")
	if instanceID == "" {
		writeProblem(w, http.StatusBadRequest, "missing-id", "Instance ID is required")
		return
	}

	// 2. Verify internal token matches the instance's stored token.
	inst, err := s.db.GetInstance(ctx, instanceID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeProblem(w, http.StatusNotFound, "not-found", "Instance not found")
			return
		}
		slog.Error("failed to get instance for ready callback",
			slog.String("instance_id", instanceID),
			slog.String("error", err.Error()),
		)
		writeProblem(w, http.StatusInternalServerError, "internal-error",
			"Failed to process callback")
		return
	}

	// The internal auth middleware already checked the Authorization header
	// against the global internal token. We also verify the instance-specific
	// token from the request body or query parameter if provided.
	instanceToken := r.URL.Query().Get("token")
	if instanceToken != "" && inst.InternalToken != nil && instanceToken != *inst.InternalToken {
		writeProblem(w, http.StatusForbidden, "forbidden",
			"Instance token mismatch")
		return
	}

	// 3. Atomically transition booting -> running.
	updated, err := s.db.SetInstanceRunning(ctx, instanceID)
	if err != nil {
		slog.Error("failed to set instance running",
			slog.String("instance_id", instanceID),
			slog.String("error", err.Error()),
		)
		writeProblem(w, http.StatusInternalServerError, "internal-error",
			"Failed to update instance status")
		return
	}

	// 4. If successfully transitioned, publish SSE event.
	if updated {
		slog.Info("instance ready",
			slog.String("instance_id", instanceID),
		)
		s.statusBroker.Publish(instanceID, StatusEvent{
			InstanceID:     instanceID,
			Status:         provision.ExternalState(provision.StateRunning),
			InternalStatus: provision.StateRunning,
			Timestamp:      time.Now().UTC().Format(time.RFC3339),
		})
	} else {
		// Already transitioned or in wrong state -- idempotent, log warning.
		slog.Warn("instance ready callback ignored (already transitioned or not booting)",
			slog.String("instance_id", instanceID),
			slog.String("current_status", inst.Status),
		)
	}

	// 5. Return 200 OK regardless (idempotent).
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleInstanceHealth handles POST /internal/instances/{id}/health.
// Called by instance health pings. Logs the health ping for Phase 6 monitoring.
func (s *Server) handleInstanceHealth(w http.ResponseWriter, r *http.Request) {
	// 1. Extract instance ID from path.
	instanceID := r.PathValue("id")
	if instanceID == "" {
		writeProblem(w, http.StatusBadRequest, "missing-id", "Instance ID is required")
		return
	}

	// 2. Verify instance exists.
	inst, err := s.db.GetInstance(r.Context(), instanceID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeProblem(w, http.StatusNotFound, "not-found", "Instance not found")
			return
		}
		slog.Error("failed to get instance for health ping",
			slog.String("instance_id", instanceID),
			slog.String("error", err.Error()),
		)
		writeProblem(w, http.StatusInternalServerError, "internal-error",
			"Failed to process health ping")
		return
	}

	// 3. Verify instance-specific token if provided.
	instanceToken := r.URL.Query().Get("token")
	if instanceToken != "" && inst.InternalToken != nil && instanceToken != *inst.InternalToken {
		writeProblem(w, http.StatusForbidden, "forbidden",
			"Instance token mismatch")
		return
	}

	// 4. Log the health ping (full last_seen update deferred to Phase 6).
	slog.Debug("health ping received",
		slog.String("instance_id", instanceID),
	)

	// 5. Return 200 OK.
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

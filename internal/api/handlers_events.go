package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gpuai/gpuctl/internal/auth"
	"github.com/gpuai/gpuctl/internal/db"
)

// handleEvents handles GET /api/v1/events.
// If ?since= is present, serves REST catch-up response (JSON array of events).
// Otherwise, serves SSE stream of per-org instance events.
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "unauthenticated", "Valid authentication required")
		return
	}

	orgID, err := s.db.GetOrgIDByClerkOrgID(ctx, claims.OrgID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeProblem(w, http.StatusNotFound, "not-found", "Organization not found")
			return
		}
		slog.Error("failed to look up org for events", slog.String("error", err.Error()))
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to process request")
		return
	}

	// If ?since= is present, serve REST catch-up response.
	if sinceStr := r.URL.Query().Get("since"); sinceStr != "" {
		s.handleListEventsREST(w, r, orgID, sinceStr)
		return
	}

	// Otherwise, serve SSE stream.
	s.handleOrgSSEStream(w, r, orgID)
}

// handleListEventsREST serves REST catch-up for GET /api/v1/events?since=<RFC3339>.
// Returns events from instance_events table as a JSON array.
func (s *Server) handleListEventsREST(w http.ResponseWriter, r *http.Request, orgID, sinceStr string) {
	ctx := r.Context()

	since, err := time.Parse(time.RFC3339, sinceStr)
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid-since",
			"since parameter must be a valid RFC3339 timestamp")
		return
	}

	// Parse optional limit (default 100, max 500).
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed < 1 {
			writeProblem(w, http.StatusBadRequest, "invalid-limit",
				"limit must be a positive integer")
			return
		}
		if parsed > 500 {
			parsed = 500
		}
		limit = parsed
	}

	events, err := s.db.ListInstanceEventsByOrg(ctx, orgID, since, limit)
	if err != nil {
		slog.Error("failed to list events", slog.String("error", err.Error()))
		writeProblem(w, http.StatusInternalServerError, "internal-error",
			"Failed to retrieve events")
		return
	}

	if events == nil {
		events = []db.InstanceEvent{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"events": events,
	})
}

// handleOrgSSEStream serves SSE for GET /api/v1/events (no ?since= param).
// Streams per-org instance events with 30s keepalive and 30min max duration.
func (s *Server) handleOrgSSEStream(w http.ResponseWriter, r *http.Request, orgID string) {
	// Check Flusher support.
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeProblem(w, http.StatusInternalServerError, "sse-not-supported",
			"Server-Sent Events not supported")
		return
	}

	ctx := r.Context()

	// Set SSE headers.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // for nginx proxies

	// Subscribe to org event broker.
	ch := s.orgEventBroker.Subscribe(orgID)
	defer s.orgEventBroker.Unsubscribe(orgID, ch)

	// Send initial connection comment.
	if _, err := fmt.Fprint(w, ": connected\n\n"); err != nil {
		return
	}
	flusher.Flush()

	// Set up keepalive ticker and max duration timer.
	keepalive := time.NewTicker(sseKeepaliveInterval)
	defer keepalive.Stop()

	maxDuration := time.NewTimer(maxSSEDuration)
	defer maxDuration.Stop()

	// Event loop.
	for {
		select {
		case <-ctx.Done():
			// Client disconnected.
			return

		case event, ok := <-ch:
			if !ok {
				// Channel closed.
				return
			}
			if err := writeSSEEvent(w, event.EventType, event); err != nil {
				return // client disconnected
			}
			flusher.Flush()

		case <-keepalive.C:
			// Write keepalive comment.
			if _, err := fmt.Fprint(w, ": keepalive\n\n"); err != nil {
				return // client disconnected
			}
			flusher.Flush()

		case <-maxDuration.C:
			// Maximum connection duration reached.
			slog.Info("SSE max duration reached for org events, closing connection",
				slog.String("org_id", orgID),
			)
			return
		}
	}
}

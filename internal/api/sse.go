package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gpuai/gpuctl/internal/auth"
	"github.com/gpuai/gpuctl/internal/db"
	"github.com/gpuai/gpuctl/internal/provision"
)

// StatusEvent represents a real-time instance status update.
type StatusEvent struct {
	InstanceID     string `json:"instance_id"`
	Status         string `json:"status"`          // external state
	InternalStatus string `json:"internal_status"` // internal state
	Timestamp      string `json:"timestamp"`       // RFC 3339
}

// StatusBroker manages SSE subscribers for instance status updates.
type StatusBroker struct {
	mu          sync.RWMutex
	subscribers map[string][]chan StatusEvent // instance_id -> channels
}

// NewStatusBroker creates a new StatusBroker.
func NewStatusBroker() *StatusBroker {
	return &StatusBroker{
		subscribers: make(map[string][]chan StatusEvent),
	}
}

// Subscribe creates a new buffered channel for receiving status events
// for the given instance. Caller must Unsubscribe when done.
func (b *StatusBroker) Subscribe(instanceID string) chan StatusEvent {
	ch := make(chan StatusEvent, 10)
	b.mu.Lock()
	b.subscribers[instanceID] = append(b.subscribers[instanceID], ch)
	b.mu.Unlock()
	return ch
}

// Unsubscribe removes a channel from the broker and closes it.
func (b *StatusBroker) Unsubscribe(instanceID string, ch chan StatusEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	subs := b.subscribers[instanceID]
	for i, sub := range subs {
		if sub == ch {
			b.subscribers[instanceID] = append(subs[:i], subs[i+1:]...)
			close(ch)
			break
		}
	}
	// Clean up empty subscriber lists.
	if len(b.subscribers[instanceID]) == 0 {
		delete(b.subscribers, instanceID)
	}
}

// Publish sends a status event to all subscribers for the given instance.
// Non-blocking: slow subscribers will miss events rather than block the publisher.
func (b *StatusBroker) Publish(instanceID string, event StatusEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subscribers[instanceID] {
		select {
		case ch <- event:
		default:
			// Subscriber is slow, skip to avoid blocking.
		}
	}
}

// maxSSEDuration is the maximum duration for an SSE connection.
// After this, the connection is closed and the client can reconnect.
const maxSSEDuration = 30 * time.Minute

// sseKeepaliveInterval is the interval between keepalive comments.
const sseKeepaliveInterval = 30 * time.Second

// handleInstanceSSE handles GET /api/v1/instances/{id}/events.
// Streams real-time status updates via Server-Sent Events with keepalive pings.
func (s *Server) handleInstanceSSE(w http.ResponseWriter, r *http.Request) {
	// 1. Check Flusher support.
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeProblem(w, http.StatusInternalServerError, "sse-not-supported",
			"Server-Sent Events not supported")
		return
	}

	ctx := r.Context()

	// 2. Extract instance ID from path.
	instanceID := r.PathValue("id")
	if instanceID == "" {
		writeProblem(w, http.StatusBadRequest, "missing-id", "Instance ID is required")
		return
	}

	// 3. Verify org ownership.
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "unauthenticated",
			"Valid authentication required")
		return
	}

	orgID, err := s.db.GetOrgIDByClerkOrgID(ctx, claims.OrgID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeProblem(w, http.StatusNotFound, "not-found", "Instance not found")
			return
		}
		slog.Error("failed to look up org for SSE", slog.String("error", err.Error()))
		writeProblem(w, http.StatusInternalServerError, "internal-error",
			"Failed to process request")
		return
	}

	inst, err := s.db.GetInstanceForOrg(ctx, instanceID, orgID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeProblem(w, http.StatusNotFound, "not-found", "Instance not found")
			return
		}
		slog.Error("failed to verify instance for SSE", slog.String("error", err.Error()))
		writeProblem(w, http.StatusInternalServerError, "internal-error",
			"Failed to process request")
		return
	}

	// 4. Set SSE response headers.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // for nginx proxies

	// 5. Subscribe to status broker.
	ch := s.statusBroker.Subscribe(instanceID)
	defer s.statusBroker.Unsubscribe(instanceID, ch)

	// 6. Send current state immediately as first event.
	initialEvent := StatusEvent{
		InstanceID:     inst.InstanceID,
		Status:         provision.ExternalState(inst.Status),
		InternalStatus: inst.Status,
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
	}
	if err := writeSSEEvent(w, "status", initialEvent); err != nil {
		return // client disconnected
	}
	flusher.Flush()

	// 7. Set up keepalive ticker and max duration timer.
	keepalive := time.NewTicker(sseKeepaliveInterval)
	defer keepalive.Stop()

	maxDuration := time.NewTimer(maxSSEDuration)
	defer maxDuration.Stop()

	// 8. Event loop.
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
			if err := writeSSEEvent(w, "status", event); err != nil {
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
			// Maximum connection duration reached. Close connection.
			// Client can reconnect.
			slog.Info("SSE max duration reached, closing connection",
				slog.String("instance_id", instanceID),
			)
			return
		}
	}
}

// writeSSEEvent writes a single SSE event to the response writer.
func writeSSEEvent(w http.ResponseWriter, eventType string, data any) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, string(jsonBytes))
	return err
}

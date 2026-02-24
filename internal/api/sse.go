package api

import (
	"net/http"
	"sync"
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

// handleInstanceSSE is a placeholder for Task 2.
func (s *Server) handleInstanceSSE(w http.ResponseWriter, r *http.Request) {
	// Implemented in Task 2.
}

package db

import (
	"context"
	"encoding/json"
	"time"
)

// InstanceEvent represents an event in the instance lifecycle.
type InstanceEvent struct {
	EventID    string          `json:"event_id"`
	InstanceID string          `json:"instance_id"`
	OrgID      string          `json:"org_id"`
	EventType  string          `json:"event_type"` // ready, interrupted, failed, terminated
	Metadata   json.RawMessage `json:"metadata,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

// CreateInstanceEvent inserts a new event into the instance_events table.
func (p *Pool) CreateInstanceEvent(ctx context.Context, event *InstanceEvent) error {
	return p.pool.QueryRow(ctx,
		`INSERT INTO instance_events (instance_id, org_id, event_type, metadata)
		 VALUES ($1, $2, $3, $4)
		 RETURNING event_id, created_at`,
		event.InstanceID, event.OrgID, event.EventType, event.Metadata,
	).Scan(&event.EventID, &event.CreatedAt)
}

// ListInstanceEventsByOrg returns events for an org since a given timestamp.
// Used by the REST catch-up endpoint GET /api/v1/events?since=<timestamp>.
// Results ordered by created_at ASC for chronological replay.
func (p *Pool) ListInstanceEventsByOrg(ctx context.Context, orgID string, since time.Time, limit int) ([]InstanceEvent, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT event_id, instance_id, org_id, event_type, metadata, created_at
		 FROM instance_events
		 WHERE org_id = $1 AND created_at > $2
		 ORDER BY created_at ASC
		 LIMIT $3`,
		orgID, since, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []InstanceEvent
	for rows.Next() {
		var e InstanceEvent
		if err := rows.Scan(&e.EventID, &e.InstanceID, &e.OrgID, &e.EventType, &e.Metadata, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// ListInstanceEventsByInstance returns all events for a specific instance.
// Used for debugging and instance detail views.
func (p *Pool) ListInstanceEventsByInstance(ctx context.Context, instanceID string) ([]InstanceEvent, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT event_id, instance_id, org_id, event_type, metadata, created_at
		 FROM instance_events
		 WHERE instance_id = $1
		 ORDER BY created_at ASC`,
		instanceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []InstanceEvent
	for rows.Next() {
		var e InstanceEvent
		if err := rows.Scan(&e.EventID, &e.InstanceID, &e.OrgID, &e.EventType, &e.Metadata, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

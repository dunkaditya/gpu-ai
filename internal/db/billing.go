package db

import (
	"context"
	"time"
)

// BillingSession represents a billing session record from the database.
// A billing session is created when an instance transitions to booting
// and closed when the instance is terminated.
type BillingSession struct {
	BillingSessionID      string
	InstanceID            string
	OrgID                 string
	GPUType               string
	GPUCount              int
	PricePerHour          float64
	StartedAt             time.Time
	EndedAt               *time.Time
	DurationSeconds       *int64
	TotalCost             *float64
	StripeReportedSeconds int64
	CreatedAt             time.Time
}

// CreateBillingSession inserts a new billing session record.
// The BillingSessionID and CreatedAt fields are populated from the RETURNING clause.
func (p *Pool) CreateBillingSession(ctx context.Context, session *BillingSession) error {
	return p.pool.QueryRow(ctx, `
		INSERT INTO billing_sessions (
			instance_id, org_id, gpu_type, gpu_count, price_per_hour, started_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING billing_session_id, created_at`,
		session.InstanceID, session.OrgID, session.GPUType,
		session.GPUCount, session.PricePerHour, session.StartedAt,
	).Scan(&session.BillingSessionID, &session.CreatedAt)
}

// CloseBillingSession closes the open billing session for an instance.
// It computes duration and cost at close time using CEIL for sub-second rounding.
// Returns nil even if no open session is found (idempotent for failed provisions).
func (p *Pool) CloseBillingSession(ctx context.Context, instanceID string, endedAt time.Time) error {
	_, err := p.pool.Exec(ctx, `
		UPDATE billing_sessions
		SET
			ended_at = $2,
			duration_seconds = CEIL(EXTRACT(EPOCH FROM ($2 - started_at)))::BIGINT,
			total_cost = CEIL(EXTRACT(EPOCH FROM ($2 - started_at))) / 3600.0 * price_per_hour * gpu_count
		WHERE instance_id = $1 AND ended_at IS NULL`,
		instanceID, endedAt,
	)
	return err
}

// GetActiveBillingSessions returns all billing sessions that have not been closed.
// Used by the billing ticker to calculate current usage and report to Stripe.
func (p *Pool) GetActiveBillingSessions(ctx context.Context) ([]BillingSession, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT billing_session_id, instance_id, org_id, gpu_type, gpu_count,
		       price_per_hour, started_at, stripe_reported_seconds, created_at
		FROM billing_sessions
		WHERE ended_at IS NULL
		ORDER BY started_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []BillingSession
	for rows.Next() {
		var s BillingSession
		if err := rows.Scan(
			&s.BillingSessionID, &s.InstanceID, &s.OrgID, &s.GPUType, &s.GPUCount,
			&s.PricePerHour, &s.StartedAt, &s.StripeReportedSeconds, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// GetBillingSessionsByOrg returns billing sessions for an organization,
// optionally filtered by date range. Returns up to 500 sessions ordered by
// most recent first.
func (p *Pool) GetBillingSessionsByOrg(ctx context.Context, orgID string, startDate, endDate *time.Time) ([]BillingSession, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT billing_session_id, instance_id, org_id, gpu_type, gpu_count,
		       price_per_hour, started_at, ended_at, duration_seconds, total_cost,
		       stripe_reported_seconds, created_at
		FROM billing_sessions
		WHERE org_id = $1
		  AND ($2::TIMESTAMPTZ IS NULL OR started_at >= $2)
		  AND ($3::TIMESTAMPTZ IS NULL OR started_at <= $3)
		ORDER BY started_at DESC
		LIMIT 500`,
		orgID, startDate, endDate,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []BillingSession
	for rows.Next() {
		var s BillingSession
		if err := rows.Scan(
			&s.BillingSessionID, &s.InstanceID, &s.OrgID, &s.GPUType, &s.GPUCount,
			&s.PricePerHour, &s.StartedAt, &s.EndedAt, &s.DurationSeconds, &s.TotalCost,
			&s.StripeReportedSeconds, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// UpdateStripeReportedSeconds increments the stripe_reported_seconds for a
// billing session by the given additional seconds. Used by the billing ticker
// to track how many GPU-seconds have already been reported to Stripe.
func (p *Pool) UpdateStripeReportedSeconds(ctx context.Context, billingSessionID string, additionalSeconds int64) error {
	_, err := p.pool.Exec(ctx, `
		UPDATE billing_sessions
		SET stripe_reported_seconds = stripe_reported_seconds + $2
		WHERE billing_session_id = $1`,
		billingSessionID, additionalSeconds,
	)
	return err
}

// GetOrgMonthSpendCents calculates the current month spend in cents for an
// organization since the given cycle start time. Uses CEIL for sub-second
// rounding and returns cents (dollars * 100) to avoid float precision issues.
func (p *Pool) GetOrgMonthSpendCents(ctx context.Context, orgID string, cycleStart time.Time) (int64, error) {
	var cents int64
	err := p.pool.QueryRow(ctx, `
		SELECT COALESCE(
			SUM(
				CEIL(
					EXTRACT(EPOCH FROM (COALESCE(ended_at, NOW()) - started_at))
				) / 3600.0 * price_per_hour * gpu_count * 100
			)::BIGINT,
			0
		)
		FROM billing_sessions
		WHERE org_id = $1 AND started_at >= $2`,
		orgID, cycleStart,
	).Scan(&cents)
	return cents, err
}

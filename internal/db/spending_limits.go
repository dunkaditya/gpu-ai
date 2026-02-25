package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// SpendingLimit represents a per-org spending limit record from the database.
type SpendingLimit struct {
	SpendingLimitID       string
	OrgID                 string
	MonthlyLimitCents     int64
	CurrentMonthSpendCents int64
	BillingCycleStart     time.Time
	Notify80Sent          bool
	Notify95Sent          bool
	LimitReachedAt        *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// GetSpendingLimit retrieves the spending limit for an organization.
// Returns ErrNotFound if no spending limit is set for the org.
func (p *Pool) GetSpendingLimit(ctx context.Context, orgID string) (*SpendingLimit, error) {
	var sl SpendingLimit
	err := p.pool.QueryRow(ctx, `
		SELECT spending_limit_id, org_id, monthly_limit_cents, current_month_spend_cents,
		       billing_cycle_start, notify_80_sent, notify_95_sent, limit_reached_at,
		       created_at, updated_at
		FROM spending_limits
		WHERE org_id = $1`,
		orgID,
	).Scan(
		&sl.SpendingLimitID, &sl.OrgID, &sl.MonthlyLimitCents, &sl.CurrentMonthSpendCents,
		&sl.BillingCycleStart, &sl.Notify80Sent, &sl.Notify95Sent, &sl.LimitReachedAt,
		&sl.CreatedAt, &sl.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &sl, nil
}

// UpsertSpendingLimit creates or updates a spending limit for an organization.
// On conflict (org already has a limit), the monthly limit is updated.
func (p *Pool) UpsertSpendingLimit(ctx context.Context, orgID string, monthlyLimitCents int64) (*SpendingLimit, error) {
	var sl SpendingLimit
	err := p.pool.QueryRow(ctx, `
		INSERT INTO spending_limits (org_id, monthly_limit_cents, billing_cycle_start)
		VALUES ($1, $2, date_trunc('month', NOW()))
		ON CONFLICT (org_id) DO UPDATE
		SET monthly_limit_cents = $2,
		    updated_at = NOW()
		RETURNING spending_limit_id, org_id, monthly_limit_cents, current_month_spend_cents,
		          billing_cycle_start, notify_80_sent, notify_95_sent, limit_reached_at,
		          created_at, updated_at`,
		orgID, monthlyLimitCents,
	).Scan(
		&sl.SpendingLimitID, &sl.OrgID, &sl.MonthlyLimitCents, &sl.CurrentMonthSpendCents,
		&sl.BillingCycleStart, &sl.Notify80Sent, &sl.Notify95Sent, &sl.LimitReachedAt,
		&sl.CreatedAt, &sl.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &sl, nil
}

// DeleteSpendingLimit removes the spending limit for an organization.
// Returns false if no limit was found to delete.
func (p *Pool) DeleteSpendingLimit(ctx context.Context, orgID string) (bool, error) {
	tag, err := p.pool.Exec(ctx, `
		DELETE FROM spending_limits WHERE org_id = $1`,
		orgID,
	)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// UpdateSpendingLimitFlags updates the notification flags and limit_reached_at
// timestamp for a spending limit. Used by the billing ticker to track threshold crossings.
func (p *Pool) UpdateSpendingLimitFlags(ctx context.Context, orgID string, notify80, notify95 bool, limitReachedAt *time.Time) error {
	_, err := p.pool.Exec(ctx, `
		UPDATE spending_limits
		SET notify_80_sent = $2,
		    notify_95_sent = $3,
		    limit_reached_at = $4,
		    updated_at = NOW()
		WHERE org_id = $1`,
		orgID, notify80, notify95, limitReachedAt,
	)
	return err
}

// ResetMonthlySpend resets the spending limit for a new billing cycle.
// Clears all notification flags, resets spend to 0, and updates the cycle start.
func (p *Pool) ResetMonthlySpend(ctx context.Context, orgID string) error {
	_, err := p.pool.Exec(ctx, `
		UPDATE spending_limits
		SET current_month_spend_cents = 0,
		    notify_80_sent = FALSE,
		    notify_95_sent = FALSE,
		    limit_reached_at = NULL,
		    billing_cycle_start = date_trunc('month', NOW()),
		    updated_at = NOW()
		WHERE org_id = $1`,
		orgID,
	)
	return err
}

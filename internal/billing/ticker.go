package billing

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/gpuai/gpuctl/internal/db"
	"github.com/gpuai/gpuctl/internal/provision"
)

// TickerDeps holds the dependencies for the billing ticker.
type TickerDeps struct {
	DB     *db.Pool
	Engine *provision.Engine
	Stripe *BillingService
	Logger *slog.Logger
}

// BillingTicker runs a 60-second loop that enforces spending limits and
// reports GPU-second usage to Stripe Billing Meters.
type BillingTicker struct {
	db     *db.Pool
	engine *provision.Engine
	stripe *BillingService
	logger *slog.Logger
}

// NewBillingTicker creates a new BillingTicker with the given dependencies.
func NewBillingTicker(deps TickerDeps) *BillingTicker {
	return &BillingTicker{
		db:     deps.DB,
		engine: deps.Engine,
		stripe: deps.Stripe,
		logger: deps.Logger,
	}
}

// Start runs the billing ticker loop. It fires every 60 seconds and stops
// when the context is cancelled. Intended to be run as a goroutine.
func (t *BillingTicker) Start(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			t.logger.Info("billing ticker stopped")
			return
		case <-ticker.C:
			t.runTick(ctx)
		}
	}
}

// runTick executes a single billing tick.
// CRITICAL ORDER: spending limits are enforced BEFORE Stripe reporting.
// This ensures limit enforcement is never delayed by Stripe API latency.
func (t *BillingTicker) runTick(ctx context.Context) {
	// Step 1: Get all active billing sessions.
	sessions, err := t.db.GetActiveBillingSessions(ctx)
	if err != nil {
		t.logger.Error("billing tick: failed to get active sessions",
			slog.String("error", err.Error()),
		)
		return
	}

	if len(sessions) == 0 {
		return
	}

	// Step 2: Group sessions by org.
	byOrg := make(map[string][]db.BillingSession)
	for _, s := range sessions {
		byOrg[s.OrgID] = append(byOrg[s.OrgID], s)
	}

	// Step 3: For each org -- LIMITS FIRST.
	for orgID, orgSessions := range byOrg {
		t.enforceSpendingLimit(ctx, orgID, orgSessions)
	}

	// Step 4: THEN Stripe reporting.
	t.reportToStripe(ctx, sessions)
}

// enforceSpendingLimit checks and enforces spending limits for a single org.
// Thresholds: 80% warning, 95% warning, 100% stop instances, 72h terminate.
func (t *BillingTicker) enforceSpendingLimit(ctx context.Context, orgID string, sessions []db.BillingSession) {
	limit, err := t.db.GetSpendingLimit(ctx, orgID)
	if errors.Is(err, db.ErrNotFound) {
		return // No limit set for this org, skip.
	}
	if err != nil {
		t.logger.Error("billing tick: failed to get spending limit",
			slog.String("org_id", orgID),
			slog.String("error", err.Error()),
		)
		return
	}

	// Check if billing cycle has rolled over to a new month.
	now := time.Now().UTC()
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	if limit.BillingCycleStart.Before(currentMonthStart) {
		if err := t.db.ResetMonthlySpend(ctx, orgID); err != nil {
			t.logger.Error("billing tick: failed to reset monthly spend",
				slog.String("org_id", orgID),
				slog.String("error", err.Error()),
			)
			return
		}
		// Re-read after reset to get clean state.
		limit, err = t.db.GetSpendingLimit(ctx, orgID)
		if err != nil {
			t.logger.Error("billing tick: failed to re-read spending limit after reset",
				slog.String("org_id", orgID),
				slog.String("error", err.Error()),
			)
			return
		}
	}

	// Get live spend from billing sessions.
	liveSpendCents, err := t.db.GetOrgMonthSpendCents(ctx, orgID, limit.BillingCycleStart)
	if err != nil {
		t.logger.Error("billing tick: failed to get org month spend",
			slog.String("org_id", orgID),
			slog.String("error", err.Error()),
		)
		return
	}

	if limit.MonthlyLimitCents <= 0 {
		return // Invalid or zero limit, skip.
	}

	pct := liveSpendCents * 100 / limit.MonthlyLimitCents

	// 72h auto-terminate: if limit was already reached and 72 hours have passed.
	if limit.LimitReachedAt != nil && time.Since(*limit.LimitReachedAt) > 72*time.Hour {
		t.logger.Warn("SPEND_LIMIT_AUTO_TERMINATE: terminating stopped instances 72h after limit reached",
			slog.String("org_id", orgID),
			slog.Int64("spend_cents", liveSpendCents),
			slog.Int64("limit_cents", limit.MonthlyLimitCents),
		)
		if err := t.engine.TerminateStoppedInstancesForOrg(ctx, orgID); err != nil {
			t.logger.Error("billing tick: failed to terminate stopped instances",
				slog.String("org_id", orgID),
				slog.String("error", err.Error()),
			)
		}
		return
	}

	// 100% threshold: stop all running instances.
	if pct >= 100 && limit.LimitReachedAt == nil {
		t.logger.Warn("SPEND_LIMIT_REACHED: stopping all running instances for org",
			slog.String("org_id", orgID),
			slog.Int64("spend_cents", liveSpendCents),
			slog.Int64("limit_cents", limit.MonthlyLimitCents),
		)
		if err := t.engine.StopInstancesForOrg(ctx, orgID); err != nil {
			t.logger.Error("billing tick: failed to stop instances for spending limit",
				slog.String("org_id", orgID),
				slog.String("error", err.Error()),
			)
		}
		nowUTC := time.Now().UTC()
		if err := t.db.UpdateSpendingLimitFlags(ctx, orgID, true, true, &nowUTC); err != nil {
			t.logger.Error("billing tick: failed to update spending limit flags at 100%",
				slog.String("org_id", orgID),
				slog.String("error", err.Error()),
			)
		}
		return
	}

	// 95% threshold: warning notification.
	if pct >= 95 && !limit.Notify95Sent {
		t.logger.Warn("SPEND_WARNING_95: org approaching spending limit (95%)",
			slog.String("org_id", orgID),
			slog.Int64("spend_cents", liveSpendCents),
			slog.Int64("limit_cents", limit.MonthlyLimitCents),
		)
		if err := t.db.UpdateSpendingLimitFlags(ctx, orgID, true, true, nil); err != nil {
			t.logger.Error("billing tick: failed to update spending limit flags at 95%",
				slog.String("org_id", orgID),
				slog.String("error", err.Error()),
			)
		}
		return
	}

	// 80% threshold: warning notification.
	if pct >= 80 && !limit.Notify80Sent {
		t.logger.Warn("SPEND_WARNING_80: org approaching spending limit (80%)",
			slog.String("org_id", orgID),
			slog.Int64("spend_cents", liveSpendCents),
			slog.Int64("limit_cents", limit.MonthlyLimitCents),
		)
		if err := t.db.UpdateSpendingLimitFlags(ctx, orgID, true, false, nil); err != nil {
			t.logger.Error("billing tick: failed to update spending limit flags at 80%",
				slog.String("org_id", orgID),
				slog.String("error", err.Error()),
			)
		}
	}
}

// reportToStripe aggregates unreported GPU-seconds by org and sends them to Stripe.
func (t *BillingTicker) reportToStripe(ctx context.Context, sessions []db.BillingSession) {
	if !t.stripe.Enabled() {
		return
	}

	now := time.Now().UTC()
	timestampBucket := now.Truncate(60 * time.Second)

	// Track unreported seconds per session for later DB update.
	type sessionDelta struct {
		billingSessionID string
		unreported       int64
	}

	// Aggregate by org.
	orgUnreported := make(map[string]int64)
	orgCustomerID := make(map[string]string)
	var deltas []sessionDelta

	for _, s := range sessions {
		elapsed := int64(now.Sub(s.StartedAt).Seconds())
		unreported := elapsed - s.StripeReportedSeconds
		if unreported <= 0 {
			continue
		}

		orgUnreported[s.OrgID] += unreported
		deltas = append(deltas, sessionDelta{
			billingSessionID: s.BillingSessionID,
			unreported:       unreported,
		})

		// Look up Stripe customer ID if not cached for this tick.
		if _, ok := orgCustomerID[s.OrgID]; !ok {
			custID, err := t.db.GetOrgStripeCustomerID(ctx, s.OrgID)
			if err != nil {
				t.logger.Error("billing tick: failed to get stripe customer ID",
					slog.String("org_id", s.OrgID),
					slog.String("error", err.Error()),
				)
				orgCustomerID[s.OrgID] = ""
			} else {
				orgCustomerID[s.OrgID] = custID
			}
		}
	}

	// Build meter event batches.
	var batches []MeterEventBatch
	for orgID, seconds := range orgUnreported {
		custID := orgCustomerID[orgID]
		if custID == "" || seconds <= 0 {
			continue
		}
		batches = append(batches, MeterEventBatch{
			StripeCustomerID: custID,
			GPUSeconds:       seconds,
			OrgID:            orgID,
			TimestampBucket:  timestampBucket,
		})
	}

	if len(batches) == 0 {
		return
	}

	// Report to Stripe.
	if err := t.stripe.ReportMeterEvents(ctx, batches); err != nil {
		t.logger.Error("billing tick: failed to report meter events",
			slog.String("error", err.Error()),
		)
		return
	}

	// Update reported seconds for each session.
	for _, d := range deltas {
		if err := t.db.UpdateStripeReportedSeconds(ctx, d.billingSessionID, d.unreported); err != nil {
			t.logger.Error("billing tick: failed to update stripe reported seconds",
				slog.String("billing_session_id", d.billingSessionID),
				slog.String("error", err.Error()),
			)
		}
	}
}

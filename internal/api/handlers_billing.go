package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gpuai/gpuctl/internal/auth"
	"github.com/gpuai/gpuctl/internal/billing"
	"github.com/gpuai/gpuctl/internal/db"
)

// BillingSessionResponse is the customer-facing JSON representation of a billing session.
type BillingSessionResponse struct {
	ID              string   `json:"id"`
	InstanceID      string   `json:"instance_id"`
	GPUType         string   `json:"gpu_type"`
	GPUCount        int      `json:"gpu_count"`
	PricePerHour    float64  `json:"price_per_hour"`
	StartedAt       string   `json:"started_at"`                       // RFC3339
	EndedAt         *string  `json:"ended_at,omitempty"`               // RFC3339, null if still running
	DurationSeconds *int64   `json:"duration_seconds,omitempty"`
	TotalCost       *float64 `json:"total_cost,omitempty"`             // null if still running
	EstimatedCost   *float64 `json:"estimated_cost,omitempty"`         // real-time for active sessions
	IsActive        bool     `json:"is_active"`
}

// UsageResponse is the JSON response for GET /api/v1/billing/usage (non-summary mode).
type UsageResponse struct {
	Sessions  []BillingSessionResponse `json:"sessions"`
	TotalCost float64                  `json:"total_cost"` // sum of completed + estimated
	Currency  string                   `json:"currency"`   // always "usd"
}

// HourlyBucket represents a single hourly aggregation bucket.
type HourlyBucket struct {
	Hour       string  `json:"hour"` // RFC3339 truncated to hour
	GPUSeconds int64   `json:"gpu_seconds"`
	Cost       float64 `json:"cost"`
}

// HourlyUsageResponse is the JSON response for GET /api/v1/billing/usage?summary=hourly.
type HourlyUsageResponse struct {
	Buckets   []HourlyBucket `json:"buckets"`
	TotalCost float64        `json:"total_cost"`
	Currency  string         `json:"currency"`
}

// SpendingLimitResponse is the JSON response for spending limit endpoints.
type SpendingLimitResponse struct {
	MonthlyLimitCents       int64   `json:"monthly_limit_cents"`
	MonthlyLimitDollars     float64 `json:"monthly_limit_dollars"`
	CurrentMonthSpendCents  int64   `json:"current_month_spend_cents"`
	CurrentMonthSpendDollars float64 `json:"current_month_spend_dollars"`
	PercentUsed             float64 `json:"percent_used"`
	BillingCycleStart       string  `json:"billing_cycle_start"` // RFC3339
	LimitReachedAt          *string `json:"limit_reached_at,omitempty"`
}

// SetSpendingLimitRequest is the JSON body for PUT /api/v1/billing/spending-limit.
type SetSpendingLimitRequest struct {
	MonthlyLimitDollars float64 `json:"monthly_limit_dollars"` // user-friendly input in dollars
}

// handleGetUsage handles GET /api/v1/billing/usage.
// Returns billing session records with optional date filtering and hourly aggregation.
func (s *Server) handleGetUsage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "unauthenticated", "Valid authentication required")
		return
	}

	orgID, err := s.db.GetOrgIDByClerkOrgID(ctx, claims.OrgID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			// Org not provisioned yet -- return empty usage.
			writeJSON(w, http.StatusOK, UsageResponse{
				Sessions:  []BillingSessionResponse{},
				TotalCost: 0,
				Currency:  "usd",
			})
			return
		}
		slog.Error("failed to look up org", slog.String("error", err.Error()))
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to process request")
		return
	}

	// Parse query params: period vs start/end (mutually exclusive).
	query := r.URL.Query()
	period := query.Get("period")
	startStr := query.Get("start")
	endStr := query.Get("end")
	summary := query.Get("summary")

	var startDate, endDate *time.Time

	if period != "" && (startStr != "" || endStr != "") {
		writeProblem(w, http.StatusBadRequest, "invalid-params",
			"'period' and 'start'/'end' are mutually exclusive")
		return
	}

	now := time.Now().UTC()

	if period != "" {
		switch period {
		case "current_month":
			start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
			startDate = &start
			endDate = &now
		case "last_30d":
			start := now.AddDate(0, 0, -30)
			startDate = &start
			endDate = &now
		default:
			writeProblem(w, http.StatusBadRequest, "invalid-params",
				"Invalid period. Accepted values: current_month, last_30d")
			return
		}
	}

	if startStr != "" {
		t, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "invalid-params",
				"Invalid 'start' parameter: must be RFC3339 format")
			return
		}
		startDate = &t
	}
	if endStr != "" {
		t, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "invalid-params",
				"Invalid 'end' parameter: must be RFC3339 format")
			return
		}
		endDate = &t
	}

	// Query billing sessions.
	sessions, err := s.db.GetBillingSessionsByOrg(ctx, orgID, startDate, endDate)
	if err != nil {
		slog.Error("failed to get billing sessions",
			slog.String("org_id", orgID),
			slog.String("error", err.Error()),
		)
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to retrieve billing usage")
		return
	}

	// Map sessions to response type and compute costs.
	respSessions := make([]BillingSessionResponse, 0, len(sessions))
	var totalCost float64

	for _, sess := range sessions {
		resp := BillingSessionResponse{
			ID:           sess.BillingSessionID,
			InstanceID:   sess.InstanceID,
			GPUType:      sess.GPUType,
			GPUCount:     sess.GPUCount,
			PricePerHour: sess.PricePerHour,
			StartedAt:    sess.StartedAt.Format(time.RFC3339),
		}

		if sess.EndedAt != nil {
			// Completed session.
			endStr := sess.EndedAt.Format(time.RFC3339)
			resp.EndedAt = &endStr
			resp.DurationSeconds = sess.DurationSeconds
			resp.TotalCost = sess.TotalCost
			resp.IsActive = false
			if sess.TotalCost != nil {
				totalCost += *sess.TotalCost
			}
		} else {
			// Active session: compute real-time estimated cost.
			elapsed := math.Ceil(time.Since(sess.StartedAt).Seconds())
			estimated := elapsed / 3600.0 * sess.PricePerHour * float64(sess.GPUCount)
			durationSec := int64(elapsed)
			resp.DurationSeconds = &durationSec
			resp.EstimatedCost = &estimated
			resp.IsActive = true
			totalCost += estimated
		}

		respSessions = append(respSessions, resp)
	}

	// If hourly summary mode, aggregate into buckets.
	if summary == "hourly" {
		buckets := aggregateHourlyBuckets(sessions, now)
		var bucketTotal float64
		for _, b := range buckets {
			bucketTotal += b.Cost
		}
		writeJSON(w, http.StatusOK, HourlyUsageResponse{
			Buckets:   buckets,
			TotalCost: bucketTotal,
			Currency:  "usd",
		})
		return
	}

	writeJSON(w, http.StatusOK, UsageResponse{
		Sessions:  respSessions,
		TotalCost: totalCost,
		Currency:  "usd",
	})
}

// aggregateHourlyBuckets groups billing sessions into hourly time buckets.
// For each session, it distributes GPU-seconds and cost across the hours the session spans.
func aggregateHourlyBuckets(sessions []db.BillingSession, now time.Time) []HourlyBucket {
	type bucket struct {
		gpuSeconds int64
		cost       float64
	}
	bucketMap := make(map[string]*bucket)

	for _, sess := range sessions {
		endTime := now
		if sess.EndedAt != nil {
			endTime = *sess.EndedAt
		}

		// Walk through each hour the session spans.
		cursor := sess.StartedAt.Truncate(time.Hour)
		for cursor.Before(endTime) {
			hourEnd := cursor.Add(time.Hour)

			// Compute overlap of session with this hour.
			overlapStart := sess.StartedAt
			if cursor.After(overlapStart) {
				overlapStart = cursor
			}
			overlapEnd := endTime
			if hourEnd.Before(overlapEnd) {
				overlapEnd = hourEnd
			}

			if overlapEnd.After(overlapStart) {
				seconds := int64(math.Ceil(overlapEnd.Sub(overlapStart).Seconds()))
				cost := float64(seconds) / 3600.0 * sess.PricePerHour * float64(sess.GPUCount)

				hourKey := cursor.Format(time.RFC3339)
				if bucketMap[hourKey] == nil {
					bucketMap[hourKey] = &bucket{}
				}
				bucketMap[hourKey].gpuSeconds += seconds * int64(sess.GPUCount)
				bucketMap[hourKey].cost += cost
			}

			cursor = hourEnd
		}
	}

	// Convert map to sorted slice.
	// Collect keys and sort.
	keys := make([]string, 0, len(bucketMap))
	for k := range bucketMap {
		keys = append(keys, k)
	}
	// Simple sort by RFC3339 string (lexicographic order works for ISO timestamps).
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	result := make([]HourlyBucket, 0, len(keys))
	for _, k := range keys {
		b := bucketMap[k]
		result = append(result, HourlyBucket{
			Hour:       k,
			GPUSeconds: b.gpuSeconds,
			Cost:       b.cost,
		})
	}
	return result
}

// handleSetSpendingLimit handles PUT /api/v1/billing/spending-limit.
// Creates or updates the spending limit for the authenticated organization.
func (s *Server) handleSetSpendingLimit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "unauthenticated", "Valid authentication required")
		return
	}

	orgID, _, err := s.db.EnsureOrgAndUser(ctx, claims.OrgID, claims.UserID, "")
	if err != nil {
		slog.Error("failed to ensure org and user",
			slog.String("clerk_org_id", claims.OrgID),
			slog.String("error", err.Error()),
		)
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to process request")
		return
	}

	var req SetSpendingLimitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid-request", "Invalid JSON request body")
		return
	}

	if req.MonthlyLimitDollars < 1.0 {
		writeProblem(w, http.StatusBadRequest, "validation-error",
			"monthly_limit_dollars must be at least 1.00")
		return
	}

	cents := int64(math.Round(req.MonthlyLimitDollars * 100))

	limit, err := s.db.UpsertSpendingLimit(ctx, orgID, cents)
	if err != nil {
		slog.Error("failed to upsert spending limit",
			slog.String("org_id", orgID),
			slog.String("error", err.Error()),
		)
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to set spending limit")
		return
	}

	writeJSON(w, http.StatusOK, spendingLimitToResponse(limit))
}

// handleGetSpendingLimit handles GET /api/v1/billing/spending-limit.
// Returns the spending limit and current usage for the authenticated organization.
func (s *Server) handleGetSpendingLimit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "unauthenticated", "Valid authentication required")
		return
	}

	orgID, err := s.db.GetOrgIDByClerkOrgID(ctx, claims.OrgID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeProblem(w, http.StatusNotFound, "no_spending_limit",
				"No spending limit configured for this organization")
			return
		}
		slog.Error("failed to look up org", slog.String("error", err.Error()))
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to process request")
		return
	}

	limit, err := s.db.GetSpendingLimit(ctx, orgID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeProblem(w, http.StatusNotFound, "no_spending_limit",
				"No spending limit configured for this organization")
			return
		}
		slog.Error("failed to get spending limit",
			slog.String("org_id", orgID),
			slog.String("error", err.Error()),
		)
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to retrieve spending limit")
		return
	}

	writeJSON(w, http.StatusOK, spendingLimitToResponse(limit))
}

// handleDeleteSpendingLimit handles DELETE /api/v1/billing/spending-limit.
// Removes the spending limit for the authenticated organization.
func (s *Server) handleDeleteSpendingLimit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "unauthenticated", "Valid authentication required")
		return
	}

	orgID, err := s.db.GetOrgIDByClerkOrgID(ctx, claims.OrgID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeProblem(w, http.StatusNotFound, "not-found", "Spending limit not found")
			return
		}
		slog.Error("failed to look up org", slog.String("error", err.Error()))
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to process request")
		return
	}

	deleted, err := s.db.DeleteSpendingLimit(ctx, orgID)
	if err != nil {
		slog.Error("failed to delete spending limit",
			slog.String("org_id", orgID),
			slog.String("error", err.Error()),
		)
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to delete spending limit")
		return
	}

	if !deleted {
		writeProblem(w, http.StatusNotFound, "not-found", "Spending limit not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// spendingLimitToResponse maps a db.SpendingLimit to a SpendingLimitResponse.
func spendingLimitToResponse(limit *db.SpendingLimit) SpendingLimitResponse {
	resp := SpendingLimitResponse{
		MonthlyLimitCents:       limit.MonthlyLimitCents,
		MonthlyLimitDollars:     float64(limit.MonthlyLimitCents) / 100.0,
		CurrentMonthSpendCents:  limit.CurrentMonthSpendCents,
		CurrentMonthSpendDollars: float64(limit.CurrentMonthSpendCents) / 100.0,
		BillingCycleStart:       limit.BillingCycleStart.Format(time.RFC3339),
	}

	// Compute percent used, clamped to max 100.0.
	if limit.MonthlyLimitCents > 0 {
		pct := float64(limit.CurrentMonthSpendCents) * 100.0 / float64(limit.MonthlyLimitCents)
		if pct > 100.0 {
			pct = 100.0
		}
		resp.PercentUsed = pct
	}

	if limit.LimitReachedAt != nil {
		s := limit.LimitReachedAt.Format(time.RFC3339)
		resp.LimitReachedAt = &s
	}

	return resp
}

// ── Credit Balance Handlers ──

// BalanceResponse is the JSON response for GET /api/v1/billing/balance.
type BalanceResponse struct {
	BalanceCents          int64   `json:"balance_cents"`
	BalanceDollars        float64 `json:"balance_dollars"`
	AutoPayEnabled        bool    `json:"auto_pay_enabled"`
	AutoPayThresholdCents int64   `json:"auto_pay_threshold_cents"`
	AutoPayAmountCents    int64   `json:"auto_pay_amount_cents"`
}

// handleGetBalance handles GET /api/v1/billing/balance.
func (s *Server) handleGetBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "unauthenticated", "Valid authentication required")
		return
	}

	orgID, _, err := s.db.EnsureOrgAndUser(ctx, claims.OrgID, claims.UserID, "")
	if err != nil {
		slog.Error("failed to ensure org", slog.String("error", err.Error()))
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to process request")
		return
	}

	bal, err := s.db.EnsureOrgBalance(ctx, orgID)
	if err != nil {
		slog.Error("failed to get balance", slog.String("error", err.Error()))
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to retrieve balance")
		return
	}

	writeJSON(w, http.StatusOK, BalanceResponse{
		BalanceCents:          bal.BalanceCents,
		BalanceDollars:        float64(bal.BalanceCents) / 100.0,
		AutoPayEnabled:        bal.AutoPayEnabled,
		AutoPayThresholdCents: bal.AutoPayThresholdCents,
		AutoPayAmountCents:    bal.AutoPayAmountCents,
	})
}

// PurchaseCreditsRequest is the JSON body for POST /api/v1/billing/credits/purchase.
type PurchaseCreditsRequest struct {
	AmountCents int64  `json:"amount_cents"`
	SuccessURL  string `json:"success_url"`
	CancelURL   string `json:"cancel_url"`
}

// PurchaseCreditsResponse is the JSON response for POST /api/v1/billing/credits/purchase.
type PurchaseCreditsResponse struct {
	CheckoutURL string `json:"checkout_url"`
	SessionID   string `json:"session_id"`
}

// handlePurchaseCredits handles POST /api/v1/billing/credits/purchase.
func (s *Server) handlePurchaseCredits(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if s.billing == nil || !s.billing.Enabled() {
		writeProblem(w, http.StatusServiceUnavailable, "stripe-not-configured",
			"Payment processing is not configured")
		return
	}

	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "unauthenticated", "Valid authentication required")
		return
	}

	orgID, _, err := s.db.EnsureOrgAndUser(ctx, claims.OrgID, claims.UserID, "")
	if err != nil {
		slog.Error("failed to ensure org", slog.String("error", err.Error()))
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to process request")
		return
	}

	var req PurchaseCreditsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid-request", "Invalid JSON request body")
		return
	}

	// Minimum $5 purchase.
	if req.AmountCents < 500 {
		writeProblem(w, http.StatusBadRequest, "validation-error", "Minimum purchase amount is $5.00")
		return
	}
	if req.SuccessURL == "" || req.CancelURL == "" {
		writeProblem(w, http.StatusBadRequest, "validation-error", "success_url and cancel_url are required")
		return
	}

	// Ensure Stripe customer exists.
	customerID, err := s.db.GetOrgStripeCustomerID(ctx, orgID)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		slog.Error("failed to get stripe customer", slog.String("error", err.Error()))
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to process request")
		return
	}

	if customerID == "" {
		email := "" // Clerk manages identity; email is optional
		customerID, err = s.billing.CreateOrGetCustomer(ctx, orgID, email)
		if err != nil {
			slog.Error("failed to create stripe customer", slog.String("error", err.Error()))
			writeProblem(w, http.StatusInternalServerError, "stripe-error", "Failed to create payment customer")
			return
		}
		if err := s.db.SetStripeCustomerID(ctx, orgID, customerID); err != nil {
			slog.Error("failed to save stripe customer id", slog.String("error", err.Error()))
			writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to process request")
			return
		}
	}

	checkoutURL, sessionID, err := s.billing.CreateCheckoutSession(ctx, customerID, req.AmountCents, orgID, req.SuccessURL, req.CancelURL)
	if err != nil {
		slog.Error("failed to create checkout session", slog.String("error", err.Error()))
		writeProblem(w, http.StatusInternalServerError, "stripe-error", "Failed to create checkout session")
		return
	}

	writeJSON(w, http.StatusOK, PurchaseCreditsResponse{
		CheckoutURL: checkoutURL,
		SessionID:   sessionID,
	})
}

// RedeemCodeRequest is the JSON body for POST /api/v1/billing/credits/redeem.
type RedeemCodeRequest struct {
	Code string `json:"code"`
}

// RedeemCodeResponse is the JSON response for POST /api/v1/billing/credits/redeem.
type RedeemCodeResponse struct {
	AmountCents    int64   `json:"amount_cents"`
	AmountDollars  float64 `json:"amount_dollars"`
	NewBalanceCents int64  `json:"new_balance_cents"`
}

// handleRedeemCreditCode handles POST /api/v1/billing/credits/redeem.
func (s *Server) handleRedeemCreditCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "unauthenticated", "Valid authentication required")
		return
	}

	orgID, _, err := s.db.EnsureOrgAndUser(ctx, claims.OrgID, claims.UserID, "")
	if err != nil {
		slog.Error("failed to ensure org", slog.String("error", err.Error()))
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to process request")
		return
	}

	var req RedeemCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid-request", "Invalid JSON request body")
		return
	}

	if req.Code == "" {
		writeProblem(w, http.StatusBadRequest, "validation-error", "code is required")
		return
	}

	code, newBalance, err := s.db.RedeemCreditCode(ctx, req.Code, orgID)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrCodeNotFound):
			writeProblem(w, http.StatusNotFound, "code-not-found", "Credit code not found")
		case errors.Is(err, db.ErrCodeAlreadyRedeemed):
			writeProblem(w, http.StatusConflict, "code-already-redeemed", "This credit code has already been redeemed")
		case errors.Is(err, db.ErrCodeExpired):
			writeProblem(w, http.StatusGone, "code-expired", "This credit code has expired")
		default:
			slog.Error("failed to redeem code", slog.String("error", err.Error()))
			writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to redeem code")
		}
		return
	}

	writeJSON(w, http.StatusOK, RedeemCodeResponse{
		AmountCents:    code.AmountCents,
		AmountDollars:  float64(code.AmountCents) / 100.0,
		NewBalanceCents: newBalance,
	})
}

// UpdateAutoPayRequest is the JSON body for PUT /api/v1/billing/auto-pay.
type UpdateAutoPayRequest struct {
	Enabled        bool  `json:"enabled"`
	ThresholdCents int64 `json:"threshold_cents"`
	AmountCents    int64 `json:"amount_cents"`
}

// handleUpdateAutoPay handles PUT /api/v1/billing/auto-pay.
func (s *Server) handleUpdateAutoPay(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "unauthenticated", "Valid authentication required")
		return
	}

	orgID, _, err := s.db.EnsureOrgAndUser(ctx, claims.OrgID, claims.UserID, "")
	if err != nil {
		slog.Error("failed to ensure org", slog.String("error", err.Error()))
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to process request")
		return
	}

	var req UpdateAutoPayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid-request", "Invalid JSON request body")
		return
	}

	if req.Enabled {
		if req.ThresholdCents < 0 {
			writeProblem(w, http.StatusBadRequest, "validation-error", "threshold_cents must be non-negative")
			return
		}
		if req.AmountCents < 500 {
			writeProblem(w, http.StatusBadRequest, "validation-error", "amount_cents must be at least 500 ($5.00)")
			return
		}
	}

	if err := s.db.UpdateAutoPay(ctx, orgID, req.Enabled, req.ThresholdCents, req.AmountCents); err != nil {
		slog.Error("failed to update auto-pay", slog.String("error", err.Error()))
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to update auto-pay settings")
		return
	}

	bal, err := s.db.EnsureOrgBalance(ctx, orgID)
	if err != nil {
		slog.Error("failed to get balance after auto-pay update", slog.String("error", err.Error()))
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to retrieve balance")
		return
	}

	writeJSON(w, http.StatusOK, BalanceResponse{
		BalanceCents:          bal.BalanceCents,
		BalanceDollars:        float64(bal.BalanceCents) / 100.0,
		AutoPayEnabled:        bal.AutoPayEnabled,
		AutoPayThresholdCents: bal.AutoPayThresholdCents,
		AutoPayAmountCents:    bal.AutoPayAmountCents,
	})
}

// TransactionResponse is the JSON representation of a single transaction.
type TransactionResponse struct {
	ID               string  `json:"id"`
	Type             string  `json:"type"`
	AmountCents      int64   `json:"amount_cents"`
	AmountDollars    float64 `json:"amount_dollars"`
	BalanceAfterCents int64  `json:"balance_after_cents"`
	Description      string  `json:"description"`
	ReferenceID      *string `json:"reference_id,omitempty"`
	CreatedAt        string  `json:"created_at"`
}

// TransactionsListResponse is the JSON response for GET /api/v1/billing/transactions.
type TransactionsListResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	HasMore      bool                  `json:"has_more"`
	NextCursor   string                `json:"next_cursor,omitempty"`
}

// handleGetTransactions handles GET /api/v1/billing/transactions.
func (s *Server) handleGetTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "unauthenticated", "Valid authentication required")
		return
	}

	orgID, err := s.db.GetOrgIDByClerkOrgID(ctx, claims.OrgID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeJSON(w, http.StatusOK, TransactionsListResponse{
				Transactions: []TransactionResponse{},
			})
			return
		}
		slog.Error("failed to look up org", slog.String("error", err.Error()))
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to process request")
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}
	before := r.URL.Query().Get("before")

	txns, hasMore, err := s.db.GetTransactions(ctx, orgID, limit, before)
	if err != nil {
		slog.Error("failed to get transactions", slog.String("error", err.Error()))
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to retrieve transactions")
		return
	}

	resp := TransactionsListResponse{
		Transactions: make([]TransactionResponse, 0, len(txns)),
		HasMore:      hasMore,
	}

	for _, t := range txns {
		resp.Transactions = append(resp.Transactions, TransactionResponse{
			ID:                t.TransactionID,
			Type:              t.Type,
			AmountCents:       t.AmountCents,
			AmountDollars:     float64(t.AmountCents) / 100.0,
			BalanceAfterCents: t.BalanceAfterCents,
			Description:       t.Description,
			ReferenceID:       t.ReferenceID,
			CreatedAt:         t.CreatedAt.Format(time.RFC3339),
		})
	}

	if hasMore && len(txns) > 0 {
		resp.NextCursor = txns[len(txns)-1].TransactionID
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleStripeWebhook handles POST /webhooks/stripe.
// No auth middleware — signature is verified via Stripe webhook secret.
func (s *Server) handleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if s.billing == nil || !s.billing.Enabled() {
		http.Error(w, "webhook not configured", http.StatusServiceUnavailable)
		return
	}

	payload, err := billing.ReadWebhookBody(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	sigHeader := r.Header.Get("Stripe-Signature")
	event, err := s.billing.VerifyWebhookSignature(payload, sigHeader)
	if err != nil {
		slog.Warn("stripe webhook signature verification failed", slog.String("error", err.Error()))
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		orgID, amountCents, sessionID, err := billing.ParseCheckoutSessionEvent(event)
		if err != nil {
			slog.Error("failed to parse checkout session event", slog.String("error", err.Error()))
			http.Error(w, "invalid event data", http.StatusBadRequest)
			return
		}

		// Deduplicate by reference_id.
		exists, err := s.db.TransactionExistsByReference(ctx, sessionID)
		if err != nil {
			slog.Error("failed to check transaction dedup", slog.String("error", err.Error()))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if exists {
			w.WriteHeader(http.StatusOK)
			return
		}

		ref := sessionID
		_, err = s.db.AddBalance(ctx, orgID, amountCents, "credit_purchase",
			fmt.Sprintf("Added $%.2f via Stripe Checkout", float64(amountCents)/100.0), &ref)
		if err != nil {
			slog.Error("failed to add balance from checkout", slog.String("error", err.Error()))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		slog.Info("checkout balance added",
			slog.String("org_id", orgID),
			slog.Int64("amount_cents", amountCents),
		)

	case "payment_intent.succeeded":
		orgID, amountCents, piID, piType, err := billing.ParsePaymentIntentEvent(event)
		if err != nil {
			slog.Error("failed to parse payment intent event", slog.String("error", err.Error()))
			http.Error(w, "invalid event data", http.StatusBadRequest)
			return
		}

		if piType == "auto_pay" {
			exists, err := s.db.TransactionExistsByReference(ctx, piID)
			if err != nil {
				slog.Error("failed to check transaction dedup", slog.String("error", err.Error()))
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			if exists {
				w.WriteHeader(http.StatusOK)
				return
			}

			ref := piID
			_, err = s.db.AddBalance(ctx, orgID, amountCents, "auto_pay",
				fmt.Sprintf("Auto-pay: added $%.2f", float64(amountCents)/100.0), &ref)
			if err != nil {
				slog.Error("failed to add auto-pay balance", slog.String("error", err.Error()))
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			slog.Info("auto-pay balance added",
				slog.String("org_id", orgID),
				slog.Int64("amount_cents", amountCents),
			)
		}

	case "payment_intent.payment_failed":
		orgID, _, _, piType, err := billing.ParsePaymentIntentEvent(event)
		if err != nil {
			slog.Warn("failed to parse failed payment intent", slog.String("error", err.Error()))
			break
		}
		if piType == "auto_pay" && orgID != "" {
			_ = s.db.UpdateAutoPayTriggered(ctx, orgID)
			slog.Warn("auto-pay payment failed", slog.String("org_id", orgID))
		}
	}

	w.WriteHeader(http.StatusOK)
}

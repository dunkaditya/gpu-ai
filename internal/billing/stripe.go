// Package billing handles Stripe integration for usage metering and payments.
package billing

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/billing/meterevent"
)

// BillingService handles Stripe Billing Meter event reporting.
// Billing sessions in PostgreSQL are the source of truth; Stripe is a reporting sink.
type BillingService struct {
	apiKey     string
	meterEvent string
	logger     *slog.Logger
	enabled    bool
}

// NewBillingService creates a BillingService.
// If apiKey is empty, the service is created but all operations are no-ops.
func NewBillingService(apiKey, meterEventName string, logger *slog.Logger) *BillingService {
	return &BillingService{
		apiKey:     apiKey,
		meterEvent: meterEventName,
		logger:     logger,
		enabled:    apiKey != "" && meterEventName != "",
	}
}

// Enabled returns whether the billing service is configured and active.
func (s *BillingService) Enabled() bool {
	return s.enabled
}

// MeterEventBatch represents aggregated GPU-seconds to report for one org.
type MeterEventBatch struct {
	StripeCustomerID string
	GPUSeconds       int64
	OrgID            string
	TimestampBucket  time.Time // truncated to 60s boundary for idempotency key
}

// ReportMeterEvents sends GPU-second usage to Stripe Billing Meters.
// Each batch is sent as a single meter event per org.
// If Stripe is not configured, this is a no-op.
func (s *BillingService) ReportMeterEvents(ctx context.Context, batches []MeterEventBatch) error {
	if !s.enabled {
		return nil
	}

	stripe.Key = s.apiKey

	for _, b := range batches {
		if b.GPUSeconds <= 0 {
			continue
		}
		identifier := fmt.Sprintf("%s:%d", b.OrgID, b.TimestampBucket.Unix())
		params := &stripe.BillingMeterEventParams{
			EventName: stripe.String(s.meterEvent),
			Payload: map[string]string{
				"stripe_customer_id": b.StripeCustomerID,
				"value":              fmt.Sprintf("%d", b.GPUSeconds),
			},
			Identifier: stripe.String(identifier),
			Timestamp:  stripe.Int64(b.TimestampBucket.Unix()),
		}
		if _, err := meterevent.New(params); err != nil {
			s.logger.Error("stripe meter event failed",
				slog.String("org_id", b.OrgID),
				slog.Int64("gpu_seconds", b.GPUSeconds),
				slog.String("error", err.Error()),
			)
			// Continue with other orgs; failed orgs retry on next tick.
		} else {
			s.logger.Info("stripe meter event reported",
				slog.String("org_id", b.OrgID),
				slog.Int64("gpu_seconds", b.GPUSeconds),
			)
		}
	}
	return nil
}

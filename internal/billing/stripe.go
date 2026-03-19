// Package billing handles Stripe integration for usage metering and payments.
package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"time"

	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/billing/meterevent"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/paymentintent"
	"github.com/stripe/stripe-go/v82/webhook"
)

// BillingService handles Stripe Billing Meter event reporting and payments.
// Billing sessions in PostgreSQL are the source of truth; Stripe is a reporting sink.
type BillingService struct {
	apiKey        string
	meterEvent    string
	webhookSecret string
	logger        *slog.Logger
	enabled       bool
}

// NewBillingService creates a BillingService.
// If apiKey is empty, the service is created but all operations are no-ops.
func NewBillingService(apiKey, meterEventName, webhookSecret string, logger *slog.Logger) *BillingService {
	return &BillingService{
		apiKey:        apiKey,
		meterEvent:    meterEventName,
		webhookSecret: webhookSecret,
		logger:        logger,
		enabled:       apiKey != "",
	}
}

// Enabled returns whether the billing service is configured and active.
func (s *BillingService) Enabled() bool {
	return s.enabled
}

// MeterEnabled returns whether Stripe meter reporting is configured.
func (s *BillingService) MeterEnabled() bool {
	return s.enabled && s.meterEvent != ""
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
	if !s.MeterEnabled() {
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
		} else {
			s.logger.Info("stripe meter event reported",
				slog.String("org_id", b.OrgID),
				slog.Int64("gpu_seconds", b.GPUSeconds),
			)
		}
	}
	return nil
}

// CreateOrGetCustomer creates a Stripe Customer for the org, or returns the existing one.
func (s *BillingService) CreateOrGetCustomer(ctx context.Context, orgID, email string) (string, error) {
	if !s.enabled {
		return "", fmt.Errorf("stripe not configured")
	}
	stripe.Key = s.apiKey

	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Metadata: map[string]string{
			"org_id": orgID,
		},
	}
	c, err := customer.New(params)
	if err != nil {
		return "", fmt.Errorf("create stripe customer: %w", err)
	}
	return c.ID, nil
}

// CreateCheckoutSession creates a Stripe Checkout Session for a one-time payment.
// Returns the checkout URL and session ID.
func (s *BillingService) CreateCheckoutSession(ctx context.Context, customerID string, amountCents int64, orgID, successURL, cancelURL string) (string, string, error) {
	if !s.enabled {
		return "", "", fmt.Errorf("stripe not configured")
	}
	stripe.Key = s.apiKey

	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency:   stripe.String("usd"),
					UnitAmount: stripe.Int64(amountCents),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("GPU.ai Credits"),
					},
				},
				Quantity: stripe.Int64(1),
			},
		},
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			SetupFutureUsage: stripe.String(string(stripe.PaymentIntentSetupFutureUsageOffSession)),
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		Metadata: map[string]string{
			"org_id":       orgID,
			"amount_cents": fmt.Sprintf("%d", amountCents),
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return "", "", fmt.Errorf("create checkout session: %w", err)
	}
	return sess.URL, sess.ID, nil
}

// VerifyWebhookSignature verifies a Stripe webhook signature and returns the parsed event.
func (s *BillingService) VerifyWebhookSignature(payload []byte, sigHeader string) (*stripe.Event, error) {
	event, err := webhook.ConstructEventWithOptions(payload, sigHeader, s.webhookSecret,
		webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true})
	if err != nil {
		return nil, fmt.Errorf("webhook signature verification failed: %w", err)
	}
	return &event, nil
}

// ChargeDefaultPaymentMethod creates and confirms a PaymentIntent using the customer's
// default payment method (off-session). Used for auto-pay.
func (s *BillingService) ChargeDefaultPaymentMethod(ctx context.Context, customerID string, amountCents int64, orgID string) (string, error) {
	if !s.enabled {
		return "", fmt.Errorf("stripe not configured")
	}
	stripe.Key = s.apiKey

	params := &stripe.PaymentIntentParams{
		Customer:       stripe.String(customerID),
		Amount:         stripe.Int64(amountCents),
		Currency:       stripe.String("usd"),
		OffSession:     stripe.Bool(true),
		Confirm:        stripe.Bool(true),
		PaymentMethodTypes: []*string{stripe.String("card")},
		Metadata: map[string]string{
			"org_id":       orgID,
			"amount_cents": fmt.Sprintf("%d", amountCents),
			"type":         "auto_pay",
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return "", fmt.Errorf("charge payment method: %w", err)
	}
	return pi.ID, nil
}

// ParseCheckoutSessionEvent extracts org_id and amount_cents from a checkout.session.completed event.
func ParseCheckoutSessionEvent(event *stripe.Event) (orgID string, amountCents int64, sessionID string, err error) {
	var sess stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
		return "", 0, "", fmt.Errorf("unmarshal checkout session: %w", err)
	}

	orgID = sess.Metadata["org_id"]
	if orgID == "" {
		return "", 0, "", fmt.Errorf("missing org_id in checkout session metadata")
	}

	amountStr := sess.Metadata["amount_cents"]
	if amountStr == "" {
		return "", 0, "", fmt.Errorf("missing amount_cents in checkout session metadata")
	}

	if _, err := fmt.Sscanf(amountStr, "%d", &amountCents); err != nil {
		return "", 0, "", fmt.Errorf("invalid amount_cents: %w", err)
	}

	return orgID, amountCents, sess.ID, nil
}

// ParsePaymentIntentEvent extracts org_id, amount_cents, and type from a payment_intent event.
func ParsePaymentIntentEvent(event *stripe.Event) (orgID string, amountCents int64, piID string, piType string, err error) {
	var pi stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
		return "", 0, "", "", fmt.Errorf("unmarshal payment intent: %w", err)
	}

	orgID = pi.Metadata["org_id"]
	amountStr := pi.Metadata["amount_cents"]
	piType = pi.Metadata["type"]

	if amountStr != "" {
		fmt.Sscanf(amountStr, "%d", &amountCents)
	} else {
		amountCents = pi.Amount
	}

	return orgID, amountCents, pi.ID, piType, nil
}

// ReadWebhookBody reads the webhook request body (max 65536 bytes).
func ReadWebhookBody(body io.Reader) ([]byte, error) {
	return io.ReadAll(io.LimitReader(body, 65536))
}

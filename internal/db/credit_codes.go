package db

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// CreditCode represents a row from the credit_codes table.
type CreditCode struct {
	CreditCodeID  string
	Code          string // stored as SHA-256 hash in DB
	AmountCents   int64
	RedeemedAt    *time.Time
	RedeemedByOrg *string
	ExpiresAt     *time.Time
	CreatedAt     time.Time
}

// ErrCodeAlreadyRedeemed is returned when a credit code has already been used.
var ErrCodeAlreadyRedeemed = errors.New("credit code already redeemed")

// ErrCodeExpired is returned when a credit code has expired.
var ErrCodeExpired = errors.New("credit code expired")

// ErrCodeNotFound is returned when a credit code doesn't exist.
var ErrCodeNotFound = errors.New("credit code not found")

// hashCode returns the SHA-256 hex digest of a normalized credit code.
func hashCode(code string) string {
	normalized := strings.ToUpper(strings.TrimSpace(code))
	h := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(h[:])
}

// RedeemCreditCode atomically redeems a credit code and adds the balance.
// The code is hashed before lookup — plaintext codes are never stored in the DB.
// Returns the code details and new balance. Single-use via redeemed_at IS NULL guard.
func (p *Pool) RedeemCreditCode(ctx context.Context, code, orgID string) (*CreditCode, int64, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback(ctx)

	codeHash := hashCode(code)

	// Look up by hash.
	var cc CreditCode
	err = tx.QueryRow(ctx, `
		SELECT credit_code_id, code, amount_cents, redeemed_at, redeemed_by_org_id, expires_at, created_at
		FROM credit_codes
		WHERE code = $1
		FOR UPDATE`,
		codeHash,
	).Scan(
		&cc.CreditCodeID, &cc.Code, &cc.AmountCents,
		&cc.RedeemedAt, &cc.RedeemedByOrg, &cc.ExpiresAt, &cc.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, 0, ErrCodeNotFound
	}
	if err != nil {
		return nil, 0, err
	}

	// Check if already redeemed.
	if cc.RedeemedAt != nil {
		return nil, 0, ErrCodeAlreadyRedeemed
	}

	// Check expiry.
	if cc.ExpiresAt != nil && cc.ExpiresAt.Before(time.Now().UTC()) {
		return nil, 0, ErrCodeExpired
	}

	// Mark as redeemed.
	now := time.Now().UTC()
	_, err = tx.Exec(ctx, `
		UPDATE credit_codes
		SET redeemed_at = $1, redeemed_by_org_id = $2
		WHERE credit_code_id = $3`,
		now, orgID, cc.CreditCodeID,
	)
	if err != nil {
		return nil, 0, err
	}
	cc.RedeemedAt = &now
	cc.RedeemedByOrg = &orgID

	// Ensure balance row exists.
	_, err = tx.Exec(ctx, `
		INSERT INTO org_balances (org_id)
		VALUES ($1)
		ON CONFLICT (org_id) DO NOTHING`,
		orgID,
	)
	if err != nil {
		return nil, 0, err
	}

	// Add balance.
	var newBalance int64
	err = tx.QueryRow(ctx, `
		UPDATE org_balances
		SET balance_cents = balance_cents + $2, updated_at = NOW()
		WHERE org_id = $1
		RETURNING balance_cents`,
		orgID, cc.AmountCents,
	).Scan(&newBalance)
	if err != nil {
		return nil, 0, err
	}

	// Record transaction (use hash as reference, not plaintext).
	_, err = tx.Exec(ctx, `
		INSERT INTO transactions (org_id, type, amount_cents, balance_after_cents, description, reference_id)
		VALUES ($1, 'credit_code', $2, $3, $4, $5)`,
		orgID, cc.AmountCents, newBalance,
		"Redeemed credit code", &codeHash,
	)
	if err != nil {
		return nil, 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, 0, err
	}

	return &cc, newBalance, nil
}

// CreateCreditCode hashes the code and inserts it into the database.
// The plaintext code is never stored.
func (p *Pool) CreateCreditCode(ctx context.Context, code string, amountCents int64, expiresAt *time.Time) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO credit_codes (code, amount_cents, expires_at)
		VALUES ($1, $2, $3)`,
		hashCode(code), amountCents, expiresAt,
	)
	return err
}

package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// OrgBalance represents a row from the org_balances table.
type OrgBalance struct {
	OrgID                   string
	BalanceCents            int64
	AutoPayEnabled          bool
	AutoPayThresholdCents   int64
	AutoPayAmountCents      int64
	AutoPayLastTriggeredAt  *time.Time
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

// Transaction represents a row from the transactions table.
type Transaction struct {
	TransactionID    string
	OrgID            string
	Type             string
	AmountCents      int64
	BalanceAfterCents int64
	Description      string
	ReferenceID      *string
	CreatedAt        time.Time
}

// EnsureOrgBalance creates an org_balances row if it doesn't exist and returns it.
func (p *Pool) EnsureOrgBalance(ctx context.Context, orgID string) (*OrgBalance, error) {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO org_balances (org_id)
		VALUES ($1)
		ON CONFLICT (org_id) DO NOTHING`,
		orgID,
	)
	if err != nil {
		return nil, err
	}
	return p.GetOrgBalance(ctx, orgID)
}

// GetOrgBalance retrieves the balance record for an organization.
func (p *Pool) GetOrgBalance(ctx context.Context, orgID string) (*OrgBalance, error) {
	var b OrgBalance
	err := p.pool.QueryRow(ctx, `
		SELECT org_id, balance_cents, auto_pay_enabled,
		       auto_pay_threshold_cents, auto_pay_amount_cents,
		       auto_pay_last_triggered_at, created_at, updated_at
		FROM org_balances
		WHERE org_id = $1`,
		orgID,
	).Scan(
		&b.OrgID, &b.BalanceCents, &b.AutoPayEnabled,
		&b.AutoPayThresholdCents, &b.AutoPayAmountCents,
		&b.AutoPayLastTriggeredAt, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// AddBalance atomically adds credits and records a transaction.
// Returns the transaction with the new balance snapshot.
func (p *Pool) AddBalance(ctx context.Context, orgID string, amountCents int64, txType, description string, referenceID *string) (*Transaction, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Ensure balance row exists.
	_, err = tx.Exec(ctx, `
		INSERT INTO org_balances (org_id)
		VALUES ($1)
		ON CONFLICT (org_id) DO NOTHING`,
		orgID,
	)
	if err != nil {
		return nil, err
	}

	// Atomically update balance and get new value.
	var newBalance int64
	err = tx.QueryRow(ctx, `
		UPDATE org_balances
		SET balance_cents = balance_cents + $2, updated_at = NOW()
		WHERE org_id = $1
		RETURNING balance_cents`,
		orgID, amountCents,
	).Scan(&newBalance)
	if err != nil {
		return nil, err
	}

	// Record transaction.
	var t Transaction
	err = tx.QueryRow(ctx, `
		INSERT INTO transactions (org_id, type, amount_cents, balance_after_cents, description, reference_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING transaction_id, org_id, type, amount_cents, balance_after_cents, description, reference_id, created_at`,
		orgID, txType, amountCents, newBalance, description, referenceID,
	).Scan(
		&t.TransactionID, &t.OrgID, &t.Type, &t.AmountCents,
		&t.BalanceAfterCents, &t.Description, &t.ReferenceID, &t.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &t, tx.Commit(ctx)
}

// DeductBalance atomically subtracts credits and records a usage_deduction transaction.
func (p *Pool) DeductBalance(ctx context.Context, orgID string, amountCents int64, description string, referenceID *string) (*Transaction, error) {
	return p.AddBalance(ctx, orgID, -amountCents, "usage_deduction", description, referenceID)
}

// GetTransactions returns paginated transactions for an org (newest first).
// Pass beforeCursor as empty string for the first page.
func (p *Pool) GetTransactions(ctx context.Context, orgID string, limit int, beforeCursor string) ([]Transaction, bool, error) {
	var rows pgx.Rows
	var err error

	if beforeCursor != "" {
		rows, err = p.pool.Query(ctx, `
			SELECT transaction_id, org_id, type, amount_cents, balance_after_cents,
			       description, reference_id, created_at
			FROM transactions
			WHERE org_id = $1 AND transaction_id < $3
			ORDER BY created_at DESC
			LIMIT $2`,
			orgID, limit+1, beforeCursor,
		)
	} else {
		rows, err = p.pool.Query(ctx, `
			SELECT transaction_id, org_id, type, amount_cents, balance_after_cents,
			       description, reference_id, created_at
			FROM transactions
			WHERE org_id = $1
			ORDER BY created_at DESC
			LIMIT $2`,
			orgID, limit+1,
		)
	}
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	var txns []Transaction
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(
			&t.TransactionID, &t.OrgID, &t.Type, &t.AmountCents,
			&t.BalanceAfterCents, &t.Description, &t.ReferenceID, &t.CreatedAt,
		); err != nil {
			return nil, false, err
		}
		txns = append(txns, t)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasMore := len(txns) > limit
	if hasMore {
		txns = txns[:limit]
	}
	return txns, hasMore, nil
}

// UpdateAutoPay configures auto-pay settings for an organization.
func (p *Pool) UpdateAutoPay(ctx context.Context, orgID string, enabled bool, thresholdCents, amountCents int64) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO org_balances (org_id, auto_pay_enabled, auto_pay_threshold_cents, auto_pay_amount_cents)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (org_id) DO UPDATE
		SET auto_pay_enabled = $2,
		    auto_pay_threshold_cents = $3,
		    auto_pay_amount_cents = $4,
		    updated_at = NOW()`,
		orgID, enabled, thresholdCents, amountCents,
	)
	return err
}

// GetOrgsNeedingAutoPay returns orgs where auto-pay is enabled and balance
// is at or below the threshold, with a 5-minute cooldown.
func (p *Pool) GetOrgsNeedingAutoPay(ctx context.Context) ([]OrgBalance, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT org_id, balance_cents, auto_pay_enabled,
		       auto_pay_threshold_cents, auto_pay_amount_cents,
		       auto_pay_last_triggered_at, created_at, updated_at
		FROM org_balances
		WHERE auto_pay_enabled
		  AND balance_cents <= auto_pay_threshold_cents
		  AND (auto_pay_last_triggered_at IS NULL
		       OR auto_pay_last_triggered_at < NOW() - INTERVAL '5 minutes')`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var balances []OrgBalance
	for rows.Next() {
		var b OrgBalance
		if err := rows.Scan(
			&b.OrgID, &b.BalanceCents, &b.AutoPayEnabled,
			&b.AutoPayThresholdCents, &b.AutoPayAmountCents,
			&b.AutoPayLastTriggeredAt, &b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, err
		}
		balances = append(balances, b)
	}
	return balances, rows.Err()
}

// UpdateAutoPayTriggered updates the last triggered timestamp for auto-pay cooldown.
func (p *Pool) UpdateAutoPayTriggered(ctx context.Context, orgID string) error {
	_, err := p.pool.Exec(ctx, `
		UPDATE org_balances
		SET auto_pay_last_triggered_at = NOW(), updated_at = NOW()
		WHERE org_id = $1`,
		orgID,
	)
	return err
}

// TransactionExistsByReference checks if a transaction with the given reference_id exists.
// Used for webhook deduplication.
func (p *Pool) TransactionExistsByReference(ctx context.Context, referenceID string) (bool, error) {
	var exists bool
	err := p.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM transactions WHERE reference_id = $1)`,
		referenceID,
	).Scan(&exists)
	return exists, err
}

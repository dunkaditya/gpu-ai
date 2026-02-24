package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// ErrIdempotencyKeyExists is returned when an idempotency key already exists
// for the given organization.
var ErrIdempotencyKeyExists = errors.New("idempotency key already exists")

// IdempotencyKey represents an idempotency key record from the database.
type IdempotencyKey struct {
	Key          string
	OrgID        string
	RequestHash  string
	ResponseCode *int
	ResponseBody []byte
	CreatedAt    time.Time
}

// GetIdempotencyKey retrieves an idempotency key for a given organization.
// Returns ErrNotFound if no matching key exists.
func (p *Pool) GetIdempotencyKey(ctx context.Context, orgID, key string) (*IdempotencyKey, error) {
	var ik IdempotencyKey
	err := p.pool.QueryRow(ctx,
		`SELECT idempotency_key, org_id, request_hash, response_code, response_body, created_at
		 FROM idempotency_keys
		 WHERE org_id = $1 AND idempotency_key = $2`,
		orgID, key,
	).Scan(&ik.Key, &ik.OrgID, &ik.RequestHash, &ik.ResponseCode, &ik.ResponseBody, &ik.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &ik, nil
}

// CreateIdempotencyKey inserts a new idempotency key for the given organization.
// Returns ErrIdempotencyKeyExists if the key already exists (conflict on org_id + idempotency_key).
func (p *Pool) CreateIdempotencyKey(ctx context.Context, orgID, key, requestHash string) error {
	_, err := p.pool.Exec(ctx,
		`INSERT INTO idempotency_keys (org_id, idempotency_key, request_hash)
		 VALUES ($1, $2, $3)`,
		orgID, key, requestHash,
	)
	if err != nil {
		// Check for unique violation (PK conflict on org_id, idempotency_key).
		if isDuplicateKeyError(err) {
			return ErrIdempotencyKeyExists
		}
		return err
	}
	return nil
}

// CompleteIdempotencyKey updates an existing idempotency key with the response
// code and body once the request has been processed.
func (p *Pool) CompleteIdempotencyKey(ctx context.Context, orgID, key string, responseCode int, responseBody []byte) error {
	_, err := p.pool.Exec(ctx,
		`UPDATE idempotency_keys SET response_code = $1, response_body = $2
		 WHERE org_id = $3 AND idempotency_key = $4`,
		responseCode, responseBody, orgID, key,
	)
	return err
}

// CleanupIdempotencyKeys deletes idempotency keys older than the specified duration.
// Returns the number of rows deleted.
func (p *Pool) CleanupIdempotencyKeys(ctx context.Context, olderThan time.Duration) (int64, error) {
	tag, err := p.pool.Exec(ctx,
		`DELETE FROM idempotency_keys WHERE created_at < NOW() - $1::interval`,
		olderThan.String(),
	)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// isDuplicateKeyError checks if the error is a PostgreSQL unique violation (SQLSTATE 23505).
func isDuplicateKeyError(err error) bool {
	// pgx wraps PostgreSQL errors with a Code() method.
	// We check for the unique_violation error code (23505).
	var pgErr interface{ SQLState() string }
	if errors.As(err, &pgErr) {
		return pgErr.SQLState() == "23505"
	}
	return false
}

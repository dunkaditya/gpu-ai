package tunnel

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// AllocatePort allocates the next available FRP remote port within the
// configured range [MinPort, MaxPort]. It must be called within a transaction.
// The function acquires a PostgreSQL advisory lock (transaction-scoped) to
// prevent concurrent allocation races, queries the current maximum
// frp_remote_port among active instances, and returns max+1.
//
// Terminated and error instances are excluded from the query, allowing their
// ports to be reclaimed (the partial unique index enforces this at the DB level).
//
// The caller is responsible for storing the allocated port in the instance row
// within the same transaction.
func AllocatePort(ctx context.Context, tx pgx.Tx) (int, error) {
	// Acquire transaction-scoped advisory lock to serialize allocations.
	_, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", portLockID)
	if err != nil {
		return 0, fmt.Errorf("tunnel: acquire port lock: %w", err)
	}

	// Query the current maximum allocated port among active instances.
	var maxPort int
	err = tx.QueryRow(ctx,
		"SELECT COALESCE(MAX(frp_remote_port), $1) FROM instances WHERE frp_remote_port IS NOT NULL AND status NOT IN ('terminated', 'error')",
		MinPort-1,
	).Scan(&maxPort)
	if err != nil {
		return 0, fmt.Errorf("tunnel: query max port: %w", err)
	}

	next := maxPort + 1
	if next > MaxPort {
		return 0, fmt.Errorf("tunnel: port range exhausted (%d-%d)", MinPort, MaxPort)
	}

	return next, nil
}

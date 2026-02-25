package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// ErrNotFound is returned when a database query yields no matching rows.
var ErrNotFound = errors.New("not found")

// Instance represents the full internal instance record from the database.
// Maps to all columns across v0+v1+v2+v3 schema versions.
type Instance struct {
	InstanceID           string
	OrgID                string
	UserID               string
	UpstreamProvider     string
	UpstreamID           string
	UpstreamIP           *string
	Hostname             string
	WGPublicKey          *string
	WGPrivateKeyEnc      *string
	WGAddress            *string
	Name                 *string
	GPUType              string
	GPUCount             int
	Tier                 string
	Region               string
	PricePerHour         float64
	UpstreamPricePerHour float64
	BillingStart         *time.Time
	BillingEnd           *time.Time
	Status               string
	ErrorReason          *string
	InternalToken        *string
	CreatedAt            time.Time
	UpdatedAt            time.Time
	ReadyAt              *time.Time
	TerminatedAt         *time.Time
}

// instanceColumns is the ordered list of columns for SELECT queries.
const instanceColumns = `instance_id, org_id, user_id, upstream_provider, upstream_id,
	upstream_ip, hostname, wg_public_key, wg_private_key_enc, wg_address,
	name, gpu_type, gpu_count, tier, region,
	price_per_hour, upstream_price_per_hour, billing_start, billing_end,
	status, error_reason, internal_token,
	created_at, updated_at, ready_at, terminated_at`

// scanInstance scans a single row into an Instance struct.
func scanInstance(row pgx.Row) (*Instance, error) {
	var i Instance
	err := row.Scan(
		&i.InstanceID, &i.OrgID, &i.UserID, &i.UpstreamProvider, &i.UpstreamID,
		&i.UpstreamIP, &i.Hostname, &i.WGPublicKey, &i.WGPrivateKeyEnc, &i.WGAddress,
		&i.Name, &i.GPUType, &i.GPUCount, &i.Tier, &i.Region,
		&i.PricePerHour, &i.UpstreamPricePerHour, &i.BillingStart, &i.BillingEnd,
		&i.Status, &i.ErrorReason, &i.InternalToken,
		&i.CreatedAt, &i.UpdatedAt, &i.ReadyAt, &i.TerminatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &i, nil
}

// CreateInstance inserts a new instance record into the database.
func (p *Pool) CreateInstance(ctx context.Context, inst *Instance) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO instances (
			instance_id, org_id, user_id, upstream_provider, upstream_id,
			upstream_ip, hostname, wg_public_key, wg_private_key_enc, wg_address,
			name, gpu_type, gpu_count, tier, region,
			price_per_hour, upstream_price_per_hour, billing_start, billing_end,
			status, error_reason, internal_token,
			ready_at, terminated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15,
			$16, $17, $18, $19,
			$20, $21, $22,
			$23, $24
		)`,
		inst.InstanceID, inst.OrgID, inst.UserID, inst.UpstreamProvider, inst.UpstreamID,
		inst.UpstreamIP, inst.Hostname, inst.WGPublicKey, inst.WGPrivateKeyEnc, inst.WGAddress,
		inst.Name, inst.GPUType, inst.GPUCount, inst.Tier, inst.Region,
		inst.PricePerHour, inst.UpstreamPricePerHour, inst.BillingStart, inst.BillingEnd,
		inst.Status, inst.ErrorReason, inst.InternalToken,
		inst.ReadyAt, inst.TerminatedAt,
	)
	return err
}

// GetInstance retrieves an instance by its ID. Returns ErrNotFound if no row exists.
func (p *Pool) GetInstance(ctx context.Context, instanceID string) (*Instance, error) {
	row := p.pool.QueryRow(ctx,
		`SELECT `+instanceColumns+` FROM instances WHERE instance_id = $1`,
		instanceID,
	)
	inst, err := scanInstance(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return inst, err
}

// GetInstanceForOrg retrieves an instance scoped to a specific organization.
// This enforces organization isolation at the query level (AUTH-03).
func (p *Pool) GetInstanceForOrg(ctx context.Context, instanceID, orgID string) (*Instance, error) {
	row := p.pool.QueryRow(ctx,
		`SELECT `+instanceColumns+` FROM instances WHERE instance_id = $1 AND org_id = $2`,
		instanceID, orgID,
	)
	inst, err := scanInstance(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return inst, err
}

// ListInstances returns instances for an organization using keyset pagination.
// The cursor is a (created_at, instance_id) tuple. Pass nil cursor for the first page.
// Returns limit+1 rows so the caller can detect has_more by checking len > limit.
func (p *Pool) ListInstances(ctx context.Context, orgID string, cursor *time.Time, cursorID string, limit int) ([]Instance, error) {
	var rows pgx.Rows
	var err error

	if cursor != nil {
		rows, err = p.pool.Query(ctx,
			`SELECT `+instanceColumns+` FROM instances
			 WHERE org_id = $1 AND (created_at, instance_id) < ($2, $3)
			 ORDER BY created_at DESC, instance_id DESC
			 LIMIT $4`,
			orgID, *cursor, cursorID, limit+1,
		)
	} else {
		rows, err = p.pool.Query(ctx,
			`SELECT `+instanceColumns+` FROM instances
			 WHERE org_id = $1
			 ORDER BY created_at DESC, instance_id DESC
			 LIMIT $2`,
			orgID, limit+1,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []Instance
	for rows.Next() {
		var inst Instance
		if err := rows.Scan(
			&inst.InstanceID, &inst.OrgID, &inst.UserID, &inst.UpstreamProvider, &inst.UpstreamID,
			&inst.UpstreamIP, &inst.Hostname, &inst.WGPublicKey, &inst.WGPrivateKeyEnc, &inst.WGAddress,
			&inst.Name, &inst.GPUType, &inst.GPUCount, &inst.Tier, &inst.Region,
			&inst.PricePerHour, &inst.UpstreamPricePerHour, &inst.BillingStart, &inst.BillingEnd,
			&inst.Status, &inst.ErrorReason, &inst.InternalToken,
			&inst.CreatedAt, &inst.UpdatedAt, &inst.ReadyAt, &inst.TerminatedAt,
		); err != nil {
			return nil, err
		}
		instances = append(instances, inst)
	}
	return instances, rows.Err()
}

// ListRunningInstancesByOrg returns all instances with status 'running' for an org.
// Used by the billing ticker to stop instances when a spending limit is reached.
func (p *Pool) ListRunningInstancesByOrg(ctx context.Context, orgID string) ([]Instance, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT `+instanceColumns+` FROM instances WHERE org_id = $1 AND status = 'running'`,
		orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []Instance
	for rows.Next() {
		var inst Instance
		if err := rows.Scan(
			&inst.InstanceID, &inst.OrgID, &inst.UserID, &inst.UpstreamProvider, &inst.UpstreamID,
			&inst.UpstreamIP, &inst.Hostname, &inst.WGPublicKey, &inst.WGPrivateKeyEnc, &inst.WGAddress,
			&inst.Name, &inst.GPUType, &inst.GPUCount, &inst.Tier, &inst.Region,
			&inst.PricePerHour, &inst.UpstreamPricePerHour, &inst.BillingStart, &inst.BillingEnd,
			&inst.Status, &inst.ErrorReason, &inst.InternalToken,
			&inst.CreatedAt, &inst.UpdatedAt, &inst.ReadyAt, &inst.TerminatedAt,
		); err != nil {
			return nil, err
		}
		instances = append(instances, inst)
	}
	return instances, rows.Err()
}

// ListStoppedInstancesByOrg returns all instances with status 'stopped' for an org.
// Used by the billing ticker to terminate instances 72h after spending limit was reached.
func (p *Pool) ListStoppedInstancesByOrg(ctx context.Context, orgID string) ([]Instance, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT `+instanceColumns+` FROM instances WHERE org_id = $1 AND status = 'stopped'`,
		orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []Instance
	for rows.Next() {
		var inst Instance
		if err := rows.Scan(
			&inst.InstanceID, &inst.OrgID, &inst.UserID, &inst.UpstreamProvider, &inst.UpstreamID,
			&inst.UpstreamIP, &inst.Hostname, &inst.WGPublicKey, &inst.WGPrivateKeyEnc, &inst.WGAddress,
			&inst.Name, &inst.GPUType, &inst.GPUCount, &inst.Tier, &inst.Region,
			&inst.PricePerHour, &inst.UpstreamPricePerHour, &inst.BillingStart, &inst.BillingEnd,
			&inst.Status, &inst.ErrorReason, &inst.InternalToken,
			&inst.CreatedAt, &inst.UpdatedAt, &inst.ReadyAt, &inst.TerminatedAt,
		); err != nil {
			return nil, err
		}
		instances = append(instances, inst)
	}
	return instances, rows.Err()
}

// UpdateInstanceStatus atomically updates an instance's status using optimistic locking.
// It only updates the row if the current status matches fromStatus, preventing race conditions.
// Returns true if the row was updated, false if the status was concurrently changed.
func (p *Pool) UpdateInstanceStatus(ctx context.Context, instanceID, fromStatus, toStatus string) (bool, error) {
	tag, err := p.pool.Exec(ctx,
		`UPDATE instances SET status = $1, updated_at = NOW() WHERE instance_id = $2 AND status = $3`,
		toStatus, instanceID, fromStatus,
	)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

// SetInstanceRunning atomically transitions an instance from booting to running,
// recording the ready_at timestamp. Returns false if the instance is not in booting state.
func (p *Pool) SetInstanceRunning(ctx context.Context, instanceID string) (bool, error) {
	tag, err := p.pool.Exec(ctx,
		`UPDATE instances SET status = 'running', ready_at = NOW(), updated_at = NOW()
		 WHERE instance_id = $1 AND status = 'booting'`,
		instanceID,
	)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

// TerminateInstance atomically terminates an instance, setting billing_end and
// terminated_at timestamps. Returns false if the instance is already in a terminal
// state (idempotent). Uses NOT IN for idempotent termination (INST-06).
func (p *Pool) TerminateInstance(ctx context.Context, instanceID string) (bool, error) {
	tag, err := p.pool.Exec(ctx,
		`UPDATE instances SET status = 'terminated', terminated_at = NOW(), billing_end = NOW(), updated_at = NOW()
		 WHERE instance_id = $1 AND status NOT IN ('terminated', 'error')`,
		instanceID,
	)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

// SetInstanceError sets an instance to error state with a human-readable reason.
func (p *Pool) SetInstanceError(ctx context.Context, instanceID, reason string) error {
	_, err := p.pool.Exec(ctx,
		`UPDATE instances SET status = 'error', error_reason = $1, updated_at = NOW() WHERE instance_id = $2`,
		reason, instanceID,
	)
	return err
}

package db

// Canonical PK column: instances.instance_id (VARCHAR(12), e.g., "gpu-4a7f")
//
// v1 schema changes affecting this table:
//   - PK renamed: id -> instance_id
//   - wg_private_key_enc column removed (security fix)
//   - internal_token column added (per-instance callback auth)
//   - updated_at column added (auto-updated via trigger)
//   - CHECK constraint on status: creating, provisioning, booting, running, stopping, terminated, error
//   - UNIQUE constraint on hostname
//   - Composite unique index on (upstream_provider, upstream_id)

// TODO: Implement instance queries using pgx:
//
// func (p *Pool) CreateInstance(ctx context.Context, inst *Instance) error
//   - INSERT into instances table
//
// func (p *Pool) GetInstance(ctx context.Context, instanceID string) (*Instance, error)
//   - SELECT instance by instance_id
//
// func (p *Pool) ListInstances(ctx context.Context, orgID string, status *string) ([]Instance, error)
//   - SELECT instances for an org, optionally filtered by status
//
// func (p *Pool) UpdateInstanceStatus(ctx context.Context, instanceID string, status string) error
//   - UPDATE instance status by instance_id
//
// func (p *Pool) TerminateInstance(ctx context.Context, instanceID string) error
//   - SET status='terminated', terminated_at=NOW(), billing_end=NOW() by instance_id
//
// type Instance struct {
//     InstanceID           string     // instances.instance_id (PK)
//     OrgID                string
//     UserID               string
//     UpstreamProvider     string
//     UpstreamID           string
//     UpstreamIP           string
//     Hostname             string
//     WireGuardPublicKey   string
//     WireGuardAddress     string
//     GPUType              string
//     GPUCount             int
//     Tier                 string
//     Region               string
//     PricePerHour         float64
//     UpstreamPricePerHour float64
//     BillingStart         *time.Time
//     BillingEnd           *time.Time
//     Status               string
//     InternalToken        string     // per-instance callback auth token
//     CreatedAt            time.Time
//     UpdatedAt            time.Time  // auto-updated via trigger
//     TerminatedAt         *time.Time
// }

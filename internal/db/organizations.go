package db

// Canonical PK column: organizations.organization_id (UUID)
// Canonical PK column: users.user_id (UUID)
//
// v1 schema changes affecting these tables:
//   - organizations PK renamed: id -> organization_id
//   - users PK renamed: id -> user_id
//   - users.org_id SET NOT NULL
//   - users.org_id ON DELETE CASCADE (org deleted = users deleted)

// TODO: Implement organization/user queries using pgx:
//
// func (p *Pool) GetOrganization(ctx context.Context, organizationID string) (*Organization, error)
// func (p *Pool) CreateOrganization(ctx context.Context, name, stripeCustomerID string) (*Organization, error)
// func (p *Pool) GetUser(ctx context.Context, userID string) (*User, error)
// func (p *Pool) GetUserByEmail(ctx context.Context, email string) (*User, error)
// func (p *Pool) CreateUser(ctx context.Context, orgID, email, name, role string) (*User, error)
//
// type Organization struct {
//     OrganizationID   string    // organizations.organization_id (PK)
//     Name             string
//     StripeCustomerID string
//     CreatedAt        time.Time
// }
//
// type User struct {
//     UserID    string    // users.user_id (PK)
//     OrgID     string
//     Email     string
//     Name      string
//     Role      string
//     CreatedAt time.Time
// }

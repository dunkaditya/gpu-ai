package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// Organization represents an organization record from the database.
type Organization struct {
	OrganizationID   string
	Name             string
	StripeCustomerID *string
	ClerkOrgID       *string
	CreatedAt        time.Time
}

// User represents a user record from the database.
type User struct {
	UserID      string
	OrgID       string
	ClerkUserID *string
	Email       string
	Name        string
	Role        string
	CreatedAt   time.Time
}

// GetOrganization retrieves an organization by its internal UUID.
// Returns ErrNotFound if no row exists.
func (p *Pool) GetOrganization(ctx context.Context, orgID string) (*Organization, error) {
	var org Organization
	err := p.pool.QueryRow(ctx,
		`SELECT organization_id, name, stripe_customer_id, clerk_org_id, created_at
		 FROM organizations WHERE organization_id = $1`,
		orgID,
	).Scan(&org.OrganizationID, &org.Name, &org.StripeCustomerID, &org.ClerkOrgID, &org.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// EnsureOrgAndUser upserts an organization (by Clerk org ID) and a user (by Clerk user ID).
// On first call, it creates the org and user. On subsequent calls, it updates the user's email.
// Returns the internal organization UUID.
func (p *Pool) EnsureOrgAndUser(ctx context.Context, clerkOrgID, clerkUserID, email string) (string, error) {
	// Upsert organization: create if not exists, otherwise no-op.
	_, err := p.pool.Exec(ctx,
		`INSERT INTO organizations (organization_id, name, clerk_org_id)
		 VALUES (gen_random_uuid(), $1, $1)
		 ON CONFLICT (clerk_org_id) DO NOTHING`,
		clerkOrgID,
	)
	if err != nil {
		return "", err
	}

	// Retrieve the internal organization UUID.
	var orgUUID string
	err = p.pool.QueryRow(ctx,
		`SELECT organization_id FROM organizations WHERE clerk_org_id = $1`,
		clerkOrgID,
	).Scan(&orgUUID)
	if err != nil {
		return "", err
	}

	// Upsert user: create if not exists, otherwise update email.
	_, err = p.pool.Exec(ctx,
		`INSERT INTO users (user_id, org_id, clerk_user_id, email, name, role)
		 VALUES (gen_random_uuid(), $1, $2, $3, '', 'member')
		 ON CONFLICT (clerk_user_id) DO UPDATE SET email = EXCLUDED.email`,
		orgUUID, clerkUserID, email,
	)
	if err != nil {
		return "", err
	}

	return orgUUID, nil
}

// GetOrgIDByClerkOrgID retrieves the internal organization UUID by Clerk org ID.
// Returns ErrNotFound if no matching organization exists.
func (p *Pool) GetOrgIDByClerkOrgID(ctx context.Context, clerkOrgID string) (string, error) {
	var orgUUID string
	err := p.pool.QueryRow(ctx,
		`SELECT organization_id FROM organizations WHERE clerk_org_id = $1`,
		clerkOrgID,
	).Scan(&orgUUID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return orgUUID, nil
}

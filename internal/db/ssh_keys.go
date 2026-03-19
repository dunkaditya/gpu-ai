package db

import (
	"context"
	"errors"
	"time"
)

// ErrDuplicateKey is returned when an SSH key with the same fingerprint
// already exists for the organization.
var ErrDuplicateKey = errors.New("ssh key already exists")

// SSHKey represents an SSH key record from the database.
type SSHKey struct {
	SSHKeyID    string
	OrgID       string
	UserID      string
	Name        string
	PublicKey   string
	Fingerprint string
	CreatedAt   time.Time
}

// CreateSSHKey inserts a new SSH key into the database.
// The ssh_key_id and created_at fields are populated by the database via RETURNING.
// Returns ErrDuplicateKey if a key with the same fingerprint already exists for the org.
func (p *Pool) CreateSSHKey(ctx context.Context, key *SSHKey) error {
	err := p.pool.QueryRow(ctx,
		`INSERT INTO ssh_keys (user_id, org_id, name, public_key, fingerprint)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING ssh_key_id, created_at`,
		key.UserID, key.OrgID, key.Name, key.PublicKey, key.Fingerprint,
	).Scan(&key.SSHKeyID, &key.CreatedAt)
	if err != nil {
		// Check for unique violation (SQLSTATE 23505) on (org_id, fingerprint).
		var pgErr interface{ SQLState() string }
		if errors.As(err, &pgErr) && pgErr.SQLState() == "23505" {
			return ErrDuplicateKey
		}
		return err
	}
	return nil
}

// ListSSHKeysByOrg retrieves all SSH keys for an organization, ordered by newest first.
func (p *Pool) ListSSHKeysByOrg(ctx context.Context, orgID string) ([]SSHKey, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT ssh_key_id, org_id, user_id, name, public_key, fingerprint, created_at
		 FROM ssh_keys WHERE org_id = $1
		 ORDER BY created_at DESC`,
		orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []SSHKey
	for rows.Next() {
		var k SSHKey
		if err := rows.Scan(&k.SSHKeyID, &k.OrgID, &k.UserID, &k.Name, &k.PublicKey, &k.Fingerprint, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

// DeleteSSHKey removes an SSH key by ID, scoped to an organization.
// Returns false (not error) when the key is not found or belongs to a different org,
// preventing existence leaking across orgs.
func (p *Pool) DeleteSSHKey(ctx context.Context, keyID, orgID string) (bool, error) {
	tag, err := p.pool.Exec(ctx,
		`DELETE FROM ssh_keys WHERE ssh_key_id = $1 AND org_id = $2`,
		keyID, orgID,
	)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// GetSSHKeysByUserID retrieves all SSH keys belonging to a specific user.
// Used for the smart provisioning default: auto-include all user keys when
// no explicit ssh_key_ids are provided.
func (p *Pool) GetSSHKeysByUserID(ctx context.Context, userID string) ([]SSHKey, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT ssh_key_id, org_id, user_id, name, public_key, fingerprint, created_at
		 FROM ssh_keys WHERE user_id = $1
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []SSHKey
	for rows.Next() {
		var k SSHKey
		if err := rows.Scan(&k.SSHKeyID, &k.OrgID, &k.UserID, &k.Name, &k.PublicKey, &k.Fingerprint, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

// GetSSHKeysByOrgID retrieves all SSH keys belonging to an organization.
// Used as fallback when the launching user has no personal keys.
func (p *Pool) GetSSHKeysByOrgID(ctx context.Context, orgID string) ([]SSHKey, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT ssh_key_id, org_id, user_id, name, public_key, fingerprint, created_at
		 FROM ssh_keys WHERE org_id = $1
		 ORDER BY created_at DESC`,
		orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []SSHKey
	for rows.Next() {
		var k SSHKey
		if err := rows.Scan(&k.SSHKeyID, &k.OrgID, &k.UserID, &k.Name, &k.PublicKey, &k.Fingerprint, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

// CountSSHKeysByOrg returns the number of SSH keys for an organization.
// Used to enforce the per-org 50-key limit.
func (p *Pool) CountSSHKeysByOrg(ctx context.Context, orgID string) (int, error) {
	var count int
	err := p.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM ssh_keys WHERE org_id = $1`,
		orgID,
	).Scan(&count)
	return count, err
}

// GetSSHKeysByIDs retrieves SSH keys by a list of IDs.
// Used by the provisioning engine to resolve ssh_key_ids to public key content.
// Returns only the keys that exist; missing IDs are silently skipped.
func (p *Pool) GetSSHKeysByIDs(ctx context.Context, ids []string) ([]SSHKey, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT ssh_key_id, org_id, user_id, name, public_key, fingerprint, created_at
		 FROM ssh_keys WHERE ssh_key_id = ANY($1)`,
		ids,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []SSHKey
	for rows.Next() {
		var k SSHKey
		if err := rows.Scan(&k.SSHKeyID, &k.OrgID, &k.UserID, &k.Name, &k.PublicKey, &k.Fingerprint, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

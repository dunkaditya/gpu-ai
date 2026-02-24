package db

import (
	"context"
	"time"
)

// SSHKey represents an SSH key record from the database.
type SSHKey struct {
	SSHKeyID    string
	UserID      string
	Name        string
	PublicKey   string
	Fingerprint string
	CreatedAt   time.Time
}

// GetSSHKeysByIDs retrieves SSH keys by a list of IDs.
// Used by the provisioning engine to resolve ssh_key_ids to public key content.
// Returns only the keys that exist; missing IDs are silently skipped.
func (p *Pool) GetSSHKeysByIDs(ctx context.Context, ids []string) ([]SSHKey, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT ssh_key_id, user_id, name, public_key, fingerprint, created_at
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
		if err := rows.Scan(&k.SSHKeyID, &k.UserID, &k.Name, &k.PublicKey, &k.Fingerprint, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

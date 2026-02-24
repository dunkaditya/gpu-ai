package db

// Canonical PK column: ssh_keys.ssh_key_id (UUID)
//
// v1 schema changes affecting this table:
//   - PK renamed: id -> ssh_key_id

// TODO: Implement SSH key queries using pgx:
//
// func (p *Pool) ListSSHKeys(ctx context.Context, userID string) ([]SSHKey, error)
// func (p *Pool) CreateSSHKey(ctx context.Context, userID, name, publicKey, fingerprint string) (*SSHKey, error)
// func (p *Pool) DeleteSSHKey(ctx context.Context, sshKeyID, userID string) error
//
// type SSHKey struct {
//     SSHKeyID    string    // ssh_keys.ssh_key_id (PK)
//     UserID      string
//     Name        string
//     PublicKey   string
//     Fingerprint string
//     CreatedAt   time.Time
// }

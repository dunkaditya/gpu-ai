package wireguard

// KeyPair holds a WireGuard key pair as base64-encoded strings.
type KeyPair struct {
	PrivateKey string // base64-encoded 32-byte Curve25519 private key
	PublicKey  string // base64-encoded 32-byte Curve25519 public key
}

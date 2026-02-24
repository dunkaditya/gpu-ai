// Package wireguard manages WireGuard peer configuration on the proxy server.
package wireguard

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// GenerateKeyPair generates a new WireGuard key pair using wgctrl-go.
// Returns the private and public keys as base64-encoded strings.
func GenerateKeyPair() (*KeyPair, error) {
	privKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("wireguard: generate private key: %w", err)
	}

	pubKey := privKey.PublicKey()

	return &KeyPair{
		PrivateKey: privKey.String(),
		PublicKey:  pubKey.String(),
	}, nil
}

// EncryptPrivateKey encrypts a WireGuard private key using AES-256-GCM.
// The encryptionKey must be exactly 32 bytes. Returns hex-encoded
// nonce+ciphertext string.
func EncryptPrivateKey(plaintext string, encryptionKey []byte) (string, error) {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("wireguard: create aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("wireguard: create gcm: %w", err)
	}

	// Generate random 12-byte nonce.
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("wireguard: generate nonce: %w", err)
	}

	// Seal prepends the nonce to the ciphertext.
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	return hex.EncodeToString(ciphertext), nil
}

// DecryptPrivateKey decrypts a hex-encoded AES-256-GCM encrypted WireGuard
// private key. The encryptionKey must be exactly 32 bytes.
func DecryptPrivateKey(ciphertextHex string, encryptionKey []byte) (string, error) {
	data, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", fmt.Errorf("wireguard: decode hex ciphertext: %w", err)
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("wireguard: create aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("wireguard: create gcm: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("wireguard: ciphertext too short (got %d bytes, need at least %d for nonce)", len(data), nonceSize)
	}

	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("wireguard: decrypt: %w", err)
	}

	return string(plaintext), nil
}

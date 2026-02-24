package wireguard

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/gpuai/gpuctl/internal/provider"
)

// testKey is a fixed 32-byte encryption key for testing.
var testKey = bytes.Repeat([]byte{0x42}, 32)

func TestGenerateKeyPair(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error: %v", err)
	}

	if kp.PrivateKey == "" {
		t.Error("PrivateKey is empty")
	}
	if kp.PublicKey == "" {
		t.Error("PublicKey is empty")
	}

	// Verify keys are valid base64 and decode to 32 bytes.
	privBytes, err := base64.StdEncoding.DecodeString(kp.PrivateKey)
	if err != nil {
		t.Fatalf("PrivateKey is not valid base64: %v", err)
	}
	if len(privBytes) != 32 {
		t.Errorf("PrivateKey decoded to %d bytes, want 32", len(privBytes))
	}

	pubBytes, err := base64.StdEncoding.DecodeString(kp.PublicKey)
	if err != nil {
		t.Fatalf("PublicKey is not valid base64: %v", err)
	}
	if len(pubBytes) != 32 {
		t.Errorf("PublicKey decoded to %d bytes, want 32", len(pubBytes))
	}
}

func TestGenerateKeyPairUniqueness(t *testing.T) {
	kp1, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() #1 error: %v", err)
	}

	kp2, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() #2 error: %v", err)
	}

	if kp1.PrivateKey == kp2.PrivateKey {
		t.Error("two generated key pairs have the same private key")
	}
	if kp1.PublicKey == kp2.PublicKey {
		t.Error("two generated key pairs have the same public key")
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	plaintext := "test-wireguard-private-key-base64-encoded"

	encrypted, err := EncryptPrivateKey(plaintext, testKey)
	if err != nil {
		t.Fatalf("EncryptPrivateKey() error: %v", err)
	}

	if encrypted == "" {
		t.Fatal("EncryptPrivateKey() returned empty string")
	}

	decrypted, err := DecryptPrivateKey(encrypted, testKey)
	if err != nil {
		t.Fatalf("DecryptPrivateKey() error: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("round-trip mismatch: got %q, want %q", decrypted, plaintext)
	}
}

func TestEncryptProducesDifferentCiphertext(t *testing.T) {
	plaintext := "same-key-data-for-both-encryptions"

	ct1, err := EncryptPrivateKey(plaintext, testKey)
	if err != nil {
		t.Fatalf("EncryptPrivateKey() #1 error: %v", err)
	}

	ct2, err := EncryptPrivateKey(plaintext, testKey)
	if err != nil {
		t.Fatalf("EncryptPrivateKey() #2 error: %v", err)
	}

	if ct1 == ct2 {
		t.Error("two encryptions of the same plaintext produced identical ciphertext (nonce reuse)")
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	plaintext := "secret-key-data"

	encrypted, err := EncryptPrivateKey(plaintext, testKey)
	if err != nil {
		t.Fatalf("EncryptPrivateKey() error: %v", err)
	}

	wrongKey := bytes.Repeat([]byte{0xFF}, 32)
	_, err = DecryptPrivateKey(encrypted, wrongKey)
	if err == nil {
		t.Error("DecryptPrivateKey() with wrong key should return error, got nil")
	}
}

func TestDecryptInvalidHex(t *testing.T) {
	_, err := DecryptPrivateKey("not-valid-hex!@#$", testKey)
	if err == nil {
		t.Error("DecryptPrivateKey() with invalid hex should return error, got nil")
	}
}

func TestDecryptTruncatedCiphertext(t *testing.T) {
	// A hex string shorter than the 12-byte nonce (24 hex chars).
	shortHex := "aabbccdd"
	_, err := DecryptPrivateKey(shortHex, testKey)
	if err == nil {
		t.Error("DecryptPrivateKey() with truncated ciphertext should return error, got nil")
	}
}

func TestEncryptWithInvalidKeyLength(t *testing.T) {
	// AES accepts 16, 24, and 32 byte keys. Use 15 bytes which is truly invalid.
	badKey := bytes.Repeat([]byte{0x42}, 15)
	_, err := EncryptPrivateKey("some-data", badKey)
	if err == nil {
		t.Error("EncryptPrivateKey() with 15-byte key should return error, got nil")
	}
}

func TestToCustomerExcludesUpstream(t *testing.T) {
	inst := &provider.Instance{
		InstanceID:       "gpu-4a7f",
		OrgID:            "org-123",
		UserID:           "user-456",
		UpstreamProvider: "runpod",
		UpstreamID:       "pod-abc123",
		UpstreamIP:       "192.168.1.100",
		Hostname:         "gpu-4a7f.gpu.ai",
		WGPublicKey:      "pubkey-base64",
		WGAddress:        "10.0.0.5",
		GPUType:          "h100_sxm",
		GPUCount:         8,
		Tier:             "on_demand",
		Region:           "us-east",
		PricePerHour:     3.50,
		Status:           "running",
	}

	customer := inst.ToCustomer()

	// Marshal to JSON to verify no upstream fields leak.
	data, err := json.Marshal(customer)
	if err != nil {
		t.Fatalf("json.Marshal(CustomerInstance) error: %v", err)
	}

	jsonStr := string(data)

	// Verify upstream fields are absent.
	upstreamFields := []string{
		"upstream_provider", "upstream_id", "upstream_ip",
		"runpod", "pod-abc123", "192.168.1.100",
		"org_id", "user_id", "wg_public_key", "wg_address",
	}
	for _, field := range upstreamFields {
		if bytes.Contains(data, []byte(field)) {
			t.Errorf("CustomerInstance JSON contains upstream field %q: %s", field, jsonStr)
		}
	}

	// Verify expected fields are present.
	expectedFields := []string{
		"id", "hostname", "ssh_command", "status",
		"gpu_type", "gpu_count", "tier", "region", "price_per_hour",
	}
	for _, field := range expectedFields {
		if !bytes.Contains(data, []byte(field)) {
			t.Errorf("CustomerInstance JSON missing expected field %q: %s", field, jsonStr)
		}
	}

	// Verify correct values.
	if customer.ID != "gpu-4a7f" {
		t.Errorf("ID = %q, want %q", customer.ID, "gpu-4a7f")
	}
	if customer.SSHCommand != "ssh root@gpu-4a7f.gpu.ai" {
		t.Errorf("SSHCommand = %q, want %q", customer.SSHCommand, "ssh root@gpu-4a7f.gpu.ai")
	}
}

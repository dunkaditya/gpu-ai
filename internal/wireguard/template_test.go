package wireguard

import (
	"strings"
	"testing"
)

// validBootstrapData returns a BootstrapData with all valid fields for testing.
func validBootstrapData() BootstrapData {
	return BootstrapData{
		InstanceID:         "gpu-4a7f",
		ProxyEndpoint:      "203.0.113.1:51820",
		ProxyPublicKey:     "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXoxMjM0NQ==",
		InstancePrivateKey: "cHJpdmF0ZWtleWRhdGFoZXJlMTIzNDU2Nzg5MGFiY2Q=",
		InstanceAddress:    "10.0.0.5/16",
		AllowedIPs:         "10.0.0.0/16",
		SSHAuthorizedKeys:  "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExampleKeyDataHere12345678901234567890 user@host",
		InternalToken:      "tok_test_abc123def456",
		Hostname:           "gpu-4a7f.gpu.ai",
		CallbackURL:        "https://api.gpu.ai/internal/instances/gpu-4a7f/ready",
	}
}

func TestRenderBootstrap(t *testing.T) {
	t.Helper()

	data := validBootstrapData()
	output, err := RenderBootstrap(data)
	if err != nil {
		t.Fatalf("RenderBootstrap failed: %v", err)
	}

	// Verify script starts with shebang
	if !strings.HasPrefix(output, "#!/bin/bash") {
		t.Error("output does not start with #!/bin/bash")
	}

	// Verify WireGuard private key
	if !strings.Contains(output, data.InstancePrivateKey) {
		t.Error("output does not contain instance private key")
	}

	// Verify proxy public key
	if !strings.Contains(output, data.ProxyPublicKey) {
		t.Error("output does not contain proxy public key")
	}

	// Verify proxy endpoint
	if !strings.Contains(output, data.ProxyEndpoint) {
		t.Error("output does not contain proxy endpoint")
	}

	// Verify instance address
	if !strings.Contains(output, data.InstanceAddress) {
		t.Error("output does not contain instance address")
	}

	// Verify SSH authorized keys
	if !strings.Contains(output, data.SSHAuthorizedKeys) {
		t.Error("output does not contain SSH authorized keys")
	}

	// Verify hostname
	if !strings.Contains(output, data.Hostname) {
		t.Error("output does not contain hostname")
	}

	// Verify callback URL with internal token
	if !strings.Contains(output, data.CallbackURL) {
		t.Error("output does not contain callback URL")
	}
	if !strings.Contains(output, data.InternalToken) {
		t.Error("output does not contain internal token")
	}

	// Verify metadata endpoint block
	if !strings.Contains(output, "169.254.169.254") {
		t.Error("output does not contain metadata endpoint block (169.254.169.254)")
	}

	// Verify NVIDIA GPU verification
	if !strings.Contains(output, "nvidia-smi") {
		t.Error("output does not contain nvidia-smi GPU verification")
	}

	// Verify unattended upgrades disabled
	if !strings.Contains(output, "unattended-upgrades") {
		t.Error("output does not contain unattended-upgrades disable section")
	}

	// Verify gpu_info in ready callback
	if !strings.Contains(output, "gpu_info") {
		t.Error("output does not contain gpu_info in ready callback")
	}
}

func TestRenderBootstrapMultipleSSHKeys(t *testing.T) {
	t.Helper()

	data := validBootstrapData()
	key1 := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExampleKeyDataHere12345678901234567890 user1@host1"
	key2 := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC1234567890abcdefghijklmnopqrstuvwxyz user2@host2"
	key3 := "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTY= user3@host3"
	data.SSHAuthorizedKeys = key1 + "\n" + key2 + "\n" + key3

	output, err := RenderBootstrap(data)
	if err != nil {
		t.Fatalf("RenderBootstrap with multiple keys failed: %v", err)
	}

	if !strings.Contains(output, key1) {
		t.Error("output does not contain SSH key 1")
	}
	if !strings.Contains(output, key2) {
		t.Error("output does not contain SSH key 2")
	}
	if !strings.Contains(output, key3) {
		t.Error("output does not contain SSH key 3")
	}
}

func TestValidateBootstrapDataValid(t *testing.T) {
	t.Helper()

	data := validBootstrapData()
	if err := ValidateBootstrapData(data); err != nil {
		t.Fatalf("valid data should pass validation, got error: %v", err)
	}
}

func TestValidateBootstrapDataEmptyFields(t *testing.T) {
	t.Helper()

	data := BootstrapData{} // all empty
	err := ValidateBootstrapData(data)
	if err == nil {
		t.Fatal("empty data should fail validation")
	}

	errMsg := err.Error()

	// Verify all required field errors are reported
	requiredFields := []string{
		"InstanceID",
		"ProxyEndpoint",
		"ProxyPublicKey",
		"InstancePrivateKey",
		"InstanceAddress",
		"AllowedIPs",
		"SSHAuthorizedKeys",
		"InternalToken",
		"Hostname",
		"CallbackURL",
	}
	for _, field := range requiredFields {
		if !strings.Contains(errMsg, field) {
			t.Errorf("error should mention %q, got: %s", field, errMsg)
		}
	}
}

func TestValidateBootstrapDataShellInjectionInstanceID(t *testing.T) {
	t.Helper()

	data := validBootstrapData()
	data.InstanceID = "$(whoami)"

	err := ValidateBootstrapData(data)
	if err == nil {
		t.Fatal("InstanceID with shell injection should fail validation")
	}

	if !strings.Contains(err.Error(), "InstanceID") {
		t.Errorf("error should mention InstanceID, got: %v", err)
	}
}

func TestValidateBootstrapDataShellInjectionSSHKey(t *testing.T) {
	t.Helper()

	data := validBootstrapData()
	data.SSHAuthorizedKeys = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExampleKeyDataHere12345678901234567890 $(malicious)"

	err := ValidateBootstrapData(data)
	if err == nil {
		t.Fatal("SSH key with shell injection should fail validation")
	}

	if !strings.Contains(err.Error(), "SSHAuthorizedKeys") {
		t.Errorf("error should mention SSHAuthorizedKeys, got: %v", err)
	}
}

func TestValidateBootstrapDataShellInjectionHostname(t *testing.T) {
	t.Helper()

	data := validBootstrapData()
	data.Hostname = "gpu;rm -rf /"

	err := ValidateBootstrapData(data)
	if err == nil {
		t.Fatal("Hostname with shell injection should fail validation")
	}

	if !strings.Contains(err.Error(), "Hostname") {
		t.Errorf("error should mention Hostname, got: %v", err)
	}
}

func TestRenderBootstrapInvalidData(t *testing.T) {
	t.Helper()

	data := BootstrapData{} // all empty -- should fail validation
	_, err := RenderBootstrap(data)
	if err == nil {
		t.Fatal("RenderBootstrap should return error for invalid data")
	}

	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("error should mention validation, got: %v", err)
	}
}

func TestBootstrapTemplateCompiles(t *testing.T) {
	t.Helper()

	if bootstrapTmpl == nil {
		t.Fatal("bootstrapTmpl should not be nil (template parse error at package init)")
	}
}

func TestRenderBootstrapOutputIsValidBash(t *testing.T) {
	t.Helper()

	data := validBootstrapData()
	output, err := RenderBootstrap(data)
	if err != nil {
		t.Fatalf("RenderBootstrap failed: %v", err)
	}

	// After rendering, no Go template syntax should remain
	if strings.Contains(output, "{{") {
		t.Error("rendered output still contains '{{' -- Go template placeholders not fully expanded")
	}
	if strings.Contains(output, "}}") {
		t.Error("rendered output still contains '}}' -- Go template placeholders not fully expanded")
	}

	// Output should start with shebang
	if !strings.HasPrefix(output, "#!/bin/bash") {
		t.Error("rendered output does not start with #!/bin/bash")
	}
}

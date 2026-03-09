package tunnel

import (
	"strings"
	"testing"
)

// validFRPBootstrapData returns a BootstrapData with all valid fields for testing.
func validFRPBootstrapData() BootstrapData {
	return BootstrapData{
		InstanceID:        "gpu-4a7f",
		ProxyHost:         "134.199.214.138",
		FRPServerPort:     7000,
		FRPToken:          "tok_frp_test_abc123def456",
		RemotePort:        10002,
		SSHAuthorizedKeys: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExampleKeyDataHere12345678901234567890 user@host",
		InternalToken:     "tok_test_abc123def456",
		Hostname:          "gpu-4a7f.gpu.ai",
		CallbackURL:       "https://api.gpu.ai/internal/instances/gpu-4a7f/ready",
	}
}

func TestRenderBootstrap(t *testing.T) {
	data := validFRPBootstrapData()
	output, err := RenderBootstrap(data)
	if err != nil {
		t.Fatalf("RenderBootstrap failed: %v", err)
	}

	// Verify script starts with shebang
	if !strings.HasPrefix(output, "#!/bin/bash") {
		t.Error("output does not start with #!/bin/bash")
	}

	// Verify frpc TOML config with correct values
	if !strings.Contains(output, `serverAddr = "134.199.214.138"`) {
		t.Error("output does not contain correct serverAddr")
	}
	if !strings.Contains(output, "serverPort = 7000") {
		t.Error("output does not contain correct serverPort")
	}
	if !strings.Contains(output, `auth.token = "tok_frp_test_abc123def456"`) {
		t.Error("output does not contain correct auth.token")
	}
	if !strings.Contains(output, "remotePort = 10002") {
		t.Error("output does not contain correct remotePort")
	}
	if !strings.Contains(output, "localPort = 22") {
		t.Error("output does not contain localPort = 22")
	}

	// Verify no Go template syntax remains
	if strings.Contains(output, "{{") {
		t.Error("rendered output still contains '{{' -- Go template placeholders not fully expanded")
	}
	if strings.Contains(output, "}}") {
		t.Error("rendered output still contains '}}' -- Go template placeholders not fully expanded")
	}
}

func TestRenderBootstrap_SSHKeys(t *testing.T) {
	data := validFRPBootstrapData()
	key1 := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExampleKeyDataHere12345678901234567890 user1@host1"
	key2 := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC1234567890abcdefghijklmnopqrstuvwxyz user2@host2"
	data.SSHAuthorizedKeys = key1 + "\n" + key2

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

	// Verify authorized_keys section exists
	if !strings.Contains(output, "authorized_keys") {
		t.Error("output does not contain authorized_keys section")
	}
}

func TestRenderBootstrap_Callback(t *testing.T) {
	data := validFRPBootstrapData()
	output, err := RenderBootstrap(data)
	if err != nil {
		t.Fatalf("RenderBootstrap failed: %v", err)
	}

	// Verify callback URL is present
	if !strings.Contains(output, data.CallbackURL) {
		t.Error("output does not contain callback URL")
	}

	// Verify internal token is used for auth
	if !strings.Contains(output, data.InternalToken) {
		t.Error("output does not contain internal token")
	}

	// Verify curl is used for callback
	if !strings.Contains(output, "curl") {
		t.Error("output does not contain curl for ready callback")
	}
}

func TestRenderBootstrap_FRPCDownload(t *testing.T) {
	data := validFRPBootstrapData()
	output, err := RenderBootstrap(data)
	if err != nil {
		t.Fatalf("RenderBootstrap failed: %v", err)
	}

	// Verify frpc binary download
	if !strings.Contains(output, "frpc") {
		t.Error("output does not contain frpc binary reference")
	}

	// Verify frpc config file
	if !strings.Contains(output, "frpc.toml") {
		t.Error("output does not contain frpc.toml config file")
	}
}

func TestRenderBootstrap_InvalidData(t *testing.T) {
	data := BootstrapData{} // all empty
	_, err := RenderBootstrap(data)
	if err == nil {
		t.Fatal("RenderBootstrap should return error for invalid data")
	}

	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("error should mention validation, got: %v", err)
	}
}

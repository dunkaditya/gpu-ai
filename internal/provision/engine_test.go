package provision

import (
	"testing"
)

// TestBuildCallbackURL verifies callback URL construction with and without
// GPUCTL_PUBLIC_URL configured.
func TestBuildCallbackURL(t *testing.T) {
	tests := []struct {
		name       string
		publicURL  string
		hostname   string
		instanceID string
		want       string
	}{
		{
			name:       "with GPUCTL_PUBLIC_URL",
			publicURL:  "https://api.gpu.ai",
			hostname:   "gpu-abc123.gpu.ai",
			instanceID: "gpu-abc123",
			want:       "https://api.gpu.ai/internal/instances/gpu-abc123/ready",
		},
		{
			name:       "with trailing slash on public URL",
			publicURL:  "https://api.gpu.ai/",
			hostname:   "gpu-abc123.gpu.ai",
			instanceID: "gpu-abc123",
			want:       "https://api.gpu.ai/internal/instances/gpu-abc123/ready",
		},
		{
			name:       "without GPUCTL_PUBLIC_URL falls back to hostname",
			publicURL:  "",
			hostname:   "gpu-abc123.gpu.ai",
			instanceID: "gpu-abc123",
			want:       "https://gpu-abc123.gpu.ai/internal/instances/gpu-abc123/ready",
		},
		{
			name:       "public URL with path prefix",
			publicURL:  "https://api.gpu.ai/v1",
			hostname:   "gpu-def456.gpu.ai",
			instanceID: "gpu-def456",
			want:       "https://api.gpu.ai/v1/internal/instances/gpu-def456/ready",
		},
		{
			name:       "public URL with multiple trailing slashes",
			publicURL:  "https://api.gpu.ai///",
			hostname:   "gpu-abc123.gpu.ai",
			instanceID: "gpu-abc123",
			want:       "https://api.gpu.ai/internal/instances/gpu-abc123/ready",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCallbackURL(tt.publicURL, tt.hostname, tt.instanceID)
			if got != tt.want {
				t.Errorf("buildCallbackURL(%q, %q, %q) = %q, want %q",
					tt.publicURL, tt.hostname, tt.instanceID, got, tt.want)
			}
		})
	}
}

// TestGenerateInstanceID verifies the instance ID format: "gpu-" + 8 hex chars.
func TestGenerateInstanceID(t *testing.T) {
	id, err := generateInstanceID()
	if err != nil {
		t.Fatalf("generateInstanceID() error: %v", err)
	}

	// Must start with "gpu-".
	if len(id) < 4 || id[:4] != "gpu-" {
		t.Errorf("generateInstanceID() = %q, want prefix 'gpu-'", id)
	}

	// Total length: "gpu-" (4) + 8 hex chars = 12.
	if len(id) != 12 {
		t.Errorf("generateInstanceID() length = %d, want 12", len(id))
	}

	// Hex suffix should only contain valid hex characters.
	for _, c := range id[4:] {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("generateInstanceID() contains non-hex char %q in suffix %q", string(c), id[4:])
		}
	}

	// Two calls should produce different IDs (extremely unlikely to collide).
	id2, err := generateInstanceID()
	if err != nil {
		t.Fatalf("generateInstanceID() second call error: %v", err)
	}
	if id == id2 {
		t.Errorf("generateInstanceID() produced duplicate IDs: %q", id)
	}
}

// TestGenerateHexToken verifies the token length matches byteLen * 2 (hex encoding).
func TestGenerateHexToken(t *testing.T) {
	tests := []struct {
		name    string
		byteLen int
		wantLen int
	}{
		{name: "16 bytes", byteLen: 16, wantLen: 32},
		{name: "32 bytes", byteLen: 32, wantLen: 64},
		{name: "1 byte", byteLen: 1, wantLen: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := generateHexToken(tt.byteLen)
			if err != nil {
				t.Fatalf("generateHexToken(%d) error: %v", tt.byteLen, err)
			}
			if len(token) != tt.wantLen {
				t.Errorf("generateHexToken(%d) length = %d, want %d", tt.byteLen, len(token), tt.wantLen)
			}

			// Verify all characters are valid hex.
			for _, c := range token {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
					t.Errorf("generateHexToken(%d) contains non-hex char %q", tt.byteLen, string(c))
				}
			}
		})
	}

	// Two calls should produce different tokens.
	t1, _ := generateHexToken(16)
	t2, _ := generateHexToken(16)
	if t1 == t2 {
		t.Errorf("generateHexToken(16) produced duplicate tokens: %q", t1)
	}
}

// TestSetOnStatusChange verifies the setter stores the callback function.
func TestSetOnStatusChange(t *testing.T) {
	engine := &Engine{}

	if engine.onStatusChange != nil {
		t.Fatal("onStatusChange should be nil before SetOnStatusChange")
	}

	var called bool
	engine.SetOnStatusChange(func(instanceID, status string) {
		called = true
	})

	if engine.onStatusChange == nil {
		t.Fatal("onStatusChange should be set after SetOnStatusChange")
	}

	// Invoke the callback to verify it's wired correctly.
	engine.onStatusChange("test-id", "provisioning")
	if !called {
		t.Error("onStatusChange callback was not invoked")
	}
}

// Note: Full progressStatus polling loop is tested via integration tests with
// real DB and provider mocks. The polling loop interacts with e.db (a *db.Pool)
// and provider.GetStatus which require either a running database or interface-based
// mocking (deferred to Phase 6 test infrastructure).

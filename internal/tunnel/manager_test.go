package tunnel

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	mgr, err := NewManager(7000, "test-token-abc123", "10000-10255", nil)
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}
	if mgr == nil {
		t.Fatal("NewManager() returned nil manager")
	}
	if mgr.frpSvc == nil {
		t.Error("NewManager() frpSvc should not be nil")
	}
	// Clean up
	_ = mgr.Close()
}

func TestNewManager_InvalidPort(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"zero port", 0},
		{"negative port", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewManager(tt.port, "test-token", "10000-10255", nil)
			if err == nil {
				t.Error("NewManager() with invalid port should return error, got nil")
			}
		})
	}
}

func TestNewManager_EmptyToken(t *testing.T) {
	// Empty token should still create a manager (token is optional for dev)
	mgr, err := NewManager(7000, "", "10000-10255", nil)
	if err != nil {
		t.Fatalf("NewManager() with empty token error: %v", err)
	}
	if mgr == nil {
		t.Fatal("NewManager() returned nil manager")
	}
	_ = mgr.Close()
}

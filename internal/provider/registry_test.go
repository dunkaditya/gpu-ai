package provider

import (
	"context"
	"testing"
)

// mockProvider is a minimal Provider implementation for testing.
// Only Name() returns a meaningful value; other methods return zero values.
type mockProvider struct {
	name string
}

func (m *mockProvider) Name() string { return m.name }

func (m *mockProvider) ListAvailable(_ context.Context) ([]GPUOffering, error) {
	return nil, nil
}

func (m *mockProvider) Provision(_ context.Context, _ ProvisionRequest) (*ProvisionResult, error) {
	return nil, nil
}

func (m *mockProvider) GetStatus(_ context.Context, _ string) (*InstanceStatus, error) {
	return nil, nil
}

func (m *mockProvider) Terminate(_ context.Context, _ string) error {
	return nil
}

func TestRegistryRegisterAndGet(t *testing.T) {
	reg := NewRegistry()
	mock := &mockProvider{name: "runpod"}

	reg.Register(mock)

	// Get existing provider
	p, ok := reg.Get("runpod")
	if !ok {
		t.Fatal("expected to find registered provider 'runpod'")
	}
	if p.Name() != "runpod" {
		t.Fatalf("expected provider name 'runpod', got %q", p.Name())
	}

	// Get unknown provider
	_, ok = reg.Get("unknown")
	if ok {
		t.Fatal("expected Get('unknown') to return false")
	}
}

func TestRegistryAll(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&mockProvider{name: "runpod"})
	reg.Register(&mockProvider{name: "e2e"})

	all := reg.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(all))
	}

	// Verify both are present (order not guaranteed)
	names := make(map[string]bool)
	for _, p := range all {
		names[p.Name()] = true
	}
	if !names["runpod"] {
		t.Error("expected 'runpod' in All() result")
	}
	if !names["e2e"] {
		t.Error("expected 'e2e' in All() result")
	}
}

func TestRegistryNames(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&mockProvider{name: "runpod"})
	reg.Register(&mockProvider{name: "e2e"})
	reg.Register(&mockProvider{name: "lambda"})

	names := reg.Names()
	if len(names) != 3 {
		t.Fatalf("expected 3 names, got %d", len(names))
	}

	// Names() returns sorted list
	expected := []string{"e2e", "lambda", "runpod"}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("expected names[%d] = %q, got %q", i, expected[i], name)
		}
	}
}

func TestRegistryReRegister(t *testing.T) {
	reg := NewRegistry()

	original := &mockProvider{name: "runpod"}
	replacement := &mockProvider{name: "runpod"}

	reg.Register(original)
	reg.Register(replacement)

	p, ok := reg.Get("runpod")
	if !ok {
		t.Fatal("expected to find provider 'runpod' after re-registration")
	}

	// Verify we got the replacement, not the original
	if p != replacement {
		t.Fatal("expected Get to return the replacement provider after re-registration")
	}

	// Should still only have one provider
	all := reg.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 provider after re-registration, got %d", len(all))
	}
}

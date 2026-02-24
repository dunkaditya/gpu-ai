package wireguard

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// mockRows implements a minimal pgx.Rows for QueryRow mock.
type mockRows struct {
	maxAddr string
	closed  bool
}

func (m *mockRows) Close()                                        {}
func (m *mockRows) Err() error                                    { return nil }
func (m *mockRows) CommandTag() pgconn.CommandTag                 { return pgconn.CommandTag{} }
func (m *mockRows) FieldDescriptions() []pgconn.FieldDescription  { return nil }
func (m *mockRows) Next() bool                                    { defer func() { m.closed = true }(); return !m.closed }
func (m *mockRows) Scan(dest ...any) error {
	if len(dest) > 0 {
		if s, ok := dest[0].(*string); ok {
			*s = m.maxAddr
		}
	}
	return nil
}
func (m *mockRows) Values() ([]any, error)                        { return nil, nil }
func (m *mockRows) RawValues() [][]byte                           { return nil }
func (m *mockRows) Conn() *pgx.Conn                               { return nil }

// mockTx implements pgx.Tx for testing IPAM allocation logic.
type mockTx struct {
	maxAddr   string // What the mock returns as max wg_address
	execErr   error  // Error to return from Exec (advisory lock)
	queryErr  error  // Error to return from QueryRow scan
	execCalls int
}

func (m *mockTx) Begin(ctx context.Context) (pgx.Tx, error) { return m, nil }
func (m *mockTx) Commit(ctx context.Context) error          { return nil }
func (m *mockTx) Rollback(ctx context.Context) error        { return nil }

func (m *mockTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}

func (m *mockTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return nil
}

func (m *mockTx) LargeObjects() pgx.LargeObjects {
	return pgx.LargeObjects{}
}

func (m *mockTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return nil, nil
}

func (m *mockTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	m.execCalls++
	if m.execErr != nil {
		return pgconn.CommandTag{}, m.execErr
	}
	return pgconn.CommandTag{}, nil
}

func (m *mockTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if m.queryErr != nil {
		return nil, m.queryErr
	}
	return &mockRows{maxAddr: m.maxAddr}, nil
}

func (m *mockTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return &mockRow{maxAddr: m.maxAddr, err: m.queryErr}
}

func (m *mockTx) Conn() *pgx.Conn { return nil }

// mockRow implements pgx.Row for QueryRow results.
type mockRow struct {
	maxAddr string
	err     error
}

func (r *mockRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) > 0 {
		if s, ok := dest[0].(*string); ok {
			*s = r.maxAddr
		}
	}
	return nil
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestNewIPAM(t *testing.T) {
	ipam, err := NewIPAM("10.0.0.0/16", testLogger())
	if err != nil {
		t.Fatalf("NewIPAM() error: %v", err)
	}

	if ipam.subnet.String() != "10.0.0.0/16" {
		t.Errorf("subnet = %q, want %q", ipam.subnet.String(), "10.0.0.0/16")
	}

	expectedProxy := net.ParseIP("10.0.0.1").To4()
	if !ipam.proxyAddr.Equal(expectedProxy) {
		t.Errorf("proxyAddr = %v, want %v", ipam.proxyAddr, expectedProxy)
	}
}

func TestNewIPAMInvalidCIDR(t *testing.T) {
	_, err := NewIPAM("not-a-cidr", testLogger())
	if err == nil {
		t.Error("NewIPAM() with invalid CIDR should return error, got nil")
	}
}

func TestIncrementIP(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"10.0.0.1", "10.0.0.2"},
		{"10.0.0.254", "10.0.0.255"},
		{"10.0.0.255", "10.0.1.0"},     // single carry
		{"10.0.255.255", "10.1.0.0"},   // double carry
		{"10.0.1.0", "10.0.1.1"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s->%s", tt.input, tt.want), func(t *testing.T) {
			input := net.ParseIP(tt.input).To4()
			got := incrementIP(input)

			want := net.ParseIP(tt.want).To4()
			if !got.Equal(want) {
				t.Errorf("incrementIP(%s) = %s, want %s", tt.input, got.String(), tt.want)
			}

			// Verify original IP was not mutated.
			originalInput := net.ParseIP(tt.input).To4()
			if !input.Equal(originalInput) {
				t.Errorf("incrementIP mutated original IP: got %s, want %s", input.String(), tt.input)
			}
		})
	}
}

func TestIsProxyAddress(t *testing.T) {
	ipam, err := NewIPAM("10.0.0.0/16", testLogger())
	if err != nil {
		t.Fatalf("NewIPAM() error: %v", err)
	}

	if !ipam.IsProxyAddress(net.ParseIP("10.0.0.1")) {
		t.Error("IsProxyAddress(10.0.0.1) = false, want true")
	}

	if ipam.IsProxyAddress(net.ParseIP("10.0.0.2")) {
		t.Error("IsProxyAddress(10.0.0.2) = true, want false")
	}

	if ipam.IsProxyAddress(net.ParseIP("192.168.1.1")) {
		t.Error("IsProxyAddress(192.168.1.1) = true, want false")
	}
}

func TestSubnetCIDR(t *testing.T) {
	ipam, err := NewIPAM("10.0.0.0/16", testLogger())
	if err != nil {
		t.Fatalf("NewIPAM() error: %v", err)
	}

	if got := ipam.SubnetCIDR(); got != "10.0.0.0/16" {
		t.Errorf("SubnetCIDR() = %q, want %q", got, "10.0.0.0/16")
	}
}

func TestAllocateAddressWithMockTx(t *testing.T) {
	ipam, err := NewIPAM("10.0.0.0/16", testLogger())
	if err != nil {
		t.Fatalf("NewIPAM() error: %v", err)
	}

	tests := []struct {
		name    string
		maxAddr string
		want    string
	}{
		{
			name:    "first allocation after proxy",
			maxAddr: "10.0.0.1",
			want:    "10.0.0.2",
		},
		{
			name:    "sequential allocation from 10.0.0.5",
			maxAddr: "10.0.0.5",
			want:    "10.0.0.6",
		},
		{
			name:    "near end of second-to-last octet",
			maxAddr: "10.0.255.254",
			want:    "10.0.255.255",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &mockTx{maxAddr: tt.maxAddr}
			got, err := ipam.AllocateAddress(context.Background(), tx)
			if err != nil {
				t.Fatalf("AllocateAddress() error: %v", err)
			}

			want := net.ParseIP(tt.want).To4()
			if !got.Equal(want) {
				t.Errorf("AllocateAddress() = %s, want %s", got.String(), tt.want)
			}

			// Verify advisory lock was acquired (Exec called).
			if tx.execCalls != 1 {
				t.Errorf("advisory lock Exec calls = %d, want 1", tx.execCalls)
			}
		})
	}
}

func TestAllocateAddressSubnetExhausted(t *testing.T) {
	ipam, err := NewIPAM("10.0.0.0/16", testLogger())
	if err != nil {
		t.Fatalf("NewIPAM() error: %v", err)
	}

	// 10.0.255.255 is the last address in 10.0.0.0/16.
	// Incrementing it yields 10.1.0.0 which is outside the subnet.
	tx := &mockTx{maxAddr: "10.0.255.255"}
	_, err = ipam.AllocateAddress(context.Background(), tx)
	if err == nil {
		t.Error("AllocateAddress() should return error when subnet exhausted, got nil")
	}

	// Verify the error mentions exhaustion.
	if err != nil && !contains(err.Error(), "exhausted") {
		t.Errorf("error should mention exhaustion, got: %v", err)
	}
}

func TestAllocateAddressAdvisoryLockFailure(t *testing.T) {
	ipam, err := NewIPAM("10.0.0.0/16", testLogger())
	if err != nil {
		t.Fatalf("NewIPAM() error: %v", err)
	}

	tx := &mockTx{
		maxAddr: "10.0.0.1",
		execErr: fmt.Errorf("connection lost"),
	}

	_, err = ipam.AllocateAddress(context.Background(), tx)
	if err == nil {
		t.Error("AllocateAddress() should return error when advisory lock fails, got nil")
	}
}

func TestAllocateAddressQueryFailure(t *testing.T) {
	ipam, err := NewIPAM("10.0.0.0/16", testLogger())
	if err != nil {
		t.Fatalf("NewIPAM() error: %v", err)
	}

	tx := &mockTx{
		maxAddr:  "10.0.0.1",
		queryErr: fmt.Errorf("query timeout"),
	}

	_, err = ipam.AllocateAddress(context.Background(), tx)
	if err == nil {
		t.Error("AllocateAddress() should return error when query fails, got nil")
	}
}

// contains is a simple helper to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

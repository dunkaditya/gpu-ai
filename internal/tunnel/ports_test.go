package tunnel

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// mockRows implements a minimal pgx.Rows for QueryRow mock.
type mockRows struct {
	maxPort int
	closed  bool
}

func (m *mockRows) Close()                                        {}
func (m *mockRows) Err() error                                    { return nil }
func (m *mockRows) CommandTag() pgconn.CommandTag                 { return pgconn.CommandTag{} }
func (m *mockRows) FieldDescriptions() []pgconn.FieldDescription  { return nil }
func (m *mockRows) Next() bool                                    { defer func() { m.closed = true }(); return !m.closed }
func (m *mockRows) Scan(dest ...any) error {
	if len(dest) > 0 {
		if s, ok := dest[0].(*int); ok {
			*s = m.maxPort
		}
	}
	return nil
}
func (m *mockRows) Values() ([]any, error) { return nil, nil }
func (m *mockRows) RawValues() [][]byte    { return nil }
func (m *mockRows) Conn() *pgx.Conn       { return nil }

// mockTx implements pgx.Tx for testing port allocation logic.
type mockTx struct {
	maxPort   int   // What the mock returns as max frp_remote_port
	execErr   error // Error to return from Exec (advisory lock)
	queryErr  error // Error to return from QueryRow scan
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
	return &mockRows{maxPort: m.maxPort}, nil
}

func (m *mockTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return &mockRow{maxPort: m.maxPort, err: m.queryErr}
}

func (m *mockTx) Conn() *pgx.Conn { return nil }

// mockRow implements pgx.Row for QueryRow results.
type mockRow struct {
	maxPort int
	err     error
}

func (r *mockRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) > 0 {
		if s, ok := dest[0].(*int); ok {
			*s = r.maxPort
		}
	}
	return nil
}

func TestAllocatePort(t *testing.T) {
	// First allocation: maxPort = MinPort-1 (9999) -> returns MinPort (10000)
	tx := &mockTx{maxPort: MinPort - 1}
	port, err := AllocatePort(context.Background(), tx)
	if err != nil {
		t.Fatalf("AllocatePort() first call error: %v", err)
	}
	if port != MinPort {
		t.Errorf("AllocatePort() first = %d, want %d", port, MinPort)
	}
	if tx.execCalls != 1 {
		t.Errorf("advisory lock Exec calls = %d, want 1", tx.execCalls)
	}

	// Second allocation: maxPort = MinPort (10000) -> returns 10001
	tx2 := &mockTx{maxPort: MinPort}
	port2, err := AllocatePort(context.Background(), tx2)
	if err != nil {
		t.Fatalf("AllocatePort() second call error: %v", err)
	}
	if port2 != MinPort+1 {
		t.Errorf("AllocatePort() second = %d, want %d", port2, MinPort+1)
	}
}

func TestAllocatePort_Exhaustion(t *testing.T) {
	// MaxPort is already allocated -> next would exceed range
	tx := &mockTx{maxPort: MaxPort}
	_, err := AllocatePort(context.Background(), tx)
	if err == nil {
		t.Error("AllocatePort() should return error when range exhausted, got nil")
	}
	if err != nil && !containsStr(err.Error(), "exhausted") {
		t.Errorf("error should mention exhaustion, got: %v", err)
	}
}

func TestAllocatePort_SkipsTerminated(t *testing.T) {
	// If terminated ports are excluded, the max in the active range is lower.
	// maxPort = 10002 means 10000, 10001, 10002 are active; next is 10003.
	tx := &mockTx{maxPort: 10002}
	port, err := AllocatePort(context.Background(), tx)
	if err != nil {
		t.Fatalf("AllocatePort() error: %v", err)
	}
	if port != 10003 {
		t.Errorf("AllocatePort() = %d, want 10003", port)
	}
}

func TestAllocatePort_AdvisoryLockFailure(t *testing.T) {
	tx := &mockTx{
		maxPort: MinPort - 1,
		execErr: fmt.Errorf("connection lost"),
	}
	_, err := AllocatePort(context.Background(), tx)
	if err == nil {
		t.Error("AllocatePort() should return error when advisory lock fails, got nil")
	}
}

func TestAllocatePort_QueryFailure(t *testing.T) {
	tx := &mockTx{
		maxPort:  MinPort - 1,
		queryErr: fmt.Errorf("query timeout"),
	}
	_, err := AllocatePort(context.Background(), tx)
	if err == nil {
		t.Error("AllocatePort() should return error when query fails, got nil")
	}
}

// containsStr is a simple helper to check if a string contains a substring.
func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

package api

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

// okHandler is a simple handler that writes 200 + "ok" body.
// Used as the "next" handler in middleware tests.
func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
}

// --- LocalhostOnly tests ---

func TestLocalhostOnly_AllowsLoopbackIPv4(t *testing.T) {
	handler := LocalhostOnly(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Errorf("expected body 'ok', got %q", rec.Body.String())
	}
}

func TestLocalhostOnly_AllowsLoopbackIPv6(t *testing.T) {
	handler := LocalhostOnly(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.RemoteAddr = net.JoinHostPort("::1", "12345")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Errorf("expected body 'ok', got %q", rec.Body.String())
	}
}

func TestLocalhostOnly_RejectsExternalIPv4(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	handler := LocalhostOnly(next)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
	if called {
		t.Error("next handler should not have been called for external IP")
	}
}

func TestLocalhostOnly_RejectsExternalIPv6(t *testing.T) {
	handler := LocalhostOnly(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.RemoteAddr = "[2001:db8::1]:12345"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestLocalhostOnly_RejectsMalformedAddr(t *testing.T) {
	handler := LocalhostOnly(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.RemoteAddr = "garbage"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// --- InternalAuthMiddleware regression tests ---

func TestInternalAuthMiddleware_ValidToken(t *testing.T) {
	token := "test-secret-token"
	handler := InternalAuthMiddleware(token, okHandler())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Errorf("expected body 'ok', got %q", rec.Body.String())
	}
}

func TestInternalAuthMiddleware_InvalidToken(t *testing.T) {
	token := "test-secret-token"
	handler := InternalAuthMiddleware(token, okHandler())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

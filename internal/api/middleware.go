package api

import (
	"encoding/json"
	"net"
	"net/http"
)

// LocalhostOnly restricts access to requests originating from loopback addresses
// (127.0.0.1 or ::1). Non-loopback requests receive a 404 Not Found to avoid
// revealing the endpoint's existence to external scanners.
func LocalhostOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		if host != "127.0.0.1" && host != "::1" {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// InternalAuthMiddleware protects internal endpoints with a shared token.
// It checks the Authorization header for "Bearer <token>" and returns
// 403 Forbidden if the token is missing or does not match.
func InternalAuthMiddleware(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer "+token {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "forbidden"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

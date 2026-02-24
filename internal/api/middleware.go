package api

import (
	"encoding/json"
	"net/http"
)

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

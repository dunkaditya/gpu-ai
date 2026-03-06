package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

func (sr *statusRecorder) Write(b []byte) (int, error) {
	if sr.status == 0 {
		sr.status = http.StatusOK
	}
	return sr.ResponseWriter.Write(b)
}

// RequestLogMiddleware logs every HTTP request with method, path, status, duration, and remote address.
func RequestLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sr := &statusRecorder{ResponseWriter: w}

		next.ServeHTTP(sr, r)

		if sr.status == 0 {
			sr.status = http.StatusOK
		}
		duration := time.Since(start)

		attrs := []slog.Attr{
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", sr.status),
			slog.Duration("duration", duration),
			slog.String("remote_addr", r.RemoteAddr),
		}

		if claims, ok := clerk.SessionClaimsFromContext(r.Context()); ok && claims != nil && claims.ActiveOrganizationID != "" {
			attrs = append(attrs, slog.String("org_id", claims.ActiveOrganizationID))
		}

		level := slog.LevelInfo
		if sr.status >= 500 {
			level = slog.LevelError
		} else if sr.status >= 400 {
			level = slog.LevelWarn
		}

		slog.LogAttrs(r.Context(), level, "http request", attrs...)
	})
}

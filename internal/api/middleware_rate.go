package api

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"golang.org/x/time/rate"
)

// limiterEntry wraps a rate.Limiter with a last-seen timestamp for cleanup.
type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
	mu       sync.Mutex
}

// OrgRateLimiter provides per-organization rate limiting using token bucket algorithm.
type OrgRateLimiter struct {
	limiters  sync.Map
	rateLimit rate.Limit
	burst     int
}

// NewOrgRateLimiter creates a new per-organization rate limiter.
// r is the rate (requests per second) and burst is the maximum burst size.
func NewOrgRateLimiter(r rate.Limit, burst int) *OrgRateLimiter {
	return &OrgRateLimiter{
		rateLimit: r,
		burst:     burst,
	}
}

// GetLimiter returns the rate.Limiter for the given organization ID,
// creating one if it does not exist.
func (o *OrgRateLimiter) GetLimiter(orgID string) *rate.Limiter {
	val, loaded := o.limiters.LoadOrStore(orgID, &limiterEntry{
		limiter:  rate.NewLimiter(o.rateLimit, o.burst),
		lastSeen: time.Now(),
	})
	entry := val.(*limiterEntry)
	if loaded {
		entry.mu.Lock()
		entry.lastSeen = time.Now()
		entry.mu.Unlock()
	}
	return entry.limiter
}

// Middleware returns an HTTP middleware that enforces per-organization rate limits.
// It extracts the organization ID from Clerk session claims. If claims are not
// available (e.g., auth middleware not applied), the request passes through --
// the auth middleware is responsible for rejecting unauthenticated requests.
func (o *OrgRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := clerk.SessionClaimsFromContext(r.Context())
		if !ok || claims == nil || claims.ActiveOrganizationID == "" {
			// No claims available -- let auth middleware handle rejection.
			next.ServeHTTP(w, r)
			return
		}

		limiter := o.GetLimiter(claims.ActiveOrganizationID)
		if !limiter.Allow() {
			w.Header().Set("Retry-After", "1")
			writeProblem(w, http.StatusTooManyRequests, "rate-limited",
				"Rate limit exceeded. Try again later.")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// StartCleanup launches a background goroutine that periodically removes
// rate limiter entries not seen in the last 10 minutes. The goroutine stops
// when the provided context is canceled.
func (o *OrgRateLimiter) StartCleanup(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				o.CleanupStale()
			}
		}
	}()
}

// CleanupStale removes limiter entries that have not been accessed in the
// last 10 minutes.
func (o *OrgRateLimiter) CleanupStale() {
	cutoff := time.Now().Add(-10 * time.Minute)
	o.limiters.Range(func(key, value any) bool {
		entry := value.(*limiterEntry)
		entry.mu.Lock()
		lastSeen := entry.lastSeen
		entry.mu.Unlock()
		if lastSeen.Before(cutoff) {
			o.limiters.Delete(key)
		}
		return true
	})
}

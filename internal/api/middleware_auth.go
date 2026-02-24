package api

import (
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
)

// ClerkAuthMiddleware returns an HTTP middleware that verifies Clerk JWTs.
//
// If clerkSecretKey is empty, the middleware always returns 401 with a
// ProblemDetail error. This prevents silent pass-through in dev environments
// without Clerk configured -- endpoints are explicitly marked as unconfigured.
//
// If clerkSecretKey is set, the middleware delegates to the Clerk SDK's
// RequireHeaderAuthorization which verifies the JWT, fetches JWKS, and
// injects SessionClaims into the request context.
func ClerkAuthMiddleware(clerkSecretKey string) func(http.Handler) http.Handler {
	if clerkSecretKey == "" {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				writeProblem(w, http.StatusUnauthorized, "auth-not-configured",
					"Authentication not configured. Set CLERK_SECRET_KEY to enable.")
			})
		}
	}

	clerk.SetKey(clerkSecretKey)
	return clerkhttp.RequireHeaderAuthorization()
}

// RequireOrg is an HTTP middleware that ensures the authenticated user has
// an active organization selected. It must be applied after ClerkAuthMiddleware.
//
// Returns 401 if session claims are missing (auth middleware not applied or failed).
// Returns 403 if ActiveOrganizationID is empty (no org selected).
func RequireOrg(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := clerk.SessionClaimsFromContext(r.Context())
		if !ok || claims == nil {
			writeProblem(w, http.StatusUnauthorized, "unauthenticated",
				"Valid authentication required")
			return
		}

		if claims.ActiveOrganizationID == "" {
			writeProblem(w, http.StatusForbidden, "org-required",
				"Active organization required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

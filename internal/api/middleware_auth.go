package api

import (
	"log/slog"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
)

// ClerkAuthMiddleware returns an HTTP middleware that verifies Clerk JWTs.
//
// If clerkSecretKey is empty, the middleware passes requests through with a
// dev user context (org "dev-org", user "dev-user"). This allows the dashboard
// to work in local dev without Clerk configured.
//
// If clerkSecretKey is set, the middleware delegates to the Clerk SDK's
// RequireHeaderAuthorization which verifies the JWT, fetches JWKS, and
// injects SessionClaims into the request context.
func ClerkAuthMiddleware(clerkSecretKey string) func(http.Handler) http.Handler {
	if clerkSecretKey == "" {
		slog.Warn("CLERK_SECRET_KEY not set — auth disabled, using dev user")
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Inject dev session claims so handlers can extract org/user IDs
				claims := clerk.SessionClaims{
					RegisteredClaims: clerk.RegisteredClaims{
						Subject: "dev-user",
					},
					Claims: clerk.Claims{
						ActiveOrganizationID: "dev-org",
					},
				}
				ctx := clerk.ContextWithSessionClaims(r.Context(), &claims)
				next.ServeHTTP(w, r.WithContext(ctx))
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

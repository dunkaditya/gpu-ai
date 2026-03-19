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
// dev user context (user "dev-user"). This allows the dashboard to work in
// local dev without Clerk configured. ActiveOrganizationID is left empty so
// that ClaimsFromContext synthesizes "personal_dev-user" as the org ID.
//
// If clerkSecretKey is set, the middleware delegates to the Clerk SDK's
// RequireHeaderAuthorization which verifies the JWT, fetches JWKS, and
// injects SessionClaims into the request context.
func ClerkAuthMiddleware(clerkSecretKey string) func(http.Handler) http.Handler {
	if clerkSecretKey == "" {
		slog.Warn("CLERK_SECRET_KEY not set — auth disabled, using dev user")
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				claims := clerk.SessionClaims{
					RegisteredClaims: clerk.RegisteredClaims{
						Subject: "dev-user",
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

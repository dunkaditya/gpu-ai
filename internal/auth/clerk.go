// Package auth handles Clerk JWT verification and user context extraction.
package auth

import (
	"context"

	"github.com/clerk/clerk-sdk-go/v2"
)

// Claims holds the authenticated user's identity extracted from a Clerk JWT.
type Claims struct {
	UserID string
	OrgID  string
}

// ClaimsFromContext extracts Claims from the request context.
// It reads the Clerk session claims set by the Clerk SDK middleware and
// maps them to the project-specific Claims struct.
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	sessionClaims, ok := clerk.SessionClaimsFromContext(ctx)
	if !ok || sessionClaims == nil {
		return nil, false
	}
	return &Claims{
		UserID: sessionClaims.Subject,
		OrgID:  sessionClaims.ActiveOrganizationID,
	}, true
}

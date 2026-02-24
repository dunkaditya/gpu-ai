package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/gpuai/gpuctl/internal/auth"
	"github.com/gpuai/gpuctl/internal/db"
)

// responseCapture wraps http.ResponseWriter to capture the status code and body.
type responseCapture struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
	written    bool
}

func (rc *responseCapture) WriteHeader(code int) {
	rc.statusCode = code
	rc.written = true
	rc.ResponseWriter.WriteHeader(code)
}

func (rc *responseCapture) Write(b []byte) (int, error) {
	rc.body.Write(b)
	if !rc.written {
		rc.statusCode = http.StatusOK
		rc.written = true
	}
	return rc.ResponseWriter.Write(b)
}

// IdempotencyMiddleware returns middleware that prevents duplicate POST requests
// using the Idempotency-Key header. If the header is absent, requests pass through.
//
// Behavior:
//   - Key found + completed (response_code set): replay stored response if request hash matches,
//     return 422 if hash mismatch (key reused with different body).
//   - Key found + in-progress (response_code nil): return 409 (concurrent duplicate).
//   - Key not found: create key, process request, store response.
func IdempotencyMiddleware(dbPool *db.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				// No idempotency key -- pass through.
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()

			// Extract org ID from claims for scoping.
			claims, ok := auth.ClaimsFromContext(ctx)
			if !ok {
				// Auth middleware should have rejected before this point.
				writeProblem(w, http.StatusUnauthorized, "unauthenticated",
					"Valid authentication required")
				return
			}

			// Resolve Clerk org ID to internal UUID for DB operations.
			internalOrgID, err := dbPool.GetOrgIDByClerkOrgID(ctx, claims.OrgID)
			if err != nil {
				slog.Error("idempotency: failed to resolve org ID",
					slog.String("clerk_org_id", claims.OrgID),
					slog.String("error", err.Error()),
				)
				writeProblem(w, http.StatusInternalServerError, "internal-error",
					"Failed to process request")
				return
			}

			// Read and hash request body, then reset for handler.
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				writeProblem(w, http.StatusBadRequest, "invalid-request",
					"Failed to read request body")
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

			hash := sha256.Sum256(bodyBytes)
			requestHash := hex.EncodeToString(hash[:])

			// Check for existing idempotency key.
			existing, err := dbPool.GetIdempotencyKey(ctx, internalOrgID, key)
			if err != nil && !errors.Is(err, db.ErrNotFound) {
				slog.Error("idempotency key lookup failed",
					slog.String("error", err.Error()),
				)
				writeProblem(w, http.StatusInternalServerError, "internal-error",
					"Failed to process request")
				return
			}

			if existing != nil {
				// Key exists -- check state.
				if existing.ResponseCode != nil {
					// Completed: verify request hash matches.
					if existing.RequestHash != requestHash {
						writeProblem(w, http.StatusUnprocessableEntity, "idempotency-mismatch",
							"Idempotency key reused with different request body")
						return
					}
					// Replay stored response.
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(*existing.ResponseCode)
					w.Write(existing.ResponseBody)
					return
				}
				// In-progress: another request is processing.
				writeProblem(w, http.StatusConflict, "idempotency-conflict",
					"Request with this idempotency key is already in progress")
				return
			}

			// Key not found: create it.
			err = dbPool.CreateIdempotencyKey(ctx, internalOrgID, key, requestHash)
			if err != nil {
				if errors.Is(err, db.ErrIdempotencyKeyExists) {
					// Race condition: another request created the key first.
					// Re-check the key state.
					existing, err = dbPool.GetIdempotencyKey(ctx, internalOrgID, key)
					if err != nil {
						slog.Error("idempotency key re-lookup failed",
							slog.String("error", err.Error()),
						)
						writeProblem(w, http.StatusInternalServerError, "internal-error",
							"Failed to process request")
						return
					}
					if existing != nil && existing.ResponseCode != nil {
						if existing.RequestHash != requestHash {
							writeProblem(w, http.StatusUnprocessableEntity, "idempotency-mismatch",
								"Idempotency key reused with different request body")
							return
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(*existing.ResponseCode)
						w.Write(existing.ResponseBody)
						return
					}
					writeProblem(w, http.StatusConflict, "idempotency-conflict",
						"Request with this idempotency key is already in progress")
					return
				}
				slog.Error("idempotency key creation failed",
					slog.String("error", err.Error()),
				)
				writeProblem(w, http.StatusInternalServerError, "internal-error",
					"Failed to process request")
				return
			}

			// Wrap ResponseWriter to capture status code and body.
			capture := &responseCapture{ResponseWriter: w}
			next.ServeHTTP(capture, r)

			// Store the response for future replays.
			if err := dbPool.CompleteIdempotencyKey(ctx, internalOrgID, key,
				capture.statusCode, capture.body.Bytes()); err != nil {
				slog.Error("failed to complete idempotency key",
					slog.String("key", key),
					slog.String("error", err.Error()),
				)
				// Response already sent -- just log the error.
			}
		})
	}
}

package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gpuai/gpuctl/internal/auth"
	"github.com/gpuai/gpuctl/internal/db"
	"golang.org/x/crypto/ssh"
)

// AddSSHKeyRequest is the JSON body for POST /api/v1/ssh-keys.
type AddSSHKeyRequest struct {
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

// SSHKeyResponse is the customer-facing JSON representation of an SSH key.
type SSHKeyResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint"`
	CreatedAt   string `json:"created_at"`
}

// allowedKeyTypes lists the SSH key types we accept.
// DSA (ssh-dss) is explicitly excluded.
var allowedKeyTypes = map[string]bool{
	"ssh-rsa":                  true,
	"ssh-ed25519":              true,
	"ecdsa-sha2-nistp256":      true,
	"ecdsa-sha2-nistp384":      true,
	"ecdsa-sha2-nistp521":      true,
}

// handleCreateSSHKey handles POST /api/v1/ssh-keys.
// Validates the key, computes its fingerprint, and stores it.
func (s *Server) handleCreateSSHKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Extract claims and ensure org+user exist.
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "unauthenticated", "Valid authentication required")
		return
	}

	orgID, userID, err := s.db.EnsureOrgAndUser(ctx, claims.OrgID, claims.UserID, "")
	if err != nil {
		slog.Error("failed to ensure org and user",
			slog.String("clerk_org_id", claims.OrgID),
			slog.String("error", err.Error()),
		)
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to process request")
		return
	}

	// 2. Decode and validate request body.
	var req AddSSHKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid-request", "Invalid JSON request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.PublicKey = strings.TrimSpace(req.PublicKey)

	if req.Name == "" {
		writeProblem(w, http.StatusBadRequest, "validation-error", "name is required")
		return
	}
	if len(req.Name) > 100 {
		writeProblem(w, http.StatusBadRequest, "validation-error", "name must be 100 characters or fewer")
		return
	}
	if req.PublicKey == "" {
		writeProblem(w, http.StatusBadRequest, "validation-error", "public_key is required")
		return
	}

	// 3. Parse and validate the SSH public key.
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(req.PublicKey))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid_ssh_key", "Could not parse SSH public key")
		return
	}

	// 4. Check key type is allowed (reject DSA and others).
	keyType := pubKey.Type()
	if !allowedKeyTypes[keyType] {
		writeProblem(w, http.StatusBadRequest, "unsupported_key_type",
			"Unsupported key type: "+keyType+". Accepted types: RSA, Ed25519, ECDSA")
		return
	}

	// 5. Compute SHA256 fingerprint.
	fingerprint := ssh.FingerprintSHA256(pubKey)

	// 6. Check org key count limit (max 50).
	count, err := s.db.CountSSHKeysByOrg(ctx, orgID)
	if err != nil {
		slog.Error("failed to count SSH keys",
			slog.String("org_id", orgID),
			slog.String("error", err.Error()),
		)
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to process request")
		return
	}
	if count >= 50 {
		writeProblem(w, http.StatusUnprocessableEntity, "key_limit_exceeded",
			"Maximum 50 SSH keys per organization")
		return
	}

	// 7. Create the SSH key record.
	key := db.SSHKey{
		UserID:      userID,
		OrgID:       orgID,
		Name:        req.Name,
		PublicKey:   req.PublicKey,
		Fingerprint: fingerprint,
	}
	if err := s.db.CreateSSHKey(ctx, &key); err != nil {
		if errors.Is(err, db.ErrDuplicateKey) {
			writeProblem(w, http.StatusConflict, "duplicate_key",
				"An SSH key with this fingerprint already exists in your organization")
			return
		}
		slog.Error("failed to create SSH key",
			slog.String("org_id", orgID),
			slog.String("error", err.Error()),
		)
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to create SSH key")
		return
	}

	// 8. Return 201 with the created key.
	writeJSON(w, http.StatusCreated, SSHKeyResponse{
		ID:          key.SSHKeyID,
		Name:        key.Name,
		Fingerprint: key.Fingerprint,
		CreatedAt:   key.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// handleListSSHKeys handles GET /api/v1/ssh-keys.
// Returns all SSH keys for the authenticated organization.
func (s *Server) handleListSSHKeys(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "unauthenticated", "Valid authentication required")
		return
	}

	orgID, err := s.db.GetOrgIDByClerkOrgID(ctx, claims.OrgID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			// Org not provisioned yet -- return empty list.
			writeJSON(w, http.StatusOK, map[string][]SSHKeyResponse{
				"ssh_keys": {},
			})
			return
		}
		slog.Error("failed to look up org", slog.String("error", err.Error()))
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to process request")
		return
	}

	keys, err := s.db.ListSSHKeysByOrg(ctx, orgID)
	if err != nil {
		slog.Error("failed to list SSH keys",
			slog.String("org_id", orgID),
			slog.String("error", err.Error()),
		)
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to list SSH keys")
		return
	}

	data := make([]SSHKeyResponse, 0, len(keys))
	for _, k := range keys {
		data = append(data, SSHKeyResponse{
			ID:          k.SSHKeyID,
			Name:        k.Name,
			Fingerprint: k.Fingerprint,
			CreatedAt:   k.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	writeJSON(w, http.StatusOK, map[string][]SSHKeyResponse{
		"ssh_keys": data,
	})
}

// handleDeleteSSHKey handles DELETE /api/v1/ssh-keys/{id}.
// Removes an SSH key scoped to the authenticated organization.
func (s *Server) handleDeleteSSHKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	keyID := r.PathValue("id")
	if keyID == "" {
		writeProblem(w, http.StatusBadRequest, "missing-id", "SSH key ID is required")
		return
	}

	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "unauthenticated", "Valid authentication required")
		return
	}

	orgID, err := s.db.GetOrgIDByClerkOrgID(ctx, claims.OrgID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeProblem(w, http.StatusNotFound, "not-found", "SSH key not found")
			return
		}
		slog.Error("failed to look up org", slog.String("error", err.Error()))
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to process request")
		return
	}

	deleted, err := s.db.DeleteSSHKey(ctx, keyID, orgID)
	if err != nil {
		slog.Error("failed to delete SSH key",
			slog.String("key_id", keyID),
			slog.String("error", err.Error()),
		)
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to delete SSH key")
		return
	}

	if !deleted {
		writeProblem(w, http.StatusNotFound, "not-found", "SSH key not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

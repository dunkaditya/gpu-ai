package api

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultPageSize is the default number of items per page.
	DefaultPageSize = 20

	// MaxPageSize is the maximum allowed number of items per page.
	MaxPageSize = 100
)

// PageParams holds parsed pagination parameters from a request.
type PageParams struct {
	Cursor string
	Limit  int
}

// PageResult is a generic paginated response envelope.
type PageResult[T any] struct {
	Data    []T    `json:"data"`
	Cursor  string `json:"cursor,omitempty"`
	HasMore bool   `json:"has_more"`
}

// ParsePageParams extracts cursor and limit from query parameters.
// Limit is clamped to [1, MaxPageSize] and defaults to DefaultPageSize.
func ParsePageParams(r *http.Request) PageParams {
	cursor := r.URL.Query().Get("cursor")

	limit := DefaultPageSize
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	// Clamp limit to valid range.
	if limit < 1 {
		limit = 1
	}
	if limit > MaxPageSize {
		limit = MaxPageSize
	}

	return PageParams{
		Cursor: cursor,
		Limit:  limit,
	}
}

// EncodeCursor encodes a (createdAt, id) tuple into a URL-safe base64 cursor string.
func EncodeCursor(createdAt time.Time, id string) string {
	raw := createdAt.Format(time.RFC3339) + "|" + id
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

// DecodeCursor decodes a cursor string into its (createdAt, id) components.
// Returns an error on any parse failure -- never panics or passes raw values to SQL.
func DecodeCursor(cursor string) (createdAt time.Time, id string, err error) {
	decoded, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("invalid cursor encoding: %w", err)
	}

	parts := strings.SplitN(string(decoded), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, "", fmt.Errorf("invalid cursor format: expected 'time|id'")
	}

	createdAt, err = time.Parse(time.RFC3339, parts[0])
	if err != nil {
		return time.Time{}, "", fmt.Errorf("invalid cursor timestamp: %w", err)
	}

	return createdAt, parts[1], nil
}

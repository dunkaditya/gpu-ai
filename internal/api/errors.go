package api

import (
	"encoding/json"
	"net/http"
)

// ProblemDetail represents an RFC 7807 Problem Details response.
type ProblemDetail struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
}

// writeProblem writes an RFC 7807 Problem Details JSON response.
// errType is appended to the base URI "https://api.gpu.ai/errors/".
func writeProblem(w http.ResponseWriter, status int, errType, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ProblemDetail{
		Type:   "https://api.gpu.ai/errors/" + errType,
		Title:  http.StatusText(status),
		Status: status,
		Detail: detail,
	})
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

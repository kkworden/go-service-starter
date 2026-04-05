package handler

import (
	"encoding/json"
	"net/http"
)

// errorResponse is the standard JSON error format returned by all endpoints.
// Machine-readable Code field allows clients to programmatically handle errors.
type errorResponse struct {
	Error string `json:"error"` // Human-readable error message.
	Code  string `json:"code"`  // Machine-readable error code (e.g., "NOT_FOUND", "INTERNAL").
}

// writeJSON serializes v as JSON and writes it to the response with the given
// HTTP status code and Content-Type: application/json header.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v) // Error is non-actionable after headers are sent.
}

// writeError writes a structured JSON error response using the standard
// errorResponse format. This is the single place all error responses flow through.
func writeError(w http.ResponseWriter, status int, msg, code string) {
	writeJSON(w, status, errorResponse{Error: msg, Code: code})
}

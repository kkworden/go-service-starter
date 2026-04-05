package domain

import "errors"

// Sentinel errors shared across all domain types.
// ErrConflict and ErrBadRequest are scaffolding — not yet used by the Item
// example, but provided for services that add validation or uniqueness checks.
var (
	ErrNotFound   = errors.New("not found")   // Requested resource does not exist.
	ErrConflict   = errors.New("conflict")    // Duplicate or conflicting resource.
	ErrBadRequest = errors.New("bad request") // Validation failure on input.
)

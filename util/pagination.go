// Package util provides small, stateless helper functions shared across
// multiple packages. Keep this package thin — only add functions that are
// genuinely reused in two or more places.
package util

// Pagination constants shared across all services.
// Scaffolding: not used by the Item example (which has no list endpoint),
// but provided for services that add paginated list queries.
const (
	DefaultPageLimit = 100 // Default number of items per page when not specified.
	MaxPageLimit     = 300 // Absolute maximum items per page; any request above this is clamped.
)

// ClampPagination normalizes limit and offset to safe bounds.
// Zero or negative limit defaults to DefaultPageLimit; values above
// MaxPageLimit are capped. Negative offsets are floored at 0.
func ClampPagination(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = DefaultPageLimit
	}
	if limit > MaxPageLimit {
		limit = MaxPageLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

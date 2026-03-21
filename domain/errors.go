package domain

import "errors"

// ErrNotFound indicates that a requested resource does not exist.
var ErrNotFound = errors.New("not found")

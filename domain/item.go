// Package domain holds plain data structs shared across all layers.
// It contains no interfaces, methods, or business logic.
package domain

// Item represents a stored data record.
type Item struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}

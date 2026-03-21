// Package store implements the persistence layer.
// Each store satisfies an interface defined by the service layer.
package store

import (
	"context"
	"fmt"
	"sync"

	"go-service-starter/domain"
)

// ItemStore is a concurrency-safe, in-memory store for Items.
// Swap this for a database-backed implementation for production use.
type ItemStore struct {
	mu    sync.RWMutex
	items map[string]domain.Item
}

// NewItemStore returns an initialised in-memory ItemStore.
func NewItemStore() *ItemStore {
	return &ItemStore{
		items: make(map[string]domain.Item),
	}
}

// Save persists an Item, overwriting any existing entry with the same ID.
func (s *ItemStore) Save(ctx context.Context, item domain.Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items[item.ID] = item
	return nil
}

// FindByID retrieves an Item by its ID. Returns an error if the ID does not exist.
func (s *ItemStore) FindByID(ctx context.Context, id string) (domain.Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[id]
	if !ok {
		return domain.Item{}, fmt.Errorf("item %s: %w", id, domain.ErrNotFound)
	}

	return item, nil
}

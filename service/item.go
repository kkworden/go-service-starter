// Package service contains business logic.
// It defines the interfaces it needs from the store layer and is
// consumed by the handler layer via handler-defined interfaces.
package service

import (
	"context"
	"fmt"

	"go-service-starter/domain"

	"github.com/google/uuid"
)

// ItemStore defines the persistence operations required by ItemService.
// Implementations live in the store package.
type ItemStore interface {
	Save(ctx context.Context, item domain.Item) error
	FindByID(ctx context.Context, id string) (domain.Item, error)
}

// ItemService implements business logic for managing Items.
type ItemService struct {
	store ItemStore
}

// NewItemService constructs an ItemService with the given store dependency.
func NewItemService(store ItemStore) *ItemService {
	return &ItemService{store: store}
}

// Create generates a new Item with a unique ID and persists it via the store.
func (s *ItemService) Create(ctx context.Context, data string) (domain.Item, error) {
	item := domain.Item{
		ID:   uuid.New().String(),
		Data: data,
	}

	err := s.store.Save(ctx, item)
	if err != nil {
		return domain.Item{}, fmt.Errorf("ItemService.Create: %w", err)
	}

	return item, nil
}

// Get retrieves an Item by ID from the store.
func (s *ItemService) Get(ctx context.Context, id string) (domain.Item, error) {
	item, err := s.store.FindByID(ctx, id)
	if err != nil {
		return domain.Item{}, fmt.Errorf("ItemService.Get: %w", err)
	}

	return item, nil
}

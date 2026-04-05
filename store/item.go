package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"go-service-starter/domain"
)

// ItemStore is a Postgres-backed store for Items.
type ItemStore struct {
	db *DB
}

// NewItemStore returns an ItemStore backed by the given database connections.
func NewItemStore(db *DB) *ItemStore {
	return &ItemStore{db: db}
}

// Save persists an Item, overwriting any existing entry with the same ID.
func (s *ItemStore) Save(ctx context.Context, item domain.Item) error {
	_, err := s.db.Writer.ExecContext(ctx,
		`INSERT INTO items (id, data) VALUES ($1, $2)
		 ON CONFLICT (id) DO UPDATE SET data = EXCLUDED.data`,
		item.ID, item.Data,
	)
	if err != nil {
		return fmt.Errorf("ItemStore.Save: %w", err)
	}
	return nil
}

// FindByID retrieves an Item by its ID. Returns domain.ErrNotFound if the ID
// does not exist.
func (s *ItemStore) FindByID(ctx context.Context, id string) (domain.Item, error) {
	var item domain.Item
	err := s.db.Reader.QueryRowContext(ctx,
		`SELECT id, data FROM items WHERE id = $1`, id,
	).Scan(&item.ID, &item.Data)

	if errors.Is(err, sql.ErrNoRows) {
		return domain.Item{}, fmt.Errorf("item %s: %w", id, domain.ErrNotFound)
	}
	if err != nil {
		return domain.Item{}, fmt.Errorf("ItemStore.FindByID: %w", err)
	}
	return item, nil
}

package store_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"

	"go-service-starter/domain"
	"go-service-starter/store"
)

// testDB opens a connection to the test database and returns a store.DB.
// Tests are skipped if TEST_DATABASE_URL is not set.
func testDB(t *testing.T) *store.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping integration test")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// Ensure the items table exists for tests.
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS items (
		id   TEXT PRIMARY KEY,
		data TEXT NOT NULL DEFAULT ''
	)`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	return &store.DB{Writer: db, Reader: db}
}

// cleanupItem removes a specific item after a test.
func cleanupItem(t *testing.T, db *store.DB, id string) {
	t.Helper()
	t.Cleanup(func() {
		db.Writer.Exec(`DELETE FROM items WHERE id = $1`, id)
	})
}

// ---------------------------------------------------------------------------
// TestItemStore_Save
// ---------------------------------------------------------------------------

func TestItemStore_Save(t *testing.T) {
	db := testDB(t)
	s := store.NewItemStore(db)

	t.Run("saves an item without error", func(t *testing.T) {
		item := domain.Item{ID: "test-save-1", Data: "hello"}
		cleanupItem(t, db, item.ID)

		err := s.Save(context.Background(), item)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("overwrites existing item with the same ID", func(t *testing.T) {
		first := domain.Item{ID: "test-save-overwrite", Data: "original"}
		second := domain.Item{ID: "test-save-overwrite", Data: "updated"}
		cleanupItem(t, db, first.ID)

		_ = s.Save(context.Background(), first)
		_ = s.Save(context.Background(), second)

		got, err := s.FindByID(context.Background(), "test-save-overwrite")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got.Data != "updated" {
			t.Errorf("expected overwritten Data %q, got %q", "updated", got.Data)
		}
	})

	t.Run("saves item with empty Data field", func(t *testing.T) {
		item := domain.Item{ID: "test-save-empty", Data: ""}
		cleanupItem(t, db, item.ID)

		err := s.Save(context.Background(), item)
		if err != nil {
			t.Fatalf("expected no error for item with empty Data, got %v", err)
		}

		got, _ := s.FindByID(context.Background(), "test-save-empty")
		if got.Data != "" {
			t.Errorf("expected empty Data, got %q", got.Data)
		}
	})
}

// ---------------------------------------------------------------------------
// TestItemStore_FindByID
// ---------------------------------------------------------------------------

func TestItemStore_FindByID(t *testing.T) {
	db := testDB(t)
	s := store.NewItemStore(db)

	t.Run("returns error for ID that was never saved", func(t *testing.T) {
		_, err := s.FindByID(context.Background(), "nonexistent")
		if err == nil {
			t.Fatal("expected an error for unknown ID, got nil")
		}
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected error chain to contain domain.ErrNotFound, got: %v", err)
		}
		if !strings.Contains(err.Error(), "nonexistent") {
			t.Errorf("expected error to mention the missing ID, got: %v", err)
		}
	})

	t.Run("returns the exact item that was saved", func(t *testing.T) {
		want := domain.Item{ID: "test-find-me", Data: "payload"}
		cleanupItem(t, db, want.ID)
		_ = s.Save(context.Background(), want)

		got, err := s.FindByID(context.Background(), "test-find-me")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got != want {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("does not return stale item after overwrite", func(t *testing.T) {
		cleanupItem(t, db, "test-overwrite-k")
		_ = s.Save(context.Background(), domain.Item{ID: "test-overwrite-k", Data: "old"})
		_ = s.Save(context.Background(), domain.Item{ID: "test-overwrite-k", Data: "new"})

		got, err := s.FindByID(context.Background(), "test-overwrite-k")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Data != "new" {
			t.Errorf("expected %q, got %q", "new", got.Data)
		}
	})
}

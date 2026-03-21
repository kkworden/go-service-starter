package store_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"go-service-starter/domain"
	"go-service-starter/store"
)

// ---------------------------------------------------------------------------
// TestItemStore_Save
// ---------------------------------------------------------------------------

func TestItemStore_Save(t *testing.T) {
	t.Run("saves an item without error", func(t *testing.T) {
		s := store.NewItemStore()
		item := domain.Item{ID: "1", Data: "hello"}

		err := s.Save(context.Background(), item)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("overwrites existing item with the same ID", func(t *testing.T) {
		s := store.NewItemStore()
		first := domain.Item{ID: "same-id", Data: "original"}
		second := domain.Item{ID: "same-id", Data: "updated"}

		_ = s.Save(context.Background(), first)
		_ = s.Save(context.Background(), second)

		got, err := s.FindByID(context.Background(), "same-id")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got.Data != "updated" {
			t.Errorf("expected overwritten Data %q, got %q", "updated", got.Data)
		}
	})

	t.Run("saves item with empty Data field", func(t *testing.T) {
		s := store.NewItemStore()
		item := domain.Item{ID: "empty-data", Data: ""}

		err := s.Save(context.Background(), item)

		if err != nil {
			t.Fatalf("expected no error for item with empty Data, got %v", err)
		}

		got, _ := s.FindByID(context.Background(), "empty-data")
		if got.Data != "" {
			t.Errorf("expected empty Data, got %q", got.Data)
		}
	})

	t.Run("saves multiple distinct items independently", func(t *testing.T) {
		s := store.NewItemStore()
		items := []domain.Item{
			{ID: "a", Data: "first"},
			{ID: "b", Data: "second"},
			{ID: "c", Data: "third"},
		}

		for _, item := range items {
			if err := s.Save(context.Background(), item); err != nil {
				t.Fatalf("Save(%q) unexpected error: %v", item.ID, err)
			}
		}

		for _, want := range items {
			got, err := s.FindByID(context.Background(), want.ID)
			if err != nil {
				t.Fatalf("FindByID(%q) unexpected error: %v", want.ID, err)
			}
			if got != want {
				t.Errorf("FindByID(%q): got %+v, want %+v", want.ID, got, want)
			}
		}
	})

	t.Run("is safe for concurrent saves", func(t *testing.T) {
		s := store.NewItemStore()
		const goroutines = 50
		var wg sync.WaitGroup

		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				item := domain.Item{
					ID:   strings.Repeat("x", n+1), // unique IDs of different lengths
					Data: "concurrent",
				}
				_ = s.Save(context.Background(), item)
			}(i)
		}

		wg.Wait()
		// If the race detector is enabled this will catch any data races.
	})
}

// ---------------------------------------------------------------------------
// TestItemStore_FindByID
// ---------------------------------------------------------------------------

func TestItemStore_FindByID(t *testing.T) {
	t.Run("returns error for ID that was never saved", func(t *testing.T) {
		s := store.NewItemStore()

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
		s := store.NewItemStore()
		want := domain.Item{ID: "find-me", Data: "payload"}
		_ = s.Save(context.Background(), want)

		got, err := s.FindByID(context.Background(), "find-me")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got != want {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("returns error for empty string ID", func(t *testing.T) {
		s := store.NewItemStore()

		_, err := s.FindByID(context.Background(), "")

		if err == nil {
			t.Fatal("expected an error for empty ID, got nil")
		}
	})

	t.Run("empty string ID key is independent of other keys", func(t *testing.T) {
		s := store.NewItemStore()
		// Save an item with empty-string ID explicitly.
		emptyKeyItem := domain.Item{ID: "", Data: "stored with empty key"}
		_ = s.Save(context.Background(), emptyKeyItem)

		got, err := s.FindByID(context.Background(), "")
		if err != nil {
			t.Fatalf("expected to find item saved under empty key, got error: %v", err)
		}
		if got.Data != "stored with empty key" {
			t.Errorf("unexpected Data %q", got.Data)
		}
	})

	t.Run("does not return stale item after overwrite", func(t *testing.T) {
		s := store.NewItemStore()
		_ = s.Save(context.Background(), domain.Item{ID: "k", Data: "old"})
		_ = s.Save(context.Background(), domain.Item{ID: "k", Data: "new"})

		got, err := s.FindByID(context.Background(), "k")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Data != "new" {
			t.Errorf("expected %q, got %q", "new", got.Data)
		}
	})

	t.Run("is safe for concurrent reads and writes", func(t *testing.T) {
		s := store.NewItemStore()
		_ = s.Save(context.Background(), domain.Item{ID: "concurrent-read", Data: "value"})

		const goroutines = 50
		var wg sync.WaitGroup

		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = s.FindByID(context.Background(), "concurrent-read")
			}()
		}

		wg.Wait()
	})
}

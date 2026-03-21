package service_test

import (
	"context"
	"errors"
	"testing"

	"go-service-starter/domain"
	"go-service-starter/service"
)

// mockItemStore is an inline mock that satisfies service.ItemStore.
type mockItemStore struct {
	saved    []domain.Item
	saveErr  error
	findItem domain.Item
	findErr  error
}

func (m *mockItemStore) Save(_ context.Context, item domain.Item) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saved = append(m.saved, item)
	return nil
}

func (m *mockItemStore) FindByID(_ context.Context, id string) (domain.Item, error) {
	if m.findErr != nil {
		return domain.Item{}, m.findErr
	}
	return m.findItem, nil
}

// ---------------------------------------------------------------------------
// TestItemService_Create
// ---------------------------------------------------------------------------

func TestItemService_Create(t *testing.T) {
	t.Run("saves item and returns it with a non-empty ID", func(t *testing.T) {
		mock := &mockItemStore{}
		svc := service.NewItemService(mock)

		item, err := svc.Create(context.Background(), "hello")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if item.ID == "" {
			t.Error("expected a non-empty ID on the returned item")
		}
		if item.Data != "hello" {
			t.Errorf("expected Data %q, got %q", "hello", item.Data)
		}
		if len(mock.saved) != 1 {
			t.Fatalf("expected 1 saved item, got %d", len(mock.saved))
		}
		if mock.saved[0].ID != item.ID {
			t.Errorf("saved item ID %q does not match returned ID %q", mock.saved[0].ID, item.ID)
		}
		if mock.saved[0].Data != "hello" {
			t.Errorf("saved item Data %q does not match input %q", mock.saved[0].Data, "hello")
		}
	})

	t.Run("each call produces a unique ID", func(t *testing.T) {
		mock := &mockItemStore{}
		svc := service.NewItemService(mock)

		first, _ := svc.Create(context.Background(), "a")
		second, _ := svc.Create(context.Background(), "b")

		if first.ID == second.ID {
			t.Errorf("expected unique IDs but both were %q", first.ID)
		}
	})

	t.Run("empty data string is accepted", func(t *testing.T) {
		mock := &mockItemStore{}
		svc := service.NewItemService(mock)

		item, err := svc.Create(context.Background(), "")

		if err != nil {
			t.Fatalf("expected no error for empty data, got %v", err)
		}
		if item.Data != "" {
			t.Errorf("expected empty Data, got %q", item.Data)
		}
	})

	t.Run("wraps store error with context", func(t *testing.T) {
		storeErr := errors.New("disk full")
		mock := &mockItemStore{saveErr: storeErr}
		svc := service.NewItemService(mock)

		_, err := svc.Create(context.Background(), "data")

		if err == nil {
			t.Fatal("expected an error when store fails, got nil")
		}
		if !errors.Is(err, storeErr) {
			t.Errorf("expected error chain to contain %v, got %v", storeErr, err)
		}
	})

	t.Run("does not save item when store returns an error", func(t *testing.T) {
		mock := &mockItemStore{saveErr: errors.New("transient failure")}
		svc := service.NewItemService(mock)

		svc.Create(context.Background(), "data") //nolint:errcheck

		if len(mock.saved) != 0 {
			t.Errorf("expected no saved items on store failure, got %d", len(mock.saved))
		}
	})
}

// ---------------------------------------------------------------------------
// TestItemService_Get
// ---------------------------------------------------------------------------

func TestItemService_Get(t *testing.T) {
	tests := []struct {
		name      string
		storeItem domain.Item
		storeErr  error
		queryID   string
		wantItem  domain.Item
		wantErr   bool
	}{
		{
			name:      "returns item when found",
			storeItem: domain.Item{ID: "abc-123", Data: "some data"},
			queryID:   "abc-123",
			wantItem:  domain.Item{ID: "abc-123", Data: "some data"},
		},
		{
			name:      "returns item with empty data field",
			storeItem: domain.Item{ID: "xyz", Data: ""},
			queryID:   "xyz",
			wantItem:  domain.Item{ID: "xyz", Data: ""},
		},
		{
			name:     "wraps store error when item not found",
			storeErr: errors.New("item not found: missing-id"),
			queryID:  "missing-id",
			wantErr:  true,
		},
		{
			name:     "wraps arbitrary store error",
			storeErr: errors.New("database unavailable"),
			queryID:  "any-id",
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockItemStore{
				findItem: tc.storeItem,
				findErr:  tc.storeErr,
			}
			svc := service.NewItemService(mock)

			item, err := svc.Get(context.Background(), tc.queryID)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected an error but got nil")
				}
				if !errors.Is(err, tc.storeErr) {
					t.Errorf("expected error chain to contain %v, got %v", tc.storeErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if item.ID != tc.wantItem.ID {
				t.Errorf("ID: got %q, want %q", item.ID, tc.wantItem.ID)
			}
			if item.Data != tc.wantItem.Data {
				t.Errorf("Data: got %q, want %q", item.Data, tc.wantItem.Data)
			}
		})
	}
}

func TestItemService_Get_errorWrapping(t *testing.T) {
	sentinel := errors.New("sentinel store error")
	mock := &mockItemStore{findErr: sentinel}
	svc := service.NewItemService(mock)

	_, err := svc.Get(context.Background(), "any")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("expected errors.Is to find sentinel in chain, got: %v", err)
	}
}
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-service-starter/domain"
	"go-service-starter/handler"

	"github.com/go-chi/chi/v5"
)

// mockItemService is an inline mock that satisfies handler.ItemService.
type mockItemService struct {
	createItem domain.Item
	createErr  error
	getItem    domain.Item
	getErr     error
	// Records the arguments received so tests can assert on them.
	lastCreateData string
	lastGetID      string
}

func (m *mockItemService) Create(_ context.Context, data string) (domain.Item, error) {
	m.lastCreateData = data
	if m.createErr != nil {
		return domain.Item{}, m.createErr
	}
	return m.createItem, nil
}

func (m *mockItemService) Get(_ context.Context, id string) (domain.Item, error) {
	m.lastGetID = id
	if m.getErr != nil {
		return domain.Item{}, m.getErr
	}
	return m.getItem, nil
}

// withChiParam injects a chi URL parameter into the request context so that
// chi.URLParam works without a running router.
func withChiParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// ---------------------------------------------------------------------------
// TestItemHandler_Create
// ---------------------------------------------------------------------------

func TestItemHandler_Create(t *testing.T) {
	t.Run("returns 201 and JSON body on success", func(t *testing.T) {
		mock := &mockItemService{
			createItem: domain.Item{ID: "new-id", Data: "hello"},
		}
		h := handler.NewItemHandler(mock)

		body, _ := json.Marshal(domain.Item{Data: "hello"})
		req := httptest.NewRequest(http.MethodPost, "/items", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("status: got %d, want %d", rec.Code, http.StatusCreated)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type: got %q, want %q", ct, "application/json")
		}

		var got domain.Item
		if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
			t.Fatalf("could not decode response body: %v", err)
		}
		if got.ID != "new-id" {
			t.Errorf("ID: got %q, want %q", got.ID, "new-id")
		}
		if got.Data != "hello" {
			t.Errorf("Data: got %q, want %q", got.Data, "hello")
		}
	})

	t.Run("passes the Data field to the service", func(t *testing.T) {
		mock := &mockItemService{
			createItem: domain.Item{ID: "x", Data: "my-data"},
		}
		h := handler.NewItemHandler(mock)

		body, _ := json.Marshal(domain.Item{Data: "my-data"})
		req := httptest.NewRequest(http.MethodPost, "/items", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if mock.lastCreateData != "my-data" {
			t.Errorf("service received Data %q, want %q", mock.lastCreateData, "my-data")
		}
	})

	t.Run("returns 400 for malformed JSON body", func(t *testing.T) {
		mock := &mockItemService{}
		h := handler.NewItemHandler(mock)

		req := httptest.NewRequest(http.MethodPost, "/items", bytes.NewReader([]byte("not-json{")))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("returns 400 for empty body", func(t *testing.T) {
		mock := &mockItemService{}
		h := handler.NewItemHandler(mock)

		req := httptest.NewRequest(http.MethodPost, "/items", bytes.NewReader([]byte("")))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("returns 500 when service returns an error", func(t *testing.T) {
		mock := &mockItemService{
			createErr: errors.New("internal failure"),
		}
		h := handler.NewItemHandler(mock)

		body, _ := json.Marshal(domain.Item{Data: "d"})
		req := httptest.NewRequest(http.MethodPost, "/items", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status: got %d, want %d", rec.Code, http.StatusInternalServerError)
		}
	})

	t.Run("returns 201 for item with empty Data field", func(t *testing.T) {
		mock := &mockItemService{
			createItem: domain.Item{ID: "empty-data-id", Data: ""},
		}
		h := handler.NewItemHandler(mock)

		body, _ := json.Marshal(domain.Item{Data: ""})
		req := httptest.NewRequest(http.MethodPost, "/items", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("status: got %d, want %d", rec.Code, http.StatusCreated)
		}
	})
}

// ---------------------------------------------------------------------------
// TestItemHandler_Get
// ---------------------------------------------------------------------------

func TestItemHandler_Get(t *testing.T) {
	t.Run("returns 200 and JSON body on success", func(t *testing.T) {
		mock := &mockItemService{
			getItem: domain.Item{ID: "abc-123", Data: "payload"},
		}
		h := handler.NewItemHandler(mock)

		req := httptest.NewRequest(http.MethodGet, "/items/abc-123", nil)
		req = withChiParam(req, "id", "abc-123")
		rec := httptest.NewRecorder()

		h.Get(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status: got %d, want %d", rec.Code, http.StatusOK)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type: got %q, want %q", ct, "application/json")
		}

		var got domain.Item
		if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
			t.Fatalf("could not decode response body: %v", err)
		}
		if got.ID != "abc-123" {
			t.Errorf("ID: got %q, want %q", got.ID, "abc-123")
		}
		if got.Data != "payload" {
			t.Errorf("Data: got %q, want %q", got.Data, "payload")
		}
	})

	t.Run("passes the URL parameter ID to the service", func(t *testing.T) {
		mock := &mockItemService{
			getItem: domain.Item{ID: "target-id", Data: "d"},
		}
		h := handler.NewItemHandler(mock)

		req := httptest.NewRequest(http.MethodGet, "/items/target-id", nil)
		req = withChiParam(req, "id", "target-id")
		rec := httptest.NewRecorder()

		h.Get(rec, req)

		if mock.lastGetID != "target-id" {
			t.Errorf("service received ID %q, want %q", mock.lastGetID, "target-id")
		}
	})

	t.Run("returns 404 when service returns ErrNotFound", func(t *testing.T) {
		mock := &mockItemService{
			getErr: fmt.Errorf("item missing-id: %w", domain.ErrNotFound),
		}
		h := handler.NewItemHandler(mock)

		req := httptest.NewRequest(http.MethodGet, "/items/missing", nil)
		req = withChiParam(req, "id", "missing")
		rec := httptest.NewRecorder()

		h.Get(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status: got %d, want %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("returns 500 for non-not-found service errors", func(t *testing.T) {
		mock := &mockItemService{
			getErr: errors.New("database connection lost"),
		}
		h := handler.NewItemHandler(mock)

		req := httptest.NewRequest(http.MethodGet, "/items/abc", nil)
		req = withChiParam(req, "id", "abc")
		rec := httptest.NewRecorder()

		h.Get(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status: got %d, want %d", rec.Code, http.StatusInternalServerError)
		}
	})

	t.Run("returns 404 for empty ID parameter", func(t *testing.T) {
		mock := &mockItemService{
			getErr: fmt.Errorf("item : %w", domain.ErrNotFound),
		}
		h := handler.NewItemHandler(mock)

		req := httptest.NewRequest(http.MethodGet, "/items/", nil)
		req = withChiParam(req, "id", "")
		rec := httptest.NewRecorder()

		h.Get(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status: got %d, want %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("does not leak internal error details in response body", func(t *testing.T) {
		mock := &mockItemService{
			getErr: fmt.Errorf("item bad-id: %w", domain.ErrNotFound),
		}
		h := handler.NewItemHandler(mock)

		req := httptest.NewRequest(http.MethodGet, "/items/bad-id", nil)
		req = withChiParam(req, "id", "bad-id")
		rec := httptest.NewRecorder()

		h.Get(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status: got %d, want %d", rec.Code, http.StatusNotFound)
		}
		body := rec.Body.String()
		if strings.Contains(body, "ItemService") {
			t.Error("response body should not contain internal type names")
		}
	})
}

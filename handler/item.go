// Package handler is the HTTP layer. It decodes requests, delegates to the
// service layer, and encodes responses. Handlers must not contain business logic.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"go-service-starter/domain"

	"github.com/go-chi/chi/v5"
)

// ItemService defines the operations the handler needs from the service layer.
// The implementation lives in the service package.
type ItemService interface {
	Create(ctx context.Context, data string) (domain.Item, error)
	Get(ctx context.Context, id string) (domain.Item, error)
}

// ItemHandler exposes HTTP endpoints for Item operations.
type ItemHandler struct {
	service ItemService
}

// NewItemHandler constructs an ItemHandler with the given service dependency.
func NewItemHandler(service ItemService) *ItemHandler {
	return &ItemHandler{service: service}
}

// Create handles POST /items. It decodes the request body, delegates creation
// to the service layer, and responds with the created Item as JSON.
func (h *ItemHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit

	var req domain.Item
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	item, err := h.service.Create(r.Context(), req.Data)
	if err != nil {
		log.Printf("Create failed: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(item) // error is non-actionable after headers are sent
}

// Get handles GET /items/{id}. It extracts the ID from the URL, retrieves
// the Item via the service layer, and responds with it as JSON.
func (h *ItemHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	item, err := h.service.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "item not found", http.StatusNotFound)
			return
		}
		log.Printf("Get failed: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(item) // error is non-actionable after headers are sent
}

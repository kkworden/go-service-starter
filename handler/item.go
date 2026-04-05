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
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}

	item, err := h.service.Create(r.Context(), req.Data)
	if err != nil {
		log.Printf("POST /items Create failed: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error", "INTERNAL")
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

// Get handles GET /items/{id}. It extracts the ID from the URL, retrieves
// the Item via the service layer, and responds with it as JSON.
func (h *ItemHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	item, err := h.service.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "item not found", "NOT_FOUND")
			return
		}
		log.Printf("GET /items/%s Get failed: %v", id, err)
		writeError(w, http.StatusInternalServerError, "internal server error", "INTERNAL")
		return
	}

	writeJSON(w, http.StatusOK, item)
}

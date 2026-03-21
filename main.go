// Package main wires dependencies together and starts the HTTP server.
// No business logic, DB calls, or conditionals belong here — just wiring and startup.
package main

import (
	"log"
	"net/http"
	"os"

	"go-service-starter/handler"
	"go-service-starter/service"
	"go-service-starter/store"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Compile-time interface checks — verifies the dependency graph is wired correctly.
var (
	_ handler.ItemService = (*service.ItemService)(nil)
	_ service.ItemStore   = (*store.ItemStore)(nil)
)

func main() {
	// Build the dependency graph: store -> service -> handler.
	itemStore := store.NewItemStore()
	itemService := service.NewItemService(itemStore)
	itemHandler := handler.NewItemHandler(itemService)

	// Configure the Chi router with standard middleware.
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Register routes.
	r.Post("/items", itemHandler.Create)
	r.Get("/items/{id}", itemHandler.Get)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

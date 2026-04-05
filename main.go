// Package main wires dependencies together and starts the HTTP server.
// No business logic, DB calls, or conditionals belong here — just wiring
// and startup. The 3-layer architecture (handler -> service -> store) is
// assembled via constructor injection with compile-time interface checks.
package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-service-starter/config"
	"go-service-starter/handler"
	"go-service-starter/middleware"
	"go-service-starter/service"
	"go-service-starter/store"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Register Postgres driver for migrate.
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// migrationsFS embeds all SQL migration files into the binary so they can be
// applied at startup without requiring the migration files on disk at runtime.
//
//go:embed migrations/*.sql
var migrationsFS embed.FS

// Compile-time interface checks — these ensure that each concrete type
// satisfies the interface expected by the layer above. If any method signature
// drifts, the build fails here with a clear error rather than at runtime.
var (
	_ handler.ItemService = (*service.ItemService)(nil)
	_ service.ItemStore   = (*store.ItemStore)(nil)
)

func main() {
	// Load configuration from environment variables.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// Open writer and reader connection pools. In local dev both point to
	// the same Postgres; in production, reader points to a read replica.
	writerDB, err := store.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("writer database: %v", err)
	}
	defer func() { _ = writerDB.Close() }()
	log.Println("Writer database connected.")

	readerDB, err := store.NewPostgresDB(cfg.DatabaseReadURL)
	if err != nil {
		log.Fatalf("reader database: %v", err)
	}
	defer func() { _ = readerDB.Close() }()
	log.Println("Reader database connected.")

	db := &store.DB{Writer: writerDB, Reader: readerDB}

	// Run database migrations on the writer only.
	d, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		log.Fatalf("migrations source: %v", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("migrations init: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migrations up: %v", err)
	}
	log.Println("Migrations applied.")

	// Build the dependency graph: store -> service -> handler.
	itemStore := store.NewItemStore(db)
	itemService := service.NewItemService(itemStore)
	itemHandler := handler.NewItemHandler(itemService)
	healthHandler := handler.NewHealthHandler(readerDB)

	// Configure the chi router with standard middleware.
	r := chi.NewRouter()
	r.Use(chimw.Logger)              // Log every request.
	r.Use(chimw.Recoverer)           // Recover from panics and return 500.
	r.Use(chimw.Compress(5))         // Gzip responses when client accepts it.
	r.Use(middleware.Decompress)     // Decompress gzip request bodies.

	// Health probes — no authentication required.
	r.Get("/healthz", healthHandler.Healthz)
	r.Get("/readyz", healthHandler.Readyz)

	// Item routes.
	r.Post("/items", itemHandler.Create)
	r.Get("/items/{id}", itemHandler.Get)

	// Graceful shutdown: lets in-flight requests finish (up to 10s) when
	// the process receives SIGINT or SIGTERM (e.g., Kubernetes rolling deploys).
	srv := &http.Server{Addr: ":" + cfg.Port, Handler: r}

	go func() {
		log.Printf("Server starting on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown: %v", err)
	}

	log.Println("Server stopped.")
}

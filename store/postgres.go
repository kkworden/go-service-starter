// Package store implements the persistence layer using raw SQL via database/sql
// with the pgx Postgres driver. Each store satisfies an interface defined by
// the service layer (consumer-defined interfaces).
package store

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib" // Register the pgx driver as "pgx" for database/sql.
)

// DB holds writer and reader connection pools. Read-only queries (SELECT) use
// the reader, which can point to a replica. Write queries (INSERT, UPDATE,
// DELETE) and transactions use the writer. In local dev both can point to the
// same Postgres instance.
type DB struct {
	Writer *sql.DB
	Reader *sql.DB
}

// NewPostgresDB opens a single connection pool to the given Postgres URL,
// pings it to verify connectivity, and returns the *sql.DB handle.
func NewPostgresDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

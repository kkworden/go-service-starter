// Package config loads application configuration from environment variables.
// Required variables (DATABASE_URL) cause an error if missing. Optional
// variables use sensible defaults but return an error if set to an unparseable
// value, preventing silent misconfiguration.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration for the server.
type Config struct {
	Port            string // HTTP listen port (default "8080").
	DatabaseURL     string // Postgres writer connection string (required).
	DatabaseReadURL string // Postgres reader replica connection string (defaults to DatabaseURL).
}

// Load reads configuration from the environment. It returns an error if any
// required variable is missing or if an optional variable is set but unparseable.
func Load() (Config, error) {
	port := envOr("PORT", "8080")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	// DATABASE_READ_URL defaults to DATABASE_URL when not set (single-instance Postgres).
	dbReadURL := envOr("DATABASE_READ_URL", dbURL)

	return Config{
		Port:            port,
		DatabaseURL:     dbURL,
		DatabaseReadURL: dbReadURL,
	}, nil
}

// envOr returns the value of the environment variable named by key,
// or fallback if the variable is empty or unset.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// envDurationOr returns the duration parsed from the named environment variable.
// If the variable is unset, it returns fallback. If the variable is set but
// cannot be parsed, it returns an error (no silent fallback).
//
// Scaffolding: not used by the starter's minimal Config, but provided for
// services that add TTL or timeout config fields (e.g., token lifetimes).
func envDurationOr(key string, fallback time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("invalid %s value %q: %w", key, v, err)
	}
	return d, nil
}

// envIntOr returns the integer parsed from the named environment variable.
// If the variable is unset, it returns fallback. If the variable is set but
// cannot be parsed, it returns an error (no silent fallback).
//
// Scaffolding: not used by the starter's minimal Config, but provided for
// services that add numeric config fields (e.g., max batch size, retry count).
func envIntOr(key string, fallback int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid %s value %q: %w", key, v, err)
	}
	return n, nil
}

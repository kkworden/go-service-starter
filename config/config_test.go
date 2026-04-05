package config_test

import (
	"testing"

	"go-service-starter/config"
)

func TestLoad_MissingDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing DATABASE_URL")
	}
}

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want %q", cfg.Port, "8080")
	}
	if cfg.DatabaseReadURL != cfg.DatabaseURL {
		t.Errorf("DatabaseReadURL = %q, want %q (same as DatabaseURL)", cfg.DatabaseReadURL, cfg.DatabaseURL)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://writer/db")
	t.Setenv("DATABASE_READ_URL", "postgres://reader/db")
	t.Setenv("PORT", "3000")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "3000" {
		t.Errorf("Port = %q, want %q", cfg.Port, "3000")
	}
	if cfg.DatabaseURL != "postgres://writer/db" {
		t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL, "postgres://writer/db")
	}
	if cfg.DatabaseReadURL != "postgres://reader/db" {
		t.Errorf("DatabaseReadURL = %q, want %q", cfg.DatabaseReadURL, "postgres://reader/db")
	}
}

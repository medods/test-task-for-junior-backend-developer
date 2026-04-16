package main

import (
    "os"
    "testing"
)

func TestLoadConfigDefaults(t *testing.T) {
    os.Unsetenv("HTTP_ADDR")
    os.Unsetenv("DATABASE_DSN")
    cfg := loadConfig()
    if cfg.HTTPAddr != ":8080" {
        t.Fatalf("unexpected default http addr: %s", cfg.HTTPAddr)
    }
    if cfg.DatabaseDSN == "" {
        t.Fatalf("expected default database dsn")
    }
}

func TestLoadConfigFromEnv(t *testing.T) {
    defer os.Unsetenv("HTTP_ADDR")
    defer os.Unsetenv("DATABASE_DSN")
    os.Setenv("HTTP_ADDR", ":9090")
    os.Setenv("DATABASE_DSN", "mydsn")
    cfg := loadConfig()
    if cfg.HTTPAddr != ":9090" {
        t.Fatalf("expected env http addr, got %s", cfg.HTTPAddr)
    }
    if cfg.DatabaseDSN != "mydsn" {
        t.Fatalf("expected env database dsn, got %s", cfg.DatabaseDSN)
    }
}

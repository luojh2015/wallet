package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.App.Env != "development" {
		t.Errorf("App.Env = %q, want %q", cfg.App.Env, "development")
	}
	if cfg.App.HTTPPort != 8080 {
		t.Errorf("App.HTTPPort = %d, want 8080", cfg.App.HTTPPort)
	}
	if cfg.App.GRPCPort != 9090 {
		t.Errorf("App.GRPCPort = %d, want 9090", cfg.App.GRPCPort)
	}
	if cfg.App.MachineID != 1 {
		t.Errorf("App.MachineID = %d, want 1", cfg.App.MachineID)
	}
	if cfg.Session.TTL != 24*time.Hour {
		t.Errorf("Session.TTL = %v, want 24h", cfg.Session.TTL)
	}
	if cfg.Postgres.Host != "localhost" {
		t.Errorf("Postgres.Host = %q, want %q", cfg.Postgres.Host, "localhost")
	}
	if cfg.Postgres.Port != 5432 {
		t.Errorf("Postgres.Port = %d, want 5432", cfg.Postgres.Port)
	}
	if cfg.Postgres.Database != "wallet" {
		t.Errorf("Postgres.Database = %q, want %q", cfg.Postgres.Database, "wallet")
	}
	if cfg.Postgres.MaxConns != 100 {
		t.Errorf("Postgres.MaxConns = %d, want 100", cfg.Postgres.MaxConns)
	}
}

func TestLoad_EmptyPath_NoFile(t *testing.T) {
	// change to a temp dir so default paths won't be found
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load('') error = %v", err)
	}
	if cfg.App.HTTPPort != 8080 {
		t.Errorf("expected default HTTPPort 8080, got %d", cfg.App.HTTPPort)
	}
}

func TestLoad_InvalidPath(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("Load() expected error for nonexistent path, got nil")
	}
}

func TestLoad_ValidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := []byte(`
app:
  env: production
  http_port: 3000
  grpc_port: 4000
  machine_id: 5
session:
  ttl: 2h
`)
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.App.Env != "production" {
		t.Errorf("App.Env = %q, want %q", cfg.App.Env, "production")
	}
	if cfg.App.HTTPPort != 3000 {
		t.Errorf("App.HTTPPort = %d, want 3000", cfg.App.HTTPPort)
	}
	if cfg.App.GRPCPort != 4000 {
		t.Errorf("App.GRPCPort = %d, want 4000", cfg.App.GRPCPort)
	}
	if cfg.App.MachineID != 5 {
		t.Errorf("App.MachineID = %d, want 5", cfg.App.MachineID)
	}
	if cfg.Session.TTL != 2*time.Hour {
		t.Errorf("Session.TTL = %v, want 2h", cfg.Session.TTL)
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	// change to a temp dir so default paths won't be found
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	t.Setenv("APP_ENV", "production")
	t.Setenv("APP_HTTP_PORT", "3000")
	t.Setenv("APP_GRPC_PORT", "4000")
	t.Setenv("APP_MACHINE_ID", "42")
	t.Setenv("POSTGRES_HOST", "db.example.com")
	t.Setenv("POSTGRES_PORT", "5433")
	t.Setenv("POSTGRES_USER", "admin")
	t.Setenv("POSTGRES_PASSWORD", "secret")
	t.Setenv("POSTGRES_DB", "walletdb")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load('') error = %v", err)
	}
	if cfg.App.Env != "production" {
		t.Errorf("App.Env = %q, want %q", cfg.App.Env, "production")
	}
	if cfg.App.HTTPPort != 3000 {
		t.Errorf("App.HTTPPort = %d, want 3000", cfg.App.HTTPPort)
	}
	if cfg.App.GRPCPort != 4000 {
		t.Errorf("App.GRPCPort = %d, want 4000", cfg.App.GRPCPort)
	}
	if cfg.App.MachineID != 42 {
		t.Errorf("App.MachineID = %d, want 42", cfg.App.MachineID)
	}
	if cfg.Postgres.Host != "db.example.com" {
		t.Errorf("Postgres.Host = %q, want %q", cfg.Postgres.Host, "db.example.com")
	}
	if cfg.Postgres.Port != 5433 {
		t.Errorf("Postgres.Port = %d, want 5433", cfg.Postgres.Port)
	}
	if cfg.Postgres.User != "admin" {
		t.Errorf("Postgres.User = %q, want %q", cfg.Postgres.User, "admin")
	}
	if cfg.Postgres.Password != "secret" {
		t.Errorf("Postgres.Password = %q, want %q", cfg.Postgres.Password, "secret")
	}
	if cfg.Postgres.Database != "walletdb" {
		t.Errorf("Postgres.Database = %q, want %q", cfg.Postgres.Database, "walletdb")
	}
}

func TestLoad_EnvOverride_InvalidPort(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	t.Setenv("APP_HTTP_PORT", "not-a-number")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load('') error = %v", err)
	}
	// invalid port should leave default
	if cfg.App.HTTPPort != 8080 {
		t.Errorf("App.HTTPPort = %d, want default 8080 (invalid env ignored)", cfg.App.HTTPPort)
	}
}

func TestPostgresConfig_DSN(t *testing.T) {
	cfg := PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "wallet",
		Password: "wallet123",
		Database: "wallet",
	}

	dsn := cfg.DSN()
	expected := "host=localhost port=5432 user=wallet password=wallet123 dbname=wallet sslmode=disable"
	if dsn != expected {
		t.Errorf("DSN() = %q, want %q", dsn, expected)
	}
}

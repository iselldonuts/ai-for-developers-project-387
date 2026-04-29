package config

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestLoadReadsDotEnv(t *testing.T) {
	withCleanEnv(t)
	withWorkingDir(t, t.TempDir())

	writeDotEnv(t, `
APP_ENV=test
APP_ADDR=:9090
APP_LOG_LEVEL=debug
DATABASE_URL=postgres://user:pass@localhost:5432/test?sslmode=disable
WEB_PORT=3000
CORS_ALLOWED_ORIGINS=http://localhost:3000, http://127.0.0.1:3000
`)

	cfg := Load()

	if cfg.Env != "test" {
		t.Fatalf("Env = %q, want %q", cfg.Env, "test")
	}
	if cfg.Addr != ":9090" {
		t.Fatalf("Addr = %q, want %q", cfg.Addr, ":9090")
	}
	if cfg.LogLevel != "debug" {
		t.Fatalf("LogLevel = %q, want %q", cfg.LogLevel, "debug")
	}
	if cfg.Storage != "postgres" {
		t.Fatalf("Storage = %q, want %q", cfg.Storage, "postgres")
	}
	if cfg.DatabaseURL != "postgres://user:pass@localhost:5432/test?sslmode=disable" {
		t.Fatalf("DatabaseURL = %q", cfg.DatabaseURL)
	}
	if cfg.WebPort != "3000" {
		t.Fatalf("WebPort = %q, want %q", cfg.WebPort, "3000")
	}

	wantOrigins := []string{"http://localhost:3000", "http://127.0.0.1:3000"}
	if !slices.Equal(cfg.CORSAllowedOrigins, wantOrigins) {
		t.Fatalf("CORSAllowedOrigins = %v, want %v", cfg.CORSAllowedOrigins, wantOrigins)
	}
}

func TestLoadDoesNotOverrideExistingEnv(t *testing.T) {
	withCleanEnv(t)
	withWorkingDir(t, t.TempDir())

	t.Setenv("APP_ADDR", ":7070")
	writeDotEnv(t, "APP_ADDR=:9090\n")

	cfg := Load()

	if cfg.Addr != ":7070" {
		t.Fatalf("Addr = %q, want existing env value %q", cfg.Addr, ":7070")
	}
}

func TestLoadUsesFallbacksWhenEnvAndDotEnvAreMissing(t *testing.T) {
	withCleanEnv(t)
	withWorkingDir(t, t.TempDir())

	cfg := Load()

	if cfg.Env != "development" {
		t.Fatalf("Env = %q, want %q", cfg.Env, "development")
	}
	if cfg.Addr != ":8080" {
		t.Fatalf("Addr = %q, want %q", cfg.Addr, ":8080")
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("LogLevel = %q, want %q", cfg.LogLevel, "info")
	}
	if cfg.Storage != "memory" {
		t.Fatalf("Storage = %q, want %q", cfg.Storage, "memory")
	}
	if cfg.DatabaseURL != "" {
		t.Fatalf("DatabaseURL = %q", cfg.DatabaseURL)
	}
	if cfg.WebPort != "5173" {
		t.Fatalf("WebPort = %q, want %q", cfg.WebPort, "5173")
	}

	wantOrigins := []string{"http://localhost:5173", "http://127.0.0.1:5173"}
	if !slices.Equal(cfg.CORSAllowedOrigins, wantOrigins) {
		t.Fatalf("CORSAllowedOrigins = %v, want %v", cfg.CORSAllowedOrigins, wantOrigins)
	}
}

func TestLoadUsesExplicitStorage(t *testing.T) {
	withCleanEnv(t)
	withWorkingDir(t, t.TempDir())

	t.Setenv("APP_STORAGE", "postgres")

	cfg := Load()

	if cfg.Storage != "postgres" {
		t.Fatalf("Storage = %q, want %q", cfg.Storage, "postgres")
	}
}

func TestLoadUsesPostgresWhenDatabaseURLExists(t *testing.T) {
	withCleanEnv(t)
	withWorkingDir(t, t.TempDir())

	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/test?sslmode=disable")

	cfg := Load()

	if cfg.Storage != "postgres" {
		t.Fatalf("Storage = %q, want %q", cfg.Storage, "postgres")
	}
}

func withCleanEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"APP_ENV",
		"APP_ADDR",
		"APP_LOG_LEVEL",
		"APP_STORAGE",
		"DATABASE_URL",
		"PORT",
		"WEB_PORT",
		"CORS_ALLOWED_ORIGINS",
	} {
		t.Setenv(key, "")
		os.Unsetenv(key)
	}
}

func TestLoadUsesPortForAddr(t *testing.T) {
	withCleanEnv(t)
	withWorkingDir(t, t.TempDir())

	t.Setenv("APP_ADDR", ":7070")
	t.Setenv("PORT", "10000")

	cfg := Load()

	if cfg.Addr != ":10000" {
		t.Fatalf("Addr = %q, want PORT-derived addr %q", cfg.Addr, ":10000")
	}
}

func withWorkingDir(t *testing.T, dir string) {
	t.Helper()

	current, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working dir: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("change working dir: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(current); err != nil {
			t.Fatalf("restore working dir: %v", err)
		}
	})
}

func writeDotEnv(t *testing.T, contents string) {
	t.Helper()

	path := filepath.Join(".", ".env")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
}

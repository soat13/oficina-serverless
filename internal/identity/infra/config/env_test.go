package config_test

import (
	"testing"

	"github.com/soat13/oficina-serverless/internal/identity/infra/config"
)

func TestEnvConfig(t *testing.T) {
	t.Run("success - Load returns values from env and Valid is true", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "super-secret")
		t.Setenv("JWT_ISSUER", "auth-service")
		t.Setenv("JWT_TTL", "7200")

		cfg := config.Load()

		if cfg.JWTSecret != "super-secret" {
			t.Fatalf("expected JWTSecret %q, got %q", "super-secret", cfg.JWTSecret)
		}
		if cfg.JWTIssuer != "auth-service" {
			t.Fatalf("expected JWTIssuer %q, got %q", "auth-service", cfg.JWTIssuer)
		}
		if cfg.JWTTTL != 7200 {
			t.Fatalf("expected JWTTTL %d, got %d", int64(7200), cfg.JWTTTL)
		}
		if !cfg.Valid() {
			t.Fatalf("expected config to be valid")
		}
	})

	t.Run("JWT_TTL empty - Load uses default 3600", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "super-secret")
		t.Setenv("JWT_ISSUER", "auth-service")
		t.Setenv("JWT_TTL", "")

		cfg := config.Load()

		if cfg.JWTTTL != 3600 {
			t.Fatalf("expected default JWTTTL %d, got %d", int64(3600), cfg.JWTTTL)
		}
		if !cfg.Valid() {
			t.Fatalf("expected config to be valid")
		}
	})

	t.Run("JWT_TTL missing - Load uses default 3600", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "super-secret")
		t.Setenv("JWT_ISSUER", "auth-service")
		// no JWT_TTL set

		cfg := config.Load()

		if cfg.JWTTTL != 3600 {
			t.Fatalf("expected default JWTTTL %d, got %d", int64(3600), cfg.JWTTTL)
		}
		if !cfg.Valid() {
			t.Fatalf("expected config to be valid")
		}
	})

	t.Run("JWT_TTL invalid - Load falls back to default 3600", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "super-secret")
		t.Setenv("JWT_ISSUER", "auth-service")
		t.Setenv("JWT_TTL", "not-a-number")

		cfg := config.Load()

		if cfg.JWTTTL != 3600 {
			t.Fatalf("expected default JWTTTL %d, got %d", int64(3600), cfg.JWTTTL)
		}
		if !cfg.Valid() {
			t.Fatalf("expected config to be valid")
		}
	})

	t.Run("Valid false - missing secret", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "")
		t.Setenv("JWT_ISSUER", "auth-service")
		t.Setenv("JWT_TTL", "3600")

		cfg := config.Load()

		if cfg.Valid() {
			t.Fatalf("expected config to be invalid")
		}
	})

	t.Run("Valid false - missing issuer", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "super-secret")
		t.Setenv("JWT_ISSUER", "")
		t.Setenv("JWT_TTL", "3600")

		cfg := config.Load()

		if cfg.Valid() {
			t.Fatalf("expected config to be invalid")
		}
	})

	t.Run("Valid false - non-positive TTL", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "super-secret")
		t.Setenv("JWT_ISSUER", "auth-service")
		t.Setenv("JWT_TTL", "0")

		cfg := config.Load()

		if cfg.Valid() {
			t.Fatalf("expected config to be invalid")
		}
	})
}

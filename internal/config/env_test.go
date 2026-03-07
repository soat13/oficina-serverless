package config_test

import (
	"strings"
	"testing"

	"github.com/soat13/oficina-serverless/internal/config"
	"github.com/stretchr/testify/require"
)

func setValidBaseEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DB_DSN", "postgres://auth:auth@localhost:5432/auth?sslmode=disable")
	t.Setenv("JWT_SECRET", "super-secret")
	t.Setenv("JWT_ISSUER", "auth-service")
}

func assertValid(t *testing.T, cfg config.Config) {
	t.Helper()
	cfg, err := config.Load()
	require.NoError(t, err)
}

func assertInvalidContains(t *testing.T, err error, expectedSubstrings ...string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected config to be invalid")
	}

	msg := err.Error()
	for _, s := range expectedSubstrings {
		if !strings.Contains(msg, s) {
			t.Fatalf("expected error to contain %q, got: %q", s, msg)
		}
	}
}

func TestEnvConfig(t *testing.T) {
	t.Run("success - Load returns values from env and validate is nil", func(t *testing.T) {
		setValidBaseEnv(t)
		t.Setenv("JWT_TTL", "7200")

		cfg, err := config.Load()
		require.NoError(t, err)

		if cfg.DBDSN == "" {
			t.Fatalf("expected DBDSN to be set")
		}
		if cfg.JWTSecret != "super-secret" {
			t.Fatalf("expected JWTSecret %q, got %q", "super-secret", cfg.JWTSecret)
		}
		if cfg.JWTIssuer != "auth-service" {
			t.Fatalf("expected JWTIssuer %q, got %q", "auth-service", cfg.JWTIssuer)
		}
		if cfg.JWTTTL != 7200 {
			t.Fatalf("expected JWTTTL %d, got %d", int64(7200), cfg.JWTTTL)
		}

		assertValid(t, cfg)
	})

	t.Run("JWT_TTL empty - Load uses default 3600", func(t *testing.T) {
		setValidBaseEnv(t)
		t.Setenv("JWT_TTL", "")

		cfg, err := config.Load()
		require.NoError(t, err)

		if cfg.JWTTTL != 3600 {
			t.Fatalf("expected default JWTTTL %d, got %d", int64(3600), cfg.JWTTTL)
		}
		assertValid(t, cfg)
	})

	t.Run("JWT_TTL missing - Load uses default 3600", func(t *testing.T) {
		setValidBaseEnv(t)

		cfg, err := config.Load()
		require.NoError(t, err)

		if cfg.JWTTTL != 3600 {
			t.Fatalf("expected default JWTTTL %d, got %d", int64(3600), cfg.JWTTTL)
		}
		assertValid(t, cfg)
	})

	t.Run("JWT_TTL invalid - Load falls back to default 3600", func(t *testing.T) {
		setValidBaseEnv(t)
		t.Setenv("JWT_TTL", "not-a-number")

		cfg, err := config.Load()
		require.NoError(t, err)

		if cfg.JWTTTL != 3600 {
			t.Fatalf("expected default JWTTTL %d, got %d", int64(3600), cfg.JWTTTL)
		}
		assertValid(t, cfg)
	})

	t.Run("invalid - missing DB_DSN", func(t *testing.T) {
		setValidBaseEnv(t)
		t.Setenv("DB_DSN", "")

		_, err := config.Load()
		assertInvalidContains(t, err, "DB_DSN")
	})

	t.Run("invalid - missing secret", func(t *testing.T) {
		setValidBaseEnv(t)
		t.Setenv("JWT_SECRET", "")

		_, err := config.Load()
		assertInvalidContains(t, err, "JWT_SECRET")
	})

	t.Run("invalid - missing issuer", func(t *testing.T) {
		setValidBaseEnv(t)
		t.Setenv("JWT_ISSUER", "")

		_, err := config.Load()
		assertInvalidContains(t, err, "JWT_ISSUER")
	})

	t.Run("invalid - non-positive TTL (validate catches it)", func(t *testing.T) {
		setValidBaseEnv(t)
		t.Setenv("JWT_TTL", "-1")

		_, err := config.Load()
		assertInvalidContains(t, err, "JWT_TTL")
	})

	t.Run("invalid - multiple missing values shows all of them", func(t *testing.T) {
		setValidBaseEnv(t)
		t.Setenv("DB_DSN", "")
		t.Setenv("JWT_SECRET", "")
		t.Setenv("JWT_ISSUER", "")

		_, err := config.Load()
		assertInvalidContains(t, err, "DB_DSN", "JWT_SECRET", "JWT_ISSUER")
	})
}

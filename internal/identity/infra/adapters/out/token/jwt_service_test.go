package tokenadapter_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/soat13/oficina-serverless/internal/identity/app/ports/out"
	tokenadapter "github.com/soat13/oficina-serverless/internal/identity/infra/adapters/out/token"
	"github.com/soat13/oficina-serverless/internal/shared/token"
)

func TestJWTService(t *testing.T) {
	ctx := context.Background()

	const (
		secret = "super-secret"
		issuer = "auth-service"
	)

	t.Run("success - Sign then Verify returns subject", func(t *testing.T) {
		svc := tokenadapter.New(secret, issuer)

		tk, err := svc.Sign(ctx, token.Subject("user-1"), token.SignOptions{TTLSeconds: 3600})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		sub, err := svc.Verify(ctx, tk)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if sub != ("user-1") {
			t.Fatalf("expected subject %q, got %q", "user-1", sub)
		}
	})

	t.Run("NewWithMethod - nil method defaults to HS256 and Sign/Verify works", func(t *testing.T) {
		svc := tokenadapter.NewWithMethod(secret, issuer, nil)

		tk, err := svc.Sign(ctx, "user-1", token.SignOptions{TTLSeconds: 10})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		_, err = svc.Verify(ctx, tk)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("expired token - Verify returns ErrTokenExpired", func(t *testing.T) {
		svc := tokenadapter.New(secret, issuer)

		now := time.Now()
		claims := jwt.MapClaims{
			"sub": "user-1",
			"iss": issuer,
			"iat": now.Add(-2 * time.Hour).Unix(),
			"exp": now.Add(-1 * time.Hour).Unix(), // expired 1 hour ago
		}

		tk := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := tk.SignedString([]byte(secret))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		_, err = svc.Verify(ctx, token.Token(signed))
		if !errors.Is(err, out.ErrTokenExpired) {
			t.Fatalf("expected ErrTokenExpired, got %v", err)
		}
	})

	t.Run("Sign - empty secret returns ErrInvalidToken", func(t *testing.T) {
		svc := tokenadapter.New("", issuer)

		_, err := svc.Sign(ctx, "user-1", token.SignOptions{TTLSeconds: 3600})
		if !errors.Is(err, out.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("Sign - empty issuer returns ErrInvalidToken", func(t *testing.T) {
		svc := tokenadapter.New(secret, "")

		_, err := svc.Sign(ctx, "user-1", token.SignOptions{TTLSeconds: 3600})
		if !errors.Is(err, out.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("Sign - empty subject returns ErrInvalidToken", func(t *testing.T) {
		svc := tokenadapter.New(secret, issuer)

		_, err := svc.Sign(ctx, "", token.SignOptions{TTLSeconds: 3600})
		if !errors.Is(err, out.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("Sign - non-positive TTL returns ErrInvalidToken", func(t *testing.T) {
		svc := tokenadapter.New(secret, issuer)

		_, err := svc.Sign(ctx, "user-1", token.SignOptions{TTLSeconds: 0})
		if !errors.Is(err, out.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("Sign - signed string error (RS256 with []byte key) returns ErrInvalidToken", func(t *testing.T) {
		svc := tokenadapter.NewWithMethod(secret, issuer, jwt.SigningMethodRS256)

		_, err := svc.Sign(ctx, "user-1", token.SignOptions{TTLSeconds: 3600})
		if !errors.Is(err, out.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("Verify - empty token returns ErrInvalidToken", func(t *testing.T) {
		svc := tokenadapter.New(secret, issuer)

		_, err := svc.Verify(ctx, "")
		if !errors.Is(err, out.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("Verify - empty secret returns ErrInvalidToken", func(t *testing.T) {
		svc := tokenadapter.New("", issuer)

		_, err := svc.Verify(ctx, "anything")
		if !errors.Is(err, out.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("Verify - empty issuer returns ErrInvalidToken", func(t *testing.T) {
		svc := tokenadapter.New(secret, "")

		_, err := svc.Verify(ctx, "anything")
		if !errors.Is(err, out.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("issuer mismatch - Verify returns ErrInvalidToken", func(t *testing.T) {
		signer := tokenadapter.New(secret, "issuer-A")
		verifier := tokenadapter.New(secret, "issuer-B")

		tk, err := signer.Sign(ctx, "user-1", token.SignOptions{TTLSeconds: 3600})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		_, err = verifier.Verify(ctx, tk)
		if !errors.Is(err, out.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("signature mismatch (wrong secret) - Verify returns ErrInvalidToken", func(t *testing.T) {
		signer := tokenadapter.New("secret-A", issuer)
		verifier := tokenadapter.New("secret-B", issuer)

		tk, err := signer.Sign(ctx, "user-1", token.SignOptions{TTLSeconds: 3600})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		_, err = verifier.Verify(ctx, tk)
		if !errors.Is(err, out.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("malformed token - Verify returns ErrInvalidToken", func(t *testing.T) {
		svc := tokenadapter.New(secret, issuer)

		_, err := svc.Verify(ctx, "not-a-jwt")
		if !errors.Is(err, out.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("missing subject - returns ErrInvalidToken", func(t *testing.T) {
		svc := tokenadapter.New(secret, issuer)

		now := time.Now()
		claims := jwt.MapClaims{
			"iss": issuer,
			"iat": now.Unix(),
			"exp": now.Add(time.Hour).Unix(),
		}

		tk := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, _ := tk.SignedString([]byte(secret))

		_, err := svc.Verify(ctx, token.Token(signed))
		if !errors.Is(err, out.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("missing issuer - returns ErrInvalidToken", func(t *testing.T) {
		svc := tokenadapter.New(secret, issuer)

		now := time.Now()
		claims := jwt.MapClaims{
			"sub": "user-1",
			"iat": now.Unix(),
			"exp": now.Add(time.Hour).Unix(),
		}

		tk := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, _ := tk.SignedString([]byte(secret))

		_, err := svc.Verify(ctx, token.Token(signed))
		if !errors.Is(err, out.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("issuer claim not a string - returns ErrInvalidToken", func(t *testing.T) {
		svc := tokenadapter.New(secret, issuer)

		now := time.Now()
		claims := jwt.MapClaims{
			"sub": "user-1",
			"iss": 123,
			"iat": now.Unix(),
			"exp": now.Add(time.Hour).Unix(),
		}

		tk := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, _ := tk.SignedString([]byte(secret))

		_, err := svc.Verify(ctx, token.Token(signed))
		if !errors.Is(err, out.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("subject claim not a string - returns ErrInvalidToken", func(t *testing.T) {
		svc := tokenadapter.New(secret, issuer)

		now := time.Now()
		claims := jwt.MapClaims{
			"sub": 999,
			"iss": issuer,
			"iat": now.Unix(),
			"exp": now.Add(time.Hour).Unix(),
		}

		tk := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, _ := tk.SignedString([]byte(secret))

		_, err := svc.Verify(ctx, token.Token(signed))
		if !errors.Is(err, out.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("algorithm mismatch - Verify returns ErrInvalidToken", func(t *testing.T) {
		signer := tokenadapter.NewWithMethod(secret, issuer, jwt.SigningMethodHS256)
		tk, err := signer.Sign(ctx, token.Subject("user-1"), token.SignOptions{TTLSeconds: 3600})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		verifier := tokenadapter.NewWithMethod(secret, issuer, jwt.SigningMethodHS512)

		_, err = verifier.Verify(ctx, tk)
		if !errors.Is(err, out.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})

}

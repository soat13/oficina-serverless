package apigw_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	identitybootstrap "github.com/soat13/oficina-serverless/internal/bootstrap/identity"
	"github.com/soat13/oficina-serverless/internal/container"
	"github.com/soat13/oficina-serverless/internal/identity/infra/adapters/in/apigw"
	"github.com/soat13/oficina-serverless/internal/identity/infra/config"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/bcrypt"
)

const (
	testSecret = "secret"
	testIssuer = "issuer"
	testTTL    = 3600
)

func setup(t *testing.T) (*bun.DB, *apigw.Handler) {
	ctx := context.Background()

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://auth:auth@localhost:5432/auth?sslmode=disable"
	}

	cfg := config.Config{
		DBDSN:     dsn,
		JWTSecret: testSecret,
		JWTIssuer: testIssuer,
		JWTTTL:    testTTL,
	}

	c, err := container.New(cfg)
	if err != nil {
		t.Fatalf("container.New: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })

	_, err = c.DB.ExecContext(ctx, `
		DROP TABLE IF EXISTS users;

		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			document TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL
		);
	`)
	if err != nil {
		t.Fatalf("reset table: %v", err)
	}

	mod, err := identitybootstrap.Setup(c)
	if err != nil {
		t.Fatalf("identitybootstrap.Setup: %v", err)
	}

	return c.DB, mod.ApiGwHandler
}

func insertUser(t *testing.T, db *bun.DB, id, cpfVal, password string) {
	ctx := context.Background()

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	_, _ = db.ExecContext(ctx, `DELETE FROM users WHERE document = ?`, cpfVal)

	_, err := db.ExecContext(ctx,
		`INSERT INTO users(id, document, password) VALUES (?,?,?)`,
		id, cpfVal, string(hash),
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
}

func post(_ *testing.T, handler *apigw.Handler, ctx context.Context, path string, body string, isBase64 bool) events.APIGatewayProxyResponse {
	resp, _ := handler.Handle(ctx, events.APIGatewayProxyRequest{
		Path:            path,
		HTTPMethod:      "POST",
		Body:            body,
		IsBase64Encoded: isBase64,
	})
	return resp
}

func TestAuthE2E(t *testing.T) {
	ctx := context.Background()
	db, handler := setup(t)

	t.Run("POST /auth/login -> success", func(t *testing.T) {
		insertUser(t, db, "user-1", "52998224725", "123")

		body, _ := json.Marshal(map[string]string{
			"cpf":      "52998224725",
			"password": "123",
		})

		resp := post(t, handler, ctx, "/auth/login", string(body), false)

		if resp.StatusCode != 200 {
			t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, resp.Body)
		}

		var out apigw.TokenResponse
		_ = json.Unmarshal([]byte(resp.Body), &out)

		if out.AccessToken == "" {
			t.Fatal("empty access token")
		}
	})

	t.Run("POST /auth/login -> cpf and password required", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"cpf":      "",
			"password": "",
		})

		resp := post(t, handler, ctx, "/auth/login", string(body), false)

		if resp.StatusCode != 400 {
			t.Fatalf("expected 400, got %d body=%s", resp.StatusCode, resp.Body)
		}

		var errResp apigw.ErrorResponse
		_ = json.Unmarshal([]byte(resp.Body), &errResp)
		if errResp.Error != "cpf_and_password_required" {
			t.Fatalf("expected cpf_and_password_required, got %q body=%s", errResp.Error, resp.Body)
		}
	})

	t.Run("POST /auth/login -> user not found", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"cpf":      "03940582166",
			"password": "123",
		})

		resp := post(t, handler, ctx, "/auth/login", string(body), false)

		if resp.StatusCode != 401 {
			t.Fatalf("expected 401, got %d body=%s", resp.StatusCode, resp.Body)
		}
	})

	t.Run("POST /auth/login -> invalid password", func(t *testing.T) {
		insertUser(t, db, "user-1", "52998224725", "123")

		body, _ := json.Marshal(map[string]string{
			"cpf":      "52998224725",
			"password": "wrong",
		})

		resp := post(t, handler, ctx, "/auth/login", string(body), false)

		if resp.StatusCode != 401 {
			t.Fatalf("expected 401, got %d body=%s", resp.StatusCode, resp.Body)
		}
	})

	t.Run("POST /auth/login -> invalid json", func(t *testing.T) {
		resp := post(t, handler, ctx, "/auth/login", "{invalid-json", false)

		if resp.StatusCode != 400 {
			t.Fatalf("expected 400, got %d body=%s", resp.StatusCode, resp.Body)
		}
	})

	t.Run("POST /auth/login -> empty body", func(t *testing.T) {
		resp := post(t, handler, ctx, "/auth/login", "", false)

		if resp.StatusCode != 400 {
			t.Fatalf("expected 400, got %d body=%s", resp.StatusCode, resp.Body)
		}
	})

	t.Run("POST /auth/login -> base64 body valid", func(t *testing.T) {
		insertUser(t, db, "user-1", "52998224725", "123")

		raw := `{"cpf":"52998224725","password":"123"}`
		encoded := base64.StdEncoding.EncodeToString([]byte(raw))

		resp := post(t, handler, ctx, "/auth/login", encoded, true)

		if resp.StatusCode != 200 {
			t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, resp.Body)
		}

		var out apigw.TokenResponse
		_ = json.Unmarshal([]byte(resp.Body), &out)

		if out.AccessToken == "" {
			t.Fatal("empty access token")
		}
	})

	t.Run("POST /auth/login -> base64 body invalid", func(t *testing.T) {
		resp := post(t, handler, ctx, "/auth/login", "###not-base64###", true)

		if resp.StatusCode != 400 {
			t.Fatalf("expected 400, got %d body=%s", resp.StatusCode, resp.Body)
		}
	})
}

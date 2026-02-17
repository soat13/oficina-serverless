package apigw_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/golang-jwt/jwt/v5"
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

func post(t *testing.T, handler *apigw.Handler, ctx context.Context, path string, body string, isBase64 bool) events.APIGatewayV2HTTPResponse {
	resp, _ := handler.Handle(ctx, events.APIGatewayV2HTTPRequest{
		RawPath: path,
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: "POST",
			},
		},
		Body:            body,
		IsBase64Encoded: isBase64,
	})
	return resp
}

func TestAuthE2E(t *testing.T) {
	ctx := context.Background()
	db, handler := setup(t)

	t.Run("POST /auth/token -> success", func(t *testing.T) {
		insertUser(t, db, "user-1", "52998224725", "123")

		body, _ := json.Marshal(map[string]string{
			"cpf":      "52998224725",
			"password": "123",
		})

		resp := post(t, handler, ctx, "/auth/token", string(body), false)

		if resp.StatusCode != 200 {
			t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, resp.Body)
		}

		var out apigw.TokenResponse
		_ = json.Unmarshal([]byte(resp.Body), &out)

		if out.AccessToken == "" {
			t.Fatal("empty access token")
		}
	})

	t.Run("POST /auth/token -> cpf and password required", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"cpf":      "",
			"password": "",
		})

		resp := post(t, handler, ctx, "/auth/token", string(body), false)

		if resp.StatusCode != 400 {
			t.Fatalf("expected 400, got %d body=%s", resp.StatusCode, resp.Body)
		}

		var errResp apigw.ErrorResponse
		_ = json.Unmarshal([]byte(resp.Body), &errResp)
		if errResp.Error != "cpf_and_password_required" {
			t.Fatalf("expected cpf_and_password_required, got %q body=%s", errResp.Error, resp.Body)
		}
	})

	t.Run("POST /auth/token -> user not found", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"cpf":      "03940582166",
			"password": "123",
		})

		resp := post(t, handler, ctx, "/auth/token", string(body), false)

		if resp.StatusCode != 401 {
			t.Fatalf("expected 401, got %d body=%s", resp.StatusCode, resp.Body)
		}
	})

	t.Run("POST /auth/token -> invalid password", func(t *testing.T) {
		insertUser(t, db, "user-1", "52998224725", "123")

		body, _ := json.Marshal(map[string]string{
			"cpf":      "52998224725",
			"password": "wrong",
		})

		resp := post(t, handler, ctx, "/auth/token", string(body), false)

		if resp.StatusCode != 401 {
			t.Fatalf("expected 401, got %d body=%s", resp.StatusCode, resp.Body)
		}
	})

	t.Run("POST /auth/token -> invalid json", func(t *testing.T) {
		resp := post(t, handler, ctx, "/auth/token", "{invalid-json", false)

		if resp.StatusCode != 400 {
			t.Fatalf("expected 400, got %d body=%s", resp.StatusCode, resp.Body)
		}
	})

	t.Run("POST /auth/token -> empty body", func(t *testing.T) {
		resp := post(t, handler, ctx, "/auth/token", "", false)

		if resp.StatusCode != 400 {
			t.Fatalf("expected 400, got %d body=%s", resp.StatusCode, resp.Body)
		}
	})

	t.Run("POST /auth/token -> base64 body valid", func(t *testing.T) {
		insertUser(t, db, "user-1", "52998224725", "123")

		raw := `{"cpf":"52998224725","password":"123"}`
		encoded := base64.StdEncoding.EncodeToString([]byte(raw))

		resp := post(t, handler, ctx, "/auth/token", encoded, true)

		if resp.StatusCode != 200 {
			t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, resp.Body)
		}

		var out apigw.TokenResponse
		_ = json.Unmarshal([]byte(resp.Body), &out)

		if out.AccessToken == "" {
			t.Fatal("empty access token")
		}
	})

	t.Run("POST /auth/token -> base64 body invalid", func(t *testing.T) {
		resp := post(t, handler, ctx, "/auth/token", "###not-base64###", true)

		if resp.StatusCode != 400 {
			t.Fatalf("expected 400, got %d body=%s", resp.StatusCode, resp.Body)
		}
	})

	t.Run("POST /auth/verify -> success", func(t *testing.T) {
		insertUser(t, db, "user-1", "52998224725", "123")

		body, _ := json.Marshal(map[string]string{
			"cpf":      "52998224725",
			"password": "123",
		})
		resp := post(t, handler, ctx, "/auth/token", string(body), false)

		if resp.StatusCode != 200 {
			t.Fatalf("expected 200 generating token, got %d body=%s", resp.StatusCode, resp.Body)
		}

		var tokenResp apigw.TokenResponse
		_ = json.Unmarshal([]byte(resp.Body), &tokenResp)

		verifyBody, _ := json.Marshal(map[string]string{
			"token": tokenResp.AccessToken,
		})
		verifyResp := post(t, handler, ctx, "/auth/verify", string(verifyBody), false)

		if verifyResp.StatusCode != 200 {
			t.Fatalf("expected 200, got %d body=%s", verifyResp.StatusCode, verifyResp.Body)
		}
	})

	t.Run("POST /auth/verify -> invalid json", func(t *testing.T) {
		verifyResp := post(t, handler, ctx, "/auth/verify", "{invalid-json", false)

		if verifyResp.StatusCode != 400 {
			t.Fatalf("expected 400, got %d body=%s", verifyResp.StatusCode, verifyResp.Body)
		}
	})

	t.Run("POST /auth/verify -> empty body", func(t *testing.T) {
		verifyResp := post(t, handler, ctx, "/auth/verify", "", false)

		if verifyResp.StatusCode != 400 {
			t.Fatalf("expected 400, got %d body=%s", verifyResp.StatusCode, verifyResp.Body)
		}
	})

	t.Run("POST /auth/verify -> missing token field", func(t *testing.T) {
		verifyBody, _ := json.Marshal(map[string]string{
			"nope": "x",
		})
		verifyResp := post(t, handler, ctx, "/auth/verify", string(verifyBody), false)

		if verifyResp.StatusCode != 401 && verifyResp.StatusCode != 400 {
			t.Fatalf("expected 401 or 400, got %d body=%s", verifyResp.StatusCode, verifyResp.Body)
		}
	})

	t.Run("POST /auth/verify -> invalid token", func(t *testing.T) {
		verifyBody, _ := json.Marshal(map[string]string{
			"token": "not-a-jwt",
		})
		verifyResp := post(t, handler, ctx, "/auth/verify", string(verifyBody), false)

		if verifyResp.StatusCode != 401 {
			t.Fatalf("expected 401, got %d body=%s", verifyResp.StatusCode, verifyResp.Body)
		}
	})

	t.Run("POST /auth/verify -> base64 body valid", func(t *testing.T) {
		insertUser(t, db, "user-1", "52998224725", "123")

		body, _ := json.Marshal(map[string]string{
			"cpf":      "52998224725",
			"password": "123",
		})
		resp := post(t, handler, ctx, "/auth/token", string(body), false)

		if resp.StatusCode != 200 {
			t.Fatalf("expected 200 generating token, got %d body=%s", resp.StatusCode, resp.Body)
		}

		var tokenResp apigw.TokenResponse
		_ = json.Unmarshal([]byte(resp.Body), &tokenResp)

		rawVerify := `{"token":"` + tokenResp.AccessToken + `"}`
		encoded := base64.StdEncoding.EncodeToString([]byte(rawVerify))

		verifyResp := post(t, handler, ctx, "/auth/verify", encoded, true)

		if verifyResp.StatusCode != 200 {
			t.Fatalf("expected 200, got %d body=%s", verifyResp.StatusCode, verifyResp.Body)
		}
	})

	t.Run("POST /auth/verify -> base64 body invalid", func(t *testing.T) {
		verifyResp := post(t, handler, ctx, "/auth/verify", "###not-base64###", true)

		if verifyResp.StatusCode != 400 {
			t.Fatalf("expected 400, got %d body=%s", verifyResp.StatusCode, verifyResp.Body)
		}
	})

	t.Run("route not found -> 404", func(t *testing.T) {
		resp, _ := handler.Handle(ctx, events.APIGatewayV2HTTPRequest{
			RawPath: "/nope",
			RequestContext: events.APIGatewayV2HTTPRequestContext{
				HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
					Method: "GET",
				},
			},
		})

		if resp.StatusCode != 404 {
			t.Fatalf("expected 404, got %d body=%s", resp.StatusCode, resp.Body)
		}
	})

	t.Run("POST /auth/verify -> expired token", func(t *testing.T) {

		claims := jwt.RegisteredClaims{
			Issuer:    testIssuer,
			Subject:   "user-1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		}

		tk := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := tk.SignedString([]byte(testSecret))
		if err != nil {
			t.Fatalf("sign token: %v", err)
		}

		verifyBody, _ := json.Marshal(map[string]string{
			"token": signed,
		})

		verifyResp := post(t, handler, ctx, "/auth/verify", string(verifyBody), false)

		if verifyResp.StatusCode != 401 {
			t.Fatalf("expected 401, got %d body=%s", verifyResp.StatusCode, verifyResp.Body)
		}

		var errResp apigw.ErrorResponse
		_ = json.Unmarshal([]byte(verifyResp.Body), &errResp)

		if errResp.Error != "token_expired" {
			t.Fatalf("expected token_expired, got %q body=%s", errResp.Error, verifyResp.Body)
		}
	})

}

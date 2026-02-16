package apigw

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/soat13/oficina-serverless/internal/identity/app"

	"github.com/soat13/oficina-serverless/internal/identity/app/ports/out"
	"github.com/soat13/oficina-serverless/internal/shared/token"
)

type Handler struct {
	authenticate app.Authenticate
	verify       app.Verify
	ttlSecs      int64
}

func NewHandler(auth app.Authenticate, verify app.Verify, ttlSecs int64) *Handler {
	return &Handler{authenticate: auth, verify: verify, ttlSecs: ttlSecs}
}

func (h *Handler) Handle(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return Router(ctx, h, req), nil
}

func (h *Handler) PostToken(ctx context.Context, req events.APIGatewayV2HTTPRequest) events.APIGatewayV2HTTPResponse {
	var body TokenRequest
	if err := decodeJSON(req, &body); err != nil {
		return jsonError(http.StatusBadRequest, "invalid_body")
	}
	if body.CPF == "" || body.Password == "" {
		return jsonError(http.StatusBadRequest, "cpf_and_password_required")
	}

	inputDTO := app.AuthenticateInput{
		CPF:      body.CPF,
		Password: body.Password,
	}

	outputDTO, err := h.authenticate.Execute(ctx, inputDTO)
	if err != nil {
		return jsonError(http.StatusUnauthorized, "invalid_credentials")
	}

	return jsonOK(http.StatusOK, TokenResponse{
		AccessToken: string(outputDTO.AccessToken),
		TokenType:   "Bearer",
		ExpiresIn:   h.ttlSecs,
	})
}

func (h *Handler) PostVerify(ctx context.Context, req events.APIGatewayV2HTTPRequest) events.APIGatewayV2HTTPResponse {
	var body VerifyRequest
	if err := decodeJSON(req, &body); err != nil {
		return jsonError(http.StatusBadRequest, "invalid_body")
	}
	if body.Token == "" {
		return jsonError(http.StatusBadRequest, "token_required")
	}

	sub, err := h.verify.Execute(ctx, token.Token(body.Token))
	if err != nil {
		if errors.Is(err, out.ErrTokenExpired) {
			return jsonError(http.StatusUnauthorized, "token_expired")
		}
		return jsonError(http.StatusUnauthorized, "invalid_token")
	}

	return jsonOK(http.StatusOK, VerifyResponse{Subject: string(sub)})
}

func decodeJSON(req events.APIGatewayV2HTTPRequest, dst any) error {
	raw := []byte(req.Body)
	if req.IsBase64Encoded {
		decoded, err := base64.StdEncoding.DecodeString(req.Body)
		if err != nil {
			return err
		}
		raw = decoded
	}
	return json.Unmarshal(raw, dst)
}

func jsonOK(status int, v any) events.APIGatewayV2HTTPResponse {
	b, _ := json.Marshal(v)
	return events.APIGatewayV2HTTPResponse{
		StatusCode: status,
		Headers: map[string]string{
			"content-type": "application/json",
		},
		Body: string(b),
	}
}

func jsonError(status int, code string) events.APIGatewayV2HTTPResponse {
	return jsonOK(status, ErrorResponse{Error: code})
}

package out

import (
	"context"
	"errors"

	"github.com/soat13/oficina-serverless/internal/shared/token"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

type TokenService interface {
	Sign(ctx context.Context, subject token.Subject, opts token.SignOptions) (token.Token, error)
	Verify(ctx context.Context, t token.Token) (token.Subject, error)
}

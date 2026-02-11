package out

import (
	"context"

	"github.com/soat13/oficina-serverless/internal/domain/shared/token"
)

type TokenService interface {
	Sign(ctx context.Context, subject token.Subject, opts token.SignOptions) (token.Token, error)
	Verify(ctx context.Context, t token.Token) (token.Subject, error)
}

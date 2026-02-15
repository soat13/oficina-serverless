package app

import (
	"context"
	"errors"

	"github.com/soat13/oficina-serverless/internal/identity/app/ports/out"
	"github.com/soat13/oficina-serverless/internal/shared/token"
)

type Verify struct {
	tokenService out.TokenService
}

func NewVerify(tokenService out.TokenService) Verify {
	return Verify{tokenService: tokenService}
}

func (u Verify) Execute(ctx context.Context, t token.Token) (token.Subject, error) {
	sub, err := u.tokenService.Verify(ctx, t)
	if err != nil {
		if errors.Is(err, out.ErrTokenExpired) {
			return "", ErrTokenExpired
		}
		return "", ErrInvalidToken
	}

	if sub == "" {
		return "", ErrInvalidToken
	}

	return sub, nil
}

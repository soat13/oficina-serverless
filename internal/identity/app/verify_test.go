package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/soat13/oficina-serverless/internal/identity/app"
	"github.com/soat13/oficina-serverless/internal/identity/app/ports/out"
	"github.com/soat13/oficina-serverless/internal/identity/app/ports/out/mocks"
	"github.com/soat13/oficina-serverless/internal/shared/token"
)

func TestVerifyExecute(t *testing.T) {
	ctx := context.Background()

	t.Run("success - returns subject", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tokenService := mocks.NewMockTokenService(ctrl)

		tokenService.EXPECT().
			Verify(ctx, token.Token("access-token")).
			Return(token.Subject("user-1"), nil)

		uc := app.NewVerify(tokenService)

		sub, err := uc.Execute(ctx, token.Token("access-token"))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if sub != token.Subject("user-1") {
			t.Fatalf("expected subject %q, got %q", "user-1", sub)
		}
	})

	t.Run("token service returns invalid token - returns ErrInvalidToken", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tokenService := mocks.NewMockTokenService(ctrl)

		tokenService.EXPECT().
			Verify(ctx, token.Token("bad-token")).
			Return(token.Subject(""), out.ErrInvalidToken)

		uc := app.NewVerify(tokenService)

		_, err := uc.Execute(ctx, "bad-token")
		if !errors.Is(err, app.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("token service returns expired token - returns ErrTokenExpired", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tokenService := mocks.NewMockTokenService(ctrl)

		tokenService.EXPECT().
			Verify(ctx, token.Token("expired-token")).
			Return(token.Subject(""), out.ErrTokenExpired)

		uc := app.NewVerify(tokenService)

		_, err := uc.Execute(ctx, "expired-token")
		if !errors.Is(err, app.ErrTokenExpired) {
			t.Fatalf("expected ErrTokenExpired, got %v", err)
		}
	})

	t.Run("token service returns empty subject - returns ErrInvalidToken", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tokenService := mocks.NewMockTokenService(ctrl)

		tokenService.EXPECT().
			Verify(ctx, token.Token("access-token")).
			Return(token.Subject(""), nil)

		uc := app.NewVerify(tokenService)

		_, err := uc.Execute(ctx, "access-token")
		if !errors.Is(err, app.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("token service returns unexpected error - returns ErrInvalidToken (no leak)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tokenService := mocks.NewMockTokenService(ctrl)
		anyErr := errors.New("jwt parse error")

		tokenService.EXPECT().
			Verify(ctx, token.Token("weird-token")).
			Return(token.Subject(""), anyErr)

		uc := app.NewVerify(tokenService)

		_, err := uc.Execute(ctx, "weird-token")
		if !errors.Is(err, app.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})
}

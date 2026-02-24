package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/soat13/oficina-serverless/internal/identity/app"
	"github.com/soat13/oficina-serverless/internal/identity/app/ports/out/mocks"
	identity "github.com/soat13/oficina-serverless/internal/identity/domain"
	"github.com/soat13/oficina-serverless/internal/shared/cpf"
	"github.com/soat13/oficina-serverless/internal/shared/token"
)

func TestAuthenticateExecute(t *testing.T) {
	ctx := context.Background()
	const ttl int64 = 3600

	t.Run("success - returns access token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		credentialRepository := mocks.NewMockCredentialRepository(ctrl)
		passwordService := mocks.NewMockPasswordService(ctrl)
		tokenService := mocks.NewMockTokenService(ctrl)

		parsedCPF, _ := cpf.Parse("52998224725")

		credentialRepository.EXPECT().
			FindByCPF(ctx, parsedCPF).
			Return(identity.Credential{ID: "user-1", PasswordHash: "hash"}, nil)

		passwordService.EXPECT().
			Compare(ctx, "plain", "hash").
			Return(true, nil)

		tokenService.EXPECT().
			Sign(ctx, token.Subject("user-1"), token.SignOptions{TTLSeconds: ttl}).
			Return(token.Token("access-token"), nil)

		uc := app.NewAuthenticate(credentialRepository, passwordService, tokenService, ttl)

		output, err := uc.Execute(ctx, app.AuthenticateInput{
			CPF:      stringToCPF(t, "529.982.247-25"),
			Password: "plain",
		})

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if output.AccessToken != "access-token" {
			t.Fatalf("expected access-token, got %q", output.AccessToken)
		}
	})

	t.Run("invalid cpf - returns ErrInvalidCredentials", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		credentialRepository := mocks.NewMockCredentialRepository(ctrl)
		passwordService := mocks.NewMockPasswordService(ctrl)
		tokenService := mocks.NewMockTokenService(ctrl)

		uc := app.NewAuthenticate(credentialRepository, passwordService, tokenService, ttl)

		_, err := uc.Execute(ctx, app.AuthenticateInput{
			CPF:      cpf.CPF{},
			Password: "plain",
		})

		if !errors.Is(err, cpf.ErrInvalidCPF) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("repo error - returns ErrInvalidCredentials", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		credentialRepository := mocks.NewMockCredentialRepository(ctrl)
		passwordService := mocks.NewMockPasswordService(ctrl)
		tokenService := mocks.NewMockTokenService(ctrl)

		parsedCPF, _ := cpf.Parse("52998224725")

		credentialRepository.EXPECT().
			FindByCPF(ctx, parsedCPF).
			Return(identity.Credential{}, errors.New("db error"))

		uc := app.NewAuthenticate(credentialRepository, passwordService, tokenService, ttl)

		_, err := uc.Execute(ctx, app.AuthenticateInput{
			CPF:      stringToCPF(t, "529.982.247-25"),
			Password: "plain",
		})
		if !errors.Is(err, app.ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("password mismatch - returns ErrInvalidCredentials", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repository := mocks.NewMockCredentialRepository(ctrl)
		passwordService := mocks.NewMockPasswordService(ctrl)
		tokenService := mocks.NewMockTokenService(ctrl)

		parsedCPF, _ := cpf.Parse("52998224725")

		repository.EXPECT().
			FindByCPF(ctx, parsedCPF).
			Return(identity.Credential{ID: "user-1", PasswordHash: "hash"}, nil)

		passwordService.EXPECT().
			Compare(ctx, "plain", "hash").
			Return(false, nil)

		uc := app.NewAuthenticate(repository, passwordService, tokenService, ttl)

		_, err := uc.Execute(ctx, app.AuthenticateInput{
			CPF:      stringToCPF(t, "529.982.247-25"),
			Password: "plain",
		})
		if !errors.Is(err, app.ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("password service error - returns ErrInvalidCredentials", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		credentialRepository := mocks.NewMockCredentialRepository(ctrl)
		passwordService := mocks.NewMockPasswordService(ctrl)
		tokenService := mocks.NewMockTokenService(ctrl)

		parsedCPF, _ := cpf.Parse("52998224725")

		credentialRepository.EXPECT().
			FindByCPF(ctx, parsedCPF).
			Return(identity.Credential{ID: "user-1", PasswordHash: "hash"}, nil)

		passwordService.EXPECT().
			Compare(ctx, "plain", "hash").
			Return(false, errors.New("hash err"))

		uc := app.NewAuthenticate(credentialRepository, passwordService, tokenService, ttl)

		_, err := uc.Execute(ctx, app.AuthenticateInput{
			CPF:      stringToCPF(t, "529.982.247-25"),
			Password: "plain",
		})
		if !errors.Is(err, app.ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("token sign error - returns original error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		credentialRepository := mocks.NewMockCredentialRepository(ctrl)
		passwordService := mocks.NewMockPasswordService(ctrl)
		tokenService := mocks.NewMockTokenService(ctrl)

		parsedCPF, _ := cpf.Parse("52998224725")
		signErr := errors.New("sign error")

		credentialRepository.EXPECT().
			FindByCPF(ctx, parsedCPF).
			Return(identity.Credential{ID: "user-1", PasswordHash: "hash"}, nil)

		passwordService.EXPECT().
			Compare(ctx, "plain", "hash").
			Return(true, nil)

		tokenService.EXPECT().
			Sign(ctx, token.Subject("user-1"), token.SignOptions{TTLSeconds: ttl}).
			Return(token.Token(""), signErr)

		uc := app.NewAuthenticate(credentialRepository, passwordService, tokenService, ttl)

		_, err := uc.Execute(ctx, app.AuthenticateInput{
			CPF:      stringToCPF(t, "529.982.247-25"),
			Password: "plain",
		})
		if !errors.Is(err, signErr) {
			t.Fatalf("expected %v, got %v", signErr, err)
		}
	})
}

func stringToCPF(t *testing.T, s string) cpf.CPF {
	parsed, err := cpf.Parse(s)
	if err != nil {
		t.Fatalf("failed to parse CPF %q: %v", s, err)
	}
	return parsed
}

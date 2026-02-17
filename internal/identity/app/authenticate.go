package app

import (
	"context"

	"github.com/soat13/oficina-serverless/internal/identity/app/ports/out"
	"github.com/soat13/oficina-serverless/internal/shared/cpf"
	"github.com/soat13/oficina-serverless/internal/shared/token"
)

type Authenticate struct {
	credentialRepository out.CredentialRepository
	passwordService      out.PasswordService
	tokenService         out.TokenService
	tokenTTLSeconds      int64
}

type AuthenticateInput struct {
	CPF      string
	Password string
}

type AuthenticateOutput struct {
	AccessToken token.Token
}

func NewAuthenticate(
	credentialRepository out.CredentialRepository,
	passwordService out.PasswordService,
	tokenService out.TokenService,
	ttl int64,
) Authenticate {
	return Authenticate{
		credentialRepository: credentialRepository,
		passwordService:      passwordService,
		tokenService:         tokenService,
		tokenTTLSeconds:      ttl,
	}
}

func (uc Authenticate) Execute(ctx context.Context, in AuthenticateInput) (AuthenticateOutput, error) {
	parsedCPF, err := cpf.Parse(in.CPF)
	if err != nil {
		return AuthenticateOutput{}, ErrInvalidCredentials
	}
	
	credential, err := uc.credentialRepository.FindByCPF(ctx, parsedCPF)
	if err != nil {
		return AuthenticateOutput{}, ErrInvalidCredentials
	}

	ok, err := uc.passwordService.Compare(ctx, in.Password, credential.PasswordHash)
	if err != nil || !ok {
		return AuthenticateOutput{}, ErrInvalidCredentials
	}

	subject := token.Subject(credential.ID)

	accessToken, err := uc.tokenService.Sign(ctx, subject, token.SignOptions{TTLSeconds: uc.tokenTTLSeconds})
	if err != nil {
		return AuthenticateOutput{}, err
	}

	return AuthenticateOutput{AccessToken: accessToken}, nil
}

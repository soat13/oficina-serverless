package app

import (
	"context"
	"log"

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
	CPF      cpf.CPF
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
	if !in.CPF.IsValid() {
		return AuthenticateOutput{}, cpf.ErrInvalidCPF
	}

	credential, err := uc.credentialRepository.FindByCPF(ctx, in.CPF)

	if err != nil {
		log.Printf("failed to find credential by CPF: %+v", err)
		return AuthenticateOutput{}, ErrInvalidCredentials
	}

	ok, err := uc.passwordService.Compare(ctx, in.Password, credential.PasswordHash)
	if err != nil || !ok {
		if err != nil {
			log.Printf("password comparison failed: %+v", err)
		}
		return AuthenticateOutput{}, ErrInvalidCredentials
	}

	subject := token.Subject(credential.ID)

	accessToken, err := uc.tokenService.Sign(ctx, subject, token.SignOptions{TTLSeconds: uc.tokenTTLSeconds})
	if err != nil {
		log.Printf("failed to sign token: %+v", err)
		return AuthenticateOutput{}, err
	}

	return AuthenticateOutput{AccessToken: accessToken}, nil
}

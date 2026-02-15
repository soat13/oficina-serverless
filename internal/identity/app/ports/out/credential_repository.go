package out

import (
	"context"
	"errors"

	"github.com/soat13/oficina-serverless/internal/identity/domain"
	"github.com/soat13/oficina-serverless/internal/shared/cpf"
)

var (
	ErrCredentialNotFound = errors.New("credential not found")
)

type CredentialRepository interface {
	FindByCPF(ctx context.Context, cpf cpf.CPF) (domain.Credential, error)
}

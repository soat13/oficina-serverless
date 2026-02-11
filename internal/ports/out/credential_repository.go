package out

import (
	"context"
	"errors"

	"github.com/soat13/oficina-serverless/internal/domain/identity"
	"github.com/soat13/oficina-serverless/internal/domain/shared/cpf"
)

var (
	ErrCredentialNotFound = errors.New("credential not found")
)

type CredentialRepository interface {
	FindByCPF(ctx context.Context, cpf cpf.CPF) (identity.Credential, error)
}

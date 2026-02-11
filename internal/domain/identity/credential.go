package identity

import (
	"github.com/soat13/oficina-serverless/internal/domain/shared/cpf"
)

type Credential struct {
	ID           string
	CPF          cpf.CPF
	PasswordHash string
}

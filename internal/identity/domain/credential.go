package domain

import (
	"github.com/soat13/oficina-serverless/internal/shared/cpf"
)

type Credential struct {
	ID           string
	CPF          cpf.CPF
	PasswordHash string
}

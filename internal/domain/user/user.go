package user

import (
	"github.com/soat13/oficina-serveless/internal/domain/shared/cpf"
)

type User struct {
	ID           string
	CPF          cpf.CPF
	PasswordHash string
}

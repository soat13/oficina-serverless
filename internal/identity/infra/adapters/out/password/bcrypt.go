package password

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/soat13/oficina-serverless/internal/identity/app/ports/out"
)

type BCryptService struct{}

func New() out.PasswordService {
	return &BCryptService{}
}

func (s *BCryptService) Compare(_ context.Context, plain string, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))

	if err == nil {
		return true, nil
	}

	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}

	return false, err
}

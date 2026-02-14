package out

import "context"

type PasswordService interface {
	Compare(ctx context.Context, plain string, hash string) (bool, error)
}

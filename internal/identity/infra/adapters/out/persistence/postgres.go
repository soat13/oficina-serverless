package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"

	"github.com/soat13/oficina-serverless/internal/identity/app/ports/out"
	"github.com/soat13/oficina-serverless/internal/identity/domain"
	"github.com/soat13/oficina-serverless/internal/shared/cpf"
)

type PostgresCredentialRepository struct {
	db *bun.DB
}

func New(db *bun.DB) out.CredentialRepository {
	return &PostgresCredentialRepository{db: db}
}

func (r *PostgresCredentialRepository) FindByCPF(ctx context.Context, c cpf.CPF) (domain.Credential, error) {
	var row credentialRow

	err := r.db.NewSelect().
		Model(&row).
		Where("document = ?", c.String()).
		Limit(1).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Credential{}, out.ErrCredentialNotFound
		}
		return domain.Credential{}, err
	}

	parsedCPF, _ := cpf.Parse(row.CPF)

	return domain.Credential{
		ID:           row.ID,
		CPF:          parsedCPF,
		PasswordHash: row.Password,
	}, nil
}

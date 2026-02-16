package container

import (
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	"github.com/soat13/oficina-serverless/internal/identity/infra/config"
)

type Container struct {
	Cfg   config.Config
	SQLDB *sql.DB
	DB    *bun.DB
}

func New(cfg config.Config) (*Container, error) {
	if !cfg.Valid() {
		return nil, fmt.Errorf("invalid configuration")
	}

	sqlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(cfg.DBDSN)))
	db := bun.NewDB(sqlDB, pgdialect.New())

	return &Container{
		Cfg:   cfg,
		SQLDB: sqlDB,
		DB:    db,
	}, nil
}

func (c *Container) Close() error {
	return c.SQLDB.Close()
}

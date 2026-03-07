package container

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/soat13/oficina-serverless/internal/config"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type Container struct {
	Cfg   config.Config
	SQLDB *sql.DB
	DB    *bun.DB
}

func New(cfg config.Config) (*Container, error) {
	sqlDB := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(cfg.DBDSN),
	))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

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

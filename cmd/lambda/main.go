package main

import (
	"context"
	"database/sql"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	"github.com/soat13/oficina-serverless/internal/identity/app"
	"github.com/soat13/oficina-serverless/internal/identity/infra/adapters/in/apigw"
	"github.com/soat13/oficina-serverless/internal/identity/infra/adapters/out/password"
	"github.com/soat13/oficina-serverless/internal/identity/infra/adapters/out/persistence"
	tokenadapter "github.com/soat13/oficina-serverless/internal/identity/infra/adapters/out/token"
	"github.com/soat13/oficina-serverless/internal/identity/infra/config"
)

func main() {
	cfg := config.Load()
	if !cfg.Valid() {
		log.Fatal("invalid configuration: JWT_SECRET, JWT_ISSUER, JWT_TTL and DBDSN are required")
	}

	openDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(cfg.DBDSN)))
	defer func() { _ = openDB.Close() }()

	credentialRepository := persistence.New(bun.NewDB(openDB, pgdialect.New()))
	jwtService := tokenadapter.New(cfg.JWTSecret, cfg.JWTIssuer)
	passwordService := password.New()

	httpHandler := apigw.NewHandler(
		app.NewAuthenticate(credentialRepository, passwordService, jwtService, cfg.JWTTTL),
		app.NewVerify(jwtService),
		cfg.JWTTTL,
	)

	lambda.Start(func(ctx context.Context, req any) (any, error) {
		lambda.Start(httpHandler.Handle)
		return httpHandler, nil
	})
}

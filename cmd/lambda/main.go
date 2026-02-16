package main

import (
	"log"

	"github.com/joho/godotenv"
	tokenadapter "github.com/soat13/oficina-serverless/internal/identity/infra/adapters/out/token"
	"github.com/soat13/oficina-serverless/internal/identity/infra/config"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	if !cfg.Valid() {
		log.Fatal("invalid configuration: JWT_SECRET, JWT_ISSUER, JWT_TTL are required")
	}

	_ = tokenadapter.New(cfg.JWTSecret, cfg.JWTIssuer)
}

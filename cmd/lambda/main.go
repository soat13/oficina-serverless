package main

import (
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	identitybootstrap "github.com/soat13/oficina-serverless/internal/bootstrap/identity"
	"github.com/soat13/oficina-serverless/internal/container"
	"github.com/soat13/oficina-serverless/internal/identity/infra/config"
)

func main() {
	cfg := config.Load()

	c, err := container.New(cfg)
	if err != nil {
		log.Printf("container init error: %v", err)
		panic(err)
	}
	defer func() { _ = c.Close() }()

	mod, err := identitybootstrap.Setup(c)
	if err != nil {
		log.Printf("bootstrap setup error: %v", err)
		panic(err)
	}

	lambda.Start(mod.ApiGwHandler.Handle)
}

package main

import (
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/soat13/oficina-serverless/internal/config"

	identitybootstrap "github.com/soat13/oficina-serverless/internal/bootstrap/identity"
	"github.com/soat13/oficina-serverless/internal/container"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	c, err := container.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = c.Close() }()

	mod, err := identitybootstrap.Setup(c)
	if err != nil {
		log.Fatal(err)
	}

	lambda.Start(mod.ApiGwHandler.Handle)
}

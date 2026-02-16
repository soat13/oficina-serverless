package identity

import (
	"fmt"

	"github.com/soat13/oficina-serverless/internal/container"
	"github.com/soat13/oficina-serverless/internal/identity/app"
	"github.com/soat13/oficina-serverless/internal/identity/infra/adapters/in/apigw"
	passwordadapter "github.com/soat13/oficina-serverless/internal/identity/infra/adapters/out/password"
	"github.com/soat13/oficina-serverless/internal/identity/infra/adapters/out/persistence"
	tokenadapter "github.com/soat13/oficina-serverless/internal/identity/infra/adapters/out/token"
)

type Module struct {
	ApiGwHandler *apigw.Handler
}

func Setup(container *container.Container) (*Module, error) {
	if container == nil || container.DB == nil {
		return nil, fmt.Errorf("container.DB is required")
	}

	credentialRepository := persistence.New(container.DB)
	passwordService := passwordadapter.New()
	jwtService := tokenadapter.New(container.Cfg.JWTSecret, container.Cfg.JWTSecret)

	h := apigw.NewHandler(
		app.NewAuthenticate(credentialRepository, passwordService, jwtService, container.Cfg.JWTTTL),
		app.NewVerify(jwtService),
	)

	return &Module{ApiGwHandler: h}, nil
}

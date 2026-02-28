# Identity Token Lambda

## Sobre o Projeto

Função **AWS Lambda** responsável por realizar a autenticação de
usuários através de **CPF e senha**, gerando um **JWT** válido para
acesso às APIs protegidas do sistema.

Essa Lambda atua como ponto de entrada de autenticação, sendo invocada
pelo API Gateway, e executa as seguintes etapas:

1.  Validação do formato e dígitos verificadores do CPF
2.  Verificação se o CPF está cadastrado no banco
3.  Validação da senha informada (bcrypt)
4.  Geração de JWT assinado (HS256)
5.  Retorno do token para o client

> A Lambda é exposta através de uma **API Gateway existente** (já provisionada).  
> O endpoint `/auth/token` é roteado para esta função via integração do API Gateway.
------------------------------------------------------------------------

## Fluxo Simplificado

```text
API Gateway
     │
     ▼
Identity Lambda
     │
     ├─ Valida CPF
     ├─ Consulta PostgreSQL
     ├─ Valida senha (bcrypt)
     ├─ Gera JWT (HS256)
     ▼
Response 200
```

------------------------------------------------------------------------

## Endpoint

POST /auth/token

### Request

``` json
{
  "cpf": "12345678909",
  "password": "minha-senha-123"
}
```

### Response (Sucesso)

``` json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer"
}
```

------------------------------------------------------------------------

## Regras de Negócio

-   CPF inválido → 400 Bad Request
-   CPF não encontrado → 401 Unauthorized
-   Senha incorreta → 401 Unauthorized
-   Sucesso → 200 OK com token JWT

------------------------------------------------------------------------

## Arquitetura

A função segue princípios de Arquitetura Hexagonal (Ports & Adapters):

-   Domain
    -   Validação de CPF
-   Application
    -   Caso de uso: GenerateToken
    -   Interface de repositório de usuário
    -   Interface de serviço JWT
-   Infrastructure
    -   Adapter PostgreSQL
    -   Adapter JWT (HS256)
    -   Adapter API Gateway

A Lambda é stateless.

------------------------------------------------------------------------

## Stack Tecnológica

-   Linguagem: Go 1.26+
-   Runtime: provided.al2 (AWS Lambda custom runtime)
-   Banco: PostgreSQL 17.6+
-   ORM: Bun
-   Hash de senha: bcrypt
-   Autenticação: JWT (HS256)
-   Infraestrutura: AWS Lambda + API Gateway

------------------------------------------------------------------------

## Variáveis de Ambiente Necessárias

### Variáveis da Aplicação (Runtime da Lambda)

- JWT_SECRET → chave utilizada para assinatura do token
- JWT_ISSUER → identificador do emissor do JWT
- JWT_TTL → tempo de expiração do token (em segundos)
- DB_DSN → string de conexão com o PostgreSQL

### Variáveis da Pipeline (GitHub Actions / Deploy)
Estas variáveis são utilizadas apenas no processo de deploy via GitHub Actions, não fazem parte da execução da aplicação:

- AWS_REGION
- AWS_ACCESS_KEY_ID
- AWS_SECRET_ACCESS_KEY
- AWS_SESSION_TOKEN
- LAMBDA_ROLE_ARN
- EXISTING_REST_API_ID
- VPC_ID
- VPC_SUBNET_IDS
- RDS_SECURITY_GROUP_ID

------------------------------------------------------------------------

## Segurança

-   Senhas armazenadas com bcrypt
-   JWT assinado com HS256
-   Tokens possuem expiração configurável
-   CPF validado antes da consulta ao banco

------------------------------------------------------------------------

## Executando Localmente

### Requisitos

-   Go instalado
-   Docker (opcional)
-   PostgreSQL

### Rodando

go run ./cmd/lambda

Ou utilizando SAM:

sam build\
sam local start-api

------------------------------------------------------------------------

## Testes

go test ./...

------------------------------------------------------------------------

## Evoluções Futuras

- Refresh token
- MFA
- Rate limit
- Instrospecção de token
- Suporte a outros métodos de autenticação

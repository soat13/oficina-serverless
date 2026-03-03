# Identity Token Lambda

## Sobre o Projeto

Função **AWS Lambda** responsável por autenticar usuários via **CPF + senha**, gerando um **JWT (HS256)** válido para acesso às APIs protegidas do sistema.

Esta função:

- É **stateless**
- Se integra a um **API Gateway REST já existente**
- Segue **Arquitetura Hexagonal (Ports & Adapters)**
- Utiliza PostgreSQL como fonte de identidade

---

# Responsabilidade

A Lambda é responsável exclusivamente por:

1. Validar CPF (formato + dígitos verificadores)
2. Verificar se o usuário existe no banco
3. Validar senha com bcrypt
4. Gerar JWT assinado (HS256)
5. Retornar token para o cliente

Ela **não realiza autorização**, apenas autenticação.

---

# Integração

A Lambda é integrada a um **API Gateway REST existente**, via integração Lambda Proxy.

Endpoint exposto:

POST `/auth/token`

---

# Fluxo Simplificado
```text
Client
│
▼
API Gateway (REST)
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


---

# Endpoint

## POST `/auth/token`

Request

```json
{
  "cpf": "12345678909",
  "password": "minha-senha-123"
}
```
Response - Http status: 200 (OK)
```json
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvbiBEb2UiLCJpYXQiOjE1MTYyMzkwMjJ9.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
    "type": "Bearer"
}
```
Response - Http status: 400 (Bad Request)
```json
{
    "error": "cpf_and_password_required" // others errors
}
```

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

| Variável | Onde Obter | Caminho no Console AWS               | Exemplo |
|-----------|------------|--------------------------------------|----------|
| EXISTING_REST_API_ID | API Gateway | API Gateway → REST APIs → Selecionar API | `a1b2c3d4e5` |
| LAMBDA_ROLE_ARN | IAM | IAM → Roles                          | `arn:aws:iam::123456789012:role/lambda-identity-role` |
| VPC_ID | VPC | VPC → Your VPCs                      | `vpc-0abc1234` |
| VPC_SUBNET_IDS | VPC | VPC → Subnets (privadas)             | `subnet-12345`, `subnet-67890` |
| RDS_SECURITY_GROUP_ID | EC2 | EC2 → Security Groups                | `sg-0123abc456` |
| DB_DSN | RDS | RDS → Databases → Endpoint           | `postgres://user:password@endpoint:5432/database?sslmode=require` |

---
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

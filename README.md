# Identity Token Lambda

## Sumário

- [Contexto](#contexto)
- [Stack Tecnológica](#stack-tecnológica)
- [Responsabilidade](#responsabilidade)
- [Integração](#integração)
- [Arquitetura](#arquitetura)
- [API](#api)
- [Execução](#execução)
- [Operação](#operação)
- [Evoluções Futuras](#evoluções-futuras)

---

## Contexto

Este serviço é uma **AWS Lambda** responsável pela autenticação de usuários utilizando **CPF + senha**.

Quando a autenticação é bem-sucedida, a função gera um **JWT (HS256)** que pode ser utilizado para acessar APIs protegidas do sistema.

Características principais:

- Serviço **stateless**
- Integrado a um **API Gateway REST**
- Implementado seguindo **Arquitetura Hexagonal (Ports & Adapters)**
- Utiliza **PostgreSQL** como fonte de identidade

---

## Stack Tecnológica

- Linguagem: Go 1.26+
- Runtime: AWS Lambda (`provided.al2`)
- Banco: PostgreSQL 17.6+
- ORM: Bun
- Hash de senha: bcrypt
- Autenticação: JWT (HS256)
- Infraestrutura: AWS Lambda + API Gateway


---

## Responsabilidade

A Lambda é responsável exclusivamente por:

1. Validar CPF (formato + dígitos verificadores)
2. Verificar se o usuário existe no banco
3. Validar senha com bcrypt
4. Gerar JWT assinado (HS256)
5. Retornar token para o cliente

Ela **não realiza autorização**, apenas autenticação.

---

## Integração

A Lambda é integrada a um **API Gateway REST existente**, via integração **Lambda Proxy**.

---

## Fluxo Simplificado
```text
Client
│
│ POST /auth/login
▼
API Gateway (REST)
│
│ Lambda Proxy Integration
▼
Identity Token Lambda
│
├─ Parse do payload
├─ Validação de CPF
├─ Consulta de credenciais no PostgreSQL
├─ Verificação da senha (bcrypt)
├─ Geração de JWT (HS256)
▼
Response 200 (JWT Token)
```


---
## API

### Endpoint

Este serviço expõe uma única operação HTTP através de **AWS API Gateway (Lambda Proxy Integration)**.

A requisição recebida pela Lambda é um evento `APIGatewayProxyRequest`.

### POST `/auth/login`

Autentica um usuário utilizando **CPF + senha** e retorna um **JWT Bearer Token** utilizado para autenticação nas APIs protegidas.

### Request

```json
{
  "cpf": "12345678909",
  "password": "minha-senha-123"
}
```
### Response — 200 OK
```json
{
    "token": "...",
    "type": "Bearer"
}
```
### Response — 400 Bad Request
```json
{
    "error": "cpf_and_password_required"
}
```
Outros códigos de erro podem ser retornados dependendo da validação realizada.

### Regras de Negócio

-   CPF inválido → 400 Bad Request
-   CPF não encontrado → 401 Unauthorized
-   Senha incorreta → 401 Unauthorized
-   Sucesso → 200 OK com token JWT

---

## Arquitetura

Este serviço segue os princípios de **Arquitetura Hexagonal (Ports & Adapters)**.

Para uma explicação detalhada da arquitetura, incluindo fluxos e organização interna do serviço, consulte: **[Architecture Documentation](docs/architecture.md)**

Resumo das camadas:

- **Domain**
    - Validação de CPF
    - Tipos e erros de domínio

- **Application**
    - Caso de uso `Authenticate`
    - Interfaces (Ports)

- **Infrastructure**
    - PostgreSQL Repository
    - JWT Service
    - Lambda Handler
---
## Execução

### Variáveis de Ambiente Necessárias

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
## Operação

### Segurança

-   Senhas armazenadas com bcrypt
-   JWT assinado com HS256
-   Tokens possuem expiração configurável
-   CPF validado antes da consulta ao banco

---

### Executando Localmente

### Requisitos

- Go instalado
- Docker e Docker Compose (para subir o PostgreSQL usado nos testes)

### Rodando

```bash
go run ./cmd/lambda
```

Ou utilizando SAM:
```bash
sam build
sam local start-api
```

---

### Testes

Os testes dependem de um banco PostgreSQL que é iniciado via Docker Compose.
Antes de executar os testes, suba o banco:

```bash
docker compose up -d
go test ./...
```

---

## Evoluções Futuras

- Refresh token
- MFA
- Rate limit
- Introspecção de token
- Suporte a outros métodos de autenticação

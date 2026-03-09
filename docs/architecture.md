# Arquitetura

## Visão Geral

O serviço **Identity Token Lambda** é uma função AWS Lambda responsável por autenticar usuários via **CPF + senha** e retornar um **JWT (HS256)** para consumo das APIs protegidas do sistema.

A solução combina:

- **AWS Serverless**
- **Arquitetura Hexagonal (Ports & Adapters)**
- **DDD tático simplificado**, com foco em isolamento das regras de negócio e baixo acoplamento

A Lambda é responsável apenas por **autenticação**.  
Ela **não realiza autorização**, controle de permissões ou refresh token.

---

## Arquitetura Serverless

A visão de infraestrutura do serviço:

```text
Client
│
│ POST /auth/login
▼
API Gateway (REST)
│
│ Lambda Proxy Integration
▼
AWS Lambda
(identity-token)
│
├─ Valida payload
├─ Valida CPF
├─ Busca credenciais
├─ Valida senha
├─ Gera JWT
▼
PostgreSQL
```

### Fluxo de Autenticação

1. O cliente envia uma requisição `POST /auth/login`
2. O **API Gateway REST** encaminha a requisição para a Lambda
3. O **Lambda Handler** faz parsing do payload
4. O caso de uso **Authenticate** é executado
5. O repositório consulta as credenciais no **PostgreSQL**
6. A senha informada é comparada com o hash usando **bcrypt**
7. Um **JWT assinado (HS256)** é gerado
8. O token é retornado ao cliente

## Arquitetura da Aplicação

Internamente, o serviço segue **Arquitetura Hexagonal (Ports & Adapters)**, separando regras de negócio, casos de uso e detalhes de infraestrutura.

```text
┌──────────────────────────────────────────────────────────────────────┐
│                           EXTERNAL WORLD                             │
│                                                                      │
│  Client  →  API Gateway REST  →  Lambda Handler                      │
└───────────────────────────────┬──────────────────────────────────────┘
                                │
                                │ calls
                                ▼
┌──────────────────────────────────────────────────────────────────────┐
│                        INBOUND ADAPTER                               │
│                                                                      │
│  Lambda / HTTP Handler                                               │
│  - recebe evento do API Gateway                                      │
│  - faz parsing/validação estrutural do payload                       │
│  - chama o caso de uso                                               │
│  - traduz resposta/erro para HTTP                                    │
└───────────────────────────────┬──────────────────────────────────────┘
                                │
                                │ uses inbound port
                                ▼
┌──────────────────────────────────────────────────────────────────────┐
│                     APPLICATION + DOMAIN CORE                        │
│                                                                      │
│  Inbound Port: AuthenticateUseCase                                   │
│                                                                      │
│  Use Case: Authenticate                                              │
│  - valida regras de entrada                                          │
│  - coordena autenticação                                             │
│  - orquestra portas de saída                                         │
│                                                                      │
│  Domain                                                              │
│  - CPF                                                               │
│  - Credential                                                        │
│  - erros de domínio                                                  │
│  - regras puras                                                      │
│                                                                      │
│  Outbound Ports:                                                     │
│  - CredentialRepository                                              │
│  - PasswordService                                                   │
│  - TokenService                                                      │
└───────────────────────────────┬──────────────────────────────────────┘
                                │
                                │ implemented by
                                ▼
┌──────────────────────────────────────────────────────────────────────┐
│                       OUTBOUND ADAPTERS                              │
│                                                                      │
│  PostgreSQL Credential Repository                                    │
│  Bcrypt Password Service                                             │
│  JWT HS256 Token Service                                             │
└───────────────────────────────┬──────────────────────────────────────┘
                                │
                                ▼
┌──────────────────────────────────────────────────────────────────────┐
│                    INFRASTRUCTURE / RESOURCES                        │
│                                                                      │
│  PostgreSQL                                                          │
│  Secrets / Env Vars                                                  │
│  AWS Lambda Runtime                                                  │
└──────────────────────────────────────────────────────────────────────┘
```





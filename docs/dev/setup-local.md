# Setup do Ambiente de Desenvolvimento

## Pré-requisitos

| Ferramenta | Versão mínima | Instalação |
|---|---|---|
| Go | 1.22+ | https://go.dev/dl/ |
| Node.js | 20+ | https://nodejs.org |
| pnpm | 9+ | `npm i -g pnpm` |
| Docker + Compose | Docker 24+ | https://docs.docker.com/get-docker/ |
| Git | 2.40+ | https://git-scm.com |
| Make | qualquer | pré-instalado no Linux/macOS |

---

## Clone e Configuração Inicial

```bash
# Clonar o repositório
git clone https://github.com/astrolink/astrolink.git
cd astrolink

# Copiar variáveis de ambiente
cp .env.example .env
# Edite o .env com suas configurações (veja seção abaixo)

# Instalar todas as dependências (Go + Node)
make install

# Subir infraestrutura de desenvolvimento
make dev-infra
# Isso sobe: PostgreSQL, Redis, RabbitMQ, pgAdmin

# Aplicar migrations do banco
make migrate

# Iniciar todos os serviços em modo dev
make dev
```

---

## Arquivo `.env` para Desenvolvimento

```bash
# .env
GO_ENV=development
LOG_LEVEL=debug

# Banco (desenvolvimento — sem senha forte)
DB_PASSWORD=devpassword
DATABASE_URL=postgres://astrolink:devpassword@localhost:5432/astrolink

# Redis
REDIS_PASSWORD=devredis
REDIS_URL=redis://:devredis@localhost:6379

# RabbitMQ
RABBITMQ_USER=astrolink
RABBITMQ_PASS=devrabbit
AMQP_URL=amqp://astrolink:devrabbit@localhost:5672/

# JWT (para dev, qualquer string serve)
JWT_SECRET=dev-jwt-secret-nao-usar-em-producao-32chars

# Mercado Pago (usar Sandbox para testes)
MP_ACCESS_TOKEN=TEST-XXXX-XXXX-XXXX
MP_PUBLIC_KEY=TEST-XXXX-XXXX-XXXX
MP_WEBHOOK_SECRET=test-webhook-secret

# Admin padrão
ADMIN_USUARIO=admin
ADMIN_SENHA=admin123

# Node identification
NODE_NAME=dev-node-01
TIMEZONE=America/Sao_Paulo

# OpenNDS/OpenWrt (desabilitado por padrao no desenvolvimento local)
OPENNDS_ENABLED=false
OPENNDS_SSH_HOST=192.168.1.1
OPENNDS_SSH_PORT=22
OPENNDS_SSH_USER=root
OPENNDS_SSH_KEY_PATH=C:\Users\charl\.ssh\id_ed25519
OPENNDS_SSH_TIMEOUT=10s
OPENNDS_AUTH_RETRIES=3
```

---

## Estrutura do Projeto

```
astrolink/
├── node/                    # Backend Go
│   ├── cmd/
│   │   └── server/
│   │       └── main.go      # Entry point
│   ├── internal/
│   │   ├── api/             # HTTP handlers (Fiber)
│   │   │   ├── portal/      # Rotas públicas do portal
│   │   │   ├── admin/       # Rotas admin (JWT)
│   │   │   └── webhooks/    # Webhooks (MP, etc.)
│   │   ├── domain/          # Lógica de negócio pura
│   │   │   ├── planos/
│   │   │   ├── vouchers/
│   │   │   ├── sessoes/
│   │   │   └── pagamentos/
│   │   ├── infra/           # Implementações externas
│   │   │   ├── db/          # SQLC queries
│   │   │   ├── redis/
│   │   │   └── amqp/
│   │   ├── network/         # SSH, ndsctl, tc
│   │   ├── scheduler/       # Jobs agendados
│   │   └── sync/            # Agente Cloud sync
│   ├── migrations/          # SQL migrations (golang-migrate)
│   ├── sqlc/
│   │   ├── schema.sql       # Schema completo
│   │   ├── query.sql        # Queries SQLC
│   │   └── sqlc.yaml        # Configuração SQLC
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
│
├── portal/                  # Portal cativo (SvelteKit)
│   ├── src/
│   │   ├── routes/          # Páginas (SvelteKit file-based routing)
│   │   │   ├── +page.svelte          # Boas-vindas
│   │   │   ├── planos/+page.svelte   # Seleção de plano
│   │   │   ├── pix/+page.svelte      # Pagamento PIX
│   │   │   └── sucesso/+page.svelte  # Acesso liberado
│   │   ├── lib/
│   │   │   ├── api.ts       # Cliente da API
│   │   │   ├── store.ts     # Estado global (Svelte stores)
│   │   │   └── components/  # Componentes reutilizáveis
│   │   └── app.css          # TailwindCSS base
│   ├── package.json
│   ├── svelte.config.js
│   └── vite.config.ts
│
├── admin/                   # Admin local (SvelteKit)
│   └── ...
│
├── cloud/                   # Painel cloud (SvelteKit + Supabase)
│   └── ...
│
├── cli/                     # CLI (Go)
│   └── ...
│
├── docs/                    # Esta documentação
├── docker-compose.dev.yml   # Infra de desenvolvimento
├── docker-compose.yml       # Produção
├── Makefile
└── .env.example
```

---

## Comandos Make

```makefile
# Makefile

.PHONY: install dev dev-infra stop migrate test build lint clean

# Instalar dependências
install:
	cd node && go mod download
	cd portal && pnpm install
	cd admin && pnpm install
	cd cloud && pnpm install

# Subir infraestrutura de dev (DB, Redis, RabbitMQ)
dev-infra:
	docker compose -f docker-compose.dev.yml up -d
	@echo "Aguardando PostgreSQL..."
	@until docker compose -f docker-compose.dev.yml exec postgres pg_isready -U astrolink 2>/dev/null; do sleep 1; done
	@echo "✅ Infra pronta!"

# Iniciar todos em modo dev (hot reload)
dev:
	@make dev-infra
	@make migrate
	@echo "Iniciando serviços..."
	# Usar concurrently ou tmux
	cd node && air &
	cd portal && pnpm dev &
	cd admin && pnpm dev --port 5001 &

# Parar infra
stop:
	docker compose -f docker-compose.dev.yml down

# Rodar migrations
migrate:
	cd node && go run cmd/migrate/main.go up

# Criar nova migration
migration NAME?=nova_migration:
	cd node && golang-migrate create -ext sql -dir migrations -seq $(NAME)

# Testes
test:
	cd node && go test ./... -v -race
	cd portal && pnpm test
	cd admin && pnpm test

# Testes com cobertura
test-coverage:
	cd node && go test ./... -coverprofile=coverage.out
	cd node && go tool cover -html=coverage.out

# Lint
lint:
	cd node && golangci-lint run
	cd portal && pnpm lint
	cd admin && pnpm lint

# Build de produção
build:
	cd node && CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/astrolink-node ./cmd/server
	cd portal && pnpm build
	cd admin && pnpm build

# Build para ARM64 (Raspberry Pi, Orange Pi)
build-arm64:
	cd node && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build \
	  -ldflags="-s -w -X main.Version=$(VERSION)" \
	  -o dist/astrolink-node-arm64 ./cmd/server

# Gerar código SQLC
sqlc:
	cd node && sqlc generate

# Limpar builds
clean:
	rm -rf node/dist portal/.svelte-kit portal/build admin/.svelte-kit admin/build
```

---

## Docker Compose de Desenvolvimento

```yaml
# docker-compose.dev.yml
version: '3.9'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: astrolink
      POSTGRES_USER: astrolink
      POSTGRES_PASSWORD: devpassword
    ports:
      - "5432:5432"
    volumes:
      - postgres_dev:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass devredis
    ports:
      - "6379:6379"

  rabbitmq:
    image: rabbitmq:3-management-alpine
    environment:
      RABBITMQ_DEFAULT_USER: astrolink
      RABBITMQ_DEFAULT_PASS: devrabbit
    ports:
      - "5672:5672"
      - "15672:15672"  # Management UI: http://localhost:15672

  pgadmin:
    image: dpage/pgadmin4:latest
    environment:
      PGADMIN_DEFAULT_EMAIL: dev@astrolink.app
      PGADMIN_DEFAULT_PASSWORD: devpassword
    ports:
      - "5050:80"     # pgAdmin UI: http://localhost:5050

volumes:
  postgres_dev:
```

---

## Hot Reload (Go)

Usando `air` para hot reload do backend Go:

```bash
# Instalar air
go install github.com/cosmtrek/air@latest

# Rodar (usa .air.toml se existir)
cd node && air
```

```toml
# node/.air.toml
[build]
  cmd = "go build -o ./tmp/main ./cmd/server"
  bin = "./tmp/main"
  include_ext = ["go", "tpl", "tmpl", "html"]
  exclude_dir = ["tmp", "vendor", "testdata"]
  delay = 300

[log]
  time = true

[color]
  main = "magenta"
  watcher = "cyan"
  build = "yellow"
  runner = "green"
```

---

## Configuração SQLC

```yaml
# node/sqlc/sqlc.yaml
version: "2"

sql:
  - engine: "postgresql"
    queries: "sqlc/query.sql"
    schema: "migrations/"
    gen:
      go:
        package: "db"
        out: "internal/infra/db"
        emit_json_tags: true
        emit_interface: true
        emit_prepared_queries: false
        emit_exact_table_names: false
```

---

## Testando o Portal Cativo em Desenvolvimento

O portal cativo espera parâmetros de MAC/IP na URL (injetados pelo OpenNDS em produção). Em desenvolvimento, use:

```
http://localhost:5000/?mac=AA:BB:CC:DD:EE:FF&ip=192.168.1.50&token=demo_token
```

O backend detecta `GO_ENV=development` e aceita qualquer MAC/token sem validação real.

---

## Extensões VSCode Recomendadas

```json
// .vscode/extensions.json
{
  "recommendations": [
    "golang.go",
    "svelte.svelte-vscode",
    "bradlc.vscode-tailwindcss",
    "ms-azuretools.vscode-docker",
    "mtxr.sqltools",
    "mtxr.sqltools-driver-pg",
    "eamodio.gitlens",
    "usernamehw.errorlens",
    "biomejs.biome"
  ]
}
```

```json
// .vscode/settings.json
{
  "editor.formatOnSave": true,
  "editor.defaultFormatter": "biomejs.biome",
  "[go]": {
    "editor.defaultFormatter": "golang.go"
  },
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "tailwindCSS.experimental.classRegex": [
    ["class\\s*=\\s*['\"`]([^'\"\\`]*)['\"`]"]
  ]
}
```

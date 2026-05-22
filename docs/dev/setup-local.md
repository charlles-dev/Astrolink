# Setup Local

## Pre-requisitos

| Ferramenta | Versao minima |
|---|---|
| Go | 1.22+ |
| Node.js | 20+ |
| npm | incluido no Node |
| Docker + Compose | Docker 24+ |
| Git | 2.40+ |
| Make | disponivel no terminal |

No Windows, este projeto usa comandos pensados para PowerShell.

## Primeiro Setup

```powershell
Copy-Item .env.example .env
make install
make dev-infra
make test
```

`make install` baixa dependencias Go e instala dependencias do portal com
`npm install`.

`make dev-infra` sobe:

- Postgres
- Redis
- RabbitMQ
- pgAdmin

## Rodando em Desenvolvimento

Abra dois terminais:

```powershell
make dev-node
```

```powershell
make dev-portal
```

URLs:

- Backend: `http://localhost:5000`
- Health: `http://localhost:5000/api/saude`
- Portal: `http://127.0.0.1:5173/?mac=AA:BB:CC:DD:EE:FF&ip=192.168.1.50&token=test`
- Painel local: `http://127.0.0.1:5173/painel`

## Variaveis de Ambiente

Use `.env.example` como base:

```env
GO_ENV=development
HTTP_ADDR=:5000
LOG_LEVEL=debug

DB_PASSWORD=devpassword
DATABASE_URL=postgres://astrolink:devpassword@localhost:5432/astrolink?sslmode=disable

REDIS_PASSWORD=devredis
REDIS_URL=redis://:devredis@localhost:6379

RABBITMQ_USER=astrolink
RABBITMQ_PASS=devrabbit
AMQP_URL=amqp://astrolink:devrabbit@localhost:5672/

JWT_SECRET=dev-jwt-secret-nao-usar-em-producao-32chars

PAYMENTS_PROVIDER=demo
MERCADOPAGO_ACCESS_TOKEN=TEST-XXXX-XXXX-XXXX
MERCADOPAGO_API_BASE_URL=
MERCADOPAGO_WEBHOOK_SECRET=test-webhook-secret
MP_PUBLIC_KEY=TEST-XXXX-XXXX-XXXX

ADMIN_USUARIO=admin
ADMIN_SENHA=admin123

NODE_NAME=dev-node-01
TIMEZONE=America/Sao_Paulo

OPENNDS_ENABLED=false
OPENNDS_SSH_HOST=192.168.1.1
OPENNDS_SSH_PORT=22
OPENNDS_SSH_USER=root
OPENNDS_SSH_KEY_PATH=C:\Users\charl\.ssh\id_ed25519
OPENNDS_SSH_TIMEOUT=10s
OPENNDS_AUTH_RETRIES=3
```

## Estrutura Atual

```text
astrolink/
  node/                    backend Go
    cmd/server/            entrada da aplicacao
    internal/api/          rotas HTTP
    internal/config/       configuracao por env
    internal/domain/       regras de negocio
    internal/gateway/      OpenNDS/SSH/no-op
    internal/infra/        memory e postgres stores
    migrations/            schema SQL

  portal/                  portal cativo SvelteKit
    src/routes/+page.svelte
    src/lib/api.ts
    src/lib/components/

  docs/                    documentacao viva
  docker-compose.dev.yml   infra local
  docker-compose.yml       compose base de producao local
  Makefile
  .env.example
```

## Comandos Make

```powershell
make install       # dependencias Go + portal npm
make dev-infra     # sobe Postgres/Redis/RabbitMQ/pgAdmin
make dev-node      # roda backend Go
make dev-portal    # roda portal em 5173
make test          # Go + Vitest
make check         # svelte-check
make build         # build backend + portal
make clean         # remove builds locais
```

## Banco Local

O Postgres dev e inicializado com as migrations de `node/migrations/`.

Para webhook Mercado Pago real, altere `PAYMENTS_PROVIDER=mercadopago` e
configure `MERCADOPAGO_ACCESS_TOKEN` junto com `MERCADOPAGO_WEBHOOK_SECRET`.

Para reiniciar a infra:

```powershell
make stop
make dev-infra
```

## Portal Cativo em Dev

Use sempre a URL com parametros simulados:

```text
http://127.0.0.1:5173/?mac=AA:BB:CC:DD:EE:FF&ip=192.168.1.50&token=test
```

Voucher demo em modo memoria:

```text
TEST-1234
```

## OpenNDS

No desenvolvimento local mantenha:

```env
OPENNDS_ENABLED=false
```

Para roteador real, configure SSH e ligue:

```env
OPENNDS_ENABLED=true
```

O backend entao usara `ndsctl auth` e `ndsctl deauth` via SSH.

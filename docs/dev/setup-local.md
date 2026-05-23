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
make install
make dev-infra
Set-Location node
go run ./cmd/setup
Set-Location ..
make test
```

`make install` baixa dependencias Go e instala dependencias do portal com
`npm install`.

`make dev-infra` sobe:

- Postgres
- Redis
- RabbitMQ
- pgAdmin

`go run ./cmd/setup` e o fluxo recomendado para criar ou atualizar o `.env`
local com dados pessoais, como Mercado Pago, admin, banco local e OpenNDS. Rode
o comando dentro de `node/`. Por padrao ele usa `.env`; para apontar outro
arquivo, use `go run ./cmd/setup -env-file caminho\\.env` ou configure
`ASTROLINK_ENV_FILE` no processo antes de iniciar o node.

Evite depender de Supabase no desenvolvimento local. O no local usa Postgres,
Redis e RabbitMQ do compose de desenvolvimento, e o provider de pagamento fica
em modo demo ate as credenciais do Mercado Pago serem configuradas.

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

Use o CLI local como caminho principal:

```powershell
Set-Location node
go run ./cmd/setup
```

O assistente atualiza somente chaves conhecidas e preserva o restante do
arquivo. Depois de alterar `.env`, reinicie o node para a configuracao entrar em
vigor.
Ao iniciar, o backend le `.env` ou o arquivo apontado por `ASTROLINK_ENV_FILE`
no processo e combina esses valores com as variaveis do processo; variaveis ja
exportadas no terminal tem prioridade.

Tambem e possivel usar `.env.example` como referencia:

```env
GO_ENV=development
HTTP_ADDR=:5000
LOG_LEVEL=debug

ASTROLINK_ALLOW_ENV_WRITE=false

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
MERCADOPAGO_PAYER_EMAIL=cliente@example.com
MERCADOPAGO_WEBHOOK_SECRET=test-webhook-secret

ADMIN_USUARIO=admin
ADMIN_SENHA=admin123
ADMIN_TOTP_SECRET=

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

`ASTROLINK_ENV_FILE` define, quando exportado no processo antes do startup, qual
arquivo `.env` sera lido e atualizado pelas ferramentas de setup; o default e
`.env`.

`ASTROLINK_ALLOW_ENV_WRITE=false` mantem a API e o painel sem permissao de
escrita no `.env`. Para permitir escrita pelo painel local, habilite
explicitamente:

```env
ASTROLINK_ALLOW_ENV_WRITE=true
```

Use essa permissao apenas em instalacoes locais confiaveis. Mesmo quando o
painel grava o arquivo, segredos nao devem ser exibidos em texto na API e o node
precisa ser reiniciado para usar os novos valores.

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

Para PIX Mercado Pago real, altere `PAYMENTS_PROVIDER=mercadopago` e
configure `MERCADOPAGO_ACCESS_TOKEN`, `MERCADOPAGO_PAYER_EMAIL` e
`MERCADOPAGO_WEBHOOK_SECRET`. O ambiente local continua offline por padrao com
`PAYMENTS_PROVIDER=demo`.

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

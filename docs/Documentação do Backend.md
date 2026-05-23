# Documentacao do Backend

## Stack

- Linguagem: Go 1.22+
- HTTP: Fiber
- Banco: Postgres ou store em memoria para desenvolvimento rapido
- Roteador: OpenNDS via SSH/`ndsctl`
- Pagamentos: provider demo por padrao, provider Mercado Pago para criacao PIX real e reconciliacao de webhook quando configurado

Codigo fonte: `node/`

## Estrutura

```text
node/
  cmd/server/              entrada da aplicacao
  cmd/setup/               assistente local para criar/atualizar .env
  internal/api/            servidor Fiber
  internal/api/portal/     rotas publicas do portal
  internal/api/admin/      rotas iniciais do admin local
  internal/config/         leitura de variaveis de ambiente
  internal/domain/         regras de planos e vouchers
  internal/gateway/        OpenNDS, SSH e no-op
  internal/infra/memory/   store em memoria
  internal/infra/postgres/ store Postgres
  internal/store/          contratos de persistencia
  migrations/              schema SQL
```

## Endpoints Implementados

### Publicos

- `GET /api/saude`
- `GET /api/settings`
- `GET /api/planos`
- `GET /api/sessao/status?mac=...`
- `POST /api/pix/gerar`
- `GET /api/pix/status/:txid`
- `GET /api/pix/aguardar/:txid`
- `POST /api/voucher/resgatar`

### Admin local inicial

- `POST /admin/auth/login`
- `POST /admin/auth/refresh`
- `POST /admin/auth/logout`
- `GET /admin/auth/me`
- `GET /admin/sistema/saude`
- `GET /admin/setup/status`
- `PUT /admin/setup/env`
- `GET /admin/planos`
- `GET /admin/usuarios`
- `GET /admin/vouchers`
- `POST /admin/vouchers/gerar`
- `POST /admin/usuarios/:mac/desconectar`

## Configuracao

Variaveis principais:

```env
GO_ENV=development
HTTP_ADDR=:5000
ASTROLINK_ALLOW_ENV_WRITE=false
DATABASE_URL=postgres://astrolink:devpassword@localhost:5432/astrolink?sslmode=disable
ADMIN_USUARIO=admin
ADMIN_SENHA=admin123
ADMIN_TOTP_SECRET=
JWT_SECRET=dev-jwt-secret-nao-usar-em-producao-32chars
PAYMENTS_PROVIDER=demo
MERCADOPAGO_ACCESS_TOKEN=TEST-XXXX-XXXX-XXXX
MERCADOPAGO_API_BASE_URL=
MERCADOPAGO_PAYER_EMAIL=cliente@example.com
MERCADOPAGO_WEBHOOK_SECRET=test-webhook-secret
NODE_NAME=dev-node-01
OPENNDS_ENABLED=false
OPENNDS_SSH_HOST=192.168.1.1
OPENNDS_SSH_PORT=22
OPENNDS_SSH_USER=root
OPENNDS_SSH_KEY_PATH=C:\Users\charl\.ssh\id_ed25519
OPENNDS_AUTH_RETRIES=3
```

Para setup local, prefira o assistente CLI:

```powershell
cd node
go run ./cmd/setup
```

O CLI cria ou atualiza `.env` por padrao. Para usar outro arquivo, rode
`go run ./cmd/setup -env-file caminho\\.env` ou defina `ASTROLINK_ENV_FILE` no
processo antes de iniciar o node. Ele e o caminho recomendado para dados
pessoais e segredos de instalacao local, incluindo Mercado Pago, admin,
banco/local e OpenNDS. No startup, o backend tambem consulta esse arquivo antes
de montar a config em memoria; variaveis ja definidas no processo continuam
tendo prioridade.

`ASTROLINK_ALLOW_ENV_WRITE=false` impede escrita de `.env` pela API/painel. Para
habilitar a escrita web em um no local confiavel, defina
`ASTROLINK_ALLOW_ENV_WRITE=true` antes de iniciar o backend. Mesmo assim, a API
de setup nunca deve retornar segredos em texto: campos sensiveis aparecem apenas
como configurados ou nao configurados.

Alteracoes no `.env` nao mudam a configuracao em memoria. Reinicie o node depois
de usar o CLI ou `PUT /admin/setup/env`.

Para criar PIX real e consultar detalhes de pagamento no webhook Mercado Pago,
use:

```env
PAYMENTS_PROVIDER=mercadopago
MERCADOPAGO_ACCESS_TOKEN=<access-token>
MERCADOPAGO_PAYER_EMAIL=<email-do-pagador-padrao>
MERCADOPAGO_WEBHOOK_SECRET=<webhook-secret>
```

`MERCADOPAGO_API_BASE_URL` e opcional e serve para testes/stubs; vazio usa
`https://api.mercadopago.com`.

## OpenNDS

Quando `OPENNDS_ENABLED=false`, o backend usa `NoopController` e nao executa
comandos no roteador. Isso deixa os testes e o desenvolvimento local simples.

Quando `OPENNDS_ENABLED=true`, o backend tenta executar:

- `ndsctl auth <mac> <duracao>`
- `ndsctl deauth <mac>`

por SSH no roteador configurado.

## Proximas Pendencias do Backend

- Agendamento automatico de jobs operacionais.
- Testes E2E com Postgres e OpenNDS simulado.

# Documentacao do Backend

## Stack

- Linguagem: Go 1.22+
- HTTP: Fiber
- Banco: Postgres ou store em memoria para desenvolvimento rapido
- Roteador: OpenNDS via SSH/`ndsctl`
- Pagamentos: provider demo por padrao, provider Mercado Pago para reconciliacao de webhook quando configurado

Codigo fonte: `node/`

## Estrutura

```text
node/
  cmd/server/              entrada da aplicacao
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
DATABASE_URL=postgres://astrolink:devpassword@localhost:5432/astrolink?sslmode=disable
ADMIN_USUARIO=admin
ADMIN_SENHA=admin123
JWT_SECRET=dev-jwt-secret-nao-usar-em-producao-32chars
PAYMENTS_PROVIDER=demo
MERCADOPAGO_ACCESS_TOKEN=TEST-XXXX-XXXX-XXXX
MERCADOPAGO_API_BASE_URL=
MERCADOPAGO_WEBHOOK_SECRET=test-webhook-secret
NODE_NAME=dev-node-01
OPENNDS_ENABLED=false
OPENNDS_SSH_HOST=192.168.1.1
OPENNDS_SSH_PORT=22
OPENNDS_SSH_USER=root
OPENNDS_SSH_KEY_PATH=C:\Users\charl\.ssh\id_ed25519
OPENNDS_AUTH_RETRIES=3
```

Para consultar detalhes reais de pagamento no webhook Mercado Pago, use:

```env
PAYMENTS_PROVIDER=mercadopago
MERCADOPAGO_ACCESS_TOKEN=<access-token>
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

- Criacao PIX real pelo Mercado Pago.
- 2FA opcional no admin local.
- Agendamento automatico de jobs operacionais.
- Testes E2E com Postgres e OpenNDS simulado.

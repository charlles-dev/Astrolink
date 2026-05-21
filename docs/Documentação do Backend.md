# Documentacao do Backend

## Stack

- Linguagem: Go 1.22+
- HTTP: Fiber
- Banco: Postgres ou store em memoria para desenvolvimento rapido
- Roteador: OpenNDS via SSH/`ndsctl`

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
NODE_NAME=dev-node-01
OPENNDS_ENABLED=false
OPENNDS_SSH_HOST=192.168.1.1
OPENNDS_SSH_PORT=22
OPENNDS_SSH_USER=root
OPENNDS_SSH_KEY_PATH=C:\Users\charl\.ssh\id_ed25519
OPENNDS_AUTH_RETRIES=3
```

## OpenNDS

Quando `OPENNDS_ENABLED=false`, o backend usa `NoopController` e nao executa
comandos no roteador. Isso deixa os testes e o desenvolvimento local simples.

Quando `OPENNDS_ENABLED=true`, o backend tenta executar:

- `ndsctl auth <mac> <duracao>`
- `ndsctl deauth <mac>`

por SSH no roteador configurado.

## Proximas Pendencias do Backend

- JWT real no admin.
- CRUD completo de planos.
- Exportacao e impressao de vouchers.
- Webhook real do Mercado Pago.
- Job de expiracao de sessoes.
- Logs de auditoria.
- Backup/restore.
- Testes E2E com Postgres e OpenNDS simulado.

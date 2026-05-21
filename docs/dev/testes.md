# Estrategia de Testes

## Comandos Principais

Na raiz do repo:

```powershell
make test
make check
make build
```

Por stack:

```powershell
cd node
go test ./...
go build ./cmd/server
```

```powershell
cd portal
npm test
npm run check
npm run build
```

## Backend Go

Testes atuais cobrem:

- handlers publicos do portal
- handlers iniciais do admin
- leitura de configuracao
- dominio de vouchers
- controller OpenNDS/no-op/SSH
- store Postgres

Arquivos relevantes:

```text
node/internal/api/admin/handlers_test.go
node/internal/api/portal/handlers_test.go
node/internal/config/config_test.go
node/internal/domain/vouchers/voucher_test.go
node/internal/gateway/opennds_test.go
node/internal/infra/postgres/store_test.go
```

Padrao esperado:

- testes de dominio sem I/O externo
- handlers testados via servidor Fiber em memoria
- OpenNDS testado com runner falso
- Postgres testado com banco isolado quando aplicavel

## Portal SvelteKit

Testes atuais cobrem:

- cliente de API
- formatadores
- componente de plano
- componente de voucher

Arquivos relevantes:

```text
portal/src/lib/api.test.ts
portal/src/lib/format.test.ts
portal/src/lib/components/PlanCard.test.ts
portal/src/lib/components/VoucherScreen.test.ts
```

## Checks Antes de Commit

Rode:

```powershell
go test ./...
npm test
npm run check
npm run build
```

Nos diretorios corretos (`node/` para Go e `portal/` para npm), ou use:

```powershell
make test
make check
make build
```

## Proximos Testes a Adicionar

- Playwright para fluxo completo do portal.
- Teste de expiracao de sessao quando o job existir.
- Testes do admin local visual quando a UI for criada.
- Teste de Mercado Pago quando o webhook real for implementado.
- Teste de health real do roteador com OpenNDS simulado.

## CI

O workflow `.github/workflows/ci.yml` executa:

- `go test ./...`
- build do backend Go
- `npm ci`
- `npm test`
- `npm run check`
- `npm run build`

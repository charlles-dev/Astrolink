# Astrolink

Astrolink e um ecossistema para vender e gerenciar acesso Wi-Fi em areas remotas usando Starlink, OpenWrt/OpenNDS, PIX e vouchers.

Esta branch esta em reconstrução para seguir a documentacao em `docs/`. A nova base do produto e:

- `node/`: backend local em Go, responsavel pelo portal cativo, painel admin, vouchers, pagamentos, OpenNDS, jobs e sync.
- `docs/`: specs, referencia de API, schema, infraestrutura, negocio e guias de desenvolvimento.
- `docker-compose.dev.yml`: Postgres, Redis, RabbitMQ e pgAdmin para desenvolvimento.

As pastas antigas `backend/`, `app/`, `admin-app/`, `cloud-app/` e `web/` ainda existem como legado enquanto a migracao e feita em fases. Elas nao sao mais a fonte de verdade arquitetural.

## Setup local

Pre-requisitos:

- Go 1.22+
- Node.js 20+ e pnpm 9+ para a proxima fase dos frontends
- Docker + Compose
- Make

```powershell
Copy-Item .env.example .env
make install
make dev-infra
make test
make dev
```

Servidor local:

```text
http://localhost:5000
```

Endpoints iniciais:

- `GET /api/saude`
- `GET /api/settings`
- `GET /api/planos`
- `GET /api/sessao/status?mac=AA:BB:CC:DD:EE:FF`
- `POST /api/pix/gerar`
- `GET /api/pix/status/:txid`
- `GET /api/pix/aguardar/:txid`
- `POST /api/voucher/resgatar`
- `POST /admin/auth/login`
- `GET /admin/sistema/saude`
- `GET /admin/planos`
- `GET /admin/usuarios`

## Testes

```powershell
make test
```

## Documentacao

Comece por:

- `docs/README.md`
- `docs/specs/portal-cativo.md`
- `docs/specs/admin-local.md`
- `docs/technical/api-reference.md`
- `docs/technical/database-schema.md`
- `docs/dev/setup-local.md`

## Proximas fases

1. Endurecer a camada Postgres com mais testes de integracao e migrations incrementais.
2. Implementar auth JWT real, refresh token e rate limiting.
3. Criar `portal/` e `admin/` em SvelteKit conforme as specs.
4. Integrar Mercado Pago, OpenNDS, Redis, RabbitMQ e jobs.
5. Remover definitivamente o legado Python/React quando a nova base cobrir os fluxos principais.

# Astrolink

Astrolink vende e gerencia acesso Wi-Fi em redes locais usando Starlink, OpenWrt/OpenNDS, PIX e vouchers.

Esta base foi limpa para manter apenas o novo stack:

- `node/`: backend local em Go para portal cativo, vouchers, pagamentos, admin local inicial e OpenNDS.
- `portal/`: portal cativo SvelteKit consumindo o backend Go.
- `docs/`: especificacoes, referencia tecnica e guias de desenvolvimento.
- `docker-compose.dev.yml`: infraestrutura local de desenvolvimento.

## Setup Local

Pre-requisitos:

- Go 1.22+
- Node.js 20+
- Docker + Compose
- Make

```powershell
Copy-Item .env.example .env
make install
make dev-infra
make test
```

Em terminais separados:

```powershell
make dev-node
make dev-portal
```

URLs locais:

- Backend Go: `http://localhost:5000`
- Portal cativo: `http://127.0.0.1:5173/?mac=AA:BB:CC:DD:EE:FF&ip=192.168.1.50&token=test`
- Painel local: `http://127.0.0.1:5173/painel`

## Endpoints Iniciais

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
- `GET /admin/vouchers`
- `POST /admin/vouchers/gerar`
- `POST /admin/usuarios/:mac/desconectar`

## Testes

```powershell
make test
make build
```

## OpenNDS

OpenNDS fica desabilitado por padrao no desenvolvimento local. Para testar em roteador real, configure no `.env`:

```env
OPENNDS_ENABLED=true
OPENNDS_SSH_HOST=192.168.1.1
OPENNDS_SSH_PORT=22
OPENNDS_SSH_USER=root
OPENNDS_SSH_KEY_PATH=C:\Users\charl\.ssh\id_ed25519
OPENNDS_AUTH_RETRIES=3
```

## Documentacao

Comece por:

- `docs/README.md`
- `docs/specs/portal-cativo.md`
- `docs/specs/admin-local.md`
- `docs/technical/openwrt-integration.md`
- `docs/technical/api-reference.md`
- `docs/technical/database-schema.md`
- `docs/dev/setup-local.md`

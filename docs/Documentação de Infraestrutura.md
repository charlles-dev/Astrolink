# Documentacao de Infraestrutura

## Ambientes

### Desenvolvimento

Rodado localmente com:

- backend Go em `node/`
- portal SvelteKit em `portal/`
- Postgres, Redis, RabbitMQ e pgAdmin via `docker-compose.dev.yml`

Comandos:

```powershell
Copy-Item .env.example .env
make install
make dev-infra
make dev-node
make dev-portal
```

### Producao Local

O `docker-compose.yml` sobe:

- `node`: backend Go
- `postgres`: banco local
- `redis`: cache/futuro runtime

O portal ainda roda como app SvelteKit separado em desenvolvimento. O empacotamento
final do portal para producao sera definido em fase posterior.

## Portas

| Servico | Porta |
|---|---|
| Backend Go | `5000` |
| Portal dev | `5173` |
| Postgres dev | `5432` |
| Redis dev | `6379` |
| RabbitMQ | `5672` |
| RabbitMQ Management | `15672` |
| pgAdmin | `5050` |

## Dados Locais

O compose de producao usa `./data` para volumes persistentes:

- `data/postgres`
- `data/redis`
- `data/uploads`
- `data/backups`
- `data/ssh-keys`

Esses diretorios ficam ignorados pelo Git.

## CI

O workflow `.github/workflows/ci.yml` valida:

- `go test ./...`
- build do backend Go
- `npm test`
- `npm run check`
- `npm run build`

## Pendencias

- Definir adapter de producao do SvelteKit.
- Empacotar portal junto do backend ou como servico separado.
- Criar script de instalacao para no local.
- Criar systemd/servico Windows para o backend.
- Definir rotina de backup/restore.

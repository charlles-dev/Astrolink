# Astrolink Docs

Esta pasta documenta a base nova do Astrolink. O codigo legado anterior foi
removido; a fonte de verdade agora e:

- `node/`: backend local em Go.
- `portal/`: portal cativo em SvelteKit.
- `docs/`: especificacoes e decisoes tecnicas.
- `docker-compose.dev.yml`: Postgres, Redis, RabbitMQ e pgAdmin para dev.

Admin cloud, app mobile, marketplace e outros produtos continuam como direcao de
produto, mas estao fora da implementacao imediata.

## Leitura Recomendada

1. [Arquitetura de Software](Arquitetura%20de%20Software.md)
2. [Setup local](dev/setup-local.md)
3. [Referencia da API](technical/api-reference.md)
4. [Schema do banco](technical/database-schema.md)
5. [Integracao OpenWrt/OpenNDS](technical/openwrt-integration.md)
6. [Portal cativo](specs/portal-cativo.md)
7. [Admin local](specs/admin-local.md)

## Specs

| Documento | Status |
|---|---|
| [portal-cativo.md](specs/portal-cativo.md) | Em implementacao |
| [admin-local.md](specs/admin-local.md) | Proxima etapa funcional |
| [painel-cloud.md](specs/painel-cloud.md) | Pausado por enquanto |
| [mapa-publico.md](specs/mapa-publico.md) | Futuro |
| [app-mobile.md](specs/app-mobile.md) | Futuro |
| [app-desktop.md](specs/app-desktop.md) | Futuro |
| [cli.md](specs/cli.md) | Futuro |
| [marketplace.md](specs/marketplace.md) | Futuro |
| [hardware-box.md](specs/hardware-box.md) | Futuro |
| [pwa-usuario.md](specs/pwa-usuario.md) | Futuro |

## Technical

| Documento | Uso |
|---|---|
| [api-reference.md](technical/api-reference.md) | Endpoints existentes e backlog |
| [database-schema.md](technical/database-schema.md) | Schema local em Postgres |
| [infraestrutura.md](technical/infraestrutura.md) | Docker, runtime e deploy local |
| [openwrt-integration.md](technical/openwrt-integration.md) | OpenNDS via SSH/`ndsctl` |
| [seguranca.md](technical/seguranca.md) | Diretrizes de seguranca |

Os documentos em `business/`, `technical/*cloud*`, `multi-tenancy`,
`sincronizacao-realtime` e `notificacoes` continuam como material de produto e
arquitetura futura. Eles nao devem guiar a implementacao desta fase.

## Dev

| Documento | Uso |
|---|---|
| [setup-local.md](dev/setup-local.md) | Como rodar o projeto |
| [testes.md](dev/testes.md) | Como validar backend e portal |
| [contribuindo.md](dev/contribuindo.md) | Fluxo de contribuicao |
| [padroes-codigo.md](dev/padroes-codigo.md) | Padroes de Go, TypeScript e SQL |

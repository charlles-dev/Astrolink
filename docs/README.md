# Documentação Astrolink

Bem-vindo à documentação do ecossistema Astrolink. Os documentos estão organizados em quatro categorias.

---

## 📐 Specs (Especificações de Produto)

Descrevem em detalhe cada produto/interface do ecossistema — telas, fluxos, comportamentos e endpoints consumidos.

| Documento | Descrição |
|---|---|
| [portal-cativo.md](specs/portal-cativo.md) | Interface que o usuário final vê ao conectar ao Wi-Fi |
| [admin-local.md](specs/admin-local.md) | Painel administrativo local do operador |
| [painel-cloud.md](specs/painel-cloud.md) | Painel SaaS multi-nó (Cloud Panel) |
| [mapa-publico.md](specs/mapa-publico.md) | Site público de descoberta de hotspots |
| [app-mobile.md](specs/app-mobile.md) | App Flutter para operadores |
| [app-desktop.md](specs/app-desktop.md) | App Tauri para instalação em campo |
| [cli.md](specs/cli.md) | CLI `astrolink` para SysAdmins |
| [marketplace.md](specs/marketplace.md) | Marketplace de temas para o portal |
| [hardware-box.md](specs/hardware-box.md) | Produto de hardware pré-configurado |
| [pwa-usuario.md](specs/pwa-usuario.md) | PWA para usuário final gerenciar sua sessão |

---

## ⚙️ Technical (Documentação Técnica)

Documentação de implementação: schemas, APIs, infraestrutura e decisões de arquitetura.

| Documento | Descrição |
|---|---|
| [database-schema.md](technical/database-schema.md) | Schema completo do banco local e Cloud |
| [api-reference.md](technical/api-reference.md) | Referência completa da API Go (local) |
| [infraestrutura.md](technical/infraestrutura.md) | Docker, deploy, systemd, CI/CD |
| [openwrt-integration.md](technical/openwrt-integration.md) | Integração com OpenNDS e roteadores |
| [seguranca.md](technical/seguranca.md) | Segurança, JWT, LGPD, OWASP |
| [multi-tenancy.md](technical/multi-tenancy.md) | RLS, isolamento de tenants, roles |
| [sincronizacao-realtime.md](technical/sincronizacao-realtime.md) | Sync Nó ↔ Cloud via RabbitMQ |
| [notificacoes.md](technical/notificacoes.md) | Sistema de notificações (push, email, WhatsApp) |

---

## 💼 Business (Documentação de Negócio)

Estratégia, monetização e go-to-market.

| Documento | Descrição |
|---|---|
| [monetizacao.md](business/monetizacao.md) | Modelo Open Core, receitas, projections |
| [precificacao.md](business/precificacao.md) | Planos detalhados, comparativo, políticas |
| [go-to-market.md](business/go-to-market.md) | ICP, canais de aquisição, lançamento |

---

## 🛠️ Dev (Guias para Desenvolvedores)

Como configurar o ambiente, contribuir e manter padrões de qualidade.

| Documento | Descrição |
|---|---|
| [setup-local.md](dev/setup-local.md) | Setup do ambiente de desenvolvimento |
| [contribuindo.md](dev/contribuindo.md) | Como contribuir com o projeto |
| [padroes-codigo.md](dev/padroes-codigo.md) | Convenções de código (Go, TypeScript, SQL) |
| [testes.md](dev/testes.md) | Estratégia de testes e cobertura |

---

## 📋 Documentos Raiz

| Documento | Descrição |
|---|---|
| [ARCHITECTURE.md](../ARCHITECTURE.md) | Visão geral da arquitetura do ecossistema |
| [ROADMAP.md](../ROADMAP.md) | Fases de desenvolvimento e marcos |


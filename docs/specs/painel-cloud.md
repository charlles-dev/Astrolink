# Spec: Painel Cloud Multi-Nó

## Visão Geral

O Painel Cloud é o produto SaaS pago da Astrolink. Permite que operadores (provedores, donos de LAN house, gestores de rede) gerenciem **múltiplos nós remotamente** de qualquer lugar, via browser ou app mobile.

**URL:** `app.astrolink.app`
**Stack:** SvelteKit + Supabase + Edge Functions (TypeScript/Deno)
**Auth:** Supabase Auth (email/senha, magic link, futuramente Google/Apple)

---

## Modelo de Dados Multi-Tenant

```sql
-- Cada operador tem um tenant
tenants (id, nome, slug, plano, status, created_at)

-- Membros de cada tenant (com roles)
tenant_members (tenant_id, user_id, role: 'owner'|'admin'|'viewer')

-- Nós vinculados ao tenant
nodes (id, tenant_id, nome, slug, lat, lng, timezone,
       status: 'online'|'offline'|'degraded',
       last_heartbeat_at, token_hash, public_listing, created_at)

-- Snapshots de métricas por nó (recebidos via RabbitMQ)
node_metrics (id, node_id, timestamp,
              users_active, revenue_today, bandwidth_down_mbps, bandwidth_up_mbps)

-- Eventos dos nós (log de tudo que aconteceu)
node_events (id, node_id, type, payload JSONB, created_at)

-- Planos de billing
billing_plans (id, nome, preco_mensal, max_nos, features JSONB)
subscriptions (id, tenant_id, plan_id, status, current_period_end, abacatepay_id)
```

**Row Level Security:** Todo `SELECT/INSERT/UPDATE/DELETE` verifica `tenant_id = auth.jwt()['tenant_id']`.

---

## Autenticação e Onboarding

### Cadastro
1. Email + senha (mínimo 8 chars)
2. Verificação de email (link de confirmação)
3. Wizard de onboarding:
   - Nome do provedor / empresa
   - Quantos nós você gerencia hoje?
   - Como você soube da Astrolink?
4. Criar workspace (slug automático: `meu-provedor.astrolink.app`)
5. Tela de "Adicione seu primeiro nó"

### Login
- Email/senha com "Lembrar por 30 dias"
- Magic link (link por email, sem senha)
- MFA: TOTP (opcional, obrigatório para roles admin)

---

## Estrutura de Navegação

```
app.astrolink.app/
├── /dashboard          Dashboard geral (todos os nós)
├── /nos                Lista de nós
│   └── /nos/:slug      Detalhes de um nó específico
├── /financeiro         Receita consolidada
├── /alertas            Central de alertas
├── /equipe             Membros do workspace
├── /configuracoes      Configurações do workspace
│   ├── /conta          Dados da conta
│   ├── /billing        Plano e cobrança
│   └── /api            API keys e webhooks
└── /suporte            Help center e tickets
```

---

## Dashboard Geral

### Header de Métricas Consolidadas

```
┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│ 🟢 Nós Online│ │ 👥 Usuários  │ │ 💰 Receita   │ │ 📈 Este mês  │
│   8 / 10     │ │  247 agora   │ │  R$ 1.234    │ │  R$ 18.450   │
│ ⚠️ 1 offline │ │              │ │    hoje      │ │              │
└──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘
```

### Mapa de Nós

Mapa interativo (Leaflet.js) com todos os nós do tenant:
- 🟢 Pin verde: online e saudável
- 🟡 Pin amarelo: degradado (latência alta, poucos usuários)
- 🔴 Pin vermelho: offline
- Clique no pin: tooltip com métricas rápidas + link para detalhes

### Lista de Nós (visão alternativa ao mapa)

| Nó | Status | Usuários | Receita hoje | Último heartbeat | Ação |
|---|---|---|---|---|---|
| Parauapebas-01 | 🟢 Online | 23 | R$ 345 | agora | [Ver] |
| Marabá-Centro | 🟢 Online | 47 | R$ 705 | 1 min | [Ver] |
| Açailândia-01 | 🔴 Offline | 0 | R$ 0 | 18 min ⚠️ | [Ver] |

### Gráfico de Receita Consolidada

- Tipo: barras agrupadas por nó (7/14/30 dias)
- Toggle: por nó individual ou total
- Hover: breakdown do dia selecionado

### Feed de Eventos Recentes (todos os nós)

```
[14:32] Parauapebas-01  💰 Pagamento R$15,00 aprovado
[14:28] Marabá-Centro   👤 Usuário conectado (AA:BB:CC...)
[14:15] Açailândia-01   ⚠️ Roteador principal offline
[13:55] Parauapebas-01  🎟️  Voucher resgatado (plano 24h)
```

---

## Detalhes de um Nó

### Header do Nó

```
← Todos os nós

🌐 Parauapebas-01
Parauapebas, PA · Criado em 10/03/2025
🟢 Online · Último heartbeat: 2s atrás

[Comandos Remotos ▼]  [Configurações]
```

### Métricas em Tempo Real (via Supabase Realtime)

```
┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│ Usuários     │ │ Receita Hoje │ │ Bandwidth ↓  │ │ Banda Total  │
│ 23 ativos    │ │  R$ 345,00   │ │  45.2 Mbps   │ │  ↓12GB ↑2GB  │
└──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘
```

### Gráfico de Usuários (última hora, 5min granularidade)

### Usuários Ativos (tabela ao vivo)

| MAC | Plano | Tempo restante | Banda atual | Ação |
|---|---|---|---|---|
| AA:BB:... | 24h | 18h 32m | 2.1 Mbps | [Desconectar] [Banir] |

### Comandos Remotos

Dropdown com ações que são enviadas via RabbitMQ ao nó:

- **Banir MAC** — modal: input MAC + motivo
- **Desconectar usuário** — por MAC
- **Estender sessão** — MAC + horas
- **Recarregar configurações** — aplica mudanças de planos/settings
- **Reiniciar OpenNDS** — reinicia o engine captivo
- **Baixar diagnóstico** — gera relatório técnico (logs, métricas, config)

Todos os comandos são logados em `node_events` com timestamp e usuário que executou.

### Histórico de Eventos do Nó

Timeline paginada com todos os eventos, filtrável por tipo.

---

## Seção Financeiro

### Receita Consolidada

```
[Período: Este mês ▼]  [Nó: Todos ▼]  [Exportar ▼]

Receita Total: R$ 18.450,00
Aprovadas: 1.230 transações | Média: R$ 15,00/transação

[Gráfico de área — receita diária por nó]

[Breakdown por nó:]
Marabá-Centro:      R$ 8.200,00  (44%)
Parauapebas-01:     R$ 6.750,00  (37%)
Açailândia-01:      R$ 3.500,00  (19%)
```

### Previsão de Receita

- Regressão linear simples baseada nos últimos 90 dias
- "No ritmo atual, você vai faturar ~R$ 19.200 este mês"
- Gráfico com linha de tendência

### Export de Relatórios

- CSV com todas as transações do período (compatível com Excel/Google Sheets)
- PDF formatado para enviar ao contador
- Filtros: período, nó, status, tipo (PIX/voucher)

---

## Central de Alertas

### Configurar Alertas

```
Nó offline por mais de: [5] minutos → [☑] Email  [☑] Push  [☐] WhatsApp

Vouchers abaixo de: [20] unidades → [☑] Email  [☐] Push

Receita diária abaixo de: R$ [50] → [☐] Email  [☑] Push

Atividade suspeita (>10 tentativas voucher inválido/hora) → [☑] Email
```

### Histórico de Alertas

| Data | Nó | Alerta | Status | Duração |
|---|---|---|---|---|
| 19/05 14:15 | Açailândia-01 | Nó offline | 🔴 Ativo | 18 min |
| 18/05 09:30 | Parauapebas-01 | Vouchers baixos (12 restantes) | ✅ Resolvido | — |
| 17/05 22:00 | Marabá-Centro | Nó offline | ✅ Resolvido | 3 min |

---

## Gestão de Equipe

### Membros do Workspace

| Nome | Email | Role | Status | Último acesso |
|---|---|---|---|---|
| João Silva | joao@... | 👑 Owner | Ativo | agora |
| Maria Souza | maria@... | 🔧 Admin | Ativo | ontem |
| Carlos Lima | carlos@... | 👁️ Viewer | Ativo | 5 dias |

**Roles:**
- **Owner:** acesso total, pode deletar workspace, gerenciar billing
- **Admin:** acesso total exceto billing e exclusão do workspace
- **Viewer:** somente leitura (dashboard, métricas, logs)

**Convidar membro:**
```
Email: [___________________]
Role: [Admin ▼]
Mensagem personalizada: [_______________] (opcional)
[Enviar convite]
```
Link de convite expira em 7 dias.

---

## Configurações do Workspace

### Informações Gerais
```
Nome do workspace: [Provedor XYZ          ]
Slug: provedor-xyz  (não alterável após criação)
Fuso horário: [America/Belem ▼]
Idioma: [Português (BR) ▼]
```

### Billing

```
Plano atual: Pro — R$ 49/mês
Próxima cobrança: 19/06/2025
Nós utilizados: 8 / 10

[Upgrade para Business — R$ 149/mês]
[Cancelar plano]

Histórico de faturas:
  19/05/2025  R$ 49,00  ✅ Pago  [Baixar PDF]
  19/04/2025  R$ 49,00  ✅ Pago  [Baixar PDF]
  19/03/2025  R$ 49,00  ✅ Pago  [Baixar PDF]
```

Cobrança via PIX automático (AbacatePay). No vencimento, PIX gerado e enviado por email. 7 dias de graça antes de suspender.

### API Keys

Para integrações externas:

```
API Keys:
  sk_prod_XXXXXXXXXXXXXX  [Criada 10/03] [Revogar]
  sk_prod_YYYYYYYYYYYYYY  [Criada 01/05] [Revogar]

[+ Criar nova API Key]
```

### Webhooks

```
Eventos disponíveis:
  [☑] node.offline        [☑] node.online
  [☑] payment.approved    [☐] user.connected
  [☐] user.expired        [☐] voucher.redeemed

URL do webhook: [https://...]
Secret: [WEBHOOK_SECRET_XXXXXXXX] [Regenerar]
[Testar webhook]  [Salvar]
```

---

## Vinculação de Nó

Processo para o operador adicionar um novo nó ao Cloud Panel:

### No Cloud Panel (passo 1):
1. Clicar em "Adicionar Nó"
2. Preencher: nome, localização (endereço ou coordenadas), fuso horário
3. Sistema gera token único: `ASTRO-XXXXXXXX-XXXXXXXXXXXX`

### No servidor local (passo 2):
```bash
# Via CLI
astrolink cloud link --token ASTRO-XXXXXXXX-XXXXXXXXXXXX

# Ou via Admin Local: Configurações → Cloud → Inserir token
```

### Confirmação:
- Nó envia heartbeat inicial ao Cloud via RabbitMQ
- Cloud Panel exibe nó como 🟢 Online
- Dados históricos do nó são sincronizados (últimas 24h)

---

## Site Público de Listagem de Nós

Cada nó pode optar por aparecer no mapa público (`astrolink.app/mapa`):

```
Aparecer no mapa público: [☑]

Informações públicas:
  Nome público: [Wi-Fi Pousada Recanto Verde]
  Descrição: [Wi-Fi de qualidade para hóspedes e visitantes]
  Foto: [Selecionar imagem]
  Endereço público: [___________________]
  Site: [___________________]
  Horário: Seg–Dom, 06:00–22:00
```

Dados sensíveis (receita, MACs, tokens) NUNCA aparecem publicamente.

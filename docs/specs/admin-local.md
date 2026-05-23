# Spec: Painel Administrativo Local

## Visão Geral

O Painel Admin Local é a interface de gerenciamento que roda **no próprio servidor do Nó**, acessível via rede local (ex: `http://192.168.1.1:5000/admin`). Funciona 100% offline, sem depender da Cloud.

**Stack:** SvelteKit + TailwindCSS + Chart.js + Lucide Icons

**Acesso:** Somente via rede local (não exposto à internet diretamente)

---

## Autenticação

### Login
- Usuário/senha configurados na instalação
- JWT com expiração de 8 horas
- Refresh token com 30 dias (rotação automática)
- 2FA via TOTP (Google Authenticator, Authy) — opcional mas recomendado
- Limite de tentativas: 5 falhas → bloqueio 10 minutos
- Log de todas as tentativas (sucesso e falha)

### Sessão
```
POST /admin/auth/login          → {access_token, refresh_token}
POST /admin/auth/refresh        → {access_token}
POST /admin/auth/logout         → invalida tokens
GET  /admin/auth/me             → dados do usuário logado
POST /admin/auth/2fa/setup      → gerar QR code TOTP
POST /admin/auth/2fa/verify     → verificar código TOTP
```

---

## Layout Geral

```
┌────────────────────────────────────────────────┐
│ [🔵 Astrolink]     Nó: Parauapebas-01  [👤] [⚙] │
├────────┬───────────────────────────────────────┤
│        │                                       │
│  NAV   │         CONTEÚDO PRINCIPAL            │
│        │                                       │
│ 📊 Dashboard      ← seção ativa                │
│ 👥 Usuários                                    │
│ 📦 Planos                                      │
│ 🎟  Vouchers                                   │
│ 💰 Pagamentos                                  │
│ 🌐 Rede                                        │
│ ⚙️  Configurações                               │
│ 📋 Logs                                        │
│                                                │
│ [v1.2.0]                                       │
└────────┴───────────────────────────────────────┘
```

---

## Seção 1: Dashboard

### Cards de Métricas (topo)

```
┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│ 👥 Conectados│ │ 💰 Hoje      │ │ 📈 Semana    │ │ 🌐 Uptime    │
│              │ │              │ │              │ │              │
│     23       │ │  R$ 345,00   │ │  R$ 1.890,00 │ │   99.8%      │
│   usuários   │ │              │ │              │ │  Roteador 1  │
│  ▲ 3 última  │ │ ▲12% vs ontem│ │              │ │  ● ● ● ○     │
│    hora      │ │              │ │              │ │  4 roteadores│
└──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘
```

### Gráfico de Receita (últimos 14 dias)
- Tipo: Área com gradiente
- Hover: tooltip com valor exato e comparativo dia anterior
- Clique no ponto: drill-down para transações do dia

### Tabela de Usuários Ativos (tempo real via WebSocket)

| MAC | Nome | Plano | Tempo restante | Banda ↓/↑ | Ação |
|---|---|---|---|---|---|
| AA:BB:... | João | 24h | 18h 32m | 2.1/0.3 Mbps | [✕] |
| CC:DD:... | — | Voucher | 45m | 0.8/0.1 Mbps | [✕] |

- Atualiza automaticamente a cada 10s
- Badge vermelho piscando quando usuário excede limite do plano
- Clique na linha: expandir detalhes (IP, histórico, localização na rede)

### Status dos Roteadores

```
┌─────────────────────────────────────────────┐
│ Roteadores                          Ver todos│
│                                             │
│ ● Roteador Principal (192.168.1.1)   23ms  │
│ ● AP Sala de Espera (192.168.1.2)    45ms  │
│ ● AP Área Externa (192.168.1.3)      67ms  │
│ ○ AP Depósito (192.168.1.4)      OFFLINE   │
│                              ⚠️ 1 problema  │
└─────────────────────────────────────────────┘
```

### Alertas Recentes

Lista das últimas notificações não lidas:
- ⚠️ "AP Depósito offline há 15 minutos"
- ✅ "Pagamento confirmado: R$ 15,00 (João Silva)"
- ℹ️ "Voucher ABCD-1234 resgatado"

---

## Seção 2: Usuários

### Lista de Usuários

**Filtros:**
- Status: Todos | Conectados agora | Expirados | Bloqueados
- Plano: Todos | Por plano específico
- Período: Hoje | Semana | Mês | Personalizado

**Tabela:**

| MAC | Nome | IP | Status | Plano | Início | Expira | Banda | Ações |
|---|---|---|---|---|---|---|---|---|
| AA:BB:CC... | João | 192.168.1.50 | 🟢 Ativo | 24h | 14:30 | amanhã 14:30 | 1.2↓ 0.2↑ | [👁][✕][+][🚫] |

**Ações por usuário:**
- 👁 **Ver detalhes**: histórico completo, sessões anteriores, total gasto
- ✕ **Desconectar**: revoga acesso imediatamente (ndsctl deauth)
- + **Estender tempo**: modal com seletor de horas/dias
- 🚫 **Banir MAC**: bloquear permanentemente com comentário opcional

### Detalhes do Usuário (modal ou página)

```
Usuário: João Silva
MAC: AA:BB:CC:DD:EE:FF
IP atual: 192.168.1.50

Sessão atual:
  Plano: Acesso 24 Horas — R$ 15,00
  Início: 19/05/2025 14:30:22
  Expira: 20/05/2025 14:30:22
  Tempo restante: 18h 32m
  Dados: ↓ 1.2GB / ↑ 234MB

Histórico de sessões (últimas 10):
  [tabela com data, plano, valor, duração]

Total histórico:
  Sessões: 12
  Total gasto: R$ 180,00
  Última visita: há 2 dias
```

---

## Seção 3: Planos

### Lista de Planos

Cards visuais mostrando exatamente como aparecem no portal:

```
┌─────────────────────┐  ┌─────────────────────┐
│ ⭐ RECOMENDADO      │  │                     │
│ Acesso 24 Horas     │  │ Acesso 1 Hora        │
│ ⏱ 24 horas         │  │ ⏱ 1 hora            │
│ ↓5 Mbps ↑2 Mbps     │  │ ↓5 Mbps ↑2 Mbps     │
│          R$ 15,00   │  │           R$ 5,00   │
│ [✏️ Editar] [🗑️]   │  │ [✏️ Editar] [🗑️]   │
│ 142 resgates        │  │ 89 resgates          │
└─────────────────────┘  └─────────────────────┘
```

### Criar / Editar Plano (modal)

```
Nome: [_____________________]
Descrição: [__________________]  (aparece no portal)
Preço: R$ [______]
Duração: [___] horas  OU  [___] GB de dados
Limite de velocidade:
  Download: [___] Mbps  (0 = ilimitado)
  Upload:   [___] Mbps  (0 = ilimitado)
Recomendado: [☑] (badge "Recomendado" no portal)
Ativo: [☑]
Visível no portal: [☑]

[Cancelar]  [Salvar]
```

---

## Seção 4: Vouchers

### Gerador de Vouchers

```
Gerar novo lote:
  Plano: [Dropdown com planos ativos]
  Quantidade: [___] vouchers
  Formato do código: [XXXX-XXXX] ▼
  Prefixo: [___] (opcional, ex: "VIP")
  Validade do voucher: [___] dias após geração
  Tipo:
    ○ Uso único (padrão)
    ○ Universal (múltiplos usos, limite: [___])

[Gerar Lote]
```

### Lista de Vouchers

**Filtros:** Status (Disponível | Usado | Expirado), Plano, Data de criação

| Código | Plano | Status | Criado | Usado em | Por |
|---|---|---|---|---|---|
| ABCD-1234 | 24h | ✅ Disponível | 10/05 | — | — |
| EFGH-5678 | 24h | 🔴 Usado | 10/05 | 15/05 14:30 | AA:BB... |
| IJKL-9012 | 1h | ⏰ Expirado | 01/05 | — | — |

**Ações:**
- Buscar voucher por código
- Desativar voucher individual
- Export CSV do lote

### Impressão de Vouchers (PDF)

Template profissional com 8 vouchers por folha A4:

```
┌────────────────────────┐  ┌────────────────────────┐
│ 🌐 Astrolink Wi-Fi     │  │ 🌐 Astrolink Wi-Fi     │
│                        │  │                        │
│ Acesso 24 Horas        │  │ Acesso 24 Horas        │
│                        │  │                        │
│ ████████████████████   │  │ ████████████████████   │
│   ABCD - 1234          │  │   EFGH - 5678          │
│ ████████████████████   │  │ ████████████████████   │
│                        │  │                        │
│ Validade: 30 dias      │  │ Validade: 30 dias      │
│ Conecte em: wifi.local │  │ Conecte em: wifi.local │
└────────────────────────┘  └────────────────────────┘
```

Configurações do PDF:
- Logo customizável
- Nome do hotspot
- URL do portal
- Instruções de uso
- Validade

---

## Seção 5: Pagamentos

### Histórico de Transações

**Filtros:** Data (intervalo), Status (Aprovado | Pendente | Cancelado), Valor mínimo/máximo

| Data | Hora | Valor | Plano | Usuário (MAC) | TxID | Status |
|---|---|---|---|---|---|---|
| 19/05 | 14:32 | R$ 15,00 | 24h | AA:BB:CC... | mp_xxxx | ✅ Aprovado |
| 19/05 | 11:15 | R$ 5,00 | 1h | DD:EE:FF... | mp_yyyy | ✅ Aprovado |
| 19/05 | 09:00 | R$ 15,00 | 24h | GG:HH:II... | mp_zzzz | ⏳ Pendente |

**Totais do período selecionado:**
```
Total: R$ 345,00  |  Aprovados: 23  |  Pendentes: 2  |  Cancelados: 1
```

**Export:**
- CSV para planilhas
- PDF relatório formatado para contador

### Configuração Mercado Pago

```
Access Token: [________________________________] [👁]
Public Key:   [________________________________] [👁]

Modo: ○ Sandbox (testes)  ● Produção

[Testar Conexão]  →  ✅ Conectado (conta: João Silva)

[Salvar]
```

---

## Seção 6: Rede

### Roteadores Vinculados

| Nome | IP | Status | Uptime | Usuários | Ações |
|---|---|---|---|---|---|
| Roteador Principal | 192.168.1.1 | 🟢 Online | 14d 6h | 18 | [📊][🔧][🗑️] |
| AP Sala de Espera | 192.168.1.2 | 🟢 Online | 14d 6h | 5 | [📊][🔧][🗑️] |
| AP Depósito | 192.168.1.4 | 🔴 Offline | — | 0 | [📊][🔧][🗑️] |

**Adicionar roteador:**
```
Nome: [___________________]
IP/Host: [_______________]
Porta SSH: [22]
Usuário SSH: [root]
Chave SSH: [Selecionar arquivo .pem] OU [Usar chave do sistema]

[Testar Conexão]  →  ✅ Conectado (OpenWrt 22.03, OpenNDS 10.1)

[Adicionar]
```

**Diagnóstico por roteador (modal):**
```
● Roteador Principal (192.168.1.1)

Ping: 23ms (média 5 tentativas)
Uptime: 14 dias, 6 horas
Versão OpenWrt: 22.03.5
Versão OpenNDS: 10.1.2
Usuários autorizados: 18
Usuários walled garden: 3

[Ping]  [Traceroute]  [Reiniciar OpenNDS]  [Reiniciar Roteador]

Teste de velocidade do uplink:
  ↓ [Testar Download]  →  45.2 Mbps (Starlink)
  ↑ [Testar Upload]    →  8.7 Mbps

Log OpenNDS (últimas 50 linhas):
[textarea com logs ao vivo via WebSocket]
```

### Blacklist de MACs

| MAC | Descrição | Bloqueado em | Bloqueado por | Ação |
|---|---|---|---|---|
| AA:BB:CC:DD:EE:FF | Uso abusivo - 200GB/dia | 15/05 | admin | [🗑️] |

```
Adicionar à blacklist:
  MAC: [__:__:__:__:__:__]
  Motivo: [___________________]
  [Adicionar]
```

### Walled Garden (Whitelist de Domínios)

Domínios acessíveis sem pagamento:

| Domínio / IP | Descrição | Ação |
|---|---|---|
| pagamentos.mercadopago.com | Necessário para PIX | 🔒 (sistema) |
| www.google.com | Verificação de conectividade | [🗑️] |
| pousadarecantoverde.com.br | Site do estabelecimento | [🗑️] |

```
Adicionar domínio:
  [___________________________]  [Adicionar]
```

---

## Seção 7: Configurações

### Geral
```
Nome do hotspot: [___________________]
Logo: [Selecionar imagem] (max 2MB, PNG/JPG/SVG)
URL do portal pós-conexão: [___________________]
Coletar nome do usuário: [☐]

Cores:
  Primária: [🎨 #06B6D4]
  Fundo: [🎨 #0F172A]
```

### Segurança
```
Alterar senha do admin: [atual] [nova] [confirmar]
2FA: [☑ Ativado] [Reconfigurar QR Code]
IPs permitidos para o admin: [___] (vazio = todos na rede local)
```

### Sincronização com Cloud
```
Status: 🟢 Conectado ao Cloud Panel
Tenant: Provedor XYZ (provedor-xyz.astrolink.app)
Nó: Parauapebas-01

Token de vinculação: [ASTRO-XXXXXX-XXXXXXXXXX] [📋 Copiar]

Última sync: há 2 minutos
Eventos pendentes: 0

[Desconectar do Cloud]
```

### Backup e Restauração
```
Backup Manual:
  [Baixar backup completo (.sql.gz)]
  Último backup: 19/05/2025 03:00 (automático)

Backups automáticos:
  Frequência: ● Diário  ○ Semanal  ○ Desativado
  Hora: [03:00]
  Reter últimos: [7] backups

Restaurar backup:
  [Selecionar arquivo .sql.gz]
  ⚠️ Isso sobrescreve TODOS os dados atuais!
  [Restaurar]
```

### Atualização de Software
```
Versão atual: v1.2.0
Última versão: v1.3.0 ✨ Atualização disponível!

Novidades na v1.3.0:
  • Suporte a pagamento por dados (GB)
  • Melhoria na velocidade do portal
  • Correção no rate limiting

[Atualizar agora] (reinicia o serviço)
[Ver changelog completo]
```

---

## Seção 8: Logs de Auditoria

| Timestamp | Usuário | Ação | Detalhes | IP |
|---|---|---|---|---|
| 19/05 14:35 | admin | BAN_MAC | AA:BB:CC... "Uso abusivo" | 192.168.1.10 |
| 19/05 14:30 | system | SESSION_EXPIRED | DD:EE:FF... plano 24h | — |
| 19/05 14:25 | admin | LOGIN | — | 192.168.1.10 |
| 19/05 14:20 | admin | PLAN_UPDATED | "24h" preço 15→18 | 192.168.1.10 |

**Filtros:** Usuário, Tipo de ação, Período
**Export:** CSV

---

## Real-time via WebSocket

Conexão: `ws://[host]/admin/ws?token=[jwt]`

**Eventos recebidos:**
```json
{"event": "user.connected", "mac": "AA:BB:CC:...", "plan": "24h"}
{"event": "user.expired", "mac": "AA:BB:CC:..."}
{"event": "payment.approved", "amount": 15.00, "plan": "24h"}
{"event": "router.offline", "router_id": 4, "name": "AP Depósito"}
{"event": "router.online", "router_id": 4}
{"event": "metrics.update", "connected": 23, "revenue_today": 345.00}
```

Reconexão automática com backoff exponencial (1s → 2s → 4s → max 30s).

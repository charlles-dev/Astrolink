# Spec: App Mobile (Flutter)

## Visão Geral

App nativo para iOS e Android destinado ao **operador/provedor** que precisa monitorar e gerenciar sua rede de hotspots de qualquer lugar. Complementa o Painel Cloud Web com uma experiência móvel otimizada e notificações push.

**Stack:** Flutter 3+ | Riverpod (state) | Supabase Flutter SDK | Firebase Cloud Messaging
**Plataformas:** Android 10+ / iOS 14+

---

## Filosofia do App

- **Monitoramento em primeiro lugar:** a maioria dos acessos é para checar status, não gerenciar
- **Ações rápidas:** as 5 ações mais comuns devem ter no máximo 2 taps
- **Push first:** o app só precisa estar aberto quando o operador quer agir; o resto é notificação
- **Offline graceful:** exibir dados cacheados quando sem internet, com indicador claro

---

## Telas

### 1. Onboarding / Login

```
[Splash com logo animado]
    ↓
[Login]
  Email: [___________________]
  Senha: [___________________]  [👁]
  [Lembrar por 30 dias]
  [Entrar]
  [Entrar com magic link →]
  [Esqueci minha senha]
```

Token JWT armazenado em Flutter Secure Storage (keychain/keystore nativo).

---

### 2. Dashboard Home

```
┌─────────────────────────────────┐
│ Bom dia, João 👋       [🔔 3]   │
│ Provedor XYZ                    │
├─────────────────────────────────┤
│ Resumo geral              Hoje  │
│                                 │
│ ┌─────────────┐ ┌─────────────┐ │
│ │ 🟢 8/10     │ │ 💰 R$1.234  │ │
│ │ Nós online  │ │ Receita     │ │
│ └─────────────┘ └─────────────┘ │
│ ┌─────────────┐ ┌─────────────┐ │
│ │ 👥 247      │ │ ⚠️ 1        │ │
│ │ Usuários    │ │ Alertas     │ │
│ └─────────────┘ └─────────────┘ │
├─────────────────────────────────┤
│ Meus Nós                 Ver →  │
│                                 │
│ ● Parauapebas-01   23u  R$345  │
│ ● Marabá-Centro    47u  R$705  │
│ ○ Açailândia-01    OFFLINE ⚠️   │
│                                 │
├─────────────────────────────────┤
│ Atividade recente               │
│                                 │
│ 💰 R$15 aprovado — 14:32       │
│ ⚠️ Açailândia offline — 14:15  │
│ 👤 Usuário conectado — 14:28   │
│                                 │
└─────────────────────────────────┘
[ 🏠 Home ] [ 📍 Nós ] [ 💰 $ ] [ 🔔 ] [ 👤 ]
```

---

### 3. Lista de Nós

```
┌─────────────────────────────────┐
│ ← Meus Nós            [+ Add]   │
├─────────────────────────────────┤
│ [🔍 Buscar nó...]               │
│                                 │
│ ┌─────────────────────────────┐ │
│ │ ● Parauapebas-01            │ │
│ │ Parauapebas, PA             │ │
│ │ 👥 23  💰 R$345  ↓45Mbps  │ │
│ └─────────────────────────────┘ │
│                                 │
│ ┌─────────────────────────────┐ │
│ │ ● Marabá-Centro             │ │
│ │ Marabá, PA                  │ │
│ │ 👥 47  💰 R$705  ↓38Mbps  │ │
│ └─────────────────────────────┘ │
│                                 │
│ ┌─────────────────────────────┐ │
│ │ ○ Açailândia-01   OFFLINE  │ │
│ │ Açailândia, MA              │ │
│ │ ⚠️ Offline há 18 minutos   │ │
│ └─────────────────────────────┘ │
└─────────────────────────────────┘
```

---

### 4. Detalhes do Nó

```
┌─────────────────────────────────┐
│ ← Parauapebas-01     [⚡ Ações] │
│ 🟢 Online · heartbeat: 3s atrás │
├─────────────────────────────────┤
│                                 │
│ Hoje                            │
│ ┌───────┐ ┌───────┐ ┌────────┐ │
│ │👥 23  │ │💰 345 │ │↓45Mbps│ │
│ │ativos │ │ hoje  │ │       │ │
│ └───────┘ └───────┘ └────────┘ │
│                                 │
│ [Gráfico usuários — última hora]│
│ ▁▂▃▅▆▇▇▅▄▃▂▂▃▄▅▆▇▆▅▄▃▄▅▆▇    │
│                                 │
├─────────────────────────────────┤
│ Usuários ativos         Ver →   │
│                                 │
│ AA:BB:CC  24h  18h restantes   │
│ DD:EE:FF  1h   23m restantes   │
│ GG:HH:II  Voucher  5h rest.   │
│                                 │
├─────────────────────────────────┤
│ Roteadores                      │
│ ● Principal (192.168.1.1) 23ms │
│ ● Sala (192.168.1.2)      45ms │
│ ○ Externo    OFFLINE ⚠️        │
└─────────────────────────────────┘
```

**Menu de Ações Rápidas (bottom sheet ao tocar ⚡):**
```
┌─────────────────────────────────┐
│ Ações — Parauapebas-01          │
│                                 │
│ [🚫 Banir MAC]                  │
│ [✂️  Desconectar usuário]        │
│ [⏱️  Estender sessão]            │
│ [🎟️  Gerar voucher]             │
│ [🔄 Reiniciar OpenNDS]          │
│ [📊 Diagnóstico completo]       │
│                                 │
│ [Cancelar]                      │
└─────────────────────────────────┘
```

---

### 5. Financeiro

```
┌─────────────────────────────────┐
│ Financeiro                      │
│ [Este mês ▼]  [Todos nós ▼]    │
├─────────────────────────────────┤
│                                 │
│ R$ 18.450,00                    │
│ 1.230 transações                │
│                                 │
│ [Gráfico barras — 30 dias]      │
│ ▁▂▃▅▆▇▇▅▄▃▂▂▃▄▅▆▇▆▅▄▃▄▅▆▇▇▆▅ │
│                                 │
│ Por nó:                         │
│ Marabá-Centro      R$ 8.200 44%│
│ Parauapebas-01     R$ 6.750 37%│
│ Açailândia-01      R$ 3.500 19%│
│                                 │
│ [📤 Exportar CSV]               │
└─────────────────────────────────┘
```

---

### 6. Alertas / Notificações

```
┌─────────────────────────────────┐
│ Alertas                [Config] │
├─────────────────────────────────┤
│                                 │
│ ⚠️ ATIVO                        │
│ Açailândia-01 offline           │
│ Há 18 minutos · 14:15           │
│ [Ver nó] [Dispensar]           │
│                                 │
│ ─────────────────────────────   │
│                                 │
│ ✅ Resolvido — 18/05 22:03      │
│ Marabá-Centro voltou online     │
│ Duração: 3 minutos              │
│                                 │
│ ─────────────────────────────   │
│                                 │
│ ℹ️ 18/05 09:30                  │
│ Parauapebas-01: vouchers baixos │
│ 12 vouchers de 24h disponíveis  │
│                                 │
└─────────────────────────────────┘
```

---

### 7. Perfil e Configurações

```
┌─────────────────────────────────┐
│ Minha conta                     │
├─────────────────────────────────┤
│ [Avatar] João Silva             │
│          joao@exemplo.com       │
│          Plano: Pro             │
├─────────────────────────────────┤
│ Workspace: Provedor XYZ         │
│                                 │
│ Notificações push               │
│   Nó offline        [🟢 ON]    │
│   Pagamento         [🟢 ON]    │
│   Vouchers baixos   [⚫ OFF]   │
│   Metas de receita  [🟢 ON]    │
│                                 │
│ Tema: [Escuro ▼]                │
│                                 │
│ [Alterar senha]                 │
│ [Sair]                         │
│ [Versão 1.2.0]                  │
└─────────────────────────────────┘
```

---

## Push Notifications (Firebase Cloud Messaging)

### Tipos de Notificação

| Evento | Título | Corpo | Ação ao tocar |
|---|---|---|---|
| Nó offline | "⚠️ {nome} offline" | "Seu nó ficou offline há {X} min" | Abre detalhes do nó |
| Nó voltou | "✅ {nome} online" | "Seu nó voltou ao ar" | Abre detalhes do nó |
| Vouchers baixos | "🎟️ Estoque baixo" | "{N} vouchers de {plano} restando" | Abre gestão de vouchers |
| Meta atingida | "🎉 Meta atingida!" | "R$ {valor} de receita hoje!" | Abre financeiro |
| Pagamento alto | "💰 R$ {valor}" | "Pagamento aprovado em {nó}" | Abre financeiro |

### Configuração de Alertas (in-app)
- Por tipo de evento: ativar/desativar
- Silenciar por período: "Não perturbe das 22:00 às 07:00"
- Por nó: ativar/desativar alertas de nós específicos
- Limiar personalizado: alertar apenas pagamentos > R$ X

---

## Realtime via Supabase

```dart
// Escutar mudanças de status dos nós em tempo real
supabase
  .from('nodes')
  .stream(primaryKey: ['id'])
  .eq('tenant_id', currentTenantId)
  .listen((data) {
    // Atualiza UI
  });

// Escutar novos eventos
supabase
  .from('node_events')
  .on(SupabaseEventTypes.insert, (payload) {
    // Nova atividade
  })
  .subscribe();
```

---

## Navegação Bottom Bar

| Tab | Ícone | Conteúdo |
|---|---|---|
| Home | 🏠 | Dashboard geral |
| Nós | 📍 | Lista e detalhes dos nós |
| Financeiro | 💰 | Receita e relatórios |
| Alertas | 🔔 | Central de alertas (badge com contagem) |
| Conta | 👤 | Perfil e configurações |

---

## Performance e UX

- Skeleton loading em todas as telas (nunca spinner branco puro)
- Pull-to-refresh em todas as listas
- Haptic feedback em ações destrutivas (ban, disconnect)
- Dados cacheados localmente com `hive` (funciona offline com último estado)
- Animações: 300ms, curva ease-out (padrão Material 3)
- Suporte a gestos: swipe para voltar (iOS), swipe em cards para ações rápidas

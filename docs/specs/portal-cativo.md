# Spec: Portal Cativo (Captive Portal)

## Visão Geral

O Portal Cativo é a interface que o usuário final vê ao conectar ao Wi-Fi e tentar acessar a internet. É o produto mais crítico do sistema — precisa ser ultra-rápido, ultra-simples e funcionar em conexões lentas e dispositivos antigos.

**Stack:** SvelteKit + Vite + TailwindCSS (build local, sem CDN)

---

## Requisitos de Performance

| Métrica | Meta |
|---|---|
| Bundle total (JS + CSS) | < 50KB gzipped |
| LCP (Largest Contentful Paint) | < 2s em 3G |
| FID (First Input Delay) | < 100ms |
| Suporte a browsers | Chrome 70+, Safari 12+, Firefox 65+ |
| Suporte a dispositivos | Android 7+, iOS 12+, qualquer desktop |
| Funciona sem JavaScript? | Não (mas degrada graciosamente com mensagem) |

---

## Fluxo de Telas

```
[Abertura]
     │
     ▼
[1. Tela de Boas-vindas]
   - Logo do hotspot (customizável)
   - Nome do hotspot
   - Status: "Você está no walled garden"
   - Botão "Ver planos de acesso"
     │
     ├──► [Tem voucher?] ──► [2b. Inserir Voucher]
     │                              │
     ▼                              ▼
[2a. Seleção de Plano]         [Validação]
   - Cards de planos                │
   - Preço em destaque         ┌────┴────┐
   - Duração clara             │         │
   - Badge "Recomendado"    Inválido  Válido
     │                        │         │
     ▼                        │         ▼
[3. Dados do Usuário] ◄───────┘    [Liberar Acesso]
   - Nome (opcional, se config)
   - (MAC já capturado pela URL)
     │
     ▼
[4. Geração do PIX]
   - QR Code grande e claro
   - Código copia-e-cola
   - Valor e descrição
   - Progresso: "Aguardando pagamento..."
   - Cancelar
     │
     ▼ (webhook confirmado → SSE)
[5. Acesso Liberado!]
   - Animação de sucesso
   - "Aproveite X horas de internet"
   - Countdown do tempo restante
   - Botão "Começar a navegar"
```

---

## Telas em Detalhe

### Tela 1 — Boas-vindas

```
┌─────────────────────────────┐
│                             │
│      [LOGO DO HOTSPOT]      │
│                             │
│    Bem-vindo ao             │
│    🌐 Astrolink Wi-Fi       │
│                             │
│  Conecte-se à internet em   │
│  poucos segundos. Sem       │
│  cadastro complicado.       │
│                             │
│  ┌───────────────────────┐  │
│  │   Ver Planos de Acesso │  │
│  └───────────────────────┘  │
│                             │
│  Já tem um voucher?         │
│  ──────── Inserir código ─  │
│                             │
└─────────────────────────────┘
```

**Comportamento:**
- Logo e nome vêm de `GET /api/settings` (com cache 5min no localStorage)
- Se usuário já tem sessão ativa: pular direto para Tela 5

### Tela 2a — Seleção de Plano

```
┌─────────────────────────────┐
│  ← Voltar                   │
│                             │
│  Escolha seu plano          │
│                             │
│  ┌───────────────────────┐  │
│  │ ⭐ RECOMENDADO        │  │
│  │ Acesso 24 Horas       │  │
│  │ ⏱ 24 horas           │  │
│  │              R$ 15,00 │  │
│  └───────────────────────┘  │
│                             │
│  ┌───────────────────────┐  │
│  │ Acesso 1 Hora         │  │
│  │ ⏱ 1 hora             │  │
│  │               R$ 5,00 │  │
│  └───────────────────────┘  │
│                             │
│  ┌───────────────────────┐  │
│  │ Pacote Semanal        │  │
│  │ ⏱ 7 dias             │  │
│  │              R$ 50,00 │  │
│  └───────────────────────┘  │
│                             │
└─────────────────────────────┘
```

**Comportamento:**
- Planos vêm de `GET /api/planos` (apenas ativos, ordenados por preço)
- Plano marcado como `recomendado=true` recebe badge
- Tap no card → avança para Tela 3
- Skeleton loading enquanto carrega

### Tela 2b — Inserir Voucher

```
┌─────────────────────────────┐
│  ← Voltar                   │
│                             │
│  🎟️ Inserir Voucher         │
│                             │
│  Digite o código do seu     │
│  cartão de acesso:          │
│                             │
│  ┌───────────────────────┐  │
│  │  _ _ _ _  _ _ _ _    │  │
│  └───────────────────────┘  │
│  (autocaps, sem espaços)    │
│                             │
│  ┌───────────────────────┐  │
│  │       Resgatar        │  │
│  └───────────────────────┘  │
│                             │
│  ⚠️ Código inválido          │  ← visível apenas se erro
│                             │
└─────────────────────────────┘
```

**Comportamento:**
- Máscara automática: 4 chars + espaço + 4 chars
- POST `/api/voucher/resgatar` → resposta imediata
- Sucesso: animação → Tela 5
- Erro: shake animation + mensagem clara

### Tela 3 — Dados do Usuário (opcional)

Só aparece se configurado `coleta_nome = true` nas Settings.

```
┌─────────────────────────────┐
│  ← Voltar                   │
│                             │
│  Quase lá! 😊               │
│                             │
│  Como podemos te chamar?    │
│  (Opcional)                 │
│                             │
│  ┌───────────────────────┐  │
│  │  Seu nome             │  │
│  └───────────────────────┘  │
│                             │
│  ┌───────────────────────┐  │
│  │     Continuar →       │  │
│  └───────────────────────┘  │
│                             │
│  [Pular]                    │
│                             │
└─────────────────────────────┘
```

### Tela 4 — Pagamento PIX

```
┌─────────────────────────────┐
│  Pague com PIX              │
│                             │
│  ┌─────────────────────┐   │
│  │                     │   │
│  │    [QR CODE]        │   │
│  │    (256x256)        │   │
│  │                     │   │
│  └─────────────────────┘   │
│                             │
│  Acesso 24 Horas — R$15,00  │
│                             │
│  ┌───────────────────────┐  │
│  │ 📋 Copiar código PIX  │  │
│  └───────────────────────┘  │
│                             │
│  ⏳ Aguardando pagamento...  │
│  ████████░░░░░░░░ (spinner) │
│                             │
│  Expira em 14:32            │
│                             │
│  [Cancelar]                 │
└─────────────────────────────┘
```

**Comportamento:**
- QR code gerado pelo backend (imagem base64 ou lib client-side)
- Botão "Copiar" usa `navigator.clipboard.writeText()`
- Feedback de sucesso ao copiar: "Copiado! ✓"
- Confirmação via **Server-Sent Events** (SSE): `GET /api/pix/aguardar/:txid`
  - Backend envia `data: {"status": "approved"}` quando webhook chegar
  - Fallback: polling a cada 5s se SSE não suportado
- Contador regressivo: 15 minutos (expiração do PIX)
- Cancelar: volta para Tela 2a

### Tela 5 — Acesso Liberado

```
┌─────────────────────────────┐
│                             │
│         ✅                  │
│                             │
│  Acesso liberado!           │
│                             │
│  Aproveite suas             │
│  24 horas de internet 🚀   │
│                             │
│  Tempo restante:            │
│  ┌───────────────────────┐  │
│  │  23h 59m 45s          │  │  ← countdown em tempo real
│  └───────────────────────┘  │
│                             │
│  ┌───────────────────────┐  │
│  │  Começar a navegar →  │  │
│  └───────────────────────┘  │
│                             │
│  💡 Salve este site para    │
│  ver seu tempo restante     │
│  a qualquer hora.           │
│                             │
└─────────────────────────────┘
```

**Comportamento:**
- Animação de confete ao entrar
- Countdown calculado com base em `fim_acesso` (UTC do backend)
- Botão "Começar a navegar" → redireciona para URL configurada (padrão: google.com)
- Link "Salvar" → instalar PWA ou favoritar URL

---

## White-Label (Customização por Nó)

Todos os elementos visuais são configuráveis via `GET /api/settings`:

```json
{
  "hotspot_nome": "Wi-Fi Pousada Recanto Verde",
  "hotspot_logo_url": "/uploads/logo.png",
  "cor_primaria": "#2ECC71",
  "cor_secundaria": "#27AE60",
  "cor_fundo": "#0D1117",
  "mensagem_boas_vindas": "Seja bem-vindo! Conecte-se e aproveite.",
  "url_pos_conexao": "https://pousadarecantoverde.com.br",
  "coleta_nome": false,
  "mostrar_velocidade": true
}
```

**CSS Variables aplicadas dinamicamente:**
```css
:root {
  --color-primary: var(--hotspot-primary, #06B6D4);
  --color-secondary: var(--hotspot-secondary, #0E7490);
  --color-bg: var(--hotspot-bg, #0F172A);
}
```

---

## Endpoints da API Utilizados

| Método | Endpoint | Descrição |
|---|---|---|
| GET | `/api/settings` | Configurações de white-label |
| GET | `/api/planos` | Lista planos ativos |
| POST | `/api/pix/gerar` | Criar cobrança PIX |
| GET | `/api/pix/aguardar/:txid` | SSE aguardar confirmação |
| POST | `/api/voucher/resgatar` | Resgatar voucher |
| GET | `/api/sessao/status` | Verificar sessão ativa pelo MAC |

---

## Captura do MAC Address

O MAC do usuário chega via query param injetada pelo OpenNDS na URL de redirecionamento:

```
https://hotspot.local/?mac=AA:BB:CC:DD:EE:FF&ip=192.168.1.100&token=abc123
```

Lógica de fallback:
1. `URLSearchParams` da URL atual
2. Se não encontrado: `mac = "00:00:00:00:00:00"` (modo demo/desenvolvimento)

---

## PWA (Progressive Web App)

```json
// manifest.json
{
  "name": "Astrolink Wi-Fi",
  "short_name": "Wi-Fi",
  "start_url": "/",
  "display": "standalone",
  "background_color": "#0F172A",
  "theme_color": "#06B6D4",
  "icons": [...]
}
```

**Service Worker:** Cache de assets estáticos (JS, CSS, logo). A página de status de sessão funciona offline com última informação cacheada.

---

## Acessibilidade

- ARIA labels em todos os botões e inputs
- Contraste mínimo WCAG AA (4.5:1)
- Navegável por teclado
- Mensagens de erro lidas por screen readers (`role="alert"`)
- Textos escaláveis (rem, não px)

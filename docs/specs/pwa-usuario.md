# Spec: PWA do Usuário Final

## Visão Geral

Progressive Web App que o usuário final instala após comprar acesso. Permite verificar o tempo restante, comprar mais tempo e receber notificações antes de expirar — sem precisar conectar ao portal cativo novamente.

**URL:** Domínio do operador (ex: `wifi.pousadarecantoverde.com.br`) ou `connect.astrolink.app`
**Stack:** SvelteKit PWA + Service Worker
**Instalação:** "Adicionar à tela inicial" no iOS/Android

---

## Por que isso importa

**Para o usuário:**
- Sem precisar acessar o portal cativo para verificar o tempo
- Comprar mais tempo é mais fácil → mais conversões
- Notificação antes de expirar → melhor experiência

**Para o operador:**
- Aumento de ~20–30% na receita por renovações facilitadas
- Redução de suporte ("quanto tempo tenho?")
- Engajamento com o estabelecimento

---

## Telas

### 1. Tela Principal (instalado, com sessão ativa)

```
┌────────────────────────────────┐
│ 🌐 Wi-Fi Pousada Recanto Verde │
│                                │
│         ✅ Conectado           │
│                                │
│      Tempo restante:           │
│   ┌──────────────────────┐    │
│   │                      │    │
│   │  🕐  18h 32m 45s     │    │  ← countdown atualiza a cada segundo
│   │                      │    │
│   └──────────────────────┘    │
│                                │
│   Plano: Acesso 24 Horas       │
│   Expira: amanhã às 14:30      │
│                                │
│   ──────────────────────────   │
│                                │
│   Dados consumidos hoje:       │
│   ↓ 1.2 GB  ↑ 234 MB          │
│                                │
│   ┌──────────────────────┐    │
│   │  + Adicionar tempo   │    │
│   └──────────────────────┘    │
│                                │
│   [Histórico de compras]       │
│                                │
└────────────────────────────────┘
```

### 2. Tela Principal (sessão expirada)

```
┌────────────────────────────────┐
│ 🌐 Wi-Fi Pousada Recanto Verde │
│                                │
│         ❌ Sem acesso          │
│                                │
│   Seu acesso expirou às        │
│   14:30 de 19/05/2025          │
│                                │
│   ┌──────────────────────┐    │
│   │  Renovar acesso →    │    │
│   └──────────────────────┘    │
│                                │
│   Planos disponíveis:          │
│   • 1 hora — R$ 5,00           │
│   • 24 horas — R$ 15,00        │
│   • 7 dias — R$ 50,00          │
│                                │
└────────────────────────────────┘
```

### 3. Comprar Mais Tempo

```
┌────────────────────────────────┐
│ ← Adicionar tempo              │
│                                │
│  Escolha um plano:             │
│                                │
│  ┌──────────────────────────┐ │
│  │ + 1 hora         R$ 5,00 │ │
│  └──────────────────────────┘ │
│                                │
│  ┌──────────────────────────┐ │
│  │ ⭐ + 24 horas   R$ 15,00 │ │  ← recomendado
│  └──────────────────────────┘ │
│                                │
│  ┌──────────────────────────┐ │
│  │ + 7 dias        R$ 50,00 │ │
│  └──────────────────────────┘ │
│                                │
│  Novo total após compra:       │
│  18h 32m + 24h = 42h 32m       │
│  (tempo é acumulado!)          │
│                                │
└────────────────────────────────┘
     ↓ (ao selecionar)
┌────────────────────────────────┐
│ Pague com PIX                  │
│                                │
│       [QR CODE]                │
│                                │
│  R$ 15,00 — +24 horas          │
│                                │
│  [📋 Copiar código PIX]        │
│                                │
│  ⏳ Aguardando pagamento...    │
│  Expira em: 14:45             │
│                                │
└────────────────────────────────┘
     ↓ (pagamento confirmado)
┌────────────────────────────────┐
│        🎉 Pago!                │
│                                │
│  +24 horas adicionados!        │
│  Novo total: 42h 32m 45s       │
│                                │
└────────────────────────────────┘
```

### 4. Histórico de Compras

```
┌────────────────────────────────┐
│ ← Histórico                    │
│                                │
│  Suas compras:                 │
│                                │
│  ┌──────────────────────────┐ │
│  │ 19/05/2025 14:30         │ │
│  │ Acesso 24 Horas          │ │
│  │ R$ 15,00 ✅ Pago via PIX │ │
│  └──────────────────────────┘ │
│                                │
│  ┌──────────────────────────┐ │
│  │ 15/05/2025 10:00         │ │
│  │ Acesso 24 Horas          │ │
│  │ R$ 15,00 ✅ Pago via PIX │ │
│  └──────────────────────────┘ │
│                                │
│  Total gasto aqui:             │
│  R$ 30,00 (2 compras)          │
│                                │
└────────────────────────────────┘
```

---

## Identificação do Usuário

O usuário é identificado pelo **MAC address** armazenado no localStorage durante a primeira conexão pelo portal cativo.

```typescript
// src/lib/store.ts
import { browser } from '$app/environment'

export const mac = {
  get: () => browser ? localStorage.getItem('astrolink_mac') : null,
  set: (value: string) => browser && localStorage.setItem('astrolink_mac', value),
}

// Durante o redirect do portal cativo, salvar o MAC:
// http://wifi.pousada.com/?mac=AA:BB:CC:DD:EE:FF&...
// O portal salva no localStorage antes de redirecionar para o sucesso
```

---

## Notificações Push (Web Push API)

```typescript
// Solicitar permissão
async function solicitarPermissaoNotificacao() {
  if (!('Notification' in window)) return

  const permission = await Notification.requestPermission()
  if (permission === 'granted') {
    const registration = await navigator.serviceWorker.ready
    const subscription = await registration.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: PUBLIC_VAPID_KEY,
    })

    // Enviar subscription para o backend
    await api.post('/api/push/subscribe', {
      mac: mac.get(),
      subscription: subscription.toJSON(),
    })
  }
}
```

**Notificações enviadas:**
- 30 minutos antes de expirar: "⏰ Seu Wi-Fi expira em 30 minutos!"
- Ao expirar: "❌ Seu acesso Wi-Fi expirou. Renove agora."
- (Somente se o usuário deu permissão)

---

## Service Worker (Offline)

```javascript
// service-worker.js
const CACHE_NAME = 'astrolink-pwa-v1'
const STATIC_ASSETS = ['/', '/historico', '/manifest.json', '/icons/...']

// Instalar: cachear assets estáticos
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => cache.addAll(STATIC_ASSETS))
  )
})

// Fetch: cache first para assets estáticos, network first para API
self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url)

  if (url.pathname.startsWith('/api/')) {
    // API: network first, fallback para cache
    event.respondWith(
      fetch(event.request)
        .then((response) => {
          caches.open(CACHE_NAME).then((c) => c.put(event.request, response.clone()))
          return response
        })
        .catch(() => caches.match(event.request))
    )
  } else {
    // Assets: cache first
    event.respondWith(
      caches.match(event.request).then((cached) => cached || fetch(event.request))
    )
  }
})

// Push: receber notificação
self.addEventListener('push', (event) => {
  const data = event.data?.json()
  event.waitUntil(
    self.registration.showNotification(data.title, {
      body: data.body,
      icon: '/icons/icon-192.png',
      badge: '/icons/badge-72.png',
      data: { url: data.url },
    })
  )
})
```

---

## Web App Manifest

```json
{
  "name": "Wi-Fi Pousada Recanto Verde",
  "short_name": "Wi-Fi",
  "description": "Gerencie seu acesso Wi-Fi",
  "start_url": "/?source=pwa",
  "display": "standalone",
  "orientation": "portrait",
  "background_color": "#0F172A",
  "theme_color": "#06B6D4",
  "categories": ["utilities"],
  "icons": [
    { "src": "/icons/icon-72.png",  "sizes": "72x72",   "type": "image/png" },
    { "src": "/icons/icon-96.png",  "sizes": "96x96",   "type": "image/png" },
    { "src": "/icons/icon-128.png", "sizes": "128x128", "type": "image/png" },
    { "src": "/icons/icon-192.png", "sizes": "192x192", "type": "image/png", "purpose": "maskable" },
    { "src": "/icons/icon-512.png", "sizes": "512x512", "type": "image/png", "purpose": "maskable" }
  ],
  "screenshots": [
    { "src": "/screenshots/home.png", "sizes": "390x844", "type": "image/png", "form_factor": "narrow" }
  ]
}
```

---

## Integração com White-Label

O PWA herda o white-label do nó:
- Nome no manifest vem de `GET /api/settings` → `hotspot_nome`
- Cores do tema vêm de `cor_primaria` e `cor_fundo`
- Logo do ícone vem de `hotspot_logo_url`
- `theme_color` = `cor_primaria`

```typescript
// src/routes/+layout.ts
export const load = async ({ fetch }) => {
  const settings = await fetch('/api/settings').then(r => r.json())
  return { settings }
}

// src/app.html — atualiza meta tags dinamicamente
// (ou usar +layout.server.ts para SSR)
```

---

## Prompt de Instalação

```svelte
<!-- src/lib/components/InstallPrompt.svelte -->
<script lang="ts">
  let deferredPrompt: BeforeInstallPromptEvent | null = null
  let showBanner = false

  onMount(() => {
    window.addEventListener('beforeinstallprompt', (e) => {
      e.preventDefault()
      deferredPrompt = e as BeforeInstallPromptEvent
      // Mostrar banner após 30s de uso
      setTimeout(() => showBanner = true, 30_000)
    })
  })

  async function instalar() {
    if (!deferredPrompt) return
    deferredPrompt.prompt()
    const { outcome } = await deferredPrompt.userChoice
    if (outcome === 'accepted') showBanner = false
    deferredPrompt = null
  }
</script>

{#if showBanner}
  <div class="fixed bottom-0 left-0 right-0 bg-slate-800 p-4 shadow-2xl">
    <p class="text-sm text-white mb-2">
      💡 Salve este app para ver seu tempo restante a qualquer hora
    </p>
    <div class="flex gap-2">
      <button onclick={instalar} class="btn-primary flex-1">Adicionar à tela inicial</button>
      <button onclick={() => showBanner = false} class="btn-ghost">Agora não</button>
    </div>
  </div>
{/if}
```

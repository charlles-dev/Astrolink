# Portal Cativo Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Build the local SvelteKit captive portal approved in `docs/superpowers/specs/2026-05-21-portal-cativo-design.md`.

**Architecture:** The portal is a standalone SvelteKit app in `portal/` that consumes the existing Go node endpoints. State is kept small and local to the route through typed API helpers, formatting utilities, and focused screen components.

**Tech Stack:** SvelteKit, Vite, TypeScript, TailwindCSS v4, Vitest, Testing Library, Playwright/browser verification.

---

### Task 1: Scaffold Portal App

**Files:**
- Create: `portal/package.json`
- Create: `portal/tsconfig.json`
- Create: `portal/vite.config.ts`
- Create: `portal/svelte.config.js`
- Create: `portal/src/app.css`
- Create: `portal/src/app.html`
- Create: `portal/src/routes/+page.svelte`
- Create: `portal/src/lib/types.ts`
- Create: `portal/src/lib/api.ts`
- Create: `portal/src/lib/format.ts`
- Create: `portal/src/lib/format.test.ts`

- [x] **Step 1: Write failing formatter tests**

Create `portal/src/lib/format.test.ts`:

```ts
import { describe, expect, it } from 'vitest'
import { formatCurrency, formatCountdown, maskVoucherCode } from './format'

describe('formatCurrency', () => {
  it('formats backend decimal strings as BRL', () => {
    expect(formatCurrency('15.00')).toBe('R$ 15,00')
  })
})

describe('formatCountdown', () => {
  it('formats seconds as compact hours and minutes', () => {
    expect(formatCountdown(86399)).toBe('23h 59m')
  })
})

describe('maskVoucherCode', () => {
  it('uppercases and groups voucher codes', () => {
    expect(maskVoucherCode('test1234')).toBe('TEST-1234')
  })
})
```

- [x] **Step 2: Run the test and verify RED**

Run:
```powershell
cd portal
npm test -- --run src/lib/format.test.ts
```

Expected: FAIL because `portal/` and `format.ts` do not exist yet.

- [x] **Step 3: Scaffold SvelteKit**

Create the app files listed above with scripts:

```json
{
  "scripts": {
    "dev": "vite --host 127.0.0.1",
    "build": "vite build",
    "preview": "vite preview --host 127.0.0.1",
    "test": "vitest run",
    "check": "svelte-check --tsconfig ./tsconfig.json"
  }
}
```

- [x] **Step 4: Implement format utilities**

Implement `formatCurrency`, `formatCountdown`, `formatDuration`, and `maskVoucherCode` in `portal/src/lib/format.ts`.

- [x] **Step 5: Run the formatter test and verify GREEN**

Run:
```powershell
cd portal
npm test -- --run src/lib/format.test.ts
```

Expected: PASS.

### Task 2: API Client

**Files:**
- Modify: `portal/src/lib/types.ts`
- Modify: `portal/src/lib/api.ts`
- Create: `portal/src/lib/api.test.ts`

- [x] **Step 1: Write failing API client tests**

Create `portal/src/lib/api.test.ts` that mocks `fetch` and verifies:

```ts
import { afterEach, describe, expect, it, vi } from 'vitest'
import { APIError, createApiClient } from './api'

afterEach(() => vi.restoreAllMocks())

describe('createApiClient', () => {
  it('loads plans from /api/planos', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({
      planos: [{ id: 1, nome: 'Acesso 24 Horas', preco: '15.00', duracao_minutos: 1440, duracao_formatada: '24 horas', dados_mb: null, velocidade_down: 10, velocidade_up: 5, recomendado: true, ativo: true, visivel_portal: true, ordem: 1 }]
    }), { status: 200 })))

    const api = createApiClient('')
    const result = await api.getPlanos()

    expect(result.planos[0].nome).toBe('Acesso 24 Horas')
    expect(fetch).toHaveBeenCalledWith('/api/planos', expect.objectContaining({ method: 'GET' }))
  })

  it('throws APIError with backend message', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({
      erro: 'nao_encontrado',
      mensagem: 'voucher nao encontrado'
    }), { status: 404 })))

    const api = createApiClient('')

    await expect(api.resgatarVoucher({ codigo: 'XXXX-9999', mac: 'AA:BB:CC:DD:EE:FF', ip: '192.168.1.50' })).rejects.toMatchObject({
      status: 404,
      code: 'nao_encontrado',
      message: 'voucher nao encontrado'
    } satisfies Partial<APIError>)
  })
})
```

- [x] **Step 2: Run API tests and verify RED**

Run:
```powershell
cd portal
npm test -- --run src/lib/api.test.ts
```

Expected: FAIL until `api.ts` exports the requested client.

- [x] **Step 3: Implement typed API client**

Implement `APIError`, `createApiClient`, and methods for:

```ts
getSettings()
getPlanos()
getSessaoStatus(mac: string)
gerarPix(body)
getPixStatus(txid: string)
resgatarVoucher(body)
```

- [x] **Step 4: Run API tests and verify GREEN**

Run:
```powershell
cd portal
npm test -- --run src/lib/api.test.ts
```

Expected: PASS.

### Task 3: Portal UI Components

**Files:**
- Create: `portal/src/lib/components/ErrorMessage.svelte`
- Create: `portal/src/lib/components/PlanCard.svelte`
- Create: `portal/src/lib/components/PortalShell.svelte`
- Create: `portal/src/lib/components/WelcomeScreen.svelte`
- Create: `portal/src/lib/components/PlanSelection.svelte`
- Create: `portal/src/lib/components/VoucherScreen.svelte`
- Create: `portal/src/lib/components/PixScreen.svelte`
- Create: `portal/src/lib/components/SuccessScreen.svelte`
- Create: `portal/src/lib/components/PlanCard.test.ts`

- [x] **Step 1: Write failing PlanCard component test**

Create `portal/src/lib/components/PlanCard.test.ts`:

```ts
import { render, screen } from '@testing-library/svelte'
import { describe, expect, it, vi } from 'vitest'
import PlanCard from './PlanCard.svelte'

describe('PlanCard', () => {
  it('renders plan name, price, duration and recommended badge', () => {
    render(PlanCard, {
      props: {
        plano: {
          id: 1,
          nome: 'Acesso 24 Horas',
          descricao: 'Um dia completo',
          preco: '15.00',
          duracao_minutos: 1440,
          duracao_formatada: '24 horas',
          dados_mb: null,
          velocidade_down: 10,
          velocidade_up: 5,
          recomendado: true,
          ativo: true,
          visivel_portal: true,
          ordem: 1
        },
        onSelect: vi.fn()
      }
    })

    expect(screen.getByText('Acesso 24 Horas')).toBeInTheDocument()
    expect(screen.getByText('R$ 15,00')).toBeInTheDocument()
    expect(screen.getByText('RECOMENDADO')).toBeInTheDocument()
  })
})
```

- [x] **Step 2: Run component test and verify RED**

Run:
```powershell
cd portal
npm test -- --run src/lib/components/PlanCard.test.ts
```

Expected: FAIL until component/test setup exists.

- [x] **Step 3: Implement components**

Build components to match the approved visual companion:

- dark welcome screen
- light plan/voucher/pix/success screens
- large tap targets
- no nested cards
- no marketing hero
- error messages with `role="alert"`

- [x] **Step 4: Run component test and verify GREEN**

Run:
```powershell
cd portal
npm test -- --run src/lib/components/PlanCard.test.ts
```

Expected: PASS.

### Task 4: Route Flow

**Files:**
- Modify: `portal/src/routes/+page.svelte`
- Modify: `portal/src/app.css`

- [x] **Step 1: Compose app flow**

Implement the route state machine:

```ts
type Step = 'welcome' | 'plans' | 'voucher' | 'pix' | 'success'
```

Behavior:

- on mount parse query params
- load cached settings, then refresh settings
- check session status
- load plans when entering plans
- create PIX when selecting plan
- redeem voucher and show success
- copy PIX code with fallback message
- SSE first, polling fallback for PIX

- [x] **Step 2: Add responsive CSS**

Use `portal/src/app.css` for global tokens, shell layout, mobile viewport, buttons, cards, forms, QR code, countdown, and accessibility focus states.

- [x] **Step 3: Run Svelte check**

Run:
```powershell
cd portal
npm run check
```

Expected: PASS.

### Task 5: Verification

**Files:**
- Verify: `portal/**`
- Verify: `node/**`

- [x] **Step 1: Run portal tests**

Run:
```powershell
cd portal
npm test
```

Expected: PASS.

- [x] **Step 2: Build portal**

Run:
```powershell
cd portal
npm run build
```

Expected: PASS.

- [x] **Step 3: Verify backend still passes**

Run:
```powershell
cd node
go test ./...
go build ./cmd/server
```

Expected: PASS.

- [x] **Step 4: Browser verification**

Run the Go backend and portal dev server, then inspect:

- desktop viewport
- mobile viewport around 390x844
- welcome to plans path
- voucher success path with `TEST-1234`
- PIX generation path

Expected: no blank screens, no overflow, and the visible UI matches the approved companion direction.


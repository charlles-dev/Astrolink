# Documentacao do Frontend

## Stack Atual

O frontend implementado nesta fase e apenas o portal cativo:

- SvelteKit
- Vite
- TypeScript
- CSS local em `portal/src/app.css`
- Vitest e Testing Library

Codigo fonte: `portal/`

## Estrutura

```text
portal/
  src/routes/+page.svelte
  src/lib/api.ts
  src/lib/types.ts
  src/lib/format.ts
  src/lib/components/
    PortalShell.svelte
    WelcomeScreen.svelte
    PlanSelection.svelte
    PlanCard.svelte
    VoucherScreen.svelte
    PixScreen.svelte
    SuccessScreen.svelte
    ErrorMessage.svelte
```

## URL de Desenvolvimento

```text
http://127.0.0.1:5173/?mac=AA:BB:CC:DD:EE:FF&ip=192.168.1.50&token=test
```

O portal espera `mac`, `ip` e `token` na query string, do mesmo jeito que o
OpenNDS injeta em producao.

## Telas Implementadas

- Boas-vindas
- Selecao de planos
- Voucher
- PIX demonstrativo
- Sucesso/acesso liberado

## APIs Consumidas

- `GET /api/settings`
- `GET /api/planos`
- `GET /api/sessao/status`
- `POST /api/pix/gerar`
- `GET /api/pix/status/:txid`
- `GET /api/pix/aguardar/:txid`
- `POST /api/voucher/resgatar`

## Proximas Pendencias do Portal

- Refinar estados vazios, loading e erro offline.
- Finalizar integracao PIX real quando o backend estiver pronto.
- Adicionar PWA/service worker.
- Cobrir fluxo completo com Playwright.
- Preparar personalizacao visual via admin local.

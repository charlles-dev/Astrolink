# Portal Cativo Design Spec

**Status:** Approved from visual companion feedback.

**Scope:** Build only the local captive portal. The Cloud admin, public map, marketplace, mobile app, desktop app, and local admin panel stay out of this phase.

## Goal

Create a lightweight SvelteKit captive portal that lets a Wi-Fi user get online quickly through one of two paths:

- select a plan and generate a PIX charge
- enter a voucher code and immediately unlock access

If the backend reports an active session for the device MAC, the portal skips directly to the success/status screen.

## Product Flow

The MVP has five screens:

1. **Welcome**
   - Loads white-label settings from `GET /api/settings`.
   - Reads `mac`, `ip`, and `token` from URL query params.
   - Calls `GET /api/sessao/status?mac=...`.
   - Shows hotspot name, walled garden status, primary action for plans, and secondary action for voucher.

2. **Plan Selection**
   - Calls `GET /api/planos`.
   - Shows active portal-visible plans as tappable cards.
   - Highlights `recomendado=true`.
   - Selecting a plan advances to optional user data or directly to PIX.

3. **Voucher**
   - Accepts uppercase voucher codes in the `XXXX-XXXX` shape, while still allowing prefixes such as `TEST-1234`.
   - Calls `POST /api/voucher/resgatar`.
   - On success, goes to success screen.
   - On failure, shows a short inline error with clear retry affordance.

4. **PIX**
   - Calls `POST /api/pix/gerar`.
   - Shows QR code, copy-and-paste code, amount, selected plan, countdown, copy button, and cancel action.
   - Uses `GET /api/pix/aguardar/:txid` SSE when available.
   - Falls back to `GET /api/pix/status/:txid` polling every 5 seconds.

5. **Success / Active Session**
   - Shows access granted state, selected plan/session name, remaining time countdown, and "Comecar a navegar".
   - Redirects to `url_pos_conexao` from settings.

## Visual Direction

Use the approved visual companion direction:

- Mobile-first, single-column, one task per screen.
- Welcome screen uses a dark branded background for identity.
- Transactional screens use light surfaces for maximum legibility.
- Large tap targets, restrained cards, and no marketing-page layout.
- White-label colors come from settings via CSS variables:
  - `--color-primary`
  - `--color-secondary`
  - `--color-bg`
- Default palette:
  - background dark: `#0F172A`
  - primary: `#38BDF8`
  - secondary/accent: `#0EA5A8`
  - success: `#22C55E`
  - light surface: `#F8FAFC`
- Use rounded corners around 14-18px for portal controls and plan cards.
- No decorative bokeh/orbs, no landing-page hero, no nested cards.

## Architecture

Create a new `portal/` SvelteKit app with Vite and TailwindCSS.

Suggested files:

- `portal/src/lib/api.ts`: typed API client and `APIError`.
- `portal/src/lib/types.ts`: shared frontend API types.
- `portal/src/lib/state.ts`: Svelte stores for device, settings, plans, selected plan, PIX transaction, active session, and current step.
- `portal/src/lib/format.ts`: currency, duration, countdown, voucher mask.
- `portal/src/lib/components/`: focused UI components:
  - `PortalShell.svelte`
  - `WelcomeScreen.svelte`
  - `PlanSelection.svelte`
  - `VoucherScreen.svelte`
  - `PixScreen.svelte`
  - `SuccessScreen.svelte`
  - `PlanCard.svelte`
  - `ErrorMessage.svelte`
- `portal/src/routes/+page.svelte`: composition and state transitions.

## Data Flow

1. On mount, parse URL query params and store device info.
2. Load settings from localStorage cache if fresh, then refresh from API.
3. Check active session by MAC.
4. If session is active, show success/status screen.
5. Otherwise wait for user action:
   - plans path loads plans and creates PIX
   - voucher path posts voucher redemption
6. Success updates local session state and starts countdown.

## Error Handling

- API unavailable: show "Nao foi possivel conectar ao servidor local" and a retry button.
- No plans: show "Nenhum plano disponivel no momento" and voucher action.
- Invalid voucher: show backend message when available; otherwise "Codigo invalido ou ja utilizado".
- PIX expired: stop SSE/polling, show expired state, allow returning to plans.
- Clipboard failure: keep the PIX code visible and show "Copie manualmente".

## Performance And Accessibility

- Keep the bundle small; avoid charting, animation, and heavy UI libraries.
- Use CSS transitions only where they clarify state changes.
- Use semantic buttons and form labels.
- All error messages use `role="alert"`.
- Ensure keyboard navigation works for plan cards and actions.
- Include a `<noscript>` message explaining that JavaScript is required for this portal.

## Tests

Minimum automated checks:

- unit tests for voucher mask, duration formatting, and countdown formatting
- component test for plan card selection
- component test for voucher error state
- build check with `npm run build`

Browser verification:

- desktop viewport
- mobile viewport around 390x844
- welcome to plans
- voucher success path using `TEST-1234`
- PIX generation path with mocked/present backend response

## Explicit Non-Goals

- No Cloud admin.
- No local admin UI yet.
- No real Mercado Pago client behavior in the frontend beyond displaying backend response.
- No offline-first service worker in this first MVP unless it falls out naturally from SvelteKit setup.
- No public landing page or public marketing content.

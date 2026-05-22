# Admin Local Parallel Roadmap Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Finish the remaining local Astrolink admin scope, keeping admin cloud paused.

**Architecture:** The current local node already has the captive portal, OpenNDS auth/deauth, admin login, health, users, plans listing, and voucher generation. The remaining work should be split into backend route modules and frontend admin panels before parallel implementation, because the current hotspots (`handlers.go`, `store.go`, `AdminDashboard.svelte`) would otherwise create conflicts.

**Tech Stack:** Go 1.22, Fiber, Postgres/memory store, SvelteKit, Vitest, svelte-check.

---

## Current Baseline

Implemented:
- Captive portal settings, plans, session status, PIX demo, voucher redeem.
- OpenNDS controller with no-op and SSH-backed auth/deauth.
- Admin local login with JWT access token, refresh token, and local lockout.
- Admin health, plans list, users list, vouchers list/generate, disconnect user.
- Portal `/painel` UI for health, users, plans, and voucher generation.
- Wave 0 split completed in commit `782d5bd`: admin backend handlers are separated by domain, and the admin dashboard is split into panel components under `portal/src/lib/components/admin/`.
- Wave 1 auth/voucher foundation completed in commit `535ef86`: JWT access tokens, refresh/logout/me endpoints, protected admin routes, session persistence, stale-token panel handling, and advanced voucher generation controls.
- Wave 2 reporting/reliability completed: admin payment history with CSV export, demo payments provider abstraction, operational logs with CSV export, backup endpoint disabled in memory/dev, and a session-expiration job hook.
- Wave 3A polish completed: printable voucher sheet from `/painel`, Mercado Pago webhook signature validation with provider status reconciliation hook, development-only PIX approval endpoint, and protected restore validation that never executes destructive restore.
- Wave 3B live operations completed: protected admin SSE snapshots, live events panel on `/painel`, and best-effort audit logs for mutating local admin actions.
- Wave 3C hardening completed: Mercado Pago payment-detail provider, PDF-ready voucher sheet, and 5-failure local admin login lockout.

Out of scope for this roadmap:
- Admin cloud.
- Multi-tenant cloud sync.
- Mobile/desktop apps.

## Dependency Rules

The following files are shared bottlenecks and should not be edited by multiple workers at once:
- `node/internal/store/store.go`
- `node/internal/api/admin/handlers.go`
- `node/internal/infra/memory/store.go`
- `node/internal/infra/postgres/store.go`
- `portal/src/lib/types.ts`
- `portal/src/lib/api.ts`
- `portal/src/lib/components/AdminDashboard.svelte`
- `portal/src/routes/painel/+page.svelte`

Before broad parallel implementation, split these into smaller modules:
- Backend admin route modules under `node/internal/api/admin/`.
- Store capability interfaces or grouped method sections.
- Frontend admin panels under `portal/src/lib/components/admin/`.
- Route orchestration in `/painel` that composes independent panels.

## Parallel Work Lanes

### Lane A: Auth, Sessions, and Audit

Purpose:
- Replace temporary base64 admin token with real JWT access token.
- Add refresh/logout/me endpoints.
- Add admin middleware for protected routes.
- Log admin auth events and mutating admin actions.

Primary ownership:
- `node/internal/api/admin/auth.go`
- `node/internal/api/admin/middleware.go`
- `node/internal/auth/`
- `node/internal/store/store.go`
- `node/internal/infra/memory/store.go`
- `node/internal/infra/postgres/store.go`
- `node/migrations/000002_admin_auth_audit.up.sql`
- `node/migrations/000002_admin_auth_audit.down.sql`
- `portal/src/routes/painel/+page.svelte`
- `portal/src/lib/api.ts`
- `portal/src/lib/types.ts`

First cut:
- JWT access token with 8h expiry.
- Refresh token persisted in local store.
- `POST /admin/auth/refresh`
- `POST /admin/auth/logout`
- `GET /admin/auth/me`
- Middleware protecting all `/admin/*` routes except login/refresh.
- JWT should use HS256, `JWT_SECRET`, and claims for user, issued-at, and expiration.
- Refresh tokens should be opaque random values, stored as hashes, and rotated on refresh.
- Apply a local 5-failure login lockout window before adding heavier rate-limit infrastructure.

Deferred:
- Audit records for login success/failure, logout, voucher generation, disconnect.
- Role, node, and session-id claims with active-session validation on every access token request.
- TOTP 2FA.
- IP allowlist.
- Password change UI.
- RBAC beyond the single local admin user.
- `ADMIN_SENHA_HASH` production hardening.

### Lane B: Plans CRUD

Purpose:
- Let the local operator create, edit, enable/disable, and reorder plans from `/painel`.

Primary ownership:
- `node/internal/api/admin/planos.go`
- `node/internal/domain/planos/plano.go`
- `node/internal/store/store.go`
- `node/internal/infra/memory/store.go`
- `node/internal/infra/postgres/store.go`
- `portal/src/lib/components/admin/AdminPlansPanel.svelte`
- `portal/src/lib/components/admin/AdminPlanForm.svelte`
- `portal/src/lib/api.ts`
- `portal/src/lib/types.ts`

First cut:
- `POST /admin/planos`
- `PUT /admin/planos/:id`
- `PATCH /admin/planos/:id/status`
- Preserve portal visibility and recommended flag.
- Validate price, duration, speeds, and display order.

Deferred:
- Drag-and-drop reorder.
- Advanced data-cap packages.

### Lane C: Voucher Operations

Purpose:
- Complete operational voucher management beyond the generator.

Primary ownership:
- `node/internal/api/admin/vouchers.go`
- `node/internal/store/store.go`
- `node/internal/infra/memory/store.go`
- `node/internal/infra/postgres/store.go`
- `portal/src/lib/components/admin/AdminVouchersPanel.svelte`
- `portal/src/lib/components/admin/AdminVoucherExportActions.svelte`
- `portal/src/lib/api.ts`
- `portal/src/lib/types.ts`

First cut:
- Filter vouchers by status, plan, code, and lot.
- `PATCH /admin/vouchers/:id/desativar`
- `GET /admin/vouchers/export.csv`
- Client action to download CSV.

Second cut:
- Printable voucher sheet page or PDF export.
- Lot-level export.

Deferred:
- Highly designed PDF ticket template if it needs a dedicated rendering pass.

### Lane D: Router Health and Network

Purpose:
- Replace placeholder router health with real or simulated OpenNDS/router status.

Primary ownership:
- `node/internal/gateway/gateway.go`
- `node/internal/gateway/opennds.go`
- `node/internal/api/admin/health.go`
- `node/internal/api/admin/roteadores.go`
- `node/internal/store/store.go`
- `node/internal/infra/memory/store.go`
- `node/internal/infra/postgres/store.go`
- `portal/src/lib/components/admin/AdminNetworkPanel.svelte`
- `portal/src/lib/types.ts`
- `portal/src/lib/api.ts`

First cut:
- `GET /admin/roteadores`
- `GET /admin/roteadores/:id/diagnostico`
- Health response calls gateway `Ping` with timeout and reports online/offline.
- Memory store returns one default router for local development.
- When `OPENNDS_ENABLED=false`, health should report router checks as disabled/dev instead of online.
- Add parser tests around `ndsctl status`, `ndsctl clients`, `ubus call system board`, and `logread -e opennds -n 50` before wiring diagnostics to HTTP.

Deferred:
- UCI auto-configuration.
- Restart OpenNDS/router commands.
- Live log streaming.

### Lane E: Payments and Reports

Purpose:
- Move from PIX demo visibility to an admin payment history and prepare Mercado Pago real.

Primary ownership:
- `node/internal/api/admin/pagamentos.go`
- `node/internal/api/portal/pix.go` or current portal handler split
- `node/internal/payments/`
- `node/internal/store/store.go`
- `node/internal/infra/memory/store.go`
- `node/internal/infra/postgres/store.go`
- `portal/src/lib/components/admin/AdminPaymentsPanel.svelte`
- `portal/src/lib/api.ts`
- `portal/src/lib/types.ts`

First cut:
- `GET /admin/pagamentos`
- Totals by status and date range.
- CSV export for payments.
- Keep PIX provider as demo unless credentials are configured.
- Add `internal/payments` with a `Provider` interface and a demo implementation preserving the current portal contract.
- Keep local development offline by default with `PAYMENTS_PROVIDER=demo`.

Second cut:
- Mercado Pago client abstraction.
- Webhook validation endpoint.
- PIX status reconciliation from provider.
- `POST /api/webhooks/mercadopago` should validate provider signature, fetch payment details from Mercado Pago, and only then update local status.
- Development-only approval endpoint may be added behind `GO_ENV=development` to simulate PIX approval without a public webhook URL.

Deferred:
- Full accounting PDF reports.

### Lane F: Jobs, Backup, and Logs

Purpose:
- Add operational reliability features for the local node.

Primary ownership:
- `node/internal/jobs/`
- `node/internal/api/admin/logs.go`
- `node/internal/api/admin/backup.go`
- `node/internal/store/store.go`
- `node/internal/infra/memory/store.go`
- `node/internal/infra/postgres/store.go`
- `portal/src/lib/components/admin/AdminLogsPanel.svelte`
- `portal/src/lib/components/admin/AdminBackupPanel.svelte`

First cut:
- Job that expires active sessions.
- `GET /admin/logs`
- CSV export for logs.
- Manual Postgres backup endpoint documented but disabled for memory store.

Deferred:
- Restore endpoint, because it is destructive.
- Automatic scheduled backup UI.
- WebSocket event feed.

## Execution Waves

### Wave 0: Split Bottleneck Files

- [x] Create backend route files: `auth.go`, `health.go`, `planos.go`, `usuarios.go`, `vouchers.go`.
- [x] Keep `Register` in `handlers.go` as the coordinator.
- [x] Split frontend `AdminDashboard.svelte` into focused panels under `portal/src/lib/components/admin/`.
- [x] Keep dashboard tests covering the extracted panels.
- [x] Run `go test ./...`, `npm test`, `npm run check`, and `npm run build`.
- [x] Commit: `refactor: split admin modules for parallel work`.

### Wave 1: Safe Parallel Implementation

Run these workers in parallel after Wave 0:

- [x] Worker A owns Lane A first cut. JWT, refresh, logout, me, and middleware are done; audit/session-bound token hardening remains deferred.
- [x] Worker B owns Lane B first cut. Backend and frontend plan CRUD are done in Wave 1B.
- [x] Worker C owns Lane C first cut. Advanced generation fields, filters, CSV export, and deactivate operation are done.
- [x] Worker D owns Lane D first cut. Parser/diagnostic foundation and admin HTTP wiring are done.

Conflict rule:
- Only one worker may edit `store.Store` at a time. If workers need new store methods, each worker should prepare a patch in its lane and the coordinator integrates the shared interface changes.

Verification per worker:
- Backend worker: `go test ./...`
- Frontend worker: `npm test -- <changed tests>`, `npm run check`
- Full integration after all workers: `go test ./...`, `npm test`, `npm run check`, `npm run build`, browser verification on `/painel`.

### Wave 2: Reporting and Reliability

Run after Wave 1 is merged:

- [x] Worker E owns Lane E first cut.
- [x] Worker F owns Lane F first cut.
- [x] Coordinator integrates shared store/API type changes.
- [x] Full verification and browser pass.
- [x] Commit: `feat: add local reporting and operations`.

### Wave 1B: Current Parallel Dispatch

Run after commit `535ef86`:

- [x] Agent Plans Backend owns backend CRUD for plans only. It may edit `node/internal/api/admin/planos.go`, `node/internal/api/admin/handlers.go`, `node/internal/domain/planos/`, `node/internal/store/store.go`, `node/internal/infra/memory/store.go`, `node/internal/infra/postgres/store.go`, and backend admin tests. It must not edit portal files.
- [x] Agent Plans Frontend owns the `/painel` plan-management UI only. It may edit `portal/src/lib/components/admin/AdminPlansPanel.svelte`, create `AdminPlanForm.svelte`, and update `portal/src/lib/api.ts`, `portal/src/lib/types.ts`, `portal/src/routes/painel/+page.svelte`, and portal tests. It must follow the backend CRUD contract from Lane B.
- [x] Agent Router Diagnostics owns parser and gateway diagnostic foundations only. It may edit `node/internal/gateway/` and add gateway tests. It must not edit admin routes or shared store files in this wave.
- [x] Coordinator integrates returned patches, resolves any shared API/type mismatch, runs full backend/frontend verification, and performs browser verification on `/painel`.

Wave 1B verification:
- `go test ./...` in `node`
- `npm test`, `npm run check`, and `npm run build` in `portal`
- `git diff --check`
- Browser verification on `http://127.0.0.1:5173/painel`: login, plan form visible, create plan success message, and plan count update.

### Wave 1C: Voucher Operations and Router HTTP

Run after commit `0de8689`:

- [x] Agent Vouchers Backend owns operational voucher filters, CSV export, and deactivate endpoint.
- [x] Agent Vouchers Frontend owns filters, export button, deactivate flow, and route orchestration for voucher operations.
- [x] Agent Router Backend owns health router status, router list, diagnostic endpoint, and OpenNDS diagnostic command wiring.
- [x] Coordinator registers the shared routes in `handlers.go`, verifies API contracts, and runs full backend/frontend/browser checks.

Wave 1C verification:
- `go test ./...` in `node`
- `npm test`, `npm run check`, and `npm run build` in `portal`
- `git diff --check`
- Browser DOM verification on `http://127.0.0.1:5173/painel`: login, voucher filters, CSV export button, and deactivate controls visible.
- Local API verification: admin login, filtered voucher query, voucher deactivation, CSV export, router health, router list, and router diagnostic.

Active agents:
- Backend plans: Dirac (`019e4b57-4344-7572-904b-474283a76b6a`)
- Frontend plans: Mendel (`019e4b57-5b8e-7db2-bb76-ef852ca2e1d5`)
- Router diagnostics: Averroes (`019e4b57-73e6-7c32-a2ae-41bc9b28cd6d`)

Plan CRUD contract for this repository:
- Create/update body fields: `nome`, `descricao`, `preco`, `duracao_minutos`, `dados_mb`, `velocidade_down`, `velocidade_up`, `recomendado`, `ativo`, `visivel_portal`, `ordem`.
- Status body: `{ "ativo": true | false }`.
- Create/update/status response: `{ "plano": <Plano> }`.

### Wave 3: Polish and Deferred Items

Run after local admin is operational:

- [x] Printable voucher sheet from the admin voucher list.
- [x] Mercado Pago webhook validation and reconciliation hook.
- [x] Development-only PIX approval endpoint behind `GO_ENV=development`.
- [x] Backup restore explicit confirmation workflow, implemented as safe validation only.
- [x] Real Mercado Pago provider client for payment detail fetches.
- [x] Highly designed PDF voucher export.
- [x] Admin SSE event stream with live local dashboard snapshot.
- [x] Best-effort local audit logs for mutating admin actions.
- [ ] Optional 2FA.

### Wave 3C: Current Parallel Dispatch

Run after commit `26c2404`:

- [x] Agent Mercado Pago Provider owned `internal/payments`, payment config, and `api.NewServer` provider construction.
- [x] Agent Voucher PDF owned the PDF/impression voucher sheet and focused Svelte component tests.
- [x] Agent Admin Login Lockout owned optional lockout store capability, memory/Postgres persistence, migration, and admin auth tests.
- [x] Coordinator added default HTTP timeout for Mercado Pago, updated env/docs, and ran full verification.

Wave 3C verification:
- `go test ./...` in `node`
- `npm test`, `npm run check`, and `npm run build` in `portal`
- `git diff --check`
- Local API verification: five invalid admin logins return `429 login_bloqueado`; valid login still returns JWT/refresh token.
- Browser DOM verification on `http://127.0.0.1:5173/painel`: voucher export button reads `Gerar folha PDF`, print sheet renders summary and customer instructions.

### Wave 3A: Current Parallel Dispatch

Run after commit `71b8ff3`:

- [x] Agent Voucher Print owned the printable voucher sheet UI, print-only component, and focused panel tests.
- [x] Agent Mercado Pago Backend owned webhook signature validation, provider status contract, development PIX approval, and store status updates.
- [x] Agent Restore Seguro owned the protected restore request path, admin UI form, API client method, and safety tests.
- [x] Coordinator integrated shared routes/types, added exact numeric webhook ID parsing, updated docs, and ran full verification.

Wave 3A verification:
- `go test ./...` in `node`
- `npm test`, `npm run check`, and `npm run build` in `portal`
- `git diff --check`
- Browser DOM verification on `http://127.0.0.1:5173/painel`: voucher print action, restore protected form, and admin dashboard render.
- Local API verification: admin login, PIX create, development approval, PIX status approved, Mercado Pago webhook without secret ignored in development, restore wrong confirmation rejected, restore confirmed returns safe `501 restore_indisponivel`.

### Wave 3B: Current Parallel Dispatch

Run after commit `383e9a2`:

- [x] Agent Backend Events owned `GET /admin/eventos`, SSE snapshot payloads, and once-mode tests.
- [x] Agent Frontend Live Events owned the compact `AdminLiveEventsPanel` and component tests.
- [x] Agent Admin Audit Logs owned optional `AppendAdminLog`, memory/Postgres log persistence, and admin mutation audit tests.
- [x] Coordinator registered the protected route, wired authenticated streaming fetch in `/painel`, integrated the live panel, added disconnect-user audit logging, updated docs, and ran full verification.

Wave 3B verification:
- `go test ./...` in `node`
- `npm test`, `npm run check`, and `npm run build` in `portal`
- `git diff --check`
- Local API verification: admin login, `GET /admin/eventos?once=1`, voucher generation audit log visible in `/admin/logs`.
- Browser DOM verification on `http://127.0.0.1:5173/painel`: live events panel visible and receiving snapshot state.

## Recommended Next Concrete Step

Move to the next Wave 3 slice after committing Wave 3C. The strongest remaining local slice is optional 2FA for the local admin or real PIX creation through Mercado Pago.

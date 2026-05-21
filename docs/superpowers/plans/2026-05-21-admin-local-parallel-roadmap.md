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
- Admin local login with temporary token.
- Admin health, plans list, users list, vouchers list/generate, disconnect user.
- Portal `/painel` UI for health, users, plans, and voucher generation.
- Wave 0 split completed in commit `782d5bd`: admin backend handlers are separated by domain, and the admin dashboard is split into panel components under `portal/src/lib/components/admin/`.

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

- [ ] Create backend route files: `auth.go`, `health.go`, `planos.go`, `usuarios.go`, `vouchers.go`.
- [ ] Keep `Register` in `handlers.go` as the coordinator.
- [ ] Split frontend `AdminDashboard.svelte` into `admin/AdminShell.svelte`, `admin/AdminMetrics.svelte`, `admin/AdminUsersPanel.svelte`, `admin/AdminPlansPanel.svelte`, and `admin/AdminVouchersPanel.svelte`.
- [ ] Move panel-specific tests next to the new components.
- [ ] Run `go test ./...`, `npm test`, `npm run check`, and `npm run build`.
- [ ] Commit: `refactor: split admin modules for parallel work`.

### Wave 1: Safe Parallel Implementation

Run these workers in parallel after Wave 0:

- [ ] Worker A owns Lane A first cut. Started as Auth P0 backend worker; audit logs are deferred to the next auth/audit pass.
- [ ] Worker B owns Lane B first cut.
- [ ] Worker C owns Lane C first cut. Started with a frontend-only voucher form expansion because the backend already accepts the extra generation fields.
- [ ] Worker D owns Lane D first cut.

Conflict rule:
- Only one worker may edit `store.Store` at a time. If workers need new store methods, each worker should prepare a patch in its lane and the coordinator integrates the shared interface changes.

Verification per worker:
- Backend worker: `go test ./...`
- Frontend worker: `npm test -- <changed tests>`, `npm run check`
- Full integration after all workers: `go test ./...`, `npm test`, `npm run check`, `npm run build`, browser verification on `/painel`.

### Wave 2: Reporting and Reliability

Run after Wave 1 is merged:

- [ ] Worker E owns Lane E first cut.
- [ ] Worker F owns Lane F first cut.
- [ ] Coordinator integrates shared store/API type changes.
- [ ] Full verification and browser pass.
- [ ] Commit: `feat: add local reporting and operations`.

### Wave 3: Polish and Deferred Items

Run after local admin is operational:

- [ ] Printable/PDF voucher sheet.
- [ ] Mercado Pago real webhook and reconciliation.
- [ ] Backup restore with explicit confirmation workflow.
- [ ] Admin WebSocket/SSE event stream.
- [ ] Optional 2FA.

## Recommended Next Concrete Step

Start with Wave 0. It is the unlocker for safe parallel work and has the lowest product risk. After Wave 0, launch four workers for lanes A-D.

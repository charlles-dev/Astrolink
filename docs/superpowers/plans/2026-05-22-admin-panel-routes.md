# Admin Panel Routes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Split the local admin panel into real SvelteKit pages while preserving the approved app shell and all current admin behavior.

**Architecture:** Move the route-owned admin state from `portal/src/routes/painel/+page.svelte` into a reusable `AdminPanelRoute.svelte` component. Extract the visual shell into `AdminShell.svelte` and the active-page renderer into `AdminPanelContent.svelte`; each route file becomes a tiny wrapper that passes its active page.

**Tech Stack:** Svelte 5, SvelteKit filesystem routing, TypeScript, Svelte Testing Library, Vitest, component-scoped CSS.

---

### Task 1: Shell And Content Split

**Files:**
- Create: `portal/src/lib/components/admin/AdminShell.svelte`
- Create: `portal/src/lib/components/admin/AdminPanelContent.svelte`
- Modify: `portal/src/lib/components/AdminDashboard.svelte`

- [ ] **Step 1: Create `AdminShell.svelte`**

Implement a presentation-only shell with these props:

```ts
export type AdminPanelPage =
  | 'overview'
  | 'usuarios'
  | 'planos'
  | 'vouchers'
  | 'pagamentos'
  | 'setup'
  | 'logs'

export let activePage: AdminPanelPage = 'overview'
export let health: AdminHealthResponse | null = null
export let usuarios: AdminUser[] = []
export let liveConnected = false
export let loading = false
export let actionMessage = ''
export let onRefresh: () => void = () => {}
export let onLogout: () => void = () => {}
```

Move the current shell markup and CSS from `AdminDashboard.svelte` into this file. Replace all anchor hashes with route links:

```ts
const navItems = [
  { id: 'overview', label: 'Visao geral', href: '/painel' },
  { id: 'usuarios', label: 'Usuarios', href: '/painel/usuarios' },
  { id: 'planos', label: 'Planos', href: '/painel/planos' },
  { id: 'vouchers', label: 'Vouchers', href: '/painel/vouchers' },
  { id: 'pagamentos', label: 'Pagamentos', href: '/painel/pagamentos' },
  { id: 'setup', label: 'Setup', href: '/painel/setup' },
  { id: 'logs', label: 'Logs', href: '/painel/logs' }
] as const
```

Use `aria-current={item.id === activePage ? 'page' : undefined}`.

- [ ] **Step 2: Create `AdminPanelContent.svelte`**

Implement a content-only renderer with the current admin props and handlers from `AdminDashboard.svelte`, plus `activePage`.

Render:

```svelte
{#if activePage === 'overview'}
  <section class="overview-stack">
    <AdminMetrics ... />
    <div class="overview-grid">
      <AdminLiveEventsPanel ... />
      <AdminUsersPanel ... />
    </div>
  </section>
{:else if activePage === 'usuarios'}
  <AdminUsersPanel ... />
{:else if activePage === 'planos'}
  <AdminPlansPanel ... />
{:else if activePage === 'vouchers'}
  <AdminVouchersPanel ... />
{:else if activePage === 'pagamentos'}
  <AdminPaymentsPanel ... />
{:else if activePage === 'setup'}
  <AdminSetupPanel ... />
{:else if activePage === 'logs'}
  <section class="logs-grid">
    <AdminLiveEventsPanel ... />
    <AdminLogsPanel ... />
    <AdminBackupPanel ... />
  </section>
{/if}
```

Use the existing child components and pass the same handlers.

- [ ] **Step 3: Make `AdminDashboard.svelte` a compatibility wrapper**

Keep the same public props as today, add `export let activePage: AdminPanelPage = 'overview'`, and render:

```svelte
<AdminShell ... {activePage}>
  <AdminPanelContent ... {activePage} />
</AdminShell>
```

Do not keep the old one-page layout inside `AdminDashboard.svelte`.

### Task 2: Shared Route Component And Real Routes

**Files:**
- Create: `portal/src/lib/components/admin/AdminPanelRoute.svelte`
- Modify: `portal/src/routes/painel/+page.svelte`
- Create: `portal/src/routes/painel/usuarios/+page.svelte`
- Create: `portal/src/routes/painel/planos/+page.svelte`
- Create: `portal/src/routes/painel/vouchers/+page.svelte`
- Create: `portal/src/routes/painel/pagamentos/+page.svelte`
- Create: `portal/src/routes/painel/setup/+page.svelte`
- Create: `portal/src/routes/painel/logs/+page.svelte`

- [ ] **Step 1: Create `AdminPanelRoute.svelte`**

Move the full script, login markup, login CSS, and admin data handlers from `portal/src/routes/painel/+page.svelte` into this component. Add:

```ts
import type { AdminPanelPage } from '$lib/components/admin/AdminShell.svelte'
import AdminDashboard from '$lib/components/AdminDashboard.svelte'

export let activePage: AdminPanelPage = 'overview'
```

When authenticated, render:

```svelte
<AdminDashboard
  {activePage}
  {health}
  {planos}
  {usuarios}
  {vouchers}
  {pagamentos}
  {pagamentosTotais}
  {logs}
  {logsTotal}
  {setupStatus}
  {liveConnected}
  {liveLastEventAt}
  {liveSnapshot}
  {liveEvents}
  {loading}
  {actionMessage}
  {backupMessage}
  {setupMessage}
  onRefresh={loadDashboard}
  onDisconnect={disconnect}
  onSavePlan={savePlan}
  onTogglePlanStatus={togglePlanStatus}
  onGenerateVouchers={generateVouchers}
  onApplyVoucherFilters={applyVoucherFilters}
  onDeactivateVoucher={deactivateVoucher}
  onExportVouchers={exportVouchers}
  onApplyPaymentFilters={applyPaymentFilters}
  onExportPayments={exportPayments}
  onApplyLogFilters={applyLogFilters}
  onExportLogs={exportLogs}
  onCreateBackup={createBackup}
  onRestoreBackup={restoreBackup}
  onSaveSetup={saveSetup}
  onLogout={logout}
/>
```

- [ ] **Step 2: Convert the route files to wrappers**

Make `portal/src/routes/painel/+page.svelte`:

```svelte
<script lang="ts">
  import AdminPanelRoute from '$lib/components/admin/AdminPanelRoute.svelte'
</script>

<AdminPanelRoute activePage="overview" />
```

Create each subroute with the same pattern and the matching `activePage`.

- [ ] **Step 3: Preserve session behavior**

Keep `TOKEN_KEY = 'astrolink.admin.token'`, `sessionStorage` persistence, `initializeDashboard`, `startLiveEvents`, `stopLiveEvents`, `expireSessionIfUnauthorized`, and `logout` unchanged in behavior.

### Task 3: Route Tests And Existing Component Tests

**Files:**
- Modify: `portal/src/lib/components/AdminDashboard.test.ts`
- Modify: `portal/src/routes/painel/page.test.ts`
- Create: `portal/src/routes/painel/usuarios/page.test.ts`
- Create: `portal/src/routes/painel/planos/page.test.ts`
- Create: `portal/src/routes/painel/vouchers/page.test.ts`
- Create: `portal/src/routes/painel/pagamentos/page.test.ts`
- Create: `portal/src/routes/painel/setup/page.test.ts`
- Create: `portal/src/routes/painel/logs/page.test.ts`

- [ ] **Step 1: Update `AdminDashboard.test.ts`**

Keep interaction coverage by rendering `AdminDashboard` with `activePage` set to the page under test:

```ts
render(AdminDashboard, { props: { activePage: 'usuarios', usuarios: [...], onDisconnect } })
render(AdminDashboard, { props: { activePage: 'vouchers', planos: [...], vouchers: [...], onGenerateVouchers } })
render(AdminDashboard, { props: { activePage: 'pagamentos', pagamentos: [...], onExportPayments } })
render(AdminDashboard, { props: { activePage: 'logs', logs: [...], onCreateBackup, onRestoreBackup } })
```

Existing expectations for buttons and handler payloads must remain the same.

- [ ] **Step 2: Update `/painel` tests**

Keep the login tests in `portal/src/routes/painel/page.test.ts`. Change the setup failure assertion from `Usuarios conectados` to a visible overview signal:

```ts
expect(screen.getByRole('heading', { name: 'Painel local' })).toBeInTheDocument()
expect(screen.getByText('Usuarios ativos')).toBeInTheDocument()
```

- [ ] **Step 3: Add route smoke tests**

Each subroute test should mock `$lib/api` like the current `/painel` test, render its local `+page.svelte`, submit login, and assert the page-specific heading:

```ts
expect(await screen.findByRole('heading', { name: 'Usuarios conectados' })).toBeInTheDocument()
expect(await screen.findByRole('heading', { name: 'Planos' })).toBeInTheDocument()
expect(await screen.findByRole('heading', { name: 'Vouchers' })).toBeInTheDocument()
expect(await screen.findByRole('heading', { name: 'Pagamentos' })).toBeInTheDocument()
expect(await screen.findByRole('heading', { name: 'Setup local' })).toBeInTheDocument()
expect(await screen.findByRole('heading', { name: 'Eventos ao vivo' })).toBeInTheDocument()
```

### Task 4: Verification

**Files:**
- No source edits unless verification exposes a defect.

- [ ] **Step 1: Run focused tests**

```powershell
cd portal
npm test -- src/lib/components/AdminDashboard.test.ts src/routes/painel/page.test.ts
```

Expected: pass.

- [ ] **Step 2: Run the full portal suite**

```powershell
cd portal
npm test
npm run check
npm run build
```

Expected: all pass.

- [ ] **Step 3: Browser verification**

Open `http://127.0.0.1:5173/painel`, log in with `admin` / `admin123`, then verify:

- `/painel` shows overview and metrics.
- `/painel/usuarios` shows only the user management page content.
- `/painel/planos` shows plans.
- `/painel/vouchers` shows vouchers.
- `/painel/pagamentos` shows payments.
- `/painel/setup` shows local setup.
- `/painel/logs` shows live events, logs and backup.
- Mobile width does not create document-level horizontal overflow.

- [ ] **Step 4: Commit**

```powershell
git add docs/superpowers/specs/2026-05-22-admin-panel-routes-design.md docs/superpowers/plans/2026-05-22-admin-panel-routes.md portal/src
git commit -m "feat: split admin panel into pages"
```

# Admin Panel App Shell Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the approved Option A visual direction for the local admin panel as a professional app shell.

**Architecture:** Keep the current Svelte component contracts intact and replace only the dashboard composition layer plus the KPI presentation. The dashboard becomes a two-column shell on desktop, a compact top navigation on smaller screens, and keeps the existing panels mounted with their existing handlers.

**Tech Stack:** Svelte 5, SvelteKit, TypeScript, component-scoped CSS, Vitest, Svelte Testing Library, svelte-check.

---

### Task 1: Dashboard Shell Composition

**Files:**
- Modify: `portal/src/lib/components/AdminDashboard.svelte`
- Test: `portal/src/lib/components/AdminDashboard.test.ts`

- [ ] **Step 1: Add derived shell state**

Add these reactive declarations after the exported props in `AdminDashboard.svelte`:

```svelte
  $: nodeStatus = health?.status ?? 'offline'
  $: databaseStatus = health?.checks.banco_dados.status ?? 'sem dados'
  $: routerOnline = health?.checks.roteadores.online ?? 0
  $: routerTotal = health?.checks.roteadores.total ?? 0
  $: activeUsers = usuarios.filter((usuario) => usuario.status === 'ativo').length
```

- [ ] **Step 2: Replace the dashboard markup**

Replace the current top-level `<section class="admin-dashboard" ...>` block with an app shell containing `admin-sidebar`, `admin-workspace`, `workspace-header`, `section-tabs`, `overview-section`, `dashboard-grid`, and `observability-grid`. Keep all existing child components and pass the same props/handlers.

- [ ] **Step 3: Run the component tests**

Run:

```powershell
cd portal
npm test -- src/lib/components/AdminDashboard.test.ts
```

Expected: all `AdminDashboard` tests pass, including the existing `Painel local`, `Usuarios conectados`, `Pagamentos`, `Logs`, and `Eventos ao vivo` assertions.

### Task 2: Shell Visual System

**Files:**
- Modify: `portal/src/lib/components/AdminDashboard.svelte`

- [ ] **Step 1: Replace dashboard CSS**

Replace the existing `.admin-dashboard`, `.admin-topbar`, `.admin-content`, and `.operations-content` styles with scoped styles for:

```css
.admin-shell
.admin-sidebar
.brand-lockup
.admin-nav
.session-card
.admin-workspace
.workspace-header
.workspace-actions
.section-tabs
.action-message
.dashboard-grid
.primary-column
.secondary-column
.observability-grid
```

Use the approved visual system: cool gray background, white surfaces, 8px radii, subtle borders, dark primary button, bordered secondary button, teal active states, and stable responsive dimensions.

- [ ] **Step 2: Add responsive behavior**

Add breakpoints for `max-width: 1080px`, `900px`, and `620px` so the sidebar becomes a top rail, grids collapse to one column, and action buttons remain readable without overflow.

- [ ] **Step 3: Run Svelte check**

Run:

```powershell
cd portal
npm run check
```

Expected: zero Svelte/TypeScript errors.

### Task 3: KPI Card Refinement

**Files:**
- Modify: `portal/src/lib/components/admin/AdminMetrics.svelte`
- Test: `portal/src/lib/components/AdminDashboard.test.ts`

- [ ] **Step 1: Refine metric markup**

Update the metric cards to include a `metric-card` class, a tiny color accent element, a `metric-label`, and a `metric-footnote`. Preserve the visible text values: `Usuarios ativos`, `Planos ativos`, `Vouchers ativos`, and `Banco`.

- [ ] **Step 2: Replace metric CSS**

Remove the fixed `max-width` and make the grid fill its parent. Use compact KPI cards with an accent strip, white surface, 8px radius, stronger number hierarchy, and responsive two-column/mobile collapse.

- [ ] **Step 3: Re-run focused tests**

Run:

```powershell
cd portal
npm test -- src/lib/components/AdminDashboard.test.ts
```

Expected: all focused component tests pass.

### Task 4: Full Verification

**Files:**
- No source edits unless verification exposes a defect.

- [ ] **Step 1: Run the route tests**

Run:

```powershell
cd portal
npm test -- src/routes/painel/page.test.ts
```

Expected: all `/painel` login/setup tests pass.

- [ ] **Step 2: Run the full portal suite**

Run:

```powershell
cd portal
npm test
npm run build
```

Expected: tests and production build pass.

- [ ] **Step 3: Browser verification**

Open `http://127.0.0.1:5173/painel`, log in with `admin` / `admin123`, and verify desktop plus mobile widths. Confirm the accepted Option A shell is visible: sidebar/top rail, workspace header, tabs, KPI band, primary/secondary content columns, and observability area.

- [ ] **Step 4: Commit**

Run:

```powershell
git add docs/superpowers/plans/2026-05-22-admin-panel-app-shell.md portal/src/lib/components/AdminDashboard.svelte portal/src/lib/components/admin/AdminMetrics.svelte
git commit -m "feat: redesign admin app shell"
```

Expected: commit succeeds on the current feature branch.

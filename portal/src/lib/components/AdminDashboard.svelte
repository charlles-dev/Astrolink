<script lang="ts">
  import AdminBackupPanel from './admin/AdminBackupPanel.svelte'
  import AdminLiveEventsPanel, {
    type AdminLiveEvent,
    type AdminLiveSnapshot
  } from './admin/AdminLiveEventsPanel.svelte'
  import AdminLogsPanel from './admin/AdminLogsPanel.svelte'
  import AdminMetrics from './admin/AdminMetrics.svelte'
  import AdminPaymentsPanel from './admin/AdminPaymentsPanel.svelte'
  import AdminPlansPanel from './admin/AdminPlansPanel.svelte'
  import AdminSetupPanel from './admin/AdminSetupPanel.svelte'
  import AdminUsersPanel from './admin/AdminUsersPanel.svelte'
  import AdminVouchersPanel from './admin/AdminVouchersPanel.svelte'
  import type {
    AdminHealthResponse,
    AdminLog,
    AdminLogFilters,
    AdminPayment,
    AdminPaymentFilters,
    AdminPaymentTotals,
    AdminPlanBody,
    AdminRestoreBackupBody,
    AdminUser,
    AdminVoucher,
    AdminVoucherFilters,
    GenerateAdminVouchersBody,
    Plano,
    SetupStatus
  } from '../types'

  export let health: AdminHealthResponse | null = null
  export let planos: Plano[] = []
  export let usuarios: AdminUser[] = []
  export let vouchers: AdminVoucher[] = []
  export let pagamentos: AdminPayment[] = []
  export let pagamentosTotais: AdminPaymentTotals = {
    pendente: 0,
    aprovado: 0,
    cancelado: 0,
    expirado: 0,
    valor_total: '0.00'
  }
  export let logs: AdminLog[] = []
  export let logsTotal = 0
  export let setupStatus: SetupStatus | null = null
  export let liveConnected = false
  export let liveLastEventAt = ''
  export let liveSnapshot: AdminLiveSnapshot | null = null
  export let liveEvents: AdminLiveEvent[] = []
  export let loading = false
  export let actionMessage = ''
  export let backupMessage = ''
  export let setupMessage = ''
  export let onRefresh: () => void = () => {}
  export let onDisconnect: (mac: string) => void = () => {}
  export let onSavePlan: (input: AdminPlanBody, id?: number) => Promise<void> | void = () => {}
  export let onTogglePlanStatus: (id: number, ativo: boolean) => Promise<void> | void = () => {}
  export let onGenerateVouchers: (input: GenerateAdminVouchersBody) => void = () => {}
  export let onApplyVoucherFilters: (filters: AdminVoucherFilters) => void = () => {}
  export let onDeactivateVoucher: (id: number) => void = () => {}
  export let onExportVouchers: (filters: AdminVoucherFilters) => void = () => {}
  export let onApplyPaymentFilters: (filters: AdminPaymentFilters) => void = () => {}
  export let onExportPayments: (filters: AdminPaymentFilters) => void = () => {}
  export let onApplyLogFilters: (filters: AdminLogFilters) => void = () => {}
  export let onExportLogs: (filters: AdminLogFilters) => void = () => {}
  export let onCreateBackup: () => void = () => {}
  export let onRestoreBackup: (input: AdminRestoreBackupBody) => void = () => {}
  export let onSaveSetup: (values: Record<string, string>) => Promise<void> | void = () => {}
  export let onLogout: () => void = () => {}

  $: nodeStatus = health?.status ?? 'offline'
  $: databaseStatus = health?.checks.banco_dados.status ?? 'sem dados'
  $: routerOnline = health?.checks.roteadores.online ?? 0
  $: routerTotal = health?.checks.roteadores.total ?? 0
  $: activeUsers = usuarios.filter((usuario) => usuario.status === 'ativo').length
</script>

<section class="admin-shell" aria-busy={loading}>
  <aside class="admin-sidebar" aria-label="Navegacao do painel local">
    <div class="brand-lockup">
      <span class="brand-mark" aria-hidden="true">A</span>
      <div>
        <strong>Astrolink</strong>
        <span>Node local</span>
      </div>
    </div>

    <nav class="admin-nav" aria-label="Areas do painel">
      <a href="#overview" aria-current="page">Visao geral</a>
      <a href="#operation">Operacao</a>
      <a href="#plans">Planos</a>
      <a href="#payments">Pagamentos</a>
      <a href="#setup">Setup</a>
      <a href="#logs">Logs</a>
    </nav>

    <div class="session-card">
      <span class="session-label">Sessao local</span>
      <strong>{nodeStatus}</strong>
      <p>Banco de dados {databaseStatus}</p>
      <p>{routerOnline}/{routerTotal} roteadores online</p>
    </div>
  </aside>

  <main class="admin-workspace">
    <header class="workspace-header">
      <div class="workspace-copy">
        <span class="workspace-status">
          <span class:online={liveConnected}></span>
          {liveConnected ? 'Tempo real ativo' : 'Tempo real offline'}
        </span>
        <h1>Painel local</h1>
        <p>
          Operacao do hotspot, planos, vouchers, pagamentos e setup do no em um workspace.
        </p>
      </div>

      <div class="workspace-summary" aria-label="Resumo operacional">
        <div>
          <span>Usuarios</span>
          <strong>{activeUsers}</strong>
        </div>
        <div>
          <span>Node</span>
          <strong>{nodeStatus}</strong>
        </div>
      </div>

      <div class="workspace-actions">
        <button type="button" class="ghost-button" onclick={onRefresh} disabled={loading}>
          {loading ? 'Atualizando' : 'Atualizar'}
        </button>
        <button type="button" class="ink-button" onclick={onLogout}>Sair</button>
      </div>
    </header>

    {#if actionMessage}
      <p class="action-message" role="status">{actionMessage}</p>
    {/if}

    <nav class="section-tabs" aria-label="Atalhos do workspace">
      <a href="#overview" aria-current="page">Visao geral</a>
      <a href="#operation">Usuarios</a>
      <a href="#plans">Planos</a>
      <a href="#payments">Pagamentos</a>
      <a href="#setup">Setup local</a>
      <a href="#logs">Observabilidade</a>
    </nav>

    <section id="overview" class="overview-section" aria-label="Metricas do painel">
      <AdminMetrics {health} {planos} {usuarios} {vouchers} />
    </section>

    <div id="operation" class="dashboard-grid">
      <div class="primary-column">
        <AdminUsersPanel {usuarios} {loading} {onDisconnect} />

        <div id="payments">
          <AdminPaymentsPanel
            {pagamentos}
            totais={pagamentosTotais}
            {loading}
            {onApplyPaymentFilters}
            {onExportPayments}
          />
        </div>
      </div>

      <aside class="secondary-column" aria-label="Configuracao e ofertas">
        <div id="plans">
          <AdminPlansPanel {planos} {loading} {onSavePlan} {onTogglePlanStatus} />
        </div>
        <div id="setup">
          <AdminSetupPanel {setupStatus} {loading} {setupMessage} {onSaveSetup} />
        </div>
        <AdminVouchersPanel
          {planos}
          {vouchers}
          {loading}
          {onGenerateVouchers}
          {onApplyVoucherFilters}
          {onDeactivateVoucher}
          {onExportVouchers}
        />
      </aside>
    </div>

    <section id="logs" class="observability-grid" aria-label="Observabilidade local">
      <AdminLiveEventsPanel
        connected={liveConnected}
        lastEventAt={liveLastEventAt}
        snapshot={liveSnapshot}
        events={liveEvents}
      />
      <AdminLogsPanel
        {logs}
        total={logsTotal}
        {loading}
        {onApplyLogFilters}
        {onExportLogs}
      />
      <AdminBackupPanel {loading} {backupMessage} {onCreateBackup} {onRestoreBackup} />
    </section>
  </main>
</section>

<style>
  .admin-shell {
    min-height: 100vh;
    display: grid;
    grid-template-columns: 248px minmax(0, 1fr);
    background: #eef4f8;
    color: var(--color-ink);
  }

  .admin-sidebar {
    position: sticky;
    top: 0;
    min-height: 100vh;
    display: grid;
    grid-template-rows: auto auto 1fr;
    gap: 22px;
    border-right: 1px solid #d8e3ef;
    padding: 24px 18px;
    background: #0b1724;
    color: white;
  }

  .brand-lockup,
  .workspace-header,
  .workspace-actions,
  .workspace-summary,
  .section-tabs {
    display: flex;
    align-items: center;
  }

  .brand-lockup {
    gap: 12px;
  }

  .brand-mark {
    width: 40px;
    height: 40px;
    display: grid;
    place-items: center;
    border-radius: 8px;
    background: #1dd3b0;
    color: #06251f;
    font-size: 1rem;
    font-weight: 950;
  }

  .brand-lockup strong,
  .brand-lockup span,
  .session-card strong,
  .session-card p,
  .session-label,
  h1,
  p {
    margin: 0;
  }

  .brand-lockup strong {
    display: block;
    font-size: 1rem;
    font-weight: 920;
    line-height: 1.1;
  }

  .brand-lockup span,
  .session-card p,
  .session-label {
    color: #9fb3c8;
  }

  .brand-lockup span {
    display: block;
    margin-top: 3px;
    font-size: 0.78rem;
    font-weight: 760;
  }

  .admin-nav {
    display: grid;
    gap: 6px;
  }

  .admin-nav a {
    min-height: 42px;
    display: flex;
    align-items: center;
    border-radius: 8px;
    padding: 0 12px;
    color: #dbeafe;
    font-size: 0.88rem;
    font-weight: 820;
    text-decoration: none;
  }

  .admin-nav a:hover,
  .admin-nav a[aria-current='page'] {
    background: rgba(29, 211, 176, 0.14);
    color: #ecfeff;
  }

  .admin-nav a[aria-current='page'] {
    box-shadow: inset 3px 0 0 #1dd3b0;
  }

  .session-card {
    align-self: end;
    display: grid;
    gap: 7px;
    border: 1px solid rgba(148, 163, 184, 0.24);
    border-radius: 8px;
    padding: 14px;
    background: rgba(15, 23, 42, 0.48);
  }

  .session-label {
    font-size: 0.72rem;
    font-weight: 860;
    text-transform: uppercase;
  }

  .session-card strong {
    font-size: 1rem;
    font-weight: 900;
  }

  .session-card p {
    font-size: 0.78rem;
    font-weight: 720;
    line-height: 1.35;
  }

  .admin-workspace {
    min-width: 0;
    display: grid;
    align-content: start;
    gap: 16px;
    padding: 28px;
  }

  .workspace-header {
    min-height: 152px;
    justify-content: space-between;
    gap: 22px;
    border: 1px solid #dbe4ef;
    border-radius: 8px;
    padding: 22px;
    background: #ffffff;
    box-shadow: 0 18px 42px rgba(15, 23, 42, 0.08);
  }

  .workspace-copy {
    min-width: 0;
    display: grid;
    gap: 8px;
  }

  .workspace-status {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    color: #475569;
    font-size: 0.78rem;
    font-weight: 860;
    text-transform: uppercase;
  }

  .workspace-status span {
    width: 9px;
    height: 9px;
    border-radius: 999px;
    background: #f59e0b;
    box-shadow: 0 0 0 4px #fef3c7;
  }

  .workspace-status span.online {
    background: #22c55e;
    box-shadow: 0 0 0 4px #dcfce7;
  }

  h1 {
    font-size: 3rem;
    font-weight: 950;
    line-height: 1;
  }

  .workspace-copy p {
    max-width: 700px;
    color: #64748b;
    font-size: 0.98rem;
    font-weight: 650;
    line-height: 1.5;
  }

  .workspace-summary {
    flex: 0 0 auto;
    gap: 10px;
  }

  .workspace-summary div {
    min-width: 112px;
    min-height: 82px;
    display: grid;
    align-content: center;
    gap: 4px;
    border: 1px solid #dbe4ef;
    border-radius: 8px;
    padding: 12px;
    background: #f8fafc;
  }

  .workspace-summary span {
    color: #64748b;
    font-size: 0.72rem;
    font-weight: 860;
    text-transform: uppercase;
  }

  .workspace-summary strong {
    min-width: 0;
    overflow-wrap: anywhere;
    font-size: 1.2rem;
    font-weight: 930;
    line-height: 1.05;
  }

  .workspace-actions {
    flex: 0 0 auto;
    gap: 10px;
  }

  .ghost-button,
  .ink-button {
    min-height: 42px;
    border-radius: 8px;
    padding: 0 14px;
    font-size: 0.86rem;
    font-weight: 850;
  }

  .ghost-button {
    border: 1px solid #cbd5e1;
    background: white;
    color: var(--color-ink);
  }

  .ink-button {
    border: 0;
    background: var(--color-ink);
    color: white;
  }

  .ghost-button:disabled {
    cursor: not-allowed;
    opacity: 0.55;
  }

  .section-tabs {
    gap: 8px;
    overflow-x: auto;
    border-bottom: 1px solid #d8e3ef;
    padding-bottom: 10px;
  }

  .admin-nav,
  .section-tabs {
    scrollbar-width: none;
  }

  .admin-nav::-webkit-scrollbar,
  .section-tabs::-webkit-scrollbar {
    display: none;
  }

  .section-tabs a {
    min-height: 38px;
    flex: 0 0 auto;
    display: inline-flex;
    align-items: center;
    border: 1px solid #dbe4ef;
    border-radius: 8px;
    padding: 0 12px;
    background: rgba(255, 255, 255, 0.78);
    color: #475569;
    font-size: 0.82rem;
    font-weight: 850;
    text-decoration: none;
  }

  .section-tabs a:hover,
  .section-tabs a[aria-current='page'] {
    border-color: #99f6e4;
    background: #ecfeff;
    color: #0f766e;
  }

  .action-message {
    margin: 0;
    border: 1px solid #bae6fd;
    border-radius: 8px;
    padding: 12px 14px;
    background: #e0f2fe;
    color: #075985;
    font-size: 0.88rem;
    font-weight: 800;
  }

  .overview-section {
    scroll-margin-top: 20px;
  }

  .dashboard-grid {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(330px, 0.42fr);
    gap: 16px;
  }

  .primary-column,
  .secondary-column {
    display: grid;
    gap: 16px;
    align-content: start;
    min-width: 0;
  }

  .observability-grid {
    display: grid;
    grid-template-columns: minmax(300px, 0.8fr) minmax(0, 1.2fr) minmax(280px, 0.72fr);
    gap: 16px;
    scroll-margin-top: 20px;
  }

  #plans,
  #payments,
  #setup {
    scroll-margin-top: 20px;
  }

  @media (max-width: 1080px) {
    .admin-shell {
      grid-template-columns: 1fr;
    }

    .admin-sidebar {
      position: static;
      min-height: auto;
      grid-template-columns: auto minmax(0, 1fr) auto;
      grid-template-rows: auto;
      align-items: center;
      gap: 14px;
      border-right: 0;
      border-bottom: 1px solid #d8e3ef;
      padding: 14px 18px;
    }

    .admin-nav {
      grid-auto-flow: column;
      grid-auto-columns: max-content;
      overflow-x: auto;
      gap: 6px;
    }

    .admin-nav a {
      min-height: 38px;
    }

    .session-card {
      align-self: auto;
      min-width: 190px;
      padding: 11px 12px;
    }

    .session-card p:last-child {
      display: none;
    }
  }

  @media (max-width: 900px) {
    .workspace-header {
      align-items: stretch;
      flex-direction: column;
      min-height: 0;
    }

    .workspace-summary {
      width: 100%;
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .dashboard-grid,
    .observability-grid {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 620px) {
    .admin-sidebar {
      grid-template-columns: 1fr;
      align-items: stretch;
    }

    .admin-workspace {
      padding: 18px;
    }

    .brand-lockup {
      justify-content: flex-start;
    }

    .admin-nav {
      margin-inline: -4px;
    }

    .session-card {
      min-width: 0;
    }

    .workspace-header {
      padding: 18px;
    }

    h1 {
      font-size: 2rem;
    }

    .workspace-actions {
      align-items: stretch;
      width: 100%;
    }

    .workspace-actions button {
      flex: 1;
    }

    .workspace-summary {
      grid-template-columns: 1fr;
    }

    .section-tabs {
      padding-bottom: 8px;
    }
  }
</style>

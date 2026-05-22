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
</script>

<section class="admin-dashboard" aria-busy={loading}>
  <header class="admin-topbar">
    <div>
      <p class="admin-kicker">Astrolink Node</p>
      <h1>Painel local</h1>
    </div>
    <div class="admin-actions">
      <button type="button" class="ghost-button" onclick={onRefresh} disabled={loading}>
        Atualizar
      </button>
      <button type="button" class="ink-button" onclick={onLogout}>Sair</button>
    </div>
  </header>

  {#if actionMessage}
    <p class="action-message" role="status">{actionMessage}</p>
  {/if}

  <AdminMetrics {health} {planos} {usuarios} {vouchers} />

  <div class="admin-content">
    <AdminUsersPanel {usuarios} {loading} {onDisconnect} />

    <aside class="side-stack">
      <AdminPlansPanel {planos} {loading} {onSavePlan} {onTogglePlanStatus} />
      <AdminSetupPanel {setupStatus} {loading} {setupMessage} {onSaveSetup} />
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

  <div class="operations-content">
    <AdminPaymentsPanel
      {pagamentos}
      totais={pagamentosTotais}
      {loading}
      {onApplyPaymentFilters}
      {onExportPayments}
    />

    <div class="operations-stack">
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
    </div>
  </div>
</section>

<style>
  .admin-dashboard {
    min-height: 100vh;
    padding: 28px;
    background: #f8fafc;
    color: var(--color-ink);
  }

  .admin-topbar,
  .admin-actions {
    display: flex;
    align-items: center;
  }

  .admin-topbar {
    justify-content: space-between;
    gap: 18px;
    margin: 0 auto 22px;
    max-width: 1180px;
  }

  .admin-kicker,
  h1,
  p {
    margin: 0;
  }

  .admin-kicker {
    color: #0f766e;
    font-size: 0.78rem;
    font-weight: 850;
    text-transform: uppercase;
  }

  h1 {
    margin-top: 4px;
    font-size: 2rem;
    font-weight: 920;
    line-height: 1.05;
  }

  .admin-actions {
    gap: 10px;
  }

  .ghost-button,
  .ink-button {
    min-height: 42px;
    border-radius: 12px;
    padding: 0 14px;
    font-size: 0.86rem;
    font-weight: 850;
  }

  .ghost-button {
    border: 1px solid var(--color-line);
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

  .action-message {
    max-width: 1180px;
    margin: 0 auto 16px;
    border: 1px solid #bae6fd;
    border-radius: 14px;
    padding: 12px 14px;
    background: #e0f2fe;
    color: #075985;
    font-size: 0.88rem;
    font-weight: 800;
  }

  .admin-content {
    max-width: 1180px;
    margin: 0 auto;
    display: grid;
    grid-template-columns: minmax(0, 1fr) 340px;
    gap: 16px;
    margin-top: 16px;
  }

  .operations-content {
    max-width: 1180px;
    margin: 16px auto 0;
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(320px, 0.72fr);
    gap: 16px;
  }

  .side-stack,
  .operations-stack {
    display: grid;
    gap: 16px;
    align-content: start;
  }

  @media (max-width: 900px) {
    .admin-content,
    .operations-content {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 620px) {
    .admin-dashboard {
      padding: 18px;
    }

    .admin-topbar,
    .admin-actions {
      align-items: stretch;
    }

    .admin-topbar {
      flex-direction: column;
    }

    .admin-actions {
      width: 100%;
    }

    .admin-actions button {
      flex: 1;
    }
  }
</style>

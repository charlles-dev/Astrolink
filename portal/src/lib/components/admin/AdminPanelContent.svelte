<script lang="ts">
  import AdminBackupPanel from './AdminBackupPanel.svelte'
  import AdminLiveEventsPanel, {
    type AdminLiveEvent,
    type AdminLiveSnapshot
  } from './AdminLiveEventsPanel.svelte'
  import AdminLogsPanel from './AdminLogsPanel.svelte'
  import AdminMetrics from './AdminMetrics.svelte'
  import AdminNetworkPanel from './AdminNetworkPanel.svelte'
  import AdminPaymentsPanel from './AdminPaymentsPanel.svelte'
  import AdminPlansPanel from './AdminPlansPanel.svelte'
  import AdminSetupPanel from './AdminSetupPanel.svelte'
  import AdminUsersPanel from './AdminUsersPanel.svelte'
  import AdminUsersSummary from './AdminUsersSummary.svelte'
  import AdminVouchersPanel from './AdminVouchersPanel.svelte'
  import type { AdminPanelPage } from './AdminShell.svelte'
  import type {
    AdminHealthResponse,
    AdminBlacklistBody,
    AdminBlacklistEntry,
    AdminLog,
    AdminLogFilters,
    AdminPayment,
    AdminPaymentFilters,
    AdminPaymentTotals,
    AdminPlanBody,
    AdminRouter,
    AdminRouterBody,
    AdminRestoreBackupBody,
    AdminUser,
    AdminWalledGardenBody,
    AdminWalledGardenEntry,
    AdminVoucher,
    AdminVoucherFilters,
    GenerateAdminVouchersBody,
    Plano,
    SetupStatus
  } from '../../types'

  export let activePage: AdminPanelPage = 'overview'
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
  export let roteadores: AdminRouter[] = []
  export let blacklist: AdminBlacklistEntry[] = []
  export let walledGarden: AdminWalledGardenEntry[] = []
  export let setupStatus: SetupStatus | null = null
  export let liveConnected = false
  export let liveLastEventAt = ''
  export let liveSnapshot: AdminLiveSnapshot | null = null
  export let liveEvents: AdminLiveEvent[] = []
  export let loading = false
  export let backupMessage = ''
  export let setupMessage = ''
  export let onDisconnect: (mac: string) => void = () => {}
  export let onExtendUser: (mac: string, minutos: number) => Promise<void> | void = () => {}
  export let onBanUser: (mac: string, motivo: string) => Promise<void> | void = () => {}
  export let onSaveRouter: (input: AdminRouterBody, id?: number) => Promise<void> | void = () => {}
  export let onDeleteRouter: (id: number) => Promise<void> | void = () => {}
  export let onDiagnoseRouter: (id: number) => Promise<void> | void = () => {}
  export let onSpeedtestRouter: (id: number) => Promise<void> | void = () => {}
  export let onAddBlacklist: (input: AdminBlacklistBody) => Promise<void> | void = () => {}
  export let onDeleteBlacklist: (mac: string) => Promise<void> | void = () => {}
  export let onAddWalledGarden: (input: AdminWalledGardenBody) => Promise<void> | void = () => {}
  export let onDeleteWalledGarden: (id: number) => Promise<void> | void = () => {}
  export let onSavePlan: (input: AdminPlanBody, id?: number) => Promise<void> | void = () => {}
  export let onTogglePlanStatus: (id: number, ativo: boolean) => Promise<void> | void = () => {}
  export let onGenerateVouchers: (input: GenerateAdminVouchersBody) => void = () => {}
  export let onApplyVoucherFilters: (filters: AdminVoucherFilters) => void = () => {}
  export let onDeactivateVoucher: (id: number) => void = () => {}
  export let onExportVouchers: (filters: AdminVoucherFilters) => void = () => {}
  export let onApplyPaymentFilters: (filters: AdminPaymentFilters) => void = () => {}
  export let onExportPayments: (filters: AdminPaymentFilters) => void = () => {}
  export let onExportPaymentReport: (filters: AdminPaymentFilters) => void = () => {}
  export let onExportPaymentReportPDF: (filters: AdminPaymentFilters) => void = () => {}
  export let onApplyLogFilters: (filters: AdminLogFilters) => void = () => {}
  export let onExportLogs: (filters: AdminLogFilters) => void = () => {}
  export let onCreateBackup: () => void = () => {}
  export let onRestoreBackup: (input: AdminRestoreBackupBody) => void = () => {}
  export let onSaveSetup: (values: Record<string, string>) => Promise<void> | void = () => {}
</script>

{#if activePage === 'overview'}
  <section class="overview-stack" aria-label="Metricas do painel">
    <AdminMetrics {health} {planos} {usuarios} {vouchers} />
    <div class="overview-grid">
      <AdminLiveEventsPanel
        connected={liveConnected}
        lastEventAt={liveLastEventAt}
        snapshot={liveSnapshot}
        events={liveEvents}
      />
      <AdminUsersSummary {usuarios} {loading} />
    </div>
  </section>
{:else if activePage === 'usuarios'}
  <AdminUsersPanel {usuarios} {loading} {onDisconnect} {onExtendUser} {onBanUser} />
{:else if activePage === 'planos'}
  <AdminPlansPanel {planos} {loading} {onSavePlan} {onTogglePlanStatus} />
{:else if activePage === 'rede'}
  <AdminNetworkPanel
    {roteadores}
    {blacklist}
    {walledGarden}
    {loading}
    {onSaveRouter}
    {onDeleteRouter}
    {onDiagnoseRouter}
    {onSpeedtestRouter}
    {onAddBlacklist}
    {onDeleteBlacklist}
    {onAddWalledGarden}
    {onDeleteWalledGarden}
  />
{:else if activePage === 'vouchers'}
  <AdminVouchersPanel
    {planos}
    {vouchers}
    {loading}
    {onGenerateVouchers}
    {onApplyVoucherFilters}
    {onDeactivateVoucher}
    {onExportVouchers}
  />
{:else if activePage === 'pagamentos'}
  <AdminPaymentsPanel
    {pagamentos}
    totais={pagamentosTotais}
    {loading}
    {onApplyPaymentFilters}
    {onExportPayments}
    {onExportPaymentReport}
    {onExportPaymentReportPDF}
  />
{:else if activePage === 'setup'}
  <AdminSetupPanel {setupStatus} {loading} {setupMessage} {onSaveSetup} />
{:else if activePage === 'logs'}
  <section class="logs-grid" aria-label="Observabilidade local">
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
{/if}

<style>
  .overview-stack {
    display: grid;
    gap: var(--admin-section-gap);
  }

  .overview-grid {
    display: grid;
    grid-template-columns: minmax(300px, 0.8fr) minmax(0, 1.2fr);
    gap: var(--admin-section-gap);
  }

  .logs-grid {
    display: grid;
    grid-template-columns: minmax(300px, 0.8fr) minmax(0, 1.2fr) minmax(280px, 0.72fr);
    gap: var(--admin-section-gap);
  }

  @media (max-width: 900px) {
    .overview-grid,
    .logs-grid {
      grid-template-columns: 1fr;
    }
  }
</style>

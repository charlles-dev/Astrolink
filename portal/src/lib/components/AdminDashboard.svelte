<script lang="ts">
  import AdminPanelContent from './admin/AdminPanelContent.svelte'
  import AdminShell, { type AdminPanelPage } from './admin/AdminShell.svelte'
  import type { AdminLiveEvent, AdminLiveSnapshot } from './admin/AdminLiveEventsPanel.svelte'
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
  } from '../types'

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
  export let actionMessage = ''
  export let backupMessage = ''
  export let setupMessage = ''
  export let onRefresh: () => void = () => {}
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
  export let onLogout: () => void = () => {}
</script>

<AdminShell
  {activePage}
  {health}
  {usuarios}
  {liveConnected}
  {loading}
  {actionMessage}
  {onRefresh}
  {onLogout}
>
  <AdminPanelContent
    {activePage}
    {health}
    {planos}
    {usuarios}
    {vouchers}
    {pagamentos}
    {pagamentosTotais}
    {logs}
    {logsTotal}
    {roteadores}
    {blacklist}
    {walledGarden}
    {setupStatus}
    {liveConnected}
    {liveLastEventAt}
    {liveSnapshot}
    {liveEvents}
    {loading}
    {backupMessage}
    {setupMessage}
    {onDisconnect}
    {onExtendUser}
    {onBanUser}
    {onSaveRouter}
    {onDeleteRouter}
    {onDiagnoseRouter}
    {onSpeedtestRouter}
    {onAddBlacklist}
    {onDeleteBlacklist}
    {onAddWalledGarden}
    {onDeleteWalledGarden}
    {onSavePlan}
    {onTogglePlanStatus}
    {onGenerateVouchers}
    {onApplyVoucherFilters}
    {onDeactivateVoucher}
    {onExportVouchers}
    {onApplyPaymentFilters}
    {onExportPayments}
    {onExportPaymentReport}
    {onExportPaymentReportPDF}
    {onApplyLogFilters}
    {onExportLogs}
    {onCreateBackup}
    {onRestoreBackup}
    {onSaveSetup}
  />
</AdminShell>

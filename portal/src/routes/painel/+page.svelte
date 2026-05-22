<script lang="ts">
  import { onDestroy, onMount } from 'svelte'

  import { APIError, api } from '$lib/api'
  import AdminDashboard from '$lib/components/AdminDashboard.svelte'
  import type {
    AdminLiveEvent,
    AdminLiveSnapshot
  } from '$lib/components/admin/AdminLiveEventsPanel.svelte'
  import type {
    AdminHealthResponse,
    AdminLog,
    AdminLogFilters,
    AdminPlanBody,
    AdminLoginResponse,
    AdminPayment,
    AdminPaymentFilters,
    AdminPaymentTotals,
    AdminUser,
    AdminVoucher,
    AdminVoucherFilters,
    AdminRestoreBackupBody,
    GenerateAdminVouchersBody,
    Plano
  } from '$lib/types'

  const TOKEN_KEY = 'astrolink.admin.token'

  let usuario = 'admin'
  let senha = 'admin123'
  let token = ''
  let health: AdminHealthResponse | null = null
  let planos: Plano[] = []
  let usuarios: AdminUser[] = []
  let vouchers: AdminVoucher[] = []
  let pagamentos: AdminPayment[] = []
  let pagamentosTotais: AdminPaymentTotals = emptyPaymentTotals()
  let logs: AdminLog[] = []
  let logsTotal = 0
  let liveConnected = false
  let liveLastEventAt = ''
  let liveSnapshot: AdminLiveSnapshot | null = null
  let liveEvents: AdminLiveEvent[] = []
  let voucherFilters: AdminVoucherFilters = { status: 'ativo' }
  let paymentFilters: AdminPaymentFilters = {}
  let logFilters: AdminLogFilters = {}
  let loading = false
  let loginLoading = false
  let loginError = ''
  let actionMessage = ''
  let backupMessage = ''
  let liveAbortController: AbortController | null = null
  let liveBuffer = ''

  onMount(() => {
    token = sessionStorage.getItem(TOKEN_KEY) || ''
    if (token) {
      void initializeDashboard()
    }
  })

  onDestroy(() => {
    stopLiveEvents()
  })

  async function login() {
    loginLoading = true
    loginError = ''
    actionMessage = ''
    try {
      const result: AdminLoginResponse = await api.loginAdmin({ usuario, senha })
      token = result.access_token
      sessionStorage.setItem(TOKEN_KEY, token)
      await loadDashboard()
      startLiveEvents()
    } catch (error) {
      loginError = messageFromError(error, 'Nao foi possivel entrar no painel')
    } finally {
      loginLoading = false
    }
  }

  async function initializeDashboard() {
    await loadDashboard()
    if (token) startLiveEvents()
  }

  async function loadDashboard() {
    if (!token) return
    loading = true
    actionMessage = ''
    try {
      const [nextHealth, nextPlanos, nextUsuarios, nextVouchers] = await Promise.all([
        api.getAdminHealth(token),
        api.getAdminPlanos(token),
        api.getAdminUsuarios(token),
        api.getAdminVouchers(token, voucherFilters)
      ])
      health = nextHealth
      planos = nextPlanos.planos
      usuarios = nextUsuarios.usuarios
      vouchers = nextVouchers.vouchers
      await loadOperations()
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Nao foi possivel carregar o painel')
    } finally {
      loading = false
    }
  }

  async function loadOperations() {
    const [paymentsResult, logsResult] = await Promise.allSettled([
      api.getAdminPagamentos(token, paymentFilters),
      api.getAdminLogs(token, logFilters)
    ])

    if (paymentsResult.status === 'fulfilled') {
      pagamentos = paymentsResult.value.pagamentos
      pagamentosTotais = paymentsResult.value.totais
    } else if (!expireSessionIfUnauthorized(paymentsResult.reason)) {
      pagamentos = []
      pagamentosTotais = emptyPaymentTotals()
      actionMessage = messageFromError(paymentsResult.reason, 'Pagamentos indisponiveis no momento')
    }

    if (logsResult.status === 'fulfilled') {
      logs = logsResult.value.logs
      logsTotal = logsResult.value.total
    } else if (!expireSessionIfUnauthorized(logsResult.reason)) {
      logs = []
      logsTotal = 0
      actionMessage = messageFromError(logsResult.reason, 'Logs indisponiveis no momento')
    }
  }

  async function disconnect(mac: string) {
    if (!token) return
    loading = true
    actionMessage = ''
    try {
      await api.disconnectAdminUsuario(token, mac)
      await loadDashboard()
      actionMessage = `${mac} desconectado do roteador`
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Nao foi possivel desconectar o usuario')
    } finally {
      loading = false
    }
  }

  async function generateVouchers(input: GenerateAdminVouchersBody) {
    if (!token) return
    loading = true
    actionMessage = ''
    try {
      const result = await api.generateAdminVouchers(token, input)
      await reloadVouchers()
      actionMessage =
        result.quantidade === 1
          ? '1 voucher gerado'
          : `${result.quantidade} vouchers gerados`
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Nao foi possivel gerar vouchers')
    } finally {
      loading = false
    }
  }

  async function applyVoucherFilters(filters: AdminVoucherFilters) {
    if (!token) return
    loading = true
    actionMessage = ''
    voucherFilters = filters
    try {
      await reloadVouchers()
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Nao foi possivel carregar vouchers')
    } finally {
      loading = false
    }
  }

  async function deactivateVoucher(id: number) {
    if (!token) return
    loading = true
    actionMessage = ''
    try {
      await api.deactivateAdminVoucher(token, id)
      await reloadVouchers()
      actionMessage = 'Voucher desativado'
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Nao foi possivel desativar o voucher')
    } finally {
      loading = false
    }
  }

  async function exportVouchers(filters: AdminVoucherFilters) {
    if (!token) return
    loading = true
    actionMessage = ''
    voucherFilters = filters
    try {
      const csv = await api.exportAdminVouchers(token, filters)
      downloadBlob(csv, 'astrolink-vouchers.csv')
      actionMessage = 'Exportacao iniciada'
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Nao foi possivel exportar vouchers')
    } finally {
      loading = false
    }
  }

  async function applyPaymentFilters(filters: AdminPaymentFilters) {
    if (!token) return
    loading = true
    actionMessage = ''
    paymentFilters = filters
    try {
      await reloadPayments()
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Nao foi possivel carregar pagamentos')
    } finally {
      loading = false
    }
  }

  async function exportPayments(filters: AdminPaymentFilters) {
    if (!token) return
    loading = true
    actionMessage = ''
    paymentFilters = filters
    try {
      const csv = await api.exportAdminPagamentos(token, filters)
      downloadBlob(csv, 'astrolink-pagamentos.csv')
      actionMessage = 'Exportacao iniciada'
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Nao foi possivel exportar pagamentos')
    } finally {
      loading = false
    }
  }

  async function applyLogFilters(filters: AdminLogFilters) {
    if (!token) return
    loading = true
    actionMessage = ''
    logFilters = filters
    try {
      await reloadLogs()
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Nao foi possivel carregar logs')
    } finally {
      loading = false
    }
  }

  async function exportLogs(filters: AdminLogFilters) {
    if (!token) return
    loading = true
    actionMessage = ''
    logFilters = filters
    try {
      const csv = await api.exportAdminLogs(token, filters)
      downloadBlob(csv, 'astrolink-logs.csv')
      actionMessage = 'Exportacao iniciada'
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Nao foi possivel exportar logs')
    } finally {
      loading = false
    }
  }

  async function createBackup() {
    if (!token) return
    loading = true
    actionMessage = ''
    backupMessage = ''
    try {
      const result = await api.createAdminBackup(token)
      backupMessage = result.mensagem || 'Backup solicitado'
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      backupMessage = messageFromError(
        error,
        'Backup indisponivel neste ambiente. Tente novamente quando o servico estiver ativo.'
      )
    } finally {
      loading = false
    }
  }

  async function restoreBackup(input: AdminRestoreBackupBody) {
    if (!token) return
    loading = true
    actionMessage = ''
    backupMessage = ''
    try {
      const result = await api.restoreAdminBackup(token, input)
      backupMessage = result.mensagem || 'Restore validado'
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      backupMessage = messageFromError(
        error,
        'Restore indisponivel neste ambiente. Nenhuma restauracao foi executada.'
      )
    } finally {
      loading = false
    }
  }

  async function reloadVouchers() {
    const result = await api.getAdminVouchers(token, voucherFilters)
    vouchers = result.vouchers
  }

  async function reloadPayments() {
    const result = await api.getAdminPagamentos(token, paymentFilters)
    pagamentos = result.pagamentos
    pagamentosTotais = result.totais
  }

  async function reloadLogs() {
    const result = await api.getAdminLogs(token, logFilters)
    logs = result.logs
    logsTotal = result.total
  }

  async function savePlan(input: AdminPlanBody, id?: number) {
    if (!token) return
    loading = true
    actionMessage = ''
    try {
      if (id) {
        await api.updateAdminPlano(token, id, input)
        actionMessage = 'Plano atualizado'
      } else {
        await api.createAdminPlano(token, input)
        actionMessage = 'Plano criado'
      }
      const result = await api.getAdminPlanos(token)
      planos = result.planos
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Nao foi possivel salvar o plano')
      throw error
    } finally {
      loading = false
    }
  }

  async function togglePlanStatus(id: number, ativo: boolean) {
    if (!token) return
    loading = true
    actionMessage = ''
    try {
      await api.updateAdminPlanoStatus(token, id, ativo)
      const result = await api.getAdminPlanos(token)
      planos = result.planos
      actionMessage = ativo ? 'Plano ativado' : 'Plano inativado'
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Nao foi possivel alterar o status do plano')
      throw error
    } finally {
      loading = false
    }
  }

  function logout() {
    resetSession()
    actionMessage = ''
    loginError = ''
  }

  function resetSession() {
    stopLiveEvents()
    sessionStorage.removeItem(TOKEN_KEY)
    token = ''
    health = null
    planos = []
    usuarios = []
    vouchers = []
    pagamentos = []
    pagamentosTotais = emptyPaymentTotals()
    logs = []
    logsTotal = 0
    liveConnected = false
    liveLastEventAt = ''
    liveSnapshot = null
    liveEvents = []
    voucherFilters = { status: 'ativo' }
    paymentFilters = {}
    logFilters = {}
    backupMessage = ''
  }

  function startLiveEvents() {
    if (!token) return

    stopLiveEvents()
    liveAbortController = new AbortController()
    liveConnected = false
    liveBuffer = ''
    void readLiveEvents(token, liveAbortController)
  }

  function stopLiveEvents() {
    liveAbortController?.abort()
    liveAbortController = null
    liveConnected = false
    liveBuffer = ''
  }

  async function readLiveEvents(currentToken: string, controller: AbortController) {
    try {
      const response = await fetch('/admin/eventos', {
        headers: { Authorization: `Bearer ${currentToken}` },
        signal: controller.signal
      })
      if (response.status === 401) {
        expireSessionIfUnauthorized(new APIError(401, 'nao_autorizado', 'Sessao expirada. Entre novamente.'))
        return
      }
      if (!response.ok || !response.body) {
        liveConnected = false
        return
      }

      liveConnected = true
      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      while (true) {
        const { done, value } = await reader.read()
        if (done) break
        liveBuffer += decoder.decode(value, { stream: true })
        drainLiveBuffer()
      }
    } catch {
      if (!controller.signal.aborted) {
        liveConnected = false
      }
    }
  }

  function drainLiveBuffer() {
    const chunks = liveBuffer.split('\n\n')
    liveBuffer = chunks.pop() ?? ''
    chunks.forEach(processLiveEventBlock)
  }

  function processLiveEventBlock(block: string) {
    const lines = block.split('\n')
    const eventName =
      lines.find((line) => line.startsWith('event:'))?.slice('event:'.length).trim() || 'message'
    const data = lines
      .filter((line) => line.startsWith('data:'))
      .map((line) => line.slice('data:'.length).trim())
      .join('\n')
    if (!data) return

    try {
      const payload = JSON.parse(data)
      if (eventName === 'snapshot') {
        applyLiveSnapshot(payload)
      } else {
        pushLiveEvent({
          tipo: eventName,
          mensagem: payload.mensagem ?? 'Evento recebido',
          timestamp: payload.timestamp ?? new Date().toISOString()
        })
      }
    } catch {
      pushLiveEvent({
        tipo: eventName,
        mensagem: data,
        timestamp: new Date().toISOString()
      })
    }
  }

  function applyLiveSnapshot(payload: Record<string, unknown>) {
    const timestamp = String(payload.timestamp ?? new Date().toISOString())
    liveLastEventAt = timestamp
    liveSnapshot = {
      usuarios: {
        ativos: Number(payload.usuarios_ativos ?? 0),
        total: Number(payload.usuarios_total ?? 0)
      },
      vouchers: {
        ativos: Number(payload.vouchers_ativos ?? 0),
        total: Number(payload.vouchers_total ?? 0)
      },
      pix: {
        pendente: Number(payload.pagamentos_pendentes ?? 0),
        aprovado: Number(payload.pagamentos_aprovados ?? 0)
      },
      logs: Number(payload.logs_total ?? 0)
    }
    pushLiveEvent({
      tipo: 'snapshot',
      mensagem: 'Estado operacional atualizado',
      timestamp
    })
  }

  function pushLiveEvent(event: AdminLiveEvent) {
    liveEvents = [
      {
        id: `${event.timestamp}-${event.tipo}-${event.mensagem}`,
        ...event
      },
      ...liveEvents
    ].slice(0, 8)
  }

  function emptyPaymentTotals(): AdminPaymentTotals {
    return {
      pendente: 0,
      aprovado: 0,
      cancelado: 0,
      expirado: 0,
      valor_total: '0.00'
    }
  }

  function downloadBlob(blob: Blob, filename: string) {
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    document.body.appendChild(link)
    link.click()
    link.remove()
    URL.revokeObjectURL(url)
  }

  function expireSessionIfUnauthorized(error: unknown) {
    if (!(error instanceof APIError) || error.status !== 401) return false
    resetSession()
    loginError = 'Sessao expirada. Entre novamente.'
    actionMessage = ''
    return true
  }

  function messageFromError(error: unknown, fallback: string) {
    if (error instanceof APIError) return error.message
    if (error instanceof Error && error.message) return error.message
    return fallback
  }
</script>

{#if token}
  <AdminDashboard
    {health}
    {planos}
    {usuarios}
    {vouchers}
    {pagamentos}
    {pagamentosTotais}
    {logs}
    {logsTotal}
    {liveConnected}
    {liveLastEventAt}
    {liveSnapshot}
    {liveEvents}
    {loading}
    {actionMessage}
    {backupMessage}
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
    onLogout={logout}
  />
{:else}
  <main class="login-screen">
    <section class="login-panel">
      <div class="brand-mark">A</div>
      <h1>Painel local</h1>
      <p>Entre para acompanhar saude do no, planos e usuarios conectados.</p>

      <form onsubmit={(event) => { event.preventDefault(); void login() }}>
        <label>
          Usuario
          <input bind:value={usuario} autocomplete="username" />
        </label>
        <label>
          Senha
          <input bind:value={senha} type="password" autocomplete="current-password" />
        </label>
        {#if loginError}
          <p class="login-error" role="alert">{loginError}</p>
        {/if}
        <button type="submit" disabled={loginLoading}>
          {loginLoading ? 'Entrando...' : 'Entrar'}
        </button>
      </form>
    </section>
  </main>
{/if}

<style>
  .login-screen {
    min-height: 100vh;
    display: grid;
    place-items: center;
    padding: 24px;
    background:
      radial-gradient(circle at top left, rgba(56, 189, 248, 0.16), transparent 32%),
      #f8fafc;
  }

  .login-panel {
    width: min(100%, 420px);
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 28px;
    background: white;
    box-shadow: 0 22px 52px rgba(15, 23, 42, 0.14);
  }

  .brand-mark {
    width: 44px;
    height: 44px;
    display: grid;
    place-items: center;
    border-radius: 8px;
    background: var(--color-ink);
    color: white;
    font-weight: 950;
  }

  h1,
  p {
    margin: 0;
  }

  h1 {
    margin-top: 18px;
    color: var(--color-ink);
    font-size: 1.8rem;
    font-weight: 920;
    line-height: 1.05;
  }

  .login-panel > p {
    margin-top: 8px;
    color: var(--color-muted);
    line-height: 1.45;
  }

  form {
    display: grid;
    gap: 14px;
    margin-top: 24px;
  }

  label {
    display: grid;
    gap: 7px;
    color: var(--color-ink);
    font-size: 0.86rem;
    font-weight: 850;
  }

  input {
    min-height: 48px;
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 0 13px;
    background: #f8fafc;
    color: var(--color-ink);
  }

  button {
    min-height: 50px;
    border: 0;
    border-radius: 8px;
    background: var(--color-ink);
    color: white;
    font-weight: 900;
  }

  button:disabled {
    cursor: wait;
    opacity: 0.7;
  }

  .login-error {
    border-radius: 8px;
    padding: 12px;
    background: #fee2e2;
    color: #991b1b;
    font-size: 0.88rem;
    font-weight: 800;
  }
</style>

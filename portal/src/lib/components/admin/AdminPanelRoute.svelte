<script lang="ts">
  import { onDestroy, onMount } from 'svelte'

  import { APIError, api } from '$lib/api'
  import AdminDashboard from '$lib/components/AdminDashboard.svelte'
  import type { AdminPanelPage } from '$lib/components/admin/AdminShell.svelte'
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
    AdminBlacklistBody,
    AdminBlacklistEntry,
    AdminRouter,
    AdminRouterBody,
    AdminUser,
    AdminWalledGardenBody,
    AdminWalledGardenEntry,
    AdminVoucher,
    AdminVoucherFilters,
    AdminRestoreBackupBody,
    GenerateAdminVouchersBody,
    Plano,
    SetupStatus
  } from '$lib/types'

  const TOKEN_KEY = 'astrolink.admin.token'

  export let activePage: AdminPanelPage = 'overview'

  let usuario = 'admin'
  let senha = 'admin123'
  let totpCodigo = ''
  let showTotp = false
  let token = ''
  let health: AdminHealthResponse | null = null
  let planos: Plano[] = []
  let usuarios: AdminUser[] = []
  let vouchers: AdminVoucher[] = []
  let pagamentos: AdminPayment[] = []
  let pagamentosTotais: AdminPaymentTotals = emptyPaymentTotals()
  let logs: AdminLog[] = []
  let logsTotal = 0
  let roteadores: AdminRouter[] = []
  let blacklist: AdminBlacklistEntry[] = []
  let walledGarden: AdminWalledGardenEntry[] = []
  let setupStatus: SetupStatus | null = null
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
  let setupMessage = ''
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
      const code = totpCodigo.trim()
      const result: AdminLoginResponse = await api.loginAdmin({
        usuario,
        senha,
        ...(showTotp && code ? { totp_codigo: code } : {})
      })
      token = result.access_token
      totpCodigo = ''
      showTotp = false
      sessionStorage.setItem(TOKEN_KEY, token)
      await loadDashboard()
      startLiveEvents()
    } catch (error) {
      if (error instanceof APIError && error.status === 428 && error.code === 'totp_obrigatorio') {
        showTotp = true
      }
      loginError = messageFromError(error, 'Não foi possível entrar no console')
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
      const optionalLoads = [loadOperations(), loadSetupStatus()]
      if (activePage === 'rede') optionalLoads.push(loadNetwork())
      await Promise.allSettled(optionalLoads)
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Não foi possível carregar o console')
    } finally {
      loading = false
    }
  }

  async function loadSetupStatus() {
    try {
      setupStatus = await api.getSetupStatus(token)
      setupMessage = ''
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) throw error
      setupStatus = null
      setupMessage = messageFromError(error, 'Setup local indisponível')
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
      actionMessage = messageFromError(paymentsResult.reason, 'Pagamentos indisponíveis no momento')
    }

    if (logsResult.status === 'fulfilled') {
      logs = logsResult.value.logs
      logsTotal = logsResult.value.total
    } else if (!expireSessionIfUnauthorized(logsResult.reason)) {
      logs = []
      logsTotal = 0
      actionMessage = messageFromError(logsResult.reason, 'Logs indisponíveis no momento')
    }
  }

  async function loadNetwork() {
    const [routersResult, blacklistResult, gardenResult] = await Promise.allSettled([
      api.getAdminRouters(token),
      api.getAdminBlacklist(token),
      api.getAdminWalledGarden(token)
    ])

    if (routersResult.status === 'fulfilled') {
      roteadores = routersResult.value.roteadores
    } else if (!expireSessionIfUnauthorized(routersResult.reason)) {
      roteadores = []
      actionMessage = messageFromError(routersResult.reason, 'Roteadores indisponiveis no momento')
    }

    if (blacklistResult.status === 'fulfilled') {
      blacklist = blacklistResult.value.blacklist
    } else if (!expireSessionIfUnauthorized(blacklistResult.reason)) {
      blacklist = []
      actionMessage = messageFromError(blacklistResult.reason, 'Blacklist indisponivel no momento')
    }

    if (gardenResult.status === 'fulfilled') {
      walledGarden = gardenResult.value.walled_garden
    } else if (!expireSessionIfUnauthorized(gardenResult.reason)) {
      walledGarden = []
      actionMessage = messageFromError(gardenResult.reason, 'Walled garden indisponivel no momento')
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
      actionMessage = messageFromError(error, 'Não foi possível desconectar o usuário')
    } finally {
      loading = false
    }
  }

  async function extendUser(mac: string, minutos: number) {
    if (!token) return
    loading = true
    actionMessage = ''
    try {
      await api.extendAdminUsuario(token, mac, { minutos })
      const result = await api.getAdminUsuarios(token)
      usuarios = result.usuarios
      actionMessage = `${minutos} minutos adicionados para ${mac}`
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Nao foi possivel estender o acesso')
    } finally {
      loading = false
    }
  }

  async function banUser(mac: string, motivo: string) {
    if (!token) return
    loading = true
    actionMessage = ''
    try {
      await api.banAdminUsuario(token, mac, { motivo })
      const [usersResult, blacklistResult] = await Promise.all([
        api.getAdminUsuarios(token),
        api.getAdminBlacklist(token)
      ])
      usuarios = usersResult.usuarios
      blacklist = blacklistResult.blacklist
      actionMessage = `${mac} bloqueado na blacklist local`
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Nao foi possivel bloquear o usuario')
    } finally {
      loading = false
    }
  }

  async function saveRouter(input: AdminRouterBody, id?: number) {
    if (!token) return
    loading = true
    actionMessage = ''
    try {
      if (id) {
        await api.updateAdminRouter(token, id, input)
        actionMessage = 'Roteador atualizado'
      } else {
        await api.createAdminRouter(token, input)
        actionMessage = 'Roteador cadastrado'
      }
      await loadNetwork()
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Nao foi possivel salvar o roteador')
      throw error
    } finally {
      loading = false
    }
  }

  async function deleteRouter(id: number) {
    if (!token) return
    loading = true
    actionMessage = ''
    try {
      await api.deleteAdminRouter(token, id)
      await loadNetwork()
      actionMessage = 'Roteador removido'
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Nao foi possivel remover o roteador')
    } finally {
      loading = false
    }
  }

  async function diagnoseRouter(id: number) {
    if (!token) return
    loading = true
    actionMessage = ''
    try {
      const result = await api.diagnoseAdminRouter(token, id)
      actionMessage =
        result.status === 'online'
          ? 'Diagnostico concluido: roteador online'
          : `Diagnostico concluido: ${result.status}`
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Nao foi possivel diagnosticar o roteador')
    } finally {
      loading = false
    }
  }

  async function speedtestRouter(id: number) {
    if (!token) return
    loading = true
    actionMessage = ''
    try {
      const result = await api.speedtestAdminRouter(token, id)
      actionMessage = result.mensagem
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Nao foi possivel medir velocidade')
    } finally {
      loading = false
    }
  }

  async function addBlacklist(input: AdminBlacklistBody) {
    if (!token) return
    loading = true
    actionMessage = ''
    try {
      await api.addAdminBlacklist(token, input)
      await loadNetwork()
      actionMessage = 'MAC adicionado a blacklist'
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Nao foi possivel adicionar a blacklist')
      throw error
    } finally {
      loading = false
    }
  }

  async function deleteBlacklist(mac: string) {
    if (!token) return
    loading = true
    actionMessage = ''
    try {
      await api.deleteAdminBlacklist(token, mac)
      await loadNetwork()
      actionMessage = 'MAC removido da blacklist'
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Nao foi possivel remover da blacklist')
    } finally {
      loading = false
    }
  }

  async function addWalledGarden(input: AdminWalledGardenBody) {
    if (!token) return
    loading = true
    actionMessage = ''
    try {
      await api.addAdminWalledGarden(token, input)
      await loadNetwork()
      actionMessage = 'Host adicionado ao walled garden'
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Nao foi possivel adicionar ao walled garden')
      throw error
    } finally {
      loading = false
    }
  }

  async function deleteWalledGarden(id: number) {
    if (!token) return
    loading = true
    actionMessage = ''
    try {
      await api.deleteAdminWalledGarden(token, id)
      await loadNetwork()
      actionMessage = 'Host removido do walled garden'
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Nao foi possivel remover do walled garden')
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
      actionMessage = messageFromError(error, 'Não foi possível gerar vouchers')
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
      actionMessage = messageFromError(error, 'Não foi possível carregar vouchers')
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
      actionMessage = messageFromError(error, 'Não foi possível desativar o voucher')
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
      actionMessage = 'Exportação iniciada'
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Não foi possível exportar vouchers')
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
      actionMessage = messageFromError(error, 'Não foi possível carregar pagamentos')
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
      actionMessage = 'Exportação iniciada'
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Não foi possível exportar pagamentos')
    } finally {
      loading = false
    }
  }

  async function exportPaymentReport(filters: AdminPaymentFilters) {
    if (!token) return
    loading = true
    actionMessage = ''
    paymentFilters = filters
    try {
      const csv = await api.exportAdminPagamentosRelatorio(token, filters)
      downloadBlob(csv, 'astrolink-relatorio-pagamentos.csv')
      actionMessage = 'Relatorio de pagamentos iniciado'
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Não foi possível exportar relatório de pagamentos')
    } finally {
      loading = false
    }
  }

  async function exportPaymentReportPDF(filters: AdminPaymentFilters) {
    if (!token) return
    loading = true
    actionMessage = ''
    paymentFilters = filters
    try {
      const pdf = await api.exportAdminPagamentosRelatorioPDF(token, filters)
      downloadBlob(pdf, 'astrolink-relatorio-pagamentos.pdf')
      actionMessage = 'Relatorio PDF iniciado'
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Não foi possível exportar relatório PDF')
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
      actionMessage = messageFromError(error, 'Não foi possível carregar logs')
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
      actionMessage = 'Exportação iniciada'
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Não foi possível exportar logs')
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
        'Backup indisponível neste ambiente. Tente novamente quando o serviço estiver ativo.'
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
        'Restore indisponível neste ambiente. Nenhuma restauração foi executada.'
      )
    } finally {
      loading = false
    }
  }

  async function saveSetup(values: Record<string, string>) {
    if (!token) return
    loading = true
    actionMessage = ''
    setupMessage = ''
    try {
      setupStatus = await api.updateSetupEnv(values, token)
      setupMessage = setupStatus.requires_restart
        ? 'Setup local salvo. Reinicie o serviço para aplicar.'
        : 'Setup local salvo'
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      setupMessage = messageFromError(error, 'Não foi possível salvar o setup local')
      throw error
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
      actionMessage = messageFromError(error, 'Não foi possível salvar o plano')
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
      actionMessage = messageFromError(error, 'Não foi possível alterar o status do plano')
      throw error
    } finally {
      loading = false
    }
  }

  function logout() {
    resetSession()
    actionMessage = ''
    loginError = ''
    totpCodigo = ''
    showTotp = false
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
    roteadores = []
    blacklist = []
    walledGarden = []
    setupStatus = null
    liveConnected = false
    liveLastEventAt = ''
    liveSnapshot = null
    liveEvents = []
    voucherFilters = { status: 'ativo' }
    paymentFilters = {}
    logFilters = {}
    backupMessage = ''
    setupMessage = ''
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
        expireSessionIfUnauthorized(new APIError(401, 'nao_autorizado', 'Sessão expirada. Entre novamente.'))
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
    loginError = 'Sessão expirada. Entre novamente.'
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
    {actionMessage}
    {backupMessage}
    {setupMessage}
    onRefresh={loadDashboard}
    onDisconnect={disconnect}
    onExtendUser={extendUser}
    onBanUser={banUser}
    onSaveRouter={saveRouter}
    onDeleteRouter={deleteRouter}
    onDiagnoseRouter={diagnoseRouter}
    onSpeedtestRouter={speedtestRouter}
    onAddBlacklist={addBlacklist}
    onDeleteBlacklist={deleteBlacklist}
    onAddWalledGarden={addWalledGarden}
    onDeleteWalledGarden={deleteWalledGarden}
    onSavePlan={savePlan}
    onTogglePlanStatus={togglePlanStatus}
    onGenerateVouchers={generateVouchers}
    onApplyVoucherFilters={applyVoucherFilters}
    onDeactivateVoucher={deactivateVoucher}
    onExportVouchers={exportVouchers}
    onApplyPaymentFilters={applyPaymentFilters}
    onExportPayments={exportPayments}
    onExportPaymentReport={exportPaymentReport}
    onExportPaymentReportPDF={exportPaymentReportPDF}
    onApplyLogFilters={applyLogFilters}
    onExportLogs={exportLogs}
    onCreateBackup={createBackup}
    onRestoreBackup={restoreBackup}
    onSaveSetup={saveSetup}
    onLogout={logout}
  />
{:else}
  <main class="login-screen" data-theme="astrolink">
    <section class="login-panel card">
      <div class="brand-mark">A</div>
      <h1>Console operacional local</h1>
      <p>Acesse o node para acompanhar saúde, sessões, planos e rotinas protegidas.</p>

      <form onsubmit={(event) => { event.preventDefault(); void login() }}>
        <label>
          Usuário
          <input class="input input-bordered" bind:value={usuario} autocomplete="username" />
        </label>
        <label>
          Senha
          <input class="input input-bordered" bind:value={senha} type="password" autocomplete="current-password" />
        </label>
        {#if showTotp}
          <label>
            Código 2FA
            <input
              bind:value={totpCodigo}
              class="input input-bordered"
              inputmode="numeric"
              autocomplete="one-time-code"
              maxlength="8"
            />
          </label>
        {/if}
        {#if loginError}
          <p class="login-error" role="alert">{loginError}</p>
        {/if}
        <button type="submit" class="btn btn-primary" disabled={loginLoading}>
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
    background: var(--color-paper);
  }

  .login-panel {
    width: min(100%, 420px);
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 32px;
    background: var(--color-surface-raised);
    box-shadow: var(--shadow-soft);
  }

  .brand-mark {
    width: 44px;
    height: 44px;
    display: grid;
    place-items: center;
    border-radius: 8px;
    background: var(--color-primary);
    color: var(--color-surface);
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
    gap: 16px;
    margin-top: 28px;
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
    border-radius: 8px;
    padding: 0 13px;
  }

  button {
    min-height: 50px;
    border-radius: 8px;
    font-weight: 900;
  }

  button:disabled {
    cursor: wait;
    opacity: 0.7;
  }

  .login-error {
    border-radius: 8px;
    padding: 12px;
    background: var(--state-error-bg);
    color: var(--state-error-text);
    font-size: 0.88rem;
    font-weight: 800;
  }
</style>

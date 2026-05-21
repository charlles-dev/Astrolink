<script lang="ts">
  import { onMount } from 'svelte'

  import '../app.css'
  import { APIError, api } from '$lib/api'
  import type {
    DeviceInfo,
    PixStatusResponse,
    PixTransaction,
    Plano,
    ResgatarVoucherResponse,
    SessaoStatus,
    Settings
  } from '$lib/types'
  import PixScreen from '$lib/components/PixScreen.svelte'
  import PlanSelection from '$lib/components/PlanSelection.svelte'
  import PortalShell from '$lib/components/PortalShell.svelte'
  import SuccessScreen from '$lib/components/SuccessScreen.svelte'
  import VoucherScreen from '$lib/components/VoucherScreen.svelte'
  import WelcomeScreen from '$lib/components/WelcomeScreen.svelte'

  type Step = 'welcome' | 'plans' | 'voucher' | 'pix' | 'success'

  const SETTINGS_CACHE_KEY = 'astrolink.portal.settings'
  const SETTINGS_CACHE_TTL = 5 * 60 * 1000

  const defaultSettings: Settings = {
    hotspot_nome: 'Astrolink Wi-Fi',
    hotspot_logo_url: '',
    cor_primaria: '#38BDF8',
    cor_secundaria: '#0EA5A8',
    cor_fundo: '#0F172A',
    mensagem_boas_vindas: 'Bem-vindo! Conecte-se e aproveite.',
    url_pos_conexao: 'https://google.com',
    coleta_nome: false,
    mostrar_velocidade: true
  }

  let step: Step = 'welcome'
  let settings = defaultSettings
  let device: DeviceInfo = { mac: '00:00:00:00:00:00', ip: '0.0.0.0', token: '' }
  let planos: Plano[] = []
  let selectedPlan: Plano | null = null
  let pix: PixTransaction | null = null
  let session: SessaoStatus | null = null

  let initialLoading = true
  let plansLoading = false
  let pixLoading = false
  let voucherLoading = false

  let welcomeError = ''
  let plansError = ''
  let pixError = ''
  let voucherError = ''
  let copyMessage = ''
  let successPlanName = 'Acesso liberado'
  let successSeconds = 0
  let pixSeconds = 0

  let pixEvents: EventSource | undefined
  let pixPoll: number | undefined
  let pixCountdown: number | undefined
  let successCountdown: number | undefined
  let copyMessageTimer: number | undefined

  onMount(() => {
    device = readDeviceFromURL()
    applySettings(settings)
    void bootstrap()

    return () => {
      cleanupTimers()
    }
  })

  async function bootstrap() {
    initialLoading = true
    welcomeError = ''
    await loadSettings()
    await checkActiveSession()
    initialLoading = false
  }

  function readDeviceFromURL(): DeviceInfo {
    const params = new URLSearchParams(window.location.search)
    return {
      mac: params.get('mac') || '00:00:00:00:00:00',
      ip: params.get('ip') || '0.0.0.0',
      token: params.get('token') || ''
    }
  }

  async function loadSettings() {
    const cached = readCachedSettings()
    if (cached) {
      settings = cached
      applySettings(settings)
    }

    try {
      const fresh = await api.getSettings()
      settings = { ...defaultSettings, ...fresh }
      applySettings(settings)
      localStorage.setItem(
        SETTINGS_CACHE_KEY,
        JSON.stringify({ storedAt: Date.now(), value: settings })
      )
    } catch (error) {
      if (!cached) {
        welcomeError = messageFromError(error, 'Nao foi possivel conectar ao servidor local')
      }
    }
  }

  function readCachedSettings(): Settings | null {
    try {
      const raw = localStorage.getItem(SETTINGS_CACHE_KEY)
      if (!raw) return null
      const parsed = JSON.parse(raw) as { storedAt?: number; value?: Settings }
      if (!parsed.storedAt || !parsed.value) return null
      if (Date.now() - parsed.storedAt > SETTINGS_CACHE_TTL) return null
      return { ...defaultSettings, ...parsed.value }
    } catch {
      return null
    }
  }

  async function checkActiveSession() {
    try {
      session = await api.getSessaoStatus(device.mac)
      if (session.ativa) {
        finishAccess(session.plano || 'Sessao ativa', session.tempo_restante_segundos || 0)
      }
    } catch (error) {
      welcomeError ||= messageFromError(error, 'Nao foi possivel verificar a sessao')
    }
  }

  async function showPlans() {
    step = 'plans'
    stopPixWatchers()
    if (planos.length === 0 || plansError) {
      await loadPlans()
    }
  }

  function showVoucher() {
    step = 'voucher'
    voucherError = ''
    stopPixWatchers()
  }

  async function loadPlans() {
    plansLoading = true
    plansError = ''
    try {
      const result = await api.getPlanos()
      planos = result.planos
    } catch (error) {
      plansError = messageFromError(error, 'Nao foi possivel carregar os planos')
    } finally {
      plansLoading = false
    }
  }

  async function selectPlan(plano: Plano) {
    selectedPlan = plano
    pix = null
    pixError = ''
    copyMessage = ''
    pixLoading = true
    step = 'pix'
    stopPixWatchers()

    try {
      const transaction = await api.gerarPix({
        plano_id: plano.id,
        mac: device.mac,
        ip: device.ip
      })
      pix = transaction
      startPixCountdown(transaction.expira_em_segundos)
      startPixWatchers(transaction.txid)
    } catch (error) {
      pixError = messageFromError(error, 'Nao foi possivel gerar o PIX')
    } finally {
      pixLoading = false
    }
  }

  async function redeemVoucher(codigo: string) {
    voucherLoading = true
    voucherError = ''
    try {
      const result = await api.resgatarVoucher({
        codigo,
        mac: device.mac,
        ip: device.ip
      })
      handleVoucherSuccess(result)
    } catch (error) {
      voucherError = messageFromError(error, 'Codigo invalido ou ja utilizado')
    } finally {
      voucherLoading = false
    }
  }

  function handleVoucherSuccess(result: ResgatarVoucherResponse) {
    finishAccess(result.plano, result.tempo_restante_segundos)
  }

  function finishAccess(planName: string, seconds: number) {
    stopPixWatchers()
    successPlanName = planName
    successSeconds = seconds
    step = 'success'
    startSuccessCountdown(seconds)
  }

  function startPixWatchers(txid: string) {
    stopPixWatchers()
    if (typeof EventSource === 'undefined') {
      startPixPolling(txid)
      return
    }

    try {
      const source = new EventSource(`/api/pix/aguardar/${encodeURIComponent(txid)}`)
      pixEvents = source
      source.addEventListener('status', (event) => {
        const payload = JSON.parse((event as MessageEvent).data) as PixStatusResponse
        handlePixStatus(payload)
      })
      source.onerror = () => {
        source.close()
        if (pixEvents === source) pixEvents = undefined
        startPixPolling(txid)
      }
    } catch {
      startPixPolling(txid)
    }
  }

  function startPixPolling(txid: string) {
    if (pixPoll) return
    pixPoll = window.setInterval(() => {
      void pollPixStatus(txid)
    }, 5000)
  }

  async function pollPixStatus(txid: string) {
    try {
      handlePixStatus(await api.getPixStatus(txid))
    } catch (error) {
      if (error instanceof APIError && error.status === 404) {
        pixError = 'Transacao PIX nao encontrada'
      }
    }
  }

  function handlePixStatus(payload: PixStatusResponse) {
    if (payload.status === 'aprovado' || payload.status === 'pago' || payload.status === 'confirmado') {
      finishAccess(selectedPlan?.nome || 'PIX confirmado', 0)
      return
    }

    if (payload.status === 'expirado' || payload.status === 'cancelado') {
      pixError = 'PIX expirado. Gere uma nova cobranca.'
      stopPixWatchers()
    }
  }

  function startPixCountdown(initialSeconds: number) {
    window.clearInterval(pixCountdown)
    pixSeconds = Math.max(0, initialSeconds)
    pixCountdown = window.setInterval(() => {
      if (pixSeconds <= 1) {
        pixSeconds = 0
        pixError ||= 'PIX expirado. Gere uma nova cobranca.'
        stopPixWatchers()
        return
      }
      pixSeconds -= 1
    }, 1000)
  }

  function startSuccessCountdown(initialSeconds: number) {
    window.clearInterval(successCountdown)
    successSeconds = Math.max(0, initialSeconds)
    successCountdown = window.setInterval(() => {
      successSeconds = Math.max(0, successSeconds - 1)
    }, 1000)
  }

  function stopPixWatchers() {
    if (pixEvents) {
      pixEvents.close()
      pixEvents = undefined
    }
    window.clearInterval(pixPoll)
    pixPoll = undefined
    window.clearInterval(pixCountdown)
    pixCountdown = undefined
  }

  function cleanupTimers() {
    stopPixWatchers()
    window.clearInterval(successCountdown)
    window.clearTimeout(copyMessageTimer)
  }

  async function copyPixCode() {
    if (!pix) return
    try {
      await navigator.clipboard.writeText(pix.pix_copia_cola)
      copyMessage = 'Codigo PIX copiado'
    } catch {
      copyMessage = 'Copie manualmente'
    }
    window.clearTimeout(copyMessageTimer)
    copyMessageTimer = window.setTimeout(() => {
      copyMessage = ''
    }, 2500)
  }

  function navigateAfterConnection() {
    window.location.assign(settings.url_pos_conexao || 'https://google.com')
  }

  function applySettings(nextSettings: Settings) {
    document.documentElement.style.setProperty('--color-primary', nextSettings.cor_primaria)
    document.documentElement.style.setProperty(
      '--color-secondary',
      nextSettings.cor_secundaria || defaultSettings.cor_secundaria || '#0EA5A8'
    )
    document.documentElement.style.setProperty('--color-bg', nextSettings.cor_fundo)
  }

  function messageFromError(error: unknown, fallback: string) {
    if (error instanceof APIError) return error.message
    if (error instanceof Error && error.message) return error.message
    return fallback
  }
</script>

<PortalShell {step} dark={step === 'welcome'}>
  {#if step === 'welcome'}
    <WelcomeScreen
      {settings}
      {device}
      loading={initialLoading}
      error={welcomeError}
      onShowPlans={showPlans}
      onShowVoucher={showVoucher}
      onRetry={bootstrap}
    />
  {:else if step === 'plans'}
    <PlanSelection
      {planos}
      loading={plansLoading}
      error={plansError}
      onBack={() => (step = 'welcome')}
      onRetry={loadPlans}
      onVoucher={showVoucher}
      onSelectPlan={selectPlan}
    />
  {:else if step === 'voucher'}
    <VoucherScreen
      submitting={voucherLoading}
      error={voucherError}
      onBack={() => (step = 'welcome')}
      onSubmit={redeemVoucher}
    />
  {:else if step === 'pix'}
    <PixScreen
      plano={selectedPlan}
      {pix}
      secondsRemaining={pixSeconds}
      loading={pixLoading}
      error={pixError}
      {copyMessage}
      onBack={showPlans}
      onCopy={copyPixCode}
    />
  {:else}
    <SuccessScreen
      {settings}
      planName={successPlanName}
      secondsRemaining={successSeconds}
      onNavigate={navigateAfterConnection}
    />
  {/if}
</PortalShell>

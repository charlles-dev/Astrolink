<script lang="ts">
  import { onMount } from 'svelte'

  import { APIError, api } from '$lib/api'
  import AdminDashboard from '$lib/components/AdminDashboard.svelte'
  import type {
    AdminHealthResponse,
    AdminLoginResponse,
    AdminUser,
    AdminVoucher,
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
  let loading = false
  let loginLoading = false
  let loginError = ''
  let actionMessage = ''

  onMount(() => {
    token = sessionStorage.getItem(TOKEN_KEY) || ''
    if (token) {
      void loadDashboard()
    }
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
    } catch (error) {
      loginError = messageFromError(error, 'Nao foi possivel entrar no painel')
    } finally {
      loginLoading = false
    }
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
        api.getAdminVouchers(token)
      ])
      health = nextHealth
      planos = nextPlanos.planos
      usuarios = nextUsuarios.usuarios
      vouchers = nextVouchers.vouchers
    } catch (error) {
      if (expireSessionIfUnauthorized(error)) return
      actionMessage = messageFromError(error, 'Nao foi possivel carregar o painel')
    } finally {
      loading = false
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
      vouchers = [...result.vouchers, ...vouchers]
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

  function logout() {
    resetSession()
    actionMessage = ''
    loginError = ''
  }

  function resetSession() {
    sessionStorage.removeItem(TOKEN_KEY)
    token = ''
    health = null
    planos = []
    usuarios = []
    vouchers = []
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
    {loading}
    {actionMessage}
    onRefresh={loadDashboard}
    onDisconnect={disconnect}
    onGenerateVouchers={generateVouchers}
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

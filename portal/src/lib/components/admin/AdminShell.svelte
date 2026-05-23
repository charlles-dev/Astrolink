<script lang="ts" module>
  export type AdminPanelPage =
    | 'overview'
    | 'usuarios'
    | 'planos'
    | 'rede'
    | 'vouchers'
    | 'pagamentos'
    | 'setup'
    | 'logs'
</script>

<script lang="ts">
  import { onMount } from 'svelte'

  import type { AdminHealthResponse, AdminUser } from '../../types'

  export let activePage: AdminPanelPage = 'overview'
  export let health: AdminHealthResponse | null = null
  export let usuarios: AdminUser[] = []
  export let liveConnected = false
  export let loading = false
  export let actionMessage = ''
  export let onRefresh: () => void = () => {}
  export let onLogout: () => void = () => {}

  type AdminTheme = 'light' | 'dark'

  const THEME_KEY = 'astrolink.admin.theme'

  const navItems = [
    {
      id: 'overview',
      label: 'Visão geral',
      href: '/painel',
      title: 'Operação local',
      description: 'Saúde do node, sessões conectadas e eventos críticos em uma leitura rápida.'
    },
    {
      id: 'usuarios',
      label: 'Usuários',
      href: '/painel/usuarios',
      title: 'Clientes ativos',
      description: 'Sessões, consumo, IP atual e ações diretas por dispositivo conectado.'
    },
    {
      id: 'planos',
      label: 'Planos',
      href: '/painel/planos',
      title: 'Catálogo de planos',
      description: 'Preço, validade, velocidade e publicação dos pacotes exibidos no portal.'
    },
    {
      id: 'rede',
      label: 'Rede',
      href: '/painel/rede',
      title: 'Rede local',
      description: 'Roteadores, OpenNDS, blacklist e walled garden controlados pelo node local.'
    },
    {
      id: 'vouchers',
      label: 'Vouchers',
      href: '/painel/vouchers',
      title: 'Emissão de vouchers',
      description: 'Lotes, filtros, impressão e status de uso para atendimento presencial.'
    },
    {
      id: 'pagamentos',
      label: 'Pagamentos',
      href: '/painel/pagamentos',
      title: 'Conciliação PIX',
      description: 'Totais, status de transações e exportações para conferência operacional.'
    },
    {
      id: 'setup',
      label: 'Setup',
      href: '/painel/setup',
      title: 'Ambiente local',
      description: 'Variáveis pessoais, credenciais e parâmetros do node sem depender do Supabase.'
    },
    {
      id: 'logs',
      label: 'Logs',
      href: '/painel/logs',
      title: 'Logs e rotinas',
      description: 'Observabilidade, filtros de incidente e rotinas protegidas de backup.'
    }
  ] as const

  let adminTheme: AdminTheme = 'light'
  let themeReady = false

  $: activeItem = navItems.find((item) => item.id === activePage) ?? navItems[0]
  $: nodeStatus = health?.status ?? 'offline'
  $: databaseStatus = health?.checks.banco_dados.status ?? 'sem dados'
  $: routerOnline = health?.checks.roteadores.online ?? 0
  $: routerTotal = health?.checks.roteadores.total ?? 0
  $: activeUsers = usuarios.filter((usuario) => usuario.status === 'ativo').length
  $: isDarkMode = adminTheme === 'dark'
  $: themeButtonLabel = isDarkMode ? 'Claro' : 'Escuro'
  $: themeButtonAria = isDarkMode ? 'Ativar modo claro' : 'Ativar modo escuro'
  $: if (themeReady) {
    applyAdminTheme(adminTheme)
  }

  onMount(() => {
    adminTheme = getInitialTheme()
    applyAdminTheme(adminTheme)
    themeReady = true
  })

  function getInitialTheme(): AdminTheme {
    const savedTheme = localStorage.getItem(THEME_KEY)
    if (isAdminTheme(savedTheme)) return savedTheme

    return window.matchMedia?.('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  }

  function isAdminTheme(value: string | null): value is AdminTheme {
    return value === 'light' || value === 'dark'
  }

  function applyAdminTheme(theme: AdminTheme) {
    document.documentElement.dataset.adminTheme = theme
    document.documentElement.style.colorScheme = theme
  }

  function toggleAdminTheme() {
    adminTheme = isDarkMode ? 'light' : 'dark'
    localStorage.setItem(THEME_KEY, adminTheme)
    applyAdminTheme(adminTheme)
  }
</script>

<section
  class="admin-shell"
  data-testid="admin-shell"
  data-theme={isDarkMode ? 'astrolink-dark' : 'astrolink'}
  data-admin-theme={adminTheme}
  aria-busy={loading}
>
  <aside class="admin-sidebar" aria-label="Navegação do painel local">
    <div class="brand-lockup">
      <span class="brand-mark" aria-hidden="true">A</span>
      <div>
        <strong>Astrolink</strong>
        <span>Console local</span>
      </div>
    </div>

    <nav class="admin-nav" aria-label="Áreas do painel">
      <ul class="menu">
        {#each navItems as item}
          <li>
            <a
              href={item.href}
              class:active-nav={item.id === activePage}
              aria-current={item.id === activePage ? 'page' : undefined}
            >
              {item.label}
            </a>
          </li>
        {/each}
      </ul>
    </nav>

    <div class="session-card card">
      <span class="session-label">Sessão local</span>
      <strong>{nodeStatus}</strong>
      <p>Banco de dados {databaseStatus}</p>
      <p>{routerOnline}/{routerTotal} roteadores online</p>
    </div>
  </aside>

  <main class="admin-workspace">
    <header class="workspace-header">
      <div class="workspace-copy">
        <div class="workspace-meta">
          <span class="workspace-status">
            <span class:online={liveConnected}></span>
            {liveConnected ? 'Tempo real ativo' : 'Tempo real offline'}
          </span>
          <span class="workspace-route">/{activeItem.id}</span>
        </div>
        <h1>{activeItem.title}</h1>
        <p>{activeItem.description}</p>
      </div>

      <div class="workspace-ops">
        <div class="workspace-summary" aria-label="Resumo operacional">
          <div>
            <span>Usuários</span>
            <strong>{activeUsers}</strong>
          </div>
          <div>
            <span>Node</span>
            <strong>{nodeStatus}</strong>
          </div>
          <div>
            <span>Roteadores</span>
            <strong>{routerOnline}/{routerTotal}</strong>
          </div>
        </div>

        <div class="workspace-actions" aria-label="Ações do console">
          <button
            type="button"
            class="btn btn-outline theme-toggle"
            onclick={toggleAdminTheme}
            aria-label={themeButtonAria}
            title={themeButtonAria}
          >
            <span class="theme-toggle-icon" aria-hidden="true"></span>
            <span>{themeButtonLabel}</span>
          </button>
          <button type="button" class="btn btn-outline refresh-button" onclick={onRefresh} disabled={loading}>
            {loading ? 'Atualizando' : 'Atualizar'}
          </button>
          <button type="button" class="btn btn-primary logout-button" onclick={onLogout}>Sair</button>
        </div>
      </div>
    </header>

    {#if actionMessage}
      <p class="action-message" role="status">{actionMessage}</p>
    {/if}

    <slot />
  </main>
</section>

<style>
  .admin-shell {
    width: 100%;
    max-width: 100vw;
    min-height: 100vh;
    display: grid;
    grid-template-columns: 268px minmax(0, 1fr);
    overflow-x: clip;
    background: var(--admin-shell-bg);
    color: var(--color-ink);
  }

  .admin-sidebar {
    position: sticky;
    top: 0;
    align-self: start;
    height: 100vh;
    height: 100dvh;
    max-height: 100vh;
    max-height: 100dvh;
    display: grid;
    grid-template-rows: auto auto 1fr;
    gap: 22px;
    border-right: 1px solid var(--admin-sidebar-line);
    padding: 24px 16px;
    overflow-y: auto;
    background: var(--admin-sidebar-bg);
    color: var(--admin-sidebar-text);
  }

  .brand-lockup,
  .workspace-header,
  .workspace-actions,
  .workspace-summary {
    display: flex;
    align-items: center;
  }

  .brand-lockup {
    gap: 12px;
    padding: 0 6px;
  }

  .brand-mark {
    width: 44px;
    height: 44px;
    display: grid;
    place-items: center;
    border-radius: 8px;
    background: var(--admin-brand-bg);
    color: var(--admin-brand-ink);
    font-size: 1rem;
    font-weight: 950;
    box-shadow: var(--shadow-hairline);
  }

  .brand-lockup strong,
  .brand-lockup div span,
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

  .brand-lockup div span,
  .session-card p,
  .session-label {
    color: var(--admin-sidebar-muted);
  }

  .brand-lockup div span {
    display: block;
    margin-top: 3px;
    font-size: 0.78rem;
    font-weight: 760;
  }

  .admin-nav {
    display: grid;
  }

  .admin-nav .menu {
    width: 100%;
    display: grid;
    gap: 4px;
    padding: 0;
    background: transparent;
  }

  .admin-nav a {
    min-height: 38px;
    display: flex;
    align-items: center;
    border-radius: 8px;
    padding: 0 12px;
    color: var(--admin-nav-text);
    font-size: 0.84rem;
    font-weight: 830;
    text-decoration: none;
    transition:
      background 160ms ease,
      color 160ms ease,
      transform 160ms ease;
  }

  .admin-nav a:hover,
  .admin-nav a.active-nav {
    background: var(--admin-nav-active-bg);
    color: var(--admin-nav-active-text);
  }

  .admin-nav a:hover {
    transform: translateX(2px);
  }

  .admin-nav a.active-nav {
    box-shadow: inset 3px 0 0 var(--admin-brand-bg);
  }

  .session-card {
    align-self: end;
    display: grid;
    gap: 7px;
    border: 1px solid var(--admin-sidebar-line);
    border-radius: 8px;
    padding: 14px;
    background: var(--admin-sidebar-card-bg);
    box-shadow: none;
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
    width: min(100%, calc(100vw - 268px));
    min-width: 0;
    display: grid;
    align-content: start;
    gap: var(--admin-section-gap);
    padding: var(--admin-page-gutter);
  }

  .workspace-header {
    min-height: 118px;
    justify-content: space-between;
    gap: 18px;
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: clamp(16px, 2.2vw, 24px);
    background: var(--color-surface-raised);
    box-shadow: var(--shadow-panel);
  }

  .workspace-copy {
    min-width: 0;
    display: grid;
    gap: 8px;
  }

  .workspace-meta {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
  }

  .workspace-status {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    color: var(--color-muted);
    font-size: 0.78rem;
    font-weight: 860;
    text-transform: uppercase;
  }

  .workspace-route {
    border: 1px solid var(--color-line);
    border-radius: 999px;
    padding: 3px 8px;
    background: var(--color-surface-subtle);
    color: var(--color-muted);
    font-size: 0.72rem;
    font-weight: 820;
  }

  .workspace-status span {
    width: 9px;
    height: 9px;
    border-radius: 999px;
    background: var(--state-warning-text);
    box-shadow: 0 0 0 4px var(--state-warning-bg);
  }

  .workspace-status span.online {
    background: var(--color-success);
    box-shadow: 0 0 0 4px var(--state-success-bg);
  }

  h1 {
    font-size: clamp(1.55rem, 2.6vw, 2.2rem);
    font-weight: 950;
    line-height: 1.05;
  }

  .workspace-copy p {
    max-width: 620px;
    color: var(--color-muted);
    font-size: 0.92rem;
    font-weight: 650;
    line-height: 1.5;
  }

  .workspace-ops {
    min-width: min(100%, 430px);
    display: grid;
    justify-items: end;
    gap: 12px;
  }

  .workspace-summary {
    width: 100%;
    gap: 8px;
  }

  .workspace-summary div {
    min-width: 0;
    min-height: 62px;
    flex: 1 1 0;
    display: grid;
    align-content: center;
    gap: 4px;
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 10px 12px;
    background: var(--color-surface-subtle);
  }

  .workspace-summary span {
    color: var(--color-muted);
    font-size: 0.72rem;
    font-weight: 860;
    text-transform: uppercase;
  }

  .workspace-summary strong {
    min-width: 0;
    overflow-wrap: anywhere;
    font-size: 1rem;
    font-weight: 930;
    line-height: 1.05;
  }

  .workspace-actions {
    flex: 0 0 auto;
    justify-content: flex-end;
    gap: 8px;
  }

  .theme-toggle {
    min-height: 38px;
    gap: 9px;
    border-radius: 8px;
    padding-inline: 12px;
    font-size: 0.82rem;
    font-weight: 850;
    white-space: nowrap;
  }

  .theme-toggle-icon {
    width: 28px;
    height: 16px;
    display: inline-flex;
    align-items: center;
    border: 1px solid var(--color-line-strong);
    border-radius: 999px;
    padding: 2px;
    background: var(--color-surface-muted);
  }

  .theme-toggle-icon::before {
    content: '';
    width: 10px;
    height: 10px;
    border-radius: 999px;
    background: var(--color-primary);
    transition: transform 160ms ease;
  }

  :global([data-admin-theme='dark']) .theme-toggle-icon::before {
    transform: translateX(10px);
  }

  .refresh-button,
  .logout-button {
    min-height: 38px;
    border-radius: 8px;
    padding-inline: 14px;
    font-size: 0.82rem;
    font-weight: 850;
  }

  .admin-nav {
    scrollbar-width: none;
  }

  .admin-nav::-webkit-scrollbar,
  .admin-nav .menu::-webkit-scrollbar {
    display: none;
  }

  .action-message {
    margin: 0;
    border: 1px solid var(--state-info-line);
    border-radius: 8px;
    padding: 14px 16px;
    background: var(--state-info-bg);
    color: var(--state-info-text);
    font-size: 0.88rem;
    font-weight: 800;
  }

  @media (max-width: 1080px) {
    .admin-shell {
      grid-template-columns: 1fr;
    }

    .admin-workspace {
      width: 100%;
    }

    .admin-sidebar {
      position: static;
      align-self: auto;
      height: auto;
      max-height: none;
      min-height: auto;
      grid-template-columns: auto minmax(0, 1fr) auto;
      grid-template-rows: auto;
      align-items: center;
      gap: 14px;
      border-right: 0;
      border-bottom: 1px solid var(--color-line);
      padding: 16px 18px;
      overflow-y: visible;
    }

    .admin-nav .menu {
      grid-auto-flow: column;
      grid-auto-columns: max-content;
      overflow-x: auto;
      gap: 6px;
    }

    .admin-nav a {
      min-height: 40px;
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

    .workspace-ops {
      width: 100%;
      min-width: 0;
      justify-items: stretch;
    }

    .workspace-summary {
      width: 100%;
      display: grid;
      grid-template-columns: repeat(3, minmax(0, 1fr));
    }

    .workspace-actions {
      justify-content: flex-start;
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

    .admin-nav .menu {
      grid-auto-flow: row;
      grid-auto-columns: auto;
      grid-template-columns: repeat(auto-fit, minmax(104px, 1fr));
      overflow-x: visible;
    }

    .admin-nav a {
      justify-content: center;
      padding-inline: 10px;
      text-align: center;
    }

    .admin-nav a:hover {
      transform: none;
    }

    .admin-nav a.active-nav {
      box-shadow: inset 0 3px 0 var(--admin-brand-bg);
    }

    .session-card {
      min-width: 0;
    }

    .workspace-header {
      padding: 20px;
    }

    h1 {
      font-size: 2rem;
    }

    .workspace-actions {
      align-items: stretch;
      width: 100%;
      gap: 8px;
    }

    .workspace-actions button {
      min-width: 0;
      flex: 1;
    }

    .theme-toggle {
      padding-inline: 10px;
    }

    .workspace-summary {
      grid-template-columns: 1fr;
    }

    .workspace-summary div {
      min-height: 54px;
    }
  }
</style>

<script lang="ts">
  import { formatCurrency, formatCountdown } from '../format'
  import type { AdminHealthResponse, AdminUser, Plano } from '../types'

  export let health: AdminHealthResponse | null = null
  export let planos: Plano[] = []
  export let usuarios: AdminUser[] = []
  export let loading = false
  export let actionMessage = ''
  export let onRefresh: () => void = () => {}
  export let onDisconnect: (mac: string) => void = () => {}
  export let onLogout: () => void = () => {}

  $: activeUsers = usuarios.filter((usuario) => usuario.status === 'ativo').length
  $: visiblePlans = planos.filter((plano) => plano.ativo).length
  $: routerOnline = health?.checks.roteadores.online ?? 0
  $: routerTotal = health?.checks.roteadores.total ?? 0
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

  <div class="metric-grid">
    <article>
      <span>Usuarios ativos</span>
      <strong>{activeUsers}</strong>
      <small>{usuarios.length} conhecidos</small>
    </article>
    <article>
      <span>Planos ativos</span>
      <strong>{visiblePlans}</strong>
      <small>{planos.length} cadastrados</small>
    </article>
    <article>
      <span>Roteadores</span>
      <strong>{routerOnline}/{routerTotal}</strong>
      <small>online agora</small>
    </article>
    <article>
      <span>Banco</span>
      <strong>{health?.checks.banco_dados.status ?? 'sem dados'}</strong>
      <small>Banco {health?.checks.banco_dados.status ?? 'desconhecido'}</small>
    </article>
  </div>

  <div class="admin-content">
    <section class="users-panel" aria-labelledby="usuarios-title">
      <div class="section-heading">
        <div>
          <h2 id="usuarios-title">Usuarios conectados</h2>
          <p>MACs conhecidos pelo no local.</p>
        </div>
        {#if loading}
          <span class="loading-chip">Carregando</span>
        {/if}
      </div>

      {#if usuarios.length === 0}
        <div class="empty-state">
          <h3>Nenhum usuario registrado</h3>
          <p>Quando alguem resgatar voucher ou iniciar PIX, aparece aqui.</p>
        </div>
      {:else}
        <div class="user-list">
          {#each usuarios as usuario (usuario.mac)}
            <article class="user-row">
              <div class="user-main">
                <span class="status-dot" class:online={usuario.status === 'ativo'}></span>
                <div>
                  <h3>{usuario.mac}</h3>
                  <p>
                    {usuario.plano?.nome ?? 'Sem plano'} · {usuario.ip_atual ?? 'sem IP'} ·
                    {usuario.status}
                  </p>
                </div>
              </div>
              <div class="user-meta">
                <span>{formatCountdown(usuario.tempo_restante_segundos)}</span>
                <small>{usuario.dados_consumidos_mb} MB</small>
              </div>
              <button
                type="button"
                class="row-button"
                aria-label={`Desconectar ${usuario.mac}`}
                disabled={loading || usuario.status !== 'ativo'}
                onclick={() => onDisconnect(usuario.mac)}
              >
                Desconectar
              </button>
            </article>
          {/each}
        </div>
      {/if}
    </section>

    <aside class="plans-panel" aria-labelledby="planos-title">
      <div class="section-heading">
        <div>
          <h2 id="planos-title">Planos</h2>
          <p>Oferta atual do portal.</p>
        </div>
      </div>

      <div class="plan-admin-list">
        {#each planos as plano (plano.id)}
          <article>
            <div>
              <h3>{plano.nome}</h3>
              <p>{plano.duracao_formatada}</p>
            </div>
            <strong>{formatCurrency(plano.preco)}</strong>
          </article>
        {:else}
          <div class="empty-state compact">
            <h3>Nenhum plano</h3>
            <p>Cadastre planos pelo backend nas proximas etapas.</p>
          </div>
        {/each}
      </div>
    </aside>
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
  .admin-actions,
  .section-heading,
  .user-main,
  .user-row,
  .plan-admin-list article {
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
  h2,
  h3,
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
  .ink-button,
  .row-button {
    min-height: 42px;
    border-radius: 12px;
    padding: 0 14px;
    font-size: 0.86rem;
    font-weight: 850;
  }

  .ghost-button,
  .row-button {
    border: 1px solid var(--color-line);
    background: white;
    color: var(--color-ink);
  }

  .ink-button {
    border: 0;
    background: var(--color-ink);
    color: white;
  }

  .ghost-button:disabled,
  .row-button:disabled {
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

  .metric-grid,
  .admin-content {
    max-width: 1180px;
    margin: 0 auto;
  }

  .metric-grid {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 12px;
  }

  .metric-grid article,
  .users-panel,
  .plans-panel {
    border: 1px solid var(--color-line);
    border-radius: 8px;
    background: white;
    box-shadow: 0 12px 28px rgba(15, 23, 42, 0.06);
  }

  .metric-grid article {
    min-height: 130px;
    display: grid;
    gap: 8px;
    padding: 18px;
  }

  .metric-grid span,
  .metric-grid small,
  .section-heading p,
  .user-row p,
  .plan-admin-list p,
  .user-meta small,
  .empty-state p {
    color: var(--color-muted);
  }

  .metric-grid span {
    font-size: 0.78rem;
    font-weight: 850;
    text-transform: uppercase;
  }

  .metric-grid strong {
    font-size: 1.65rem;
    font-weight: 930;
    line-height: 1;
  }

  .metric-grid small,
  .user-meta small {
    font-size: 0.78rem;
    font-weight: 750;
  }

  .admin-content {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 340px;
    gap: 16px;
    margin-top: 16px;
  }

  .users-panel,
  .plans-panel {
    padding: 18px;
  }

  .section-heading {
    justify-content: space-between;
    gap: 14px;
    margin-bottom: 16px;
  }

  h2 {
    font-size: 1.05rem;
    font-weight: 900;
  }

  .section-heading p {
    margin-top: 4px;
    font-size: 0.88rem;
  }

  .loading-chip {
    border-radius: 999px;
    padding: 7px 10px;
    background: #f1f5f9;
    color: #475569;
    font-size: 0.75rem;
    font-weight: 850;
  }

  .user-list,
  .plan-admin-list {
    display: grid;
    gap: 10px;
  }

  .user-row {
    min-height: 82px;
    justify-content: space-between;
    gap: 14px;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    padding: 13px;
    background: #fcfdff;
  }

  .user-main {
    min-width: 0;
    flex: 1;
    gap: 10px;
  }

  .user-main h3,
  .plan-admin-list h3,
  .empty-state h3 {
    margin: 0;
    font-size: 0.95rem;
    font-weight: 900;
  }

  .user-main p,
  .plan-admin-list p {
    margin-top: 4px;
    font-size: 0.82rem;
    line-height: 1.35;
  }

  .status-dot {
    width: 10px;
    height: 10px;
    flex: 0 0 auto;
    border-radius: 999px;
    background: #94a3b8;
  }

  .status-dot.online {
    background: var(--color-success);
    box-shadow: 0 0 0 5px #dcfce7;
  }

  .user-meta {
    width: 90px;
    display: grid;
    gap: 3px;
    text-align: right;
  }

  .user-meta span {
    font-size: 0.9rem;
    font-weight: 900;
  }

  .row-button {
    white-space: nowrap;
  }

  .plan-admin-list article {
    justify-content: space-between;
    gap: 12px;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    padding: 13px;
    background: #fcfdff;
  }

  .plan-admin-list strong {
    white-space: nowrap;
    font-size: 0.95rem;
    font-weight: 920;
  }

  .empty-state {
    border: 1px dashed var(--color-line);
    border-radius: 8px;
    padding: 18px;
    background: #f8fafc;
  }

  .empty-state.compact {
    padding: 14px;
  }

  .empty-state p {
    margin-top: 6px;
    font-size: 0.88rem;
  }

  @media (max-width: 900px) {
    .metric-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .admin-content {
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

    .metric-grid {
      grid-template-columns: 1fr;
    }

    .user-row {
      align-items: stretch;
      flex-direction: column;
    }

    .user-meta {
      width: auto;
      text-align: left;
    }
  }
</style>

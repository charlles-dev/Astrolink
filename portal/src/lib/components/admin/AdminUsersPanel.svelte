<script lang="ts">
  import { formatCountdown } from '../../format'
  import type { AdminUser } from '../../types'

  export let usuarios: AdminUser[] = []
  export let loading = false
  export let onDisconnect: (mac: string) => void = () => {}
</script>

<section class="users-panel card" aria-labelledby="usuarios-title">
  <div class="section-heading">
    <div>
      <h2 id="usuarios-title">Usuários conectados</h2>
      <p>MACs conhecidos pelo nó local.</p>
    </div>
    {#if loading}
      <span class="loading-chip badge badge-soft">Carregando</span>
    {/if}
  </div>

  {#if loading && usuarios.length === 0}
    <div class="loading-state" role="status">
      <span class="loading loading-spinner loading-sm"></span>
      Carregando usuários
    </div>
  {:else if usuarios.length === 0}
    <div class="empty-state">
      <h3>Nenhum usuário registrado</h3>
      <p>Quando alguém resgatar voucher ou iniciar PIX, aparece aqui.</p>
    </div>
  {:else}
    <div class="user-list" role="table" aria-label="Usuários conectados">
      <div class="user-head" role="row">
        <span role="columnheader">Cliente</span>
        <span role="columnheader">Plano / IP</span>
        <span role="columnheader">Tempo</span>
        <span role="columnheader">Consumo</span>
        <span role="columnheader">Ação</span>
      </div>

      {#each usuarios as usuario (usuario.mac)}
        <div class="user-row" role="row">
          <div class="user-main" role="cell">
            <span class="status-dot" class:online={usuario.status === 'ativo'}></span>
            <div>
              <h3>{usuario.mac}</h3>
              <p>{usuario.status}</p>
            </div>
          </div>

          <div class="plan-cell" role="cell">
            <span>{usuario.plano?.nome ?? 'Sem plano'}</span>
            <small>{usuario.ip_atual ?? 'sem IP'}</small>
          </div>

          <div class="metric-cell" role="cell">
            <span>{formatCountdown(usuario.tempo_restante_segundos)}</span>
          </div>

          <div class="metric-cell" role="cell">
            <span>{usuario.dados_consumidos_mb} MB</span>
          </div>

          <div class="action-cell" role="cell">
            <button
              type="button"
              class="btn btn-outline row-button"
              aria-label={`Desconectar ${usuario.mac}`}
              disabled={loading || usuario.status !== 'ativo'}
              onclick={() => onDisconnect(usuario.mac)}
            >
              Desconectar
            </button>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</section>

<style>
  .users-panel {
    border: 1px solid var(--color-line);
    border-radius: var(--admin-panel-radius);
    padding: 0;
    overflow: hidden;
    background: var(--color-surface-raised);
    box-shadow: var(--shadow-panel);
  }

  .section-heading,
  .user-main,
  .loading-state {
    display: flex;
    align-items: center;
  }

  h2,
  h3,
  p {
    margin: 0;
  }

  .section-heading {
    justify-content: space-between;
    gap: 16px;
    border-bottom: 1px solid var(--color-line);
    padding: 15px 16px;
  }

  h2 {
    font-size: 1rem;
    font-weight: 900;
    line-height: 1.2;
  }

  .section-heading p,
  .user-main p,
  .plan-cell small,
  .empty-state p,
  .user-head {
    color: var(--color-muted);
  }

  .section-heading p {
    margin-top: 3px;
    font-size: 0.8rem;
    line-height: 1.3;
  }

  .loading-chip {
    border-radius: 999px;
    padding: 7px 10px;
    background: var(--state-neutral-bg);
    color: var(--state-neutral-text);
    font-size: 0.74rem;
    font-weight: 850;
  }

  .user-list {
    display: grid;
  }

  .user-head,
  .user-row {
    display: grid;
    grid-template-columns: minmax(180px, 1.35fr) minmax(160px, 1fr) minmax(92px, auto) minmax(88px, auto) minmax(112px, auto);
    align-items: center;
    column-gap: 14px;
  }

  .user-head {
    border-bottom: 1px solid var(--color-line);
    padding: 9px 16px;
    background: var(--color-surface-subtle);
    font-size: 0.7rem;
    font-weight: 900;
    text-transform: uppercase;
  }

  .user-head span:nth-child(n + 3) {
    text-align: right;
  }

  .user-row {
    min-height: 64px;
    border-bottom: 1px solid var(--color-line);
    padding: 10px 16px;
    background: var(--color-row);
  }

  .user-row:last-child {
    border-bottom: 0;
  }

  .user-main {
    min-width: 0;
    gap: 10px;
  }

  .user-main div,
  .plan-cell {
    min-width: 0;
  }

  .user-main h3,
  .empty-state h3 {
    overflow-wrap: anywhere;
    font-size: 0.88rem;
    font-weight: 900;
    line-height: 1.25;
  }

  .user-main p {
    margin-top: 2px;
    font-size: 0.74rem;
    font-weight: 800;
    line-height: 1.25;
    text-transform: uppercase;
  }

  .plan-cell {
    display: grid;
    gap: 2px;
  }

  .plan-cell span,
  .metric-cell span {
    overflow-wrap: anywhere;
    font-size: 0.82rem;
    font-weight: 850;
    line-height: 1.25;
  }

  .plan-cell small {
    overflow-wrap: anywhere;
    font-size: 0.74rem;
    font-weight: 750;
    line-height: 1.25;
  }

  .metric-cell,
  .action-cell {
    min-width: 0;
    text-align: right;
  }

  .status-dot {
    width: 8px;
    height: 8px;
    flex: 0 0 auto;
    border-radius: 999px;
    background: var(--color-muted);
  }

  .status-dot.online {
    background: var(--color-success);
    box-shadow: 0 0 0 4px var(--state-success-bg);
  }

  .row-button {
    min-height: 34px;
    border-radius: 6px;
    padding: 0 11px;
    font-size: 0.78rem;
    font-weight: 850;
    white-space: nowrap;
  }

  .row-button:disabled {
    cursor: not-allowed;
    opacity: 0.55;
  }

  .empty-state {
    padding: 22px 16px;
    background: var(--color-surface-subtle);
  }

  .loading-state {
    min-height: 96px;
    justify-content: center;
    gap: 10px;
    background: var(--color-surface-subtle);
    color: var(--color-muted);
    font-size: 0.88rem;
    font-weight: 850;
  }

  .empty-state p {
    margin-top: 6px;
    font-size: 0.84rem;
    line-height: 1.4;
  }

  @media (max-width: 820px) {
    .user-head {
      display: none;
    }

    .user-row {
      grid-template-columns: minmax(0, 1fr) auto;
      row-gap: 9px;
    }

    .user-main,
    .plan-cell {
      grid-column: 1 / -1;
    }

    .metric-cell,
    .action-cell {
      text-align: left;
    }

    .action-cell {
      text-align: right;
    }
  }

  @media (max-width: 520px) {
    .section-heading {
      align-items: flex-start;
      flex-direction: column;
    }

    .user-row {
      grid-template-columns: 1fr;
    }

    .action-cell {
      text-align: left;
    }

    .row-button {
      width: 100%;
    }
  }
</style>

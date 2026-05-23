<script lang="ts">
  import { formatCountdown } from '../../format'
  import type { AdminUser } from '../../types'

  export let usuarios: AdminUser[] = []
  export let loading = false

  $: visibleUsers = usuarios.slice(0, 4)
  $: hiddenCount = Math.max(usuarios.length - visibleUsers.length, 0)
</script>

<section class="users-summary card" aria-labelledby="usuarios-summary-title">
  <div class="section-heading">
    <div>
      <h2 id="usuarios-summary-title">Usuários conectados</h2>
      <p>Sessões recentes no nó local.</p>
    </div>
    <a class="btn btn-outline summary-link" href="/painel/usuarios">Ver usuários</a>
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
    <div class="summary-list" role="list">
      {#each visibleUsers as usuario (usuario.mac)}
        <article class="summary-row" role="listitem">
          <div class="identity-cell">
            <span class="status-dot" class:online={usuario.status === 'ativo'}></span>
            <div>
              <h3>{usuario.mac}</h3>
              <p>{usuario.plano?.nome ?? 'Sem plano'}</p>
            </div>
          </div>
          <span class="ip-cell">{usuario.ip_atual ?? 'sem IP'}</span>
          <strong>{formatCountdown(usuario.tempo_restante_segundos)}</strong>
        </article>
      {/each}
    </div>

    {#if hiddenCount > 0}
      <p class="summary-footnote">+{hiddenCount} {hiddenCount === 1 ? 'usuário' : 'usuários'} na lista completa.</p>
    {/if}
  {/if}
</section>

<style>
  .users-summary {
    border: 1px solid var(--color-line);
    border-radius: var(--admin-panel-radius);
    padding: 0;
    overflow: hidden;
    background: var(--color-surface-raised);
    box-shadow: var(--shadow-panel);
  }

  .section-heading,
  .summary-row,
  .identity-cell,
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
    gap: 14px;
    border-bottom: 1px solid var(--color-line);
    padding: 14px 16px;
  }

  h2 {
    font-size: 0.96rem;
    font-weight: 900;
    line-height: 1.2;
  }

  .section-heading p,
  .summary-list p,
  .summary-footnote,
  .empty-state p,
  .loading-state,
  .ip-cell {
    color: var(--color-muted);
  }

  .section-heading p {
    margin-top: 3px;
    font-size: 0.78rem;
    line-height: 1.3;
  }

  .summary-link {
    min-height: 32px;
    border-radius: 6px;
    padding-inline: 10px;
    font-size: 0.76rem;
    font-weight: 850;
    white-space: nowrap;
  }

  .summary-list {
    display: grid;
  }

  .summary-row {
    min-height: 58px;
    gap: 12px;
    border-bottom: 1px solid var(--color-line);
    padding: 10px 16px;
    background: var(--color-row);
  }

  .summary-row:last-child {
    border-bottom: 0;
  }

  .identity-cell {
    min-width: 0;
    flex: 1;
    gap: 10px;
  }

  .identity-cell div {
    min-width: 0;
  }

  .summary-list h3,
  .empty-state h3 {
    overflow-wrap: anywhere;
    font-size: 0.86rem;
    font-weight: 900;
    line-height: 1.25;
  }

  .summary-list p,
  .ip-cell {
    overflow-wrap: anywhere;
    font-size: 0.76rem;
    font-weight: 750;
    line-height: 1.3;
  }

  .summary-list p {
    margin-top: 2px;
  }

  .ip-cell {
    flex: 0 1 112px;
    text-align: right;
  }

  .summary-row strong {
    flex: 0 0 82px;
    font-size: 0.82rem;
    font-weight: 900;
    line-height: 1.2;
    text-align: right;
    white-space: nowrap;
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

  .summary-footnote {
    border-top: 1px solid var(--color-line);
    padding: 10px 16px;
    background: var(--color-surface-subtle);
    font-size: 0.76rem;
    font-weight: 800;
  }

  .loading-state {
    min-height: 88px;
    justify-content: center;
    gap: 10px;
    background: var(--color-surface-subtle);
    font-size: 0.84rem;
    font-weight: 850;
  }

  .empty-state {
    padding: 20px 16px;
    background: var(--color-surface-subtle);
  }

  .empty-state p {
    margin-top: 5px;
    font-size: 0.82rem;
    line-height: 1.4;
  }

  @media (max-width: 620px) {
    .section-heading {
      align-items: stretch;
      flex-direction: column;
    }

    .summary-link {
      width: 100%;
    }

    .summary-row {
      display: grid;
      grid-template-columns: 1fr auto;
    }

    .identity-cell {
      grid-column: 1 / -1;
    }

    .ip-cell {
      flex-basis: auto;
      text-align: left;
    }

    .summary-row strong {
      flex-basis: auto;
    }
  }
</style>

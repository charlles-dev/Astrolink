<script lang="ts">
  import { formatCountdown } from '../../format'
  import type { AdminUser } from '../../types'

  export let usuarios: AdminUser[] = []
  export let loading = false
  export let onDisconnect: (mac: string) => void = () => {}
</script>

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
                {usuario.plano?.nome ?? 'Sem plano'} - {usuario.ip_atual ?? 'sem IP'} -
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

<style>
  .users-panel {
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 18px;
    background: white;
    box-shadow: 0 12px 28px rgba(15, 23, 42, 0.06);
  }

  .section-heading,
  .user-main,
  .user-row {
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
    margin-bottom: 16px;
  }

  h2 {
    font-size: 1.05rem;
    font-weight: 900;
  }

  .section-heading p,
  .user-row p,
  .user-meta small,
  .empty-state p {
    color: var(--color-muted);
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

  .user-list {
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
  .empty-state h3 {
    font-size: 0.95rem;
    font-weight: 900;
  }

  .user-main p {
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

  .user-meta small {
    font-size: 0.78rem;
    font-weight: 750;
  }

  .row-button {
    min-height: 42px;
    border: 1px solid var(--color-line);
    border-radius: 12px;
    padding: 0 14px;
    background: white;
    color: var(--color-ink);
    font-size: 0.86rem;
    font-weight: 850;
    white-space: nowrap;
  }

  .row-button:disabled {
    cursor: not-allowed;
    opacity: 0.55;
  }

  .empty-state {
    border: 1px dashed var(--color-line);
    border-radius: 8px;
    padding: 18px;
    background: #f8fafc;
  }

  .empty-state p {
    margin-top: 6px;
    font-size: 0.88rem;
  }

  @media (max-width: 620px) {
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

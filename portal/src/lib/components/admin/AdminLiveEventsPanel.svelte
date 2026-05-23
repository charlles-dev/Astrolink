<script module lang="ts">
  export type AdminLiveSnapshot = {
    usuarios: {
      ativos: number
      total: number
    }
    vouchers: {
      ativos: number
      total: number
    }
    pix: {
      pendente: number
      aprovado: number
    }
    logs: number
  }

  export type AdminLiveEvent = {
    id?: string | number
    tipo: string
    mensagem: string
    timestamp: string
  }
</script>

<script lang="ts">
  export let connected = false
  export let lastEventAt = ''
  export let snapshot: AdminLiveSnapshot | null = null
  export let events: AdminLiveEvent[] = []

  $: statusText = connected ? 'Conectado' : 'Desconectado'
  $: statusMessage = connected
    ? 'Recebendo eventos em tempo real.'
    : 'Aguardando reconexao do canal ao vivo.'
  $: lastUpdateText = lastEventAt
    ? `Última atualização: ${formatDate(lastEventAt)}`
    : 'Última atualização: sem dados'

  function formatDate(value: string) {
    if (!value) return 'sem data'
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) return value
    return date.toLocaleString('pt-BR', {
      day: '2-digit',
      month: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    })
  }
</script>

<section class="live-events-panel card" aria-labelledby="live-events-title">
  <div class="section-heading">
    <div>
      <h2 id="live-events-title">Eventos ao vivo</h2>
      <p>{statusMessage}</p>
    </div>

    <span class="status-pill badge" class:online={connected} class:offline={!connected}>
      <span aria-hidden="true"></span>
      {statusText}
    </span>
  </div>

  <div class="live-toolbar">
    <p class="last-update">{lastUpdateText}</p>
    <a class="btn btn-outline logs-link" href="/painel/logs">Ver logs</a>
  </div>

  {#if snapshot}
    <dl class="snapshot-grid" aria-label="Snapshot operacional">
      <div>
        <dt>Usuários</dt>
        <dd>{snapshot.usuarios.ativos}/{snapshot.usuarios.total}</dd>
        <small>ativos/total</small>
      </div>

      <div>
        <dt>Vouchers</dt>
        <dd>{snapshot.vouchers.ativos}/{snapshot.vouchers.total}</dd>
        <small>ativos/total</small>
      </div>

      <div>
        <dt>PIX</dt>
        <dd>{snapshot.pix.pendente} pendente</dd>
        <small>{snapshot.pix.aprovado} aprovado</small>
      </div>

      <div>
        <dt>Logs</dt>
        <dd>{snapshot.logs}</dd>
        <small>registros</small>
      </div>
    </dl>
  {:else}
    <div class="empty-state compact">
      <h3>Sem snapshot recebido.</h3>
      <p>Os contadores aparecem quando o canal publicar o primeiro estado.</p>
    </div>
  {/if}

  <div class="events-section" aria-live="polite">
    <div class="events-head">
      <h3>Últimos eventos</h3>
      <span>{events.length} recentes</span>
    </div>

    <div class="event-list">
      {#each events.slice(0, 5) as event (event.id ?? `${event.timestamp}-${event.tipo}-${event.mensagem}`)}
        <article aria-label={event.tipo}>
          <div class="event-meta">
            <strong>{event.tipo}</strong>
            <small>{formatDate(event.timestamp)}</small>
          </div>
          <p>{event.mensagem}</p>
        </article>
      {:else}
        <div class="empty-state">
          <h3>Nenhum evento recebido</h3>
          <p>Aguardando atividade operacional.</p>
        </div>
      {/each}
    </div>
  </div>
</section>

<style>
  .live-events-panel {
    container-type: inline-size;
    border: 1px solid var(--color-line);
    border-radius: var(--admin-panel-radius);
    padding: 0;
    overflow: hidden;
    background: var(--color-surface-raised);
    box-shadow: var(--shadow-panel);
  }

  .section-heading,
  .status-pill,
  .live-toolbar,
  .events-head,
  .event-meta {
    display: flex;
    align-items: center;
  }

  h2,
  h3,
  p,
  dl,
  dd {
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
  .last-update,
  .snapshot-grid small,
  .events-head span,
  .event-meta small,
  .event-list p,
  .empty-state p {
    color: var(--color-muted);
  }

  .section-heading p {
    margin-top: 3px;
    font-size: 0.8rem;
    line-height: 1.3;
  }

  .status-pill {
    flex: 0 0 auto;
    gap: 7px;
    border: 1px solid var(--color-line);
    border-radius: 999px;
    padding: 7px 9px;
    background: var(--color-surface-subtle);
    color: var(--color-muted);
    font-size: 0.72rem;
    font-weight: 900;
    line-height: 1;
    white-space: nowrap;
  }

  .status-pill span {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: currentColor;
  }

  .status-pill.online {
    border-color: var(--state-success-line);
    background: var(--state-success-bg);
    color: var(--state-success-text);
  }

  .status-pill.offline {
    border-color: var(--state-error-line);
    background: var(--state-error-bg);
    color: var(--state-error-text);
  }

  .live-toolbar {
    justify-content: space-between;
    gap: 12px;
    border-bottom: 1px solid var(--color-line);
    padding: 10px 16px;
    background: var(--color-surface-subtle);
  }

  .last-update {
    font-size: 0.76rem;
    font-weight: 800;
    line-height: 1.35;
  }

  .logs-link {
    min-height: 30px;
    border-radius: 6px;
    padding-inline: 9px;
    font-size: 0.72rem;
    font-weight: 900;
    white-space: nowrap;
  }

  .snapshot-grid {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    border-bottom: 1px solid var(--color-line);
  }

  .snapshot-grid div {
    min-width: 0;
    display: grid;
    gap: 3px;
    border-right: 1px solid var(--color-line);
    padding: 12px 16px;
    background: var(--color-row);
  }

  .snapshot-grid div:last-child {
    border-right: 0;
  }

  .snapshot-grid dt {
    color: var(--color-muted);
    font-size: 0.7rem;
    font-weight: 900;
    line-height: 1.2;
    text-transform: uppercase;
  }

  .snapshot-grid dd {
    color: var(--color-ink);
    font-size: 1rem;
    font-weight: 950;
    line-height: 1.15;
    overflow-wrap: anywhere;
  }

  .snapshot-grid small {
    font-size: 0.74rem;
    font-weight: 750;
    line-height: 1.25;
    overflow-wrap: anywhere;
  }

  .events-section {
    display: grid;
  }

  .events-head {
    justify-content: space-between;
    gap: 10px;
    border-bottom: 1px solid var(--color-line);
    padding: 12px 16px;
  }

  h3 {
    color: var(--color-ink);
    font-size: 0.9rem;
    font-weight: 900;
    line-height: 1.25;
  }

  .events-head span {
    flex: 0 0 auto;
    font-size: 0.72rem;
    font-weight: 850;
  }

  .event-list {
    display: grid;
  }

  .event-list article {
    min-width: 0;
    display: grid;
    gap: 4px;
    border-bottom: 1px solid var(--color-line);
    padding: 10px 16px 10px 14px;
    box-shadow: inset 3px 0 0 var(--color-secondary);
  }

  .event-list article:last-child {
    border-bottom: 0;
  }

  .event-meta {
    min-width: 0;
    gap: 8px;
    justify-content: space-between;
  }

  .event-meta strong {
    min-width: 0;
    color: var(--color-ink);
    font-size: 0.8rem;
    font-weight: 900;
    line-height: 1.25;
    overflow-wrap: anywhere;
  }

  .event-meta small {
    flex: 0 0 auto;
    font-size: 0.72rem;
    font-weight: 800;
    line-height: 1.25;
    white-space: nowrap;
  }

  .event-list p {
    font-size: 0.8rem;
    line-height: 1.35;
    overflow-wrap: anywhere;
  }

  .empty-state {
    padding: 18px 16px;
    background: var(--color-surface-subtle);
  }

  .empty-state.compact {
    border-bottom: 1px solid var(--color-line);
  }

  .empty-state h3 {
    font-size: 0.88rem;
  }

  .empty-state p {
    margin-top: 5px;
    font-size: 0.8rem;
    line-height: 1.35;
  }

  @container (max-width: 430px) {
    .snapshot-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .snapshot-grid div:nth-child(2n) {
      border-right: 0;
    }

    .snapshot-grid div:nth-child(-n + 2) {
      border-bottom: 1px solid var(--color-line);
    }
  }

  @media (max-width: 760px) {
    .snapshot-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .snapshot-grid div:nth-child(2n) {
      border-right: 0;
    }

    .snapshot-grid div:nth-child(-n + 2) {
      border-bottom: 1px solid var(--color-line);
    }
  }

  @media (max-width: 520px) {
    .section-heading,
    .live-toolbar,
    .events-head,
    .event-meta {
      align-items: flex-start;
      flex-direction: column;
    }

    .status-pill {
      align-self: flex-start;
    }

    .snapshot-grid {
      grid-template-columns: 1fr;
    }

    .snapshot-grid div,
    .snapshot-grid div:nth-child(2n) {
      border-right: 0;
    }

    .snapshot-grid div:nth-child(n) {
      border-bottom: 1px solid var(--color-line);
    }

    .snapshot-grid div:last-child {
      border-bottom: 0;
    }
  }
</style>

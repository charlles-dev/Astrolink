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
    ? `Ultima atualizacao: ${formatDate(lastEventAt)}`
    : 'Ultima atualizacao: sem dados'

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

<section class="live-events-panel" aria-labelledby="live-events-title">
  <div class="section-heading">
    <div>
      <h2 id="live-events-title">Eventos ao vivo</h2>
      <p>{statusMessage}</p>
    </div>

    <span class="status-pill" class:online={connected} class:offline={!connected}>
      <span aria-hidden="true"></span>
      {statusText}
    </span>
  </div>

  <p class="last-update">{lastUpdateText}</p>

  {#if snapshot}
    <div class="snapshot-grid" aria-label="Snapshot operacional">
      <div>
        <span>Usuarios</span>
        <strong>{snapshot.usuarios.ativos}/{snapshot.usuarios.total}</strong>
        <small>ativos/total</small>
      </div>

      <div>
        <span>Vouchers</span>
        <strong>{snapshot.vouchers.ativos}/{snapshot.vouchers.total}</strong>
        <small>ativos/total</small>
      </div>

      <div>
        <span>PIX</span>
        <strong>{snapshot.pix.pendente} pendente</strong>
        <small>{snapshot.pix.aprovado} aprovado</small>
      </div>

      <div>
        <span>Logs</span>
        <strong>{snapshot.logs}</strong>
        <small>registros</small>
      </div>
    </div>
  {:else}
    <div class="empty-state compact">
      <h3>Sem snapshot recebido.</h3>
      <p>Os contadores aparecem quando o canal publicar o primeiro estado.</p>
    </div>
  {/if}

  <div class="events-section" aria-live="polite">
    <div class="events-head">
      <h3>Ultimos eventos</h3>
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
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 18px;
    background: white;
    box-shadow: 0 12px 28px rgba(15, 23, 42, 0.06);
  }

  .section-heading,
  .status-pill,
  .events-head,
  .event-meta {
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
    margin-bottom: 10px;
  }

  h2 {
    font-size: 1.05rem;
    font-weight: 900;
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
    margin-top: 4px;
    font-size: 0.88rem;
    line-height: 1.35;
  }

  .status-pill {
    flex: 0 0 auto;
    gap: 7px;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    padding: 7px 9px;
    background: #f8fafc;
    color: var(--color-muted);
    font-size: 0.74rem;
    font-weight: 900;
    line-height: 1;
    white-space: nowrap;
  }

  .status-pill span {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: currentColor;
  }

  .status-pill.online {
    border-color: #bbf7d0;
    background: #f0fdf4;
    color: #166534;
  }

  .status-pill.offline {
    border-color: #fecaca;
    background: #fef2f2;
    color: #991b1b;
  }

  .last-update {
    border-top: 1px solid var(--color-line);
    padding-top: 12px;
    font-size: 0.78rem;
    font-weight: 800;
    line-height: 1.35;
  }

  .snapshot-grid {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    margin-top: 12px;
    border-top: 1px solid #e2e8f0;
    border-left: 1px solid #e2e8f0;
  }

  .snapshot-grid div {
    min-width: 0;
    display: grid;
    gap: 4px;
    border-right: 1px solid #e2e8f0;
    border-bottom: 1px solid #e2e8f0;
    padding: 10px;
    background: #fcfdff;
  }

  .snapshot-grid span {
    color: var(--color-muted);
    font-size: 0.72rem;
    font-weight: 900;
    text-transform: uppercase;
  }

  .snapshot-grid strong {
    color: var(--color-ink);
    font-size: 1rem;
    font-weight: 950;
    line-height: 1.1;
    overflow-wrap: anywhere;
  }

  .snapshot-grid small {
    font-size: 0.75rem;
    font-weight: 750;
    overflow-wrap: anywhere;
  }

  .events-section {
    display: grid;
    gap: 10px;
    margin-top: 14px;
    border-top: 1px solid var(--color-line);
    padding-top: 14px;
  }

  .events-head {
    justify-content: space-between;
    gap: 10px;
  }

  h3 {
    color: var(--color-ink);
    font-size: 0.92rem;
    font-weight: 900;
  }

  .events-head span {
    flex: 0 0 auto;
    font-size: 0.74rem;
    font-weight: 850;
  }

  .event-list {
    display: grid;
    gap: 8px;
  }

  .event-list article {
    min-width: 0;
    border-left: 3px solid #38bdf8;
    padding: 9px 0 9px 10px;
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
    overflow-wrap: anywhere;
  }

  .event-meta small {
    flex: 0 0 auto;
    font-size: 0.72rem;
    font-weight: 800;
    white-space: nowrap;
  }

  .event-list p {
    margin-top: 4px;
    font-size: 0.82rem;
    line-height: 1.35;
    overflow-wrap: anywhere;
  }

  .empty-state {
    border: 1px dashed var(--color-line);
    border-radius: 8px;
    padding: 14px;
    background: #f8fafc;
  }

  .empty-state.compact {
    margin-top: 12px;
  }

  .empty-state h3 {
    font-size: 0.9rem;
  }

  .empty-state p {
    margin-top: 5px;
    font-size: 0.82rem;
    line-height: 1.35;
  }

  @media (max-width: 760px) {
    .snapshot-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }

  @media (max-width: 520px) {
    .section-heading,
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
  }
</style>

<script lang="ts">
  import type { AdminLog, AdminLogFilters } from '../../types'

  export let logs: AdminLog[] = []
  export let total = 0
  export let loading = false
  export let onApplyLogFilters: (filters: AdminLogFilters) => void = () => {}
  export let onExportLogs: (filters: AdminLogFilters) => void = () => {}

  let nivel = ''
  let tipo = ''
  let texto = ''

  $: visibleLogs = logs.slice(0, 12)
  $: logsTruncated = total > visibleLogs.length || logs.length > visibleLogs.length

  function currentFilters(): AdminLogFilters {
    const filters: AdminLogFilters = {}
    if (nivel) filters.nivel = nivel
    if (tipo.trim()) filters.tipo = tipo.trim()
    if (texto.trim()) filters.texto = texto.trim()
    return filters
  }

  function applyFilters() {
    onApplyLogFilters(currentFilters())
  }

  function clearFilters() {
    nivel = ''
    tipo = ''
    texto = ''
    onApplyLogFilters({})
  }

  function exportCsv() {
    onExportLogs(currentFilters())
  }

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

  function levelLabel(value: string) {
    const labels: Record<string, string> = {
      info: 'Info',
      aviso: 'Aviso',
      erro: 'Erro'
    }
    return labels[value] ?? value
  }
</script>

<section class="logs-panel card" aria-labelledby="logs-title">
  <div class="section-heading">
    <div>
      <h2 id="logs-title">Logs</h2>
      <p>Observabilidade local: {total} registros indexados.</p>
    </div>
  </div>

  <form
    class="filter-form"
    aria-label="Filtros de logs"
    onsubmit={(event) => {
      event.preventDefault()
      applyFilters()
    }}
  >
    <div class="filter-grid">
      <label class="field">
        Nível
        <select class="select select-bordered" bind:value={nivel} disabled={loading}>
          <option value="">Todos</option>
          <option value="info">Info</option>
          <option value="aviso">Aviso</option>
          <option value="erro">Erro</option>
        </select>
      </label>
      <label class="field">
        Tipo
        <input class="input input-bordered" bind:value={tipo} autocomplete="off" disabled={loading} />
      </label>
      <label class="field wide">
        Buscar texto
        <input class="input input-bordered" bind:value={texto} autocomplete="off" disabled={loading} />
      </label>
    </div>

    <div class="filter-actions">
      <button
        type="submit"
        class="btn btn-primary ink-button"
        disabled={loading}
        aria-label="Aplicar filtros de logs"
      >
        Aplicar filtros
      </button>
      <button
        type="button"
        class="btn btn-outline ghost-button"
        onclick={clearFilters}
        disabled={loading}
        aria-label="Limpar filtros de logs"
      >
        Limpar filtros
      </button>
      <button
        type="button"
        class="btn btn-outline ghost-button"
        onclick={exportCsv}
        disabled={loading}
        aria-label="Exportar logs CSV"
      >
        Exportar CSV
      </button>
    </div>
  </form>

  <p class="list-count">
    {#if logsTruncated}
      Mostrando {visibleLogs.length} de {total || logs.length} logs (limite de 12).
    {:else}
      Mostrando {visibleLogs.length} {visibleLogs.length === 1 ? 'log' : 'logs'}.
    {/if}
  </p>

  <div class="log-list">
    {#each visibleLogs as log (`${log.timestamp}-${log.tipo}-${log.mensagem}`)}
      <article>
        <div class="log-main">
          <div class="log-head">
            <span class={`badge level ${log.nivel}`}>{levelLabel(log.nivel)}</span>
            <strong>{log.tipo}</strong>
            <small>{formatDate(log.timestamp)}</small>
          </div>
          <p>{log.mensagem}</p>
          {#if log.detalhes}
            <div class="details-block">
              <span>Detalhes</span>
              <code>{JSON.stringify(log.detalhes)}</code>
            </div>
          {/if}
        </div>
      </article>
    {:else}
      <div class="empty-state">
        <h3>Nenhum log encontrado</h3>
        <p>Ajuste filtros ou atualize a coleta local.</p>
      </div>
    {/each}
  </div>
</section>

<style>
  .logs-panel {
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: var(--admin-panel-padding);
    background: var(--color-surface-raised);
    box-shadow: var(--shadow-panel);
  }

  .section-heading,
  .filter-actions,
  .log-head {
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
    gap: 18px;
    margin-bottom: 20px;
  }

  h2 {
    font-size: 1.05rem;
    font-weight: 900;
  }

  .section-heading p,
  .log-head small,
  .list-count,
  .empty-state p {
    color: var(--color-muted);
  }

  .section-heading p {
    margin-top: 4px;
    font-size: 0.88rem;
  }

  .filter-form {
    display: grid;
    gap: 14px;
    margin-bottom: 20px;
    border-top: 1px solid var(--color-line);
    padding-top: 20px;
  }

  .filter-grid {
    display: grid;
    grid-template-columns: 120px 150px minmax(0, 1fr);
    gap: 14px;
  }

  .field {
    display: grid;
    gap: 6px;
    color: var(--color-ink);
    font-size: 0.78rem;
    font-weight: 850;
  }

  .field input,
  .field select {
    width: 100%;
    min-height: 42px;
    border-radius: 8px;
    padding: 0 11px;
  }

  .filter-actions {
    gap: 10px;
    flex-wrap: wrap;
  }

  .list-count {
    margin: 0 0 12px;
    font-size: 0.78rem;
    font-weight: 850;
  }

  .ink-button,
  .ghost-button {
    min-height: 42px;
    border-radius: 8px;
    padding: 0 14px;
    font-size: 0.86rem;
    font-weight: 850;
  }

  .log-list {
    display: grid;
    gap: var(--admin-row-gap);
  }

  .log-list article {
    position: relative;
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 16px 16px 16px 18px;
    background: var(--color-row);
  }

  .log-list article::before {
    content: '';
    position: absolute;
    inset: 12px auto 12px 0;
    width: 3px;
    border-radius: 999px;
    background: var(--state-info-text);
  }

  .log-main {
    min-width: 0;
  }

  .log-head {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr) auto;
    gap: 10px;
    width: 100%;
  }

  .log-head strong {
    font-size: 0.82rem;
    font-weight: 900;
    overflow-wrap: anywhere;
  }

  .log-head small {
    font-size: 0.75rem;
    font-weight: 750;
  }

  .log-main p {
    margin-top: 7px;
    color: var(--color-ink);
    font-size: 0.86rem;
    line-height: 1.35;
    overflow-wrap: anywhere;
  }

  .details-block {
    display: grid;
    gap: 6px;
    margin-top: 10px;
  }

  .details-block span {
    color: var(--color-muted);
    font-size: 0.68rem;
    font-weight: 900;
    text-transform: uppercase;
  }

  code {
    display: block;
    border-radius: 8px;
    padding: 10px;
    background: var(--state-neutral-bg);
    color: var(--state-neutral-text);
    font-size: 0.75rem;
    line-height: 1.35;
    overflow-wrap: anywhere;
    white-space: pre-wrap;
  }

  .level {
    background: var(--state-info-bg);
    color: var(--state-info-text);
    font-size: 0.7rem;
    font-weight: 900;
  }

  .level.erro {
    background: var(--state-error-bg);
    color: var(--state-error-text);
  }

  .level.aviso {
    background: var(--state-warning-bg);
    color: var(--state-warning-text);
  }

  .log-list article:has(.level.erro)::before {
    background: var(--state-error-text);
  }

  .log-list article:has(.level.aviso)::before {
    background: var(--state-warning-text);
  }

  button:disabled {
    cursor: not-allowed;
    opacity: 0.58;
  }

  .empty-state {
    border: 1px dashed var(--color-line);
    border-radius: 8px;
    padding: 22px;
    background: var(--color-surface-subtle);
  }

  .empty-state h3 {
    font-size: 0.95rem;
    font-weight: 900;
  }

  .empty-state p {
    margin-top: 6px;
    font-size: 0.88rem;
  }

  @media (max-width: 760px) {
    .filter-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .field.wide {
      grid-column: 1 / -1;
    }
  }

  @media (max-width: 520px) {
    .filter-grid {
      grid-template-columns: 1fr;
    }

    .field.wide {
      grid-column: auto;
    }

    .filter-actions {
      align-items: stretch;
      flex-direction: column;
    }

    .log-head {
      grid-template-columns: 1fr;
    }
  }
</style>

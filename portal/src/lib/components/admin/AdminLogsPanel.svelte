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
</script>

<section class="logs-panel" aria-labelledby="logs-title">
  <div class="section-heading">
    <div>
      <h2 id="logs-title">Logs</h2>
      <p>{total} registros operacionais.</p>
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
        Nivel
        <select bind:value={nivel} disabled={loading}>
          <option value="">Todos</option>
          <option value="info">Info</option>
          <option value="aviso">Aviso</option>
          <option value="erro">Erro</option>
        </select>
      </label>
      <label class="field">
        Tipo
        <input bind:value={tipo} autocomplete="off" disabled={loading} />
      </label>
      <label class="field wide">
        Texto
        <input bind:value={texto} autocomplete="off" disabled={loading} />
      </label>
    </div>

    <div class="filter-actions">
      <button
        type="submit"
        class="ink-button"
        disabled={loading}
        aria-label="Aplicar filtros de logs"
      >
        Aplicar filtros
      </button>
      <button
        type="button"
        class="ghost-button"
        onclick={exportCsv}
        disabled={loading}
        aria-label="Exportar logs CSV"
      >
        Exportar CSV
      </button>
    </div>
  </form>

  <div class="log-list">
    {#each logs.slice(0, 12) as log (`${log.timestamp}-${log.tipo}-${log.mensagem}`)}
      <article>
        <div class="log-main">
          <div class="log-head">
            <span class={`level ${log.nivel}`}>{log.nivel}</span>
            <strong>{log.tipo}</strong>
            <small>{formatDate(log.timestamp)}</small>
          </div>
          <p>{log.mensagem}</p>
          {#if log.detalhes}
            <code>{JSON.stringify(log.detalhes)}</code>
          {/if}
        </div>
      </article>
    {:else}
      <div class="empty-state">
        <h3>Nenhum log encontrado</h3>
        <p>Ajuste filtros ou atualize o painel.</p>
      </div>
    {/each}
  </div>
</section>

<style>
  .logs-panel {
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 18px;
    background: white;
    box-shadow: 0 12px 28px rgba(15, 23, 42, 0.06);
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
    gap: 14px;
    margin-bottom: 14px;
  }

  h2 {
    font-size: 1.05rem;
    font-weight: 900;
  }

  .section-heading p,
  .log-head small,
  .empty-state p {
    color: var(--color-muted);
  }

  .section-heading p {
    margin-top: 4px;
    font-size: 0.88rem;
  }

  .filter-form {
    display: grid;
    gap: 10px;
    margin-bottom: 14px;
    border-top: 1px solid var(--color-line);
    padding-top: 14px;
  }

  .filter-grid {
    display: grid;
    grid-template-columns: 120px 150px minmax(0, 1fr);
    gap: 10px;
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
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 0 11px;
    background: #f8fafc;
    color: var(--color-ink);
  }

  .filter-actions {
    gap: 8px;
  }

  .ink-button,
  .ghost-button {
    min-height: 42px;
    border-radius: 12px;
    padding: 0 14px;
    font-size: 0.86rem;
    font-weight: 850;
  }

  .ink-button {
    border: 0;
    background: var(--color-ink);
    color: white;
  }

  .ghost-button {
    border: 1px solid var(--color-line);
    background: white;
    color: var(--color-ink);
  }

  .log-list {
    display: grid;
    gap: 10px;
  }

  .log-list article {
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    padding: 13px;
    background: #fcfdff;
  }

  .log-main {
    min-width: 0;
  }

  .log-head {
    gap: 8px;
    flex-wrap: wrap;
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

  code {
    display: block;
    margin-top: 8px;
    border-radius: 8px;
    padding: 8px;
    background: #f1f5f9;
    color: #334155;
    font-size: 0.75rem;
    line-height: 1.35;
    overflow-wrap: anywhere;
    white-space: pre-wrap;
  }

  .level {
    border-radius: 999px;
    padding: 5px 8px;
    background: #e0f2fe;
    color: #075985;
    font-size: 0.7rem;
    font-weight: 900;
  }

  .level.erro {
    background: #fee2e2;
    color: #991b1b;
  }

  .level.aviso {
    background: #fef9c3;
    color: #854d0e;
  }

  button:disabled {
    cursor: not-allowed;
    opacity: 0.58;
  }

  .empty-state {
    border: 1px dashed var(--color-line);
    border-radius: 8px;
    padding: 18px;
    background: #f8fafc;
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
  }
</style>

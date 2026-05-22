<script lang="ts">
  import type { AdminPayment, AdminPaymentFilters, AdminPaymentTotals } from '../../types'

  export let pagamentos: AdminPayment[] = []
  export let totais: AdminPaymentTotals = {
    pendente: 0,
    aprovado: 0,
    cancelado: 0,
    expirado: 0,
    valor_total: '0.00'
  }
  export let loading = false
  export let onApplyPaymentFilters: (filters: AdminPaymentFilters) => void = () => {}
  export let onExportPayments: (filters: AdminPaymentFilters) => void = () => {}

  let status: AdminPaymentFilters['status'] = ''
  let inicio = ''
  let fim = ''

  function currentFilters(): AdminPaymentFilters {
    const filters: AdminPaymentFilters = {}
    if (status) filters.status = status
    if (inicio) filters.inicio = inicio
    if (fim) filters.fim = fim
    return filters
  }

  function applyFilters() {
    onApplyPaymentFilters(currentFilters())
  }

  function exportCsv() {
    onExportPayments(currentFilters())
  }

  function formatMoney(value: string) {
    return `R$ ${value || '0.00'}`
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

<section class="payments-panel" aria-labelledby="pagamentos-title">
  <div class="section-heading">
    <div>
      <h2 id="pagamentos-title">Pagamentos</h2>
      <p>PIX e liberacoes recentes.</p>
    </div>
    <strong>{formatMoney(totais.valor_total)}</strong>
  </div>

  <div class="total-grid">
    <span><b>{totais.pendente}</b> pendente</span>
    <span><b>{totais.aprovado}</b> aprovado</span>
    <span><b>{totais.cancelado}</b> cancelado</span>
    <span><b>{totais.expirado}</b> expirado</span>
  </div>

  <form
    class="filter-form"
    aria-label="Filtros de pagamentos"
    onsubmit={(event) => {
      event.preventDefault()
      applyFilters()
    }}
  >
    <div class="filter-grid">
      <label class="field">
        Status do pagamento
        <select bind:value={status} disabled={loading}>
          <option value="">Todos</option>
          <option value="pendente">Pendente</option>
          <option value="aprovado">Aprovado</option>
          <option value="cancelado">Cancelado</option>
          <option value="expirado">Expirado</option>
        </select>
      </label>
      <label class="field">
        Inicio
        <input bind:value={inicio} type="date" disabled={loading} />
      </label>
      <label class="field">
        Fim
        <input bind:value={fim} type="date" disabled={loading} />
      </label>
    </div>

    <div class="filter-actions">
      <button
        type="submit"
        class="ink-button"
        disabled={loading}
        aria-label="Aplicar filtros de pagamentos"
      >
        Aplicar filtros
      </button>
      <button
        type="button"
        class="ghost-button"
        onclick={exportCsv}
        disabled={loading}
        aria-label="Exportar pagamentos CSV"
      >
        Exportar CSV
      </button>
    </div>
  </form>

  <div class="payment-list">
    {#each pagamentos.slice(0, 10) as pagamento (pagamento.txid)}
      <article>
        <div class="payment-main">
          <h3>{pagamento.txid}</h3>
          <p>{pagamento.plano?.nome ?? pagamento.descricao} - {pagamento.mac}</p>
        </div>
        <div class="payment-meta">
          <span class={`status ${pagamento.status}`}>{pagamento.status}</span>
          <strong>{formatMoney(pagamento.valor)}</strong>
          <small>{formatDate(pagamento.created_at)}</small>
        </div>
      </article>
    {:else}
      <div class="empty-state">
        <h3>Nenhum pagamento</h3>
        <p>Ajuste filtros ou aguarde novas transacoes.</p>
      </div>
    {/each}
  </div>
</section>

<style>
  .payments-panel {
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 18px;
    background: white;
    box-shadow: 0 12px 28px rgba(15, 23, 42, 0.06);
  }

  .section-heading,
  .filter-actions,
  .payment-list article {
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
  .payment-list p,
  .payment-meta small,
  .empty-state p {
    color: var(--color-muted);
  }

  .section-heading p {
    margin-top: 4px;
    font-size: 0.88rem;
  }

  .section-heading strong {
    flex: 0 0 auto;
    font-size: 1.12rem;
    font-weight: 950;
  }

  .total-grid {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 8px;
    margin-bottom: 14px;
  }

  .total-grid span {
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    padding: 9px;
    background: #f8fafc;
    color: var(--color-muted);
    font-size: 0.76rem;
    font-weight: 800;
    overflow-wrap: anywhere;
  }

  .total-grid b {
    display: block;
    color: var(--color-ink);
    font-size: 1rem;
    font-weight: 950;
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
    grid-template-columns: 1fr 150px 150px;
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

  .payment-list {
    display: grid;
    gap: 10px;
  }

  .payment-list article {
    justify-content: space-between;
    gap: 12px;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    padding: 13px;
    background: #fcfdff;
  }

  .payment-main {
    min-width: 0;
  }

  .payment-main h3 {
    font-family: ui-monospace, "SFMono-Regular", Consolas, monospace;
    font-size: 0.92rem;
    font-weight: 900;
    letter-spacing: 0;
    overflow-wrap: anywhere;
  }

  .payment-main p {
    margin-top: 4px;
    font-size: 0.8rem;
    line-height: 1.35;
    overflow-wrap: anywhere;
  }

  .payment-meta {
    display: grid;
    gap: 4px;
    justify-items: end;
    text-align: right;
  }

  .payment-meta strong {
    font-size: 0.92rem;
    font-weight: 950;
  }

  .payment-meta small {
    font-size: 0.76rem;
    font-weight: 750;
  }

  .status {
    border-radius: 999px;
    padding: 5px 8px;
    background: #f1f5f9;
    color: #475569;
    font-size: 0.7rem;
    font-weight: 900;
  }

  .status.aprovado {
    background: #dcfce7;
    color: #166534;
  }

  .status.pendente {
    background: #fef9c3;
    color: #854d0e;
  }

  .status.cancelado,
  .status.expirado {
    background: #fee2e2;
    color: #991b1b;
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
    .total-grid,
    .filter-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }

  @media (max-width: 520px) {
    .section-heading,
    .filter-actions,
    .payment-list article {
      align-items: stretch;
      flex-direction: column;
    }

    .filter-grid,
    .total-grid {
      grid-template-columns: 1fr;
    }

    .payment-meta {
      justify-items: start;
      text-align: left;
    }
  }
</style>

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

  $: visiblePayments = pagamentos.slice(0, 10)
  $: paymentsTruncated = pagamentos.length > visiblePayments.length

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

  function clearFilters() {
    status = ''
    inicio = ''
    fim = ''
    onApplyPaymentFilters({})
  }

  function exportCsv() {
    onExportPayments(currentFilters())
  }

  function formatMoney(value: string) {
    const amount = Number(value || 0)
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL'
    })
      .format(Number.isFinite(amount) ? amount : 0)
      .replace(/\u00a0/g, ' ')
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

  function statusLabel(value: string) {
    const labels: Record<string, string> = {
      pendente: 'Pendente',
      aprovado: 'Aprovado',
      cancelado: 'Cancelado',
      expirado: 'Expirado'
    }
    return labels[value] ?? value
  }
</script>

<section class="payments-panel card" aria-labelledby="pagamentos-title">
  <div class="section-heading">
    <div>
      <h2 id="pagamentos-title">Pagamentos</h2>
      <p>Conciliação PIX local, liberações e divergências recentes.</p>
    </div>
    <div class="headline-total">
      <span>Total conciliado</span>
      <strong>{formatMoney(totais.valor_total)}</strong>
    </div>
  </div>

  <div class="total-grid">
    <span class="pending"><b>{totais.pendente}</b> pendentes</span>
    <span class="approved"><b>{totais.aprovado}</b> aprovados</span>
    <span class="canceled"><b>{totais.cancelado}</b> cancelados</span>
    <span class="expired"><b>{totais.expirado}</b> expirados</span>
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
        <select class="select select-bordered" bind:value={status} disabled={loading}>
          <option value="">Todos</option>
          <option value="pendente">Pendente</option>
          <option value="aprovado">Aprovado</option>
          <option value="cancelado">Cancelado</option>
          <option value="expirado">Expirado</option>
        </select>
      </label>
      <label class="field">
        Início
        <input class="input input-bordered" bind:value={inicio} type="date" disabled={loading} />
      </label>
      <label class="field">
        Fim
        <input class="input input-bordered" bind:value={fim} type="date" disabled={loading} />
      </label>
    </div>

    <div class="filter-actions">
      <button
        type="submit"
        class="btn btn-primary ink-button"
        disabled={loading}
        aria-label="Aplicar filtros de pagamentos"
      >
        Aplicar filtros
      </button>
      <button
        type="button"
        class="btn btn-outline ghost-button"
        onclick={clearFilters}
        disabled={loading}
        aria-label="Limpar filtros de pagamentos"
      >
        Limpar filtros
      </button>
      <button
        type="button"
        class="btn btn-outline ghost-button"
        onclick={exportCsv}
        disabled={loading}
        aria-label="Exportar pagamentos CSV"
      >
        Exportar CSV
      </button>
    </div>
  </form>

  <p class="list-count">
    {#if paymentsTruncated}
      Mostrando {visiblePayments.length} de {pagamentos.length} pagamentos (limite de 10).
    {:else}
      Mostrando {visiblePayments.length} {visiblePayments.length === 1 ? 'pagamento' : 'pagamentos'}.
    {/if}
  </p>

  <div class="payment-list">
    {#each visiblePayments as pagamento (pagamento.txid)}
      <article>
        <div class="payment-main">
          <div class="payment-id">
            <span>TXID</span>
            <h3>{pagamento.txid}</h3>
          </div>
          <p>{pagamento.plano?.nome ?? pagamento.descricao}</p>
          <small>{pagamento.mac}</small>
        </div>
        <div class="payment-meta">
          <span class={`badge status ${pagamento.status}`}>{statusLabel(pagamento.status)}</span>
          <strong>{formatMoney(pagamento.valor)}</strong>
          <small>{formatDate(pagamento.created_at)}</small>
        </div>
      </article>
    {:else}
      <div class="empty-state">
        <h3>Nenhum pagamento</h3>
        <p>Ajuste filtros ou aguarde novas transações do gateway local.</p>
      </div>
    {/each}
  </div>
</section>

<style>
  .payments-panel {
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: var(--admin-panel-padding);
    background: var(--color-surface-raised);
    box-shadow: var(--shadow-panel);
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
    gap: 18px;
    margin-bottom: 20px;
  }

  h2 {
    font-size: 1.05rem;
    font-weight: 900;
  }

  .section-heading p,
  .payment-list p,
  .payment-meta small,
  .list-count,
  .empty-state p {
    color: var(--color-muted);
  }

  .section-heading p {
    margin-top: 4px;
    font-size: 0.88rem;
  }

  .headline-total {
    display: grid;
    gap: 2px;
    justify-items: end;
    text-align: right;
  }

  .headline-total span {
    color: var(--color-muted);
    font-size: 0.72rem;
    font-weight: 850;
    text-transform: uppercase;
  }

  .headline-total strong {
    flex: 0 0 auto;
    font-size: 1.12rem;
    font-weight: 950;
  }

  .total-grid {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 12px;
    margin-bottom: 20px;
  }

  .total-grid span {
    position: relative;
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 12px 12px 12px 14px;
    background: var(--color-surface-subtle);
    color: var(--color-muted);
    font-size: 0.76rem;
    font-weight: 800;
    overflow-wrap: anywhere;
  }

  .total-grid span::before {
    content: '';
    position: absolute;
    inset: 10px auto 10px 0;
    width: 3px;
    border-radius: 999px;
    background: var(--state-neutral-text);
  }

  .total-grid .approved::before {
    background: var(--state-success-text);
  }

  .total-grid .pending::before {
    background: var(--state-warning-text);
  }

  .total-grid .canceled::before,
  .total-grid .expired::before {
    background: var(--state-error-text);
  }

  .total-grid b {
    display: block;
    color: var(--color-ink);
    font-size: 1rem;
    font-weight: 950;
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
    grid-template-columns: 1fr 150px 150px;
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

  .payment-list {
    display: grid;
    gap: var(--admin-row-gap);
  }

  .payment-list article {
    justify-content: space-between;
    gap: 16px;
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 16px;
    background: var(--color-row);
  }

  .payment-list article:hover {
    border-color: var(--color-muted);
  }

  .payment-main {
    min-width: 0;
  }

  .payment-id {
    display: grid;
    gap: 3px;
  }

  .payment-id span {
    color: var(--color-muted);
    font-size: 0.66rem;
    font-weight: 900;
    letter-spacing: 0;
    text-transform: uppercase;
  }

  .payment-main h3 {
    font-family: ui-monospace, "SFMono-Regular", Consolas, monospace;
    font-size: 0.92rem;
    font-weight: 900;
    letter-spacing: 0;
    overflow-wrap: anywhere;
  }

  .payment-main p {
    margin-top: 8px;
    font-size: 0.8rem;
    line-height: 1.35;
    overflow-wrap: anywhere;
  }

  .payment-main small {
    display: block;
    margin-top: 3px;
    color: var(--color-muted);
    font-family: ui-monospace, "SFMono-Regular", Consolas, monospace;
    font-size: 0.74rem;
    font-weight: 750;
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
    background: var(--state-neutral-bg);
    color: var(--state-neutral-text);
    font-size: 0.7rem;
    font-weight: 900;
  }

  .status.aprovado {
    background: var(--state-success-bg);
    color: var(--state-success-text);
  }

  .status.pendente {
    background: var(--state-warning-bg);
    color: var(--state-warning-text);
  }

  .status.cancelado,
  .status.expirado {
    background: var(--state-error-bg);
    color: var(--state-error-text);
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

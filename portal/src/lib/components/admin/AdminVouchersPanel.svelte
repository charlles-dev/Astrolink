<script lang="ts">
  import { tick } from 'svelte'
  import type {
    AdminVoucher,
    AdminVoucherFilters,
    AdminVoucherStatusFilter,
    GenerateAdminVouchersBody,
    Plano
  } from '../../types'
  import AdminVoucherPrintSheet from './AdminVoucherPrintSheet.svelte'

  export let planos: Plano[] = []
  export let vouchers: AdminVoucher[] = []
  export let loading = false
  export let onGenerateVouchers: (input: GenerateAdminVouchersBody) => void = () => {}
  export let onApplyVoucherFilters: (filters: AdminVoucherFilters) => void = () => {}
  export let onDeactivateVoucher: (id: number) => void = () => {}
  export let onExportVouchers: (filters: AdminVoucherFilters) => void = () => {}

  let voucherPlanoID = 0
  let voucherQuantidade = 1
  let voucherPrefixo = 'VIP'
  let voucherTipo: 'single_use' | 'universal' = 'single_use'
  let voucherUsosMaximos = 1
  let voucherValidadeDias: number | '' = ''
  let filtroStatus: AdminVoucherStatusFilter = 'ativo'
  let filtroPlanoID = ''
  let filtroCodigo = ''
  let filtroLoteID = ''
  let confirmDeactivateID: number | null = null
  let printSheetReady = false

  $: if (planos.length > 0 && !planos.some((plano) => plano.id === voucherPlanoID)) {
    voucherPlanoID = planos[0].id
  }

  function submitVoucherForm() {
    const body: GenerateAdminVouchersBody = {
      plano_id: Number(voucherPlanoID),
      quantidade: Number(voucherQuantidade),
      tipo: voucherTipo,
      prefixo: voucherPrefixo.trim().toUpperCase()
    }

    const validadeDias = Number(voucherValidadeDias)
    if (Number.isFinite(validadeDias) && validadeDias > 0) {
      body.validade_dias = validadeDias
    }

    if (voucherTipo === 'universal') {
      body.usos_maximos = Number(voucherUsosMaximos)
    }

    onGenerateVouchers(body)
  }

  function currentFilters(): AdminVoucherFilters {
    const filters: AdminVoucherFilters = {}
    if (filtroStatus) filters.status = filtroStatus

    const planoID = Number(filtroPlanoID)
    if (Number.isFinite(planoID) && planoID > 0) filters.plano_id = planoID

    const codigo = filtroCodigo.trim().toUpperCase()
    if (codigo) filters.codigo = codigo

    const loteID = Number(filtroLoteID)
    if (Number.isFinite(loteID) && loteID > 0) filters.lote_id = loteID

    return filters
  }

  function applyFilters() {
    confirmDeactivateID = null
    onApplyVoucherFilters(currentFilters())
  }

  function exportCsv() {
    onExportVouchers(currentFilters())
  }

  async function printSheet() {
    if (vouchers.length === 0) return

    printSheetReady = true
    await tick()

    if (typeof window !== 'undefined' && typeof window.print === 'function') {
      window.print()
    }
  }

  function requestDeactivate(voucher: AdminVoucher) {
    if (!voucher.ativo) return
    if (confirmDeactivateID === voucher.id) {
      confirmDeactivateID = null
      onDeactivateVoucher(voucher.id)
      return
    }
    confirmDeactivateID = voucher.id
  }
</script>

<section class="vouchers-panel" aria-labelledby="vouchers-title">
  <div class="voucher-operational">
    <div class="section-heading">
      <div>
        <h2 id="vouchers-title">Vouchers</h2>
        <p>Codigos para vender em dinheiro.</p>
      </div>
    </div>

  <form
    class="voucher-form"
    onsubmit={(event) => {
      event.preventDefault()
      submitVoucherForm()
    }}
  >
    <label class="field">
      Plano
      <select bind:value={voucherPlanoID} disabled={loading || planos.length === 0}>
        {#each planos as plano (plano.id)}
          <option value={plano.id}>{plano.nome}</option>
        {/each}
      </select>
    </label>

    <fieldset class="type-control">
      <legend>Tipo</legend>
      <div class="type-options">
        <label class:active={voucherTipo === 'single_use'}>
          <input type="radio" bind:group={voucherTipo} value="single_use" disabled={loading} />
          <span>Uso unico</span>
        </label>
        <label class:active={voucherTipo === 'universal'}>
          <input type="radio" bind:group={voucherTipo} value="universal" disabled={loading} />
          <span>Universal</span>
        </label>
      </div>
    </fieldset>

    <div class="form-grid">
      <label class="field">
        Prefixo
        <input bind:value={voucherPrefixo} maxlength="6" autocomplete="off" />
      </label>
      <label class="field">
        Quantidade
        <input bind:value={voucherQuantidade} min="1" max="200" type="number" />
      </label>
    </div>

    <div class="form-grid" class:single-field={voucherTipo !== 'universal'}>
      <label class="field">
        Validade (dias)
        <input bind:value={voucherValidadeDias} min="1" type="number" placeholder="Opcional" />
      </label>
      {#if voucherTipo === 'universal'}
        <label class="field">
          Usos maximos
          <input bind:value={voucherUsosMaximos} min="1" type="number" />
        </label>
      {/if}
    </div>

    <button type="submit" class="ink-button wide" disabled={loading || planos.length === 0}>
      Gerar vouchers
    </button>
  </form>

  <form
    class="filter-form"
    aria-label="Filtros de vouchers"
    onsubmit={(event) => {
      event.preventDefault()
      applyFilters()
    }}
  >
    <div class="filter-grid">
      <label class="field">
        Status
        <select bind:value={filtroStatus} disabled={loading}>
          <option value="ativo">Ativos</option>
          <option value="inativo">Inativos</option>
          <option value="todos">Todos</option>
        </select>
      </label>

      <label class="field">
        Plano do filtro
        <select bind:value={filtroPlanoID} disabled={loading}>
          <option value="">Todos</option>
          {#each planos as plano (plano.id)}
            <option value={String(plano.id)}>{plano.nome}</option>
          {/each}
        </select>
      </label>

      <label class="field">
        Codigo
        <input bind:value={filtroCodigo} autocomplete="off" maxlength="32" />
      </label>

      <label class="field">
        Lote
        <input bind:value={filtroLoteID} min="1" type="number" inputmode="numeric" />
      </label>
    </div>

    <div class="filter-actions">
      <button type="submit" class="ink-button" disabled={loading}>Aplicar filtros</button>
      <button type="button" class="ghost-button" onclick={exportCsv} disabled={loading}>
        Exportar CSV
      </button>
      <button
        type="button"
        class="ghost-button"
        onclick={printSheet}
        disabled={loading || vouchers.length === 0}
      >
        Imprimir folha
      </button>
    </div>
  </form>

  <div class="voucher-list">
    {#each vouchers.slice(0, 8) as voucher (voucher.id)}
      <article>
        <div class="voucher-main">
          <div>
            <h3>{voucher.codigo}</h3>
            <p>
              {voucher.plano.nome} - {voucher.usos_atuais}/{voucher.usos_maximos ?? 1} uso
              {#if voucher.lote_id}
                - lote {voucher.lote_id}
              {/if}
            </p>
          </div>
        </div>
        <div class="voucher-actions">
          <span class:inactive={!voucher.ativo}>{voucher.ativo ? 'ativo' : 'inativo'}</span>
          {#if voucher.ativo}
            <button
              type="button"
              class:confirming={confirmDeactivateID === voucher.id}
              onclick={() => requestDeactivate(voucher)}
              disabled={loading}
              aria-label={confirmDeactivateID === voucher.id
                ? `Confirmar desativacao ${voucher.codigo}`
                : `Desativar ${voucher.codigo}`}
            >
              {confirmDeactivateID === voucher.id ? 'Confirmar' : 'Desativar'}
            </button>
          {/if}
        </div>
      </article>
    {:else}
      <div class="empty-state compact">
        <h3>Nenhum voucher emitido</h3>
        <p>Gere um lote para venda presencial.</p>
      </div>
    {/each}
  </div>
  </div>

  {#if printSheetReady}
    <div class="print-shell">
      <AdminVoucherPrintSheet {vouchers} />
    </div>
  {/if}
</section>

<style>
  .vouchers-panel {
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 18px;
    background: white;
    box-shadow: 0 12px 28px rgba(15, 23, 42, 0.06);
  }

  .section-heading,
  .voucher-list article {
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
  .empty-state p {
    color: var(--color-muted);
  }

  .section-heading p {
    margin-top: 4px;
    font-size: 0.88rem;
  }

  .voucher-form {
    display: grid;
    gap: 10px;
    margin-bottom: 14px;
  }

  .voucher-form .field,
  .filter-form .field,
  .type-control legend {
    display: grid;
    gap: 6px;
    color: var(--color-ink);
    font-size: 0.78rem;
    font-weight: 850;
  }

  .voucher-form input:not([type='radio']),
  .voucher-form select,
  .filter-form input,
  .filter-form select {
    width: 100%;
    min-height: 42px;
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 0 11px;
    background: #f8fafc;
    color: var(--color-ink);
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
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 10px;
  }

  .filter-actions,
  .voucher-actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .filter-actions {
    justify-content: stretch;
    flex-wrap: wrap;
  }

  .filter-actions button {
    flex: 1;
    min-width: 132px;
  }

  .print-shell {
    position: fixed;
    top: 0;
    left: -10000px;
    width: 210mm;
    background: white;
  }

  .form-grid {
    display: grid;
    grid-template-columns: 1fr 116px;
    gap: 10px;
  }

  .form-grid.single-field {
    grid-template-columns: 1fr;
  }

  .type-control {
    display: grid;
    gap: 6px;
    border: 0;
    margin: 0;
    padding: 0;
  }

  .type-options {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8px;
  }

  .type-options label {
    position: relative;
    display: flex;
    min-height: 42px;
    align-items: center;
    justify-content: center;
    border: 1px solid var(--color-line);
    border-radius: 8px;
    background: #f8fafc;
    color: var(--color-muted);
    cursor: pointer;
    font-size: 0.82rem;
    font-weight: 900;
  }

  .type-options label.active {
    border-color: #0f766e;
    background: #ecfdf5;
    color: #0f766e;
  }

  .type-options input {
    position: absolute;
    opacity: 0;
  }

  .ink-button {
    min-height: 42px;
    border: 0;
    border-radius: 12px;
    padding: 0 14px;
    background: var(--color-ink);
    color: white;
    font-size: 0.86rem;
    font-weight: 850;
  }

  .ghost-button {
    min-height: 42px;
    border: 1px solid var(--color-line);
    border-radius: 12px;
    padding: 0 14px;
    background: white;
    color: var(--color-ink);
    font-size: 0.86rem;
    font-weight: 850;
  }

  .wide {
    width: 100%;
  }

  .voucher-list {
    display: grid;
    gap: 10px;
  }

  .voucher-list article {
    justify-content: space-between;
    gap: 12px;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    padding: 13px;
    background: #fcfdff;
  }

  .voucher-main {
    min-width: 0;
  }

  .voucher-list h3 {
    font-family: ui-monospace, "SFMono-Regular", Consolas, monospace;
    font-size: 0.92rem;
    font-weight: 900;
    letter-spacing: 0;
  }

  .voucher-list p {
    margin-top: 4px;
    color: var(--color-muted);
    font-size: 0.8rem;
  }

  .voucher-list span {
    border-radius: 999px;
    padding: 6px 8px;
    background: #dcfce7;
    color: #166534;
    font-size: 0.72rem;
    font-weight: 900;
  }

  .voucher-list span.inactive {
    background: #f1f5f9;
    color: #64748b;
  }

  .voucher-actions {
    flex-shrink: 0;
    justify-content: flex-end;
  }

  .voucher-actions button {
    min-height: 34px;
    border: 1px solid #fecaca;
    border-radius: 8px;
    padding: 0 9px;
    background: #fff1f2;
    color: #9f1239;
    font-size: 0.72rem;
    font-weight: 900;
  }

  .voucher-actions button.confirming {
    border-color: #fb923c;
    background: #ffedd5;
    color: #9a3412;
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

  .empty-state.compact {
    padding: 14px;
  }

  .empty-state h3 {
    font-size: 0.95rem;
    font-weight: 900;
  }

  .empty-state p {
    margin-top: 6px;
    font-size: 0.88rem;
  }

  @media (max-width: 420px) {
    .form-grid,
    .type-options,
    .filter-grid,
    .filter-actions {
      grid-template-columns: 1fr;
    }

    .filter-actions,
    .voucher-list article,
    .voucher-actions {
      align-items: stretch;
    }

    .voucher-list article,
    .voucher-actions {
      flex-direction: column;
    }

    .filter-actions {
      flex-direction: column;
    }

    .voucher-actions {
      width: 100%;
    }
  }

  @page {
    size: A4;
    margin: 10mm;
  }

  @media print {
    :global(body *) {
      visibility: hidden;
    }

    .vouchers-panel {
      border: 0;
      padding: 0;
      box-shadow: none;
    }

    .voucher-operational {
      display: none;
    }

    .print-shell,
    .print-shell :global(*) {
      visibility: visible;
    }

    .print-shell {
      position: absolute;
      inset: 0;
      display: block;
      margin: 0;
      border: 0;
      border-radius: 0;
      overflow: visible;
    }
  }
</style>

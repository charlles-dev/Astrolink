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

  $: activeVouchers = vouchers.filter((voucher) => voucher.ativo).length
  $: inactiveVouchers = vouchers.length - activeVouchers
  $: universalVouchers = vouchers.filter((voucher) => voucher.tipo === 'universal').length

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
    confirmDeactivateID = voucher.id
  }

  function confirmDeactivate(voucher: AdminVoucher) {
    confirmDeactivateID = null
    onDeactivateVoucher(voucher.id)
  }

  function cancelDeactivate() {
    confirmDeactivateID = null
  }
</script>

<section class="vouchers-panel card" aria-labelledby="vouchers-title">
  <div class="voucher-operational">
    <div class="section-heading">
      <div>
        <h2 id="vouchers-title">Vouchers</h2>
        <p>Códigos para venda presencial e ativação assistida.</p>
      </div>
    </div>

    <div class="voucher-toolbar" aria-label="Resumo de vouchers">
      <span><strong>{vouchers.length}</strong> no filtro</span>
      <span><strong>{activeVouchers}</strong> ativos</span>
      <span><strong>{inactiveVouchers}</strong> inativos</span>
      <span><strong>{universalVouchers}</strong> universais</span>
    </div>

  <form
    class="voucher-form workbench"
    onsubmit={(event) => {
      event.preventDefault()
      submitVoucherForm()
    }}
  >
    <div class="workbench-head">
      <div>
        <h3>Emitir lote</h3>
        <p>Configure prefixo, plano e limites antes de imprimir ou exportar.</p>
      </div>
    </div>

    <label class="field">
      Plano
      <select class="select select-bordered" bind:value={voucherPlanoID} disabled={loading || planos.length === 0}>
        {#each planos as plano (plano.id)}
          <option value={plano.id}>{plano.nome}</option>
        {/each}
      </select>
    </label>

    <fieldset class="type-control">
      <legend>Tipo</legend>
      <div class="type-options">
        <label class:active={voucherTipo === 'single_use'}>
          <input class="radio radio-primary" type="radio" bind:group={voucherTipo} value="single_use" disabled={loading} />
          <span>Uso unico</span>
        </label>
        <label class:active={voucherTipo === 'universal'}>
          <input class="radio radio-primary" type="radio" bind:group={voucherTipo} value="universal" disabled={loading} />
          <span>Universal</span>
        </label>
      </div>
    </fieldset>

    <div class="form-grid">
      <label class="field">
        Prefixo
        <input class="input input-bordered" bind:value={voucherPrefixo} maxlength="6" autocomplete="off" disabled={loading} />
      </label>
      <label class="field">
        Quantidade
        <input class="input input-bordered" bind:value={voucherQuantidade} min="1" max="200" type="number" disabled={loading} />
      </label>
    </div>

    <div class="form-grid" class:single-field={voucherTipo !== 'universal'}>
      <label class="field">
        Validade (dias)
        <input class="input input-bordered" bind:value={voucherValidadeDias} min="1" type="number" placeholder="Opcional" disabled={loading} />
      </label>
      {#if voucherTipo === 'universal'}
        <label class="field">
          Usos máximos
          <input class="input input-bordered" bind:value={voucherUsosMaximos} min="1" type="number" disabled={loading} />
        </label>
      {/if}
    </div>

    <button type="submit" class="btn btn-primary ink-button wide" disabled={loading || planos.length === 0}>
      Gerar vouchers
    </button>
  </form>

  <form
    class="filter-form workbench"
    aria-label="Filtros de vouchers"
    onsubmit={(event) => {
      event.preventDefault()
      applyFilters()
    }}
  >
    <div class="workbench-head">
      <div>
        <h3>Consulta operacional</h3>
        <p>Filtre por status, plano, código ou lote.</p>
      </div>
    </div>

    <div class="filter-grid">
      <label class="field">
        Status
        <select class="select select-bordered" bind:value={filtroStatus} disabled={loading}>
          <option value="ativo">Ativos</option>
          <option value="inativo">Inativos</option>
          <option value="todos">Todos</option>
        </select>
      </label>

      <label class="field">
        Plano do filtro
        <select class="select select-bordered" bind:value={filtroPlanoID} disabled={loading}>
          <option value="">Todos</option>
          {#each planos as plano (plano.id)}
            <option value={String(plano.id)}>{plano.nome}</option>
          {/each}
        </select>
      </label>

      <label class="field">
        Código
        <input class="input input-bordered" bind:value={filtroCodigo} autocomplete="off" maxlength="32" disabled={loading} />
      </label>

      <label class="field">
        Lote
        <input class="input input-bordered" bind:value={filtroLoteID} min="1" type="number" inputmode="numeric" disabled={loading} />
      </label>
    </div>

    <div class="filter-actions">
      <button type="submit" class="btn btn-primary ink-button" disabled={loading}>Aplicar filtros</button>
      <button type="button" class="btn btn-outline ghost-button" onclick={exportCsv} disabled={loading}>
        Exportar CSV
      </button>
      <button
        type="button"
        class="btn btn-outline ghost-button"
        onclick={printSheet}
        disabled={loading || vouchers.length === 0}
      >
        Gerar folha PDF
      </button>
    </div>
  </form>

  {#if vouchers.length > 0}
    <div class="list-summary" class:limited={vouchers.length > 8}>
      <strong>Mostrando {Math.min(vouchers.length, 8)} de {vouchers.length} vouchers</strong>
      {#if vouchers.length > 8}
        <span>Há mais {vouchers.length - 8} vouchers nos filtros atuais.</span>
      {/if}
    </div>
  {/if}

  <div class="voucher-list">
    {#if vouchers.length > 0}
      <div class="voucher-head" aria-hidden="true">
        <span>Código</span>
        <span>Plano e uso</span>
        <span>Estado</span>
        <span>Ações</span>
      </div>
    {/if}

    {#each vouchers.slice(0, 8) as voucher (voucher.id)}
      <article aria-label={`Voucher ${voucher.codigo}`}>
        <div class="voucher-main">
          <div>
            <h3>{voucher.codigo}</h3>
            <p>{voucher.tipo === 'universal' ? 'Universal' : 'Uso único'}</p>
          </div>
        </div>

        <div class="voucher-usage">
          <strong>{voucher.plano.nome}</strong>
          <span>
            {voucher.usos_atuais}/{voucher.usos_maximos ?? 1} uso
            {#if voucher.lote_id}
              - lote {voucher.lote_id}
            {/if}
          </span>
        </div>

        <div class="voucher-state">
          <span class="badge" class:inactive={!voucher.ativo}>{voucher.ativo ? 'ativo' : 'inativo'}</span>
          {#if voucher.validade_em}
            <small>validade registrada</small>
          {/if}
        </div>

        <div class="voucher-actions">
          {#if voucher.ativo}
            {#if confirmDeactivateID === voucher.id}
              <div class="confirm-box">
                <p>Desativar {voucher.codigo}? O código não poderá ser usado em novas ativações.</p>
                <div>
                  <button
                    type="button"
                    class="btn btn-outline danger-row-button confirming"
                    onclick={() => confirmDeactivate(voucher)}
                    disabled={loading}
                    aria-label={`Confirmar desativação de ${voucher.codigo}`}
                  >
                    Confirmar
                  </button>
                  <button
                    type="button"
                    class="btn btn-outline ghost-button"
                    onclick={cancelDeactivate}
                    disabled={loading}
                    aria-label={`Cancelar desativação de ${voucher.codigo}`}
                  >
                    Cancelar
                  </button>
                </div>
              </div>
            {:else}
              <button
                type="button"
                class="btn btn-outline danger-row-button"
                onclick={() => requestDeactivate(voucher)}
                disabled={loading}
                aria-label={`Desativar ${voucher.codigo}`}
              >
                Desativar
              </button>
            {/if}
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
    padding: var(--admin-panel-padding);
    background: var(--color-surface-raised);
    box-shadow: var(--shadow-panel);
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
    gap: 18px;
    margin-bottom: 14px;
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

  .voucher-toolbar {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 1px;
    margin-bottom: 14px;
    overflow: hidden;
    border: 1px solid var(--color-line);
    border-radius: 8px;
    background: var(--color-line);
  }

  .voucher-toolbar span {
    min-width: 0;
    display: grid;
    gap: 2px;
    padding: 10px 12px;
    background: var(--color-surface-subtle);
    color: var(--color-muted);
    font-size: 0.72rem;
    font-weight: 850;
    text-transform: uppercase;
  }

  .voucher-toolbar strong {
    color: var(--color-ink);
    font-size: 1rem;
    line-height: 1;
  }

  .workbench {
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 12px;
    background: var(--color-surface-subtle);
  }

  .workbench-head {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    border-bottom: 1px solid var(--color-line);
    padding-bottom: 10px;
  }

  .workbench-head h3 {
    color: var(--color-ink);
    font-size: 0.9rem;
    font-weight: 950;
  }

  .workbench-head p {
    margin-top: 3px;
    color: var(--color-muted);
    font-size: 0.76rem;
    font-weight: 800;
  }

  .voucher-form {
    display: grid;
    gap: 12px;
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
    min-height: 38px;
    border-radius: 8px;
    padding: 0 10px;
  }

  .filter-form {
    display: grid;
    gap: 12px;
    margin-bottom: 14px;
  }

  .filter-grid {
    display: grid;
    grid-template-columns: 120px minmax(160px, 1fr) minmax(140px, 1fr) 110px;
    gap: 10px;
  }

  .filter-actions,
  .voucher-actions {
    display: flex;
    align-items: center;
    gap: 10px;
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
    background: var(--color-surface);
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
    gap: 8px;
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
    min-height: 38px;
    align-items: center;
    justify-content: center;
    border-radius: 8px;
    background: var(--color-surface-subtle);
    color: var(--color-muted);
    cursor: pointer;
    font-size: 0.82rem;
    font-weight: 900;
  }

  .type-options label.active {
    border-color: var(--color-primary);
    background: var(--state-success-bg);
    color: var(--color-primary-strong);
  }

  .type-options input {
    position: absolute;
    opacity: 0;
  }

  .ink-button {
    min-height: 38px;
    border-radius: 8px;
    padding: 0 14px;
    font-size: 0.86rem;
    font-weight: 850;
  }

  .ghost-button {
    min-height: 38px;
    border-radius: 8px;
    padding: 0 14px;
    font-size: 0.86rem;
    font-weight: 850;
  }

  .wide {
    width: 100%;
  }

  .voucher-list {
    display: grid;
    gap: 0;
    overflow: hidden;
    border: 1px solid var(--color-line);
    border-radius: 8px;
    background: var(--color-line);
  }

  .voucher-head {
    display: grid;
    grid-template-columns: minmax(150px, 0.8fr) minmax(180px, 1.1fr) minmax(120px, 0.6fr) minmax(190px, 0.85fr);
    gap: 12px;
    padding: 9px 12px;
    background: var(--color-surface-muted);
    color: var(--color-muted);
    font-size: 0.68rem;
    font-weight: 950;
    letter-spacing: 0.02em;
    text-transform: uppercase;
  }

  .list-summary {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 12px;
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 10px 12px;
    background: var(--color-surface-subtle);
    color: var(--color-ink);
    font-size: 0.8rem;
  }

  .list-summary.limited {
    border-color: var(--state-warning-line);
    background: var(--state-warning-bg);
  }

  .list-summary strong,
  .list-summary span {
    font-weight: 850;
  }

  .list-summary span {
    color: var(--state-warning-text);
  }

  .voucher-list article {
    display: grid;
    grid-template-columns: minmax(150px, 0.8fr) minmax(180px, 1.1fr) minmax(120px, 0.6fr) minmax(190px, 0.85fr);
    align-items: center;
    gap: 12px;
    border: 0;
    border-top: 1px solid var(--color-line);
    border-radius: 0;
    padding: 12px;
    background: var(--color-row);
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

  .voucher-usage {
    display: grid;
    gap: 3px;
    min-width: 0;
  }

  .voucher-usage strong {
    color: var(--color-ink);
    font-size: 0.82rem;
    font-weight: 900;
    overflow-wrap: anywhere;
  }

  .voucher-usage span,
  .voucher-state small {
    color: var(--color-muted);
    font-size: 0.76rem;
    font-weight: 800;
  }

  .voucher-state {
    display: grid;
    justify-items: start;
    gap: 5px;
  }

  .voucher-state .badge {
    background: var(--state-success-bg);
    color: var(--state-success-text);
    font-size: 0.72rem;
    font-weight: 900;
  }

  .voucher-state .badge.inactive {
    background: var(--state-neutral-bg);
    color: var(--state-neutral-text);
  }

  .voucher-actions {
    flex-shrink: 0;
    justify-content: flex-end;
  }

  .voucher-actions button {
    min-height: 36px;
    border: 1px solid var(--state-error-line);
    border-radius: 8px;
    padding: 0 9px;
    background: var(--state-error-bg);
    color: var(--state-error-text);
    font-size: 0.72rem;
    font-weight: 900;
  }

  .voucher-actions button.confirming {
    border-color: var(--state-warning-line);
    background: var(--state-warning-bg);
    color: var(--state-warning-text);
  }

  .confirm-box {
    display: grid;
    gap: 8px;
    max-width: 320px;
    border: 1px solid var(--state-warning-line);
    border-radius: 8px;
    padding: 10px;
    background: var(--state-warning-bg);
  }

  .confirm-box p {
    color: var(--state-warning-text);
    font-size: 0.78rem;
    font-weight: 850;
    line-height: 1.35;
  }

  .confirm-box div {
    display: flex;
    gap: 8px;
    justify-content: flex-end;
  }

  .confirm-box .ghost-button {
    min-height: 36px;
    border-color: var(--color-line-strong);
    background: var(--color-surface);
    color: var(--color-ink);
    font-size: 0.72rem;
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

  @media (max-width: 640px) {
    .voucher-toolbar,
    .form-grid,
    .type-options,
    .filter-grid,
    .filter-actions,
    .voucher-list article {
      grid-template-columns: 1fr;
    }

    .voucher-head {
      display: none;
    }

    .filter-actions,
    .voucher-actions {
      align-items: stretch;
    }

    .voucher-actions {
      flex-direction: column;
    }

    .filter-actions {
      flex-direction: column;
    }

    .voucher-actions {
      width: 100%;
    }

    .list-summary,
    .confirm-box div {
      align-items: stretch;
      flex-direction: column;
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

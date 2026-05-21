<script lang="ts">
  import type { AdminVoucher, GenerateAdminVouchersBody, Plano } from '../../types'

  export let planos: Plano[] = []
  export let vouchers: AdminVoucher[] = []
  export let loading = false
  export let onGenerateVouchers: (input: GenerateAdminVouchersBody) => void = () => {}

  let voucherPlanoID = 0
  let voucherQuantidade = 1
  let voucherPrefixo = 'VIP'

  $: if (planos.length > 0 && !planos.some((plano) => plano.id === voucherPlanoID)) {
    voucherPlanoID = planos[0].id
  }

  function submitVoucherForm() {
    onGenerateVouchers({
      plano_id: Number(voucherPlanoID),
      quantidade: Number(voucherQuantidade),
      prefixo: voucherPrefixo.trim().toUpperCase()
    })
  }
</script>

<section class="vouchers-panel" aria-labelledby="vouchers-title">
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
    <label>
      Plano
      <select bind:value={voucherPlanoID} disabled={loading || planos.length === 0}>
        {#each planos as plano (plano.id)}
          <option value={plano.id}>{plano.nome}</option>
        {/each}
      </select>
    </label>
    <div class="form-grid">
      <label>
        Prefixo
        <input bind:value={voucherPrefixo} maxlength="6" autocomplete="off" />
      </label>
      <label>
        Quantidade
        <input bind:value={voucherQuantidade} min="1" max="200" type="number" />
      </label>
    </div>
    <button type="submit" class="ink-button wide" disabled={loading || planos.length === 0}>
      Gerar vouchers
    </button>
  </form>

  <div class="voucher-list">
    {#each vouchers.slice(0, 8) as voucher (voucher.id)}
      <article>
        <div>
          <h3>{voucher.codigo}</h3>
          <p>
            {voucher.plano.nome} - {voucher.usos_atuais}/{voucher.usos_maximos ?? 1} uso
          </p>
        </div>
        <span class:inactive={!voucher.ativo}>{voucher.ativo ? 'ativo' : 'inativo'}</span>
      </article>
    {:else}
      <div class="empty-state compact">
        <h3>Nenhum voucher emitido</h3>
        <p>Gere um lote para venda presencial.</p>
      </div>
    {/each}
  </div>
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

  .voucher-form label {
    display: grid;
    gap: 6px;
    color: var(--color-ink);
    font-size: 0.78rem;
    font-weight: 850;
  }

  .voucher-form input,
  .voucher-form select {
    width: 100%;
    min-height: 42px;
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 0 11px;
    background: #f8fafc;
    color: var(--color-ink);
  }

  .form-grid {
    display: grid;
    grid-template-columns: 1fr 116px;
    gap: 10px;
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
</style>

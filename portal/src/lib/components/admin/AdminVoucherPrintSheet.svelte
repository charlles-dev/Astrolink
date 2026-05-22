<script lang="ts">
  import type { AdminVoucher } from '../../types'

  export let vouchers: AdminVoucher[] = []

  const dateFormatter = new Intl.DateTimeFormat('pt-BR', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    timeZone: 'UTC'
  })

  function voucherTypeLabel(tipo: string) {
    return tipo === 'universal' ? 'Universal' : 'Uso unico'
  }

  function maxUses(voucher: AdminVoucher) {
    return voucher.usos_maximos ?? 1
  }

  function formatValidity(value?: string | null) {
    if (!value) return ''

    const date = new Date(value)
    if (Number.isNaN(date.getTime())) return ''

    return dateFormatter.format(date)
  }
</script>

<section class="voucher-print-sheet" data-testid="voucher-print-sheet" aria-label="Folha de vouchers">
  <header>
    <h2>Vouchers Astrolink</h2>
    <p>{vouchers.length} codigo{vouchers.length === 1 ? '' : 's'}</p>
  </header>

  <div class="ticket-grid">
    {#each vouchers as voucher (voucher.id)}
      <article class="ticket">
        <div class="ticket-code">{voucher.codigo}</div>
        <div class="ticket-plan">{voucher.plano.nome}</div>

        <dl>
          <div>
            <dt>Tipo</dt>
            <dd>{voucherTypeLabel(voucher.tipo)}</dd>
          </div>
          <div>
            <dt>Uso</dt>
            <dd>{voucher.usos_atuais}/{maxUses(voucher)}</dd>
          </div>
          {#if voucher.lote_id}
            <div>
              <dt>Lote</dt>
              <dd>Lote {voucher.lote_id}</dd>
            </div>
          {/if}
          {#if formatValidity(voucher.validade_em)}
            <div>
              <dt>Validade</dt>
              <dd>{formatValidity(voucher.validade_em)}</dd>
            </div>
          {/if}
        </dl>
      </article>
    {/each}
  </div>
</section>

<style>
  .voucher-print-sheet {
    display: grid;
    gap: 12px;
    padding: 16px;
    background: white;
    color: #0f172a;
  }

  header {
    display: flex;
    align-items: end;
    justify-content: space-between;
    gap: 12px;
    border-bottom: 1px solid #cbd5e1;
    padding-bottom: 10px;
  }

  h2,
  p,
  dl,
  dd {
    margin: 0;
  }

  h2 {
    font-size: 1rem;
    font-weight: 900;
  }

  p {
    color: #475569;
    font-size: 0.78rem;
    font-weight: 800;
  }

  .ticket-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 10px;
  }

  .ticket {
    break-inside: avoid;
    border: 1px solid #94a3b8;
    border-radius: 8px;
    padding: 12px;
    background: #ffffff;
  }

  .ticket-code {
    overflow-wrap: anywhere;
    font-family: ui-monospace, "SFMono-Regular", Consolas, monospace;
    font-size: 1.1rem;
    font-weight: 900;
    letter-spacing: 0;
  }

  .ticket-plan {
    margin-top: 4px;
    color: #334155;
    font-size: 0.82rem;
    font-weight: 850;
  }

  dl {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8px;
    margin-top: 12px;
  }

  dt {
    color: #64748b;
    font-size: 0.62rem;
    font-weight: 900;
    text-transform: uppercase;
  }

  dd {
    margin-top: 2px;
    color: #0f172a;
    font-size: 0.8rem;
    font-weight: 900;
  }

  @media print {
    .voucher-print-sheet {
      padding: 0;
    }

    .ticket-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 8mm;
    }

    .ticket {
      min-height: 42mm;
      border-color: #475569;
      box-shadow: none;
    }
  }
</style>

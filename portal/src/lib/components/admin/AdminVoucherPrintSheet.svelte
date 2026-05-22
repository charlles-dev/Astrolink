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

  function uniqueValues(values: string[]) {
    return Array.from(new Set(values.filter(Boolean)))
  }

  $: lotes = uniqueValues(vouchers.map((voucher) => voucher.lote_id ? `Lote ${voucher.lote_id}` : ''))
  $: planos = uniqueValues(vouchers.map((voucher) => voucher.plano.nome))
  $: usosDisponiveis = vouchers.reduce((total, voucher) => {
    return total + Math.max(maxUses(voucher) - voucher.usos_atuais, 0)
  }, 0)

  function formatValidity(value?: string | null) {
    if (!value) return ''

    const date = new Date(value)
    if (Number.isNaN(date.getTime())) return ''

    return dateFormatter.format(date)
  }
</script>

<section class="voucher-print-sheet" data-testid="voucher-print-sheet" aria-label="Folha de vouchers">
  <header class="sheet-header">
    <div>
      <p class="sheet-kicker">Folha para PDF ou impressao</p>
      <h2>Astrolink</h2>
      <p class="sheet-subtitle">Vouchers prontos para venda presencial</p>
    </div>

    <dl class="sheet-summary" aria-label="Resumo da folha">
      <div>
        <dt>Quantidade</dt>
        <dd>{vouchers.length} voucher{vouchers.length === 1 ? '' : 's'}</dd>
      </div>
      <div>
        <dt>Lote</dt>
        <dd>{lotes.length ? lotes.join(', ') : 'Sem lote'}</dd>
      </div>
      <div>
        <dt>Plano</dt>
        <dd>{planos.length ? planos.join(', ') : 'Sem plano'}</dd>
      </div>
      <div>
        <dt>Usos livres</dt>
        <dd>{usosDisponiveis}</dd>
      </div>
    </dl>
  </header>

  <p class="cut-note">Recorte nas linhas pontilhadas</p>

  <div class="ticket-grid">
    {#each vouchers as voucher (voucher.id)}
      <article class="ticket">
        <div class="ticket-brand">Astrolink</div>
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

        <ol>
          <li>Digite este codigo no portal do Wi-Fi.</li>
          <li>Use antes de entregar ou vender a outro cliente.</li>
          {#if formatValidity(voucher.validade_em)}
            <li>Valido somente ate a data indicada.</li>
          {/if}
        </ol>
      </article>
    {/each}
  </div>
</section>

<style>
  .voucher-print-sheet {
    display: grid;
    width: 210mm;
    min-height: 297mm;
    gap: 10mm;
    padding: 12mm;
    background: white;
    color: #0f172a;
    font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  }

  .sheet-header {
    display: grid;
    grid-template-columns: minmax(0, 0.9fr) minmax(0, 1.1fr);
    align-items: start;
    gap: 10mm;
    border-bottom: 2px solid #0f172a;
    padding-bottom: 7mm;
  }

  h2,
  p,
  dl,
  dd,
  ol {
    margin: 0;
  }

  h2 {
    color: #0f172a;
    font-size: 2.2rem;
    font-weight: 900;
    line-height: 1;
  }

  .sheet-kicker,
  .sheet-subtitle,
  .cut-note {
    color: #475569;
    font-size: 0.78rem;
    font-weight: 800;
  }

  .sheet-kicker {
    margin-bottom: 4px;
    text-transform: uppercase;
  }

  .sheet-subtitle {
    margin-top: 7px;
  }

  .sheet-summary {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 4mm;
  }

  .sheet-summary div {
    border-left: 3px solid #14b8a6;
    padding-left: 3mm;
  }

  .sheet-summary dd {
    overflow-wrap: anywhere;
  }

  .cut-note {
    border: 1px dashed #94a3b8;
    padding: 3mm 4mm;
    text-align: center;
  }

  .ticket-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8mm;
  }

  .ticket {
    position: relative;
    break-inside: avoid;
    min-height: 68mm;
    border: 1px dashed #64748b;
    border-radius: 2mm;
    padding: 6mm;
    background: #ffffff;
  }

  .ticket::before,
  .ticket::after {
    position: absolute;
    color: #64748b;
    font-size: 0.72rem;
    font-weight: 900;
  }

  .ticket::before {
    content: '+';
    top: -4mm;
    left: -3mm;
  }

  .ticket::after {
    content: '+';
    right: -3mm;
    bottom: -4mm;
  }

  .ticket-brand {
    color: #0f766e;
    font-size: 0.72rem;
    font-weight: 900;
    text-transform: uppercase;
  }

  .ticket-code {
    overflow-wrap: anywhere;
    font-family: ui-monospace, "SFMono-Regular", Consolas, monospace;
    margin-top: 4mm;
    color: #0f172a;
    font-size: 1.55rem;
    font-weight: 900;
    letter-spacing: 0;
    line-height: 1.1;
  }

  .ticket-plan {
    margin-top: 2mm;
    color: #334155;
    font-size: 0.9rem;
    font-weight: 850;
  }

  .ticket dl {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 3mm;
    margin-top: 5mm;
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

  ol {
    display: grid;
    gap: 1.5mm;
    margin-top: 5mm;
    padding-left: 4mm;
    color: #334155;
    font-size: 0.72rem;
    font-weight: 750;
    line-height: 1.35;
  }

  @media (max-width: 760px) {
    .voucher-print-sheet {
      width: 100%;
      min-height: auto;
      padding: 18px;
      gap: 18px;
    }

    .sheet-header,
    .ticket-grid {
      grid-template-columns: 1fr;
    }

    .ticket {
      min-height: auto;
    }
  }

  @media print {
    .voucher-print-sheet {
      width: auto;
      min-height: auto;
      padding: 0;
      gap: 7mm;
    }

    .ticket-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 8mm;
    }

    .ticket {
      min-height: 68mm;
      box-shadow: none;
    }
  }
</style>

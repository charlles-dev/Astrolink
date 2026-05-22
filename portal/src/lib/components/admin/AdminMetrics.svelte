<script lang="ts">
  import type { AdminHealthResponse, AdminUser, AdminVoucher, Plano } from '../../types'

  export let health: AdminHealthResponse | null = null
  export let planos: Plano[] = []
  export let usuarios: AdminUser[] = []
  export let vouchers: AdminVoucher[] = []

  $: activeUsers = usuarios.filter((usuario) => usuario.status === 'ativo').length
  $: visiblePlans = planos.filter((plano) => plano.ativo).length
  $: activeVouchers = vouchers.filter((voucher) => voucher.ativo).length
  $: routerOnline = health?.checks.roteadores.online ?? 0
  $: routerTotal = health?.checks.roteadores.total ?? 0
</script>

<div class="metric-grid">
  <article class="metric-card tone-teal">
    <span class="metric-accent" aria-hidden="true"></span>
    <span class="metric-label">Usuarios ativos</span>
    <strong>{activeUsers}</strong>
    <small class="metric-footnote">{usuarios.length} conhecidos</small>
  </article>
  <article class="metric-card tone-blue">
    <span class="metric-accent" aria-hidden="true"></span>
    <span class="metric-label">Planos ativos</span>
    <strong>{visiblePlans}</strong>
    <small class="metric-footnote">{planos.length} cadastrados</small>
  </article>
  <article class="metric-card tone-green">
    <span class="metric-accent" aria-hidden="true"></span>
    <span class="metric-label">Vouchers ativos</span>
    <strong>{activeVouchers}</strong>
    <small class="metric-footnote">{vouchers.length} emitidos</small>
  </article>
  <article class="metric-card tone-slate">
    <span class="metric-accent" aria-hidden="true"></span>
    <span class="metric-label">Banco</span>
    <strong>{health?.checks.banco_dados.status ?? 'sem dados'}</strong>
    <small class="metric-footnote">{routerOnline}/{routerTotal} roteador online</small>
  </article>
</div>

<style>
  .metric-grid {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 12px;
    width: 100%;
  }

  .metric-card {
    position: relative;
    min-height: 132px;
    display: grid;
    gap: 8px;
    align-content: end;
    overflow: hidden;
    border: 1px solid #dbe4ef;
    border-radius: 8px;
    padding: 18px;
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 252, 0.96));
    box-shadow: 0 14px 30px rgba(15, 23, 42, 0.06);
  }

  .metric-accent {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 4px;
    background: #14b8a6;
  }

  .tone-blue .metric-accent {
    background: #38bdf8;
  }

  .tone-green .metric-accent {
    background: #22c55e;
  }

  .tone-slate .metric-accent {
    background: #64748b;
  }

  .metric-label,
  .metric-footnote {
    color: #64748b;
  }

  .metric-label {
    font-size: 0.78rem;
    font-weight: 850;
    text-transform: uppercase;
  }

  .metric-card strong {
    min-width: 0;
    overflow-wrap: anywhere;
    font-size: 1.82rem;
    font-weight: 930;
    line-height: 1;
  }

  .metric-footnote {
    font-size: 0.78rem;
    font-weight: 750;
  }

  @media (max-width: 900px) {
    .metric-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }

  @media (max-width: 620px) {
    .metric-grid {
      grid-template-columns: 1fr;
    }
  }
</style>

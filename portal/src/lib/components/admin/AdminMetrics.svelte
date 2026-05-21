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
  <article>
    <span>Usuarios ativos</span>
    <strong>{activeUsers}</strong>
    <small>{usuarios.length} conhecidos</small>
  </article>
  <article>
    <span>Planos ativos</span>
    <strong>{visiblePlans}</strong>
    <small>{planos.length} cadastrados</small>
  </article>
  <article>
    <span>Vouchers ativos</span>
    <strong>{activeVouchers}</strong>
    <small>{vouchers.length} emitidos</small>
  </article>
  <article>
    <span>Banco</span>
    <strong>{health?.checks.banco_dados.status ?? 'sem dados'}</strong>
    <small>{routerOnline}/{routerTotal} roteador online</small>
  </article>
</div>

<style>
  .metric-grid {
    max-width: 1180px;
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 12px;
    margin: 0 auto;
  }

  .metric-grid article {
    min-height: 130px;
    display: grid;
    gap: 8px;
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 18px;
    background: white;
    box-shadow: 0 12px 28px rgba(15, 23, 42, 0.06);
  }

  .metric-grid span,
  .metric-grid small {
    color: var(--color-muted);
  }

  .metric-grid span {
    font-size: 0.78rem;
    font-weight: 850;
    text-transform: uppercase;
  }

  .metric-grid strong {
    font-size: 1.65rem;
    font-weight: 930;
    line-height: 1;
  }

  .metric-grid small {
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

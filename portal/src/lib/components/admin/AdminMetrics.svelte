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

<section class="metric-grid" aria-label="Resumo operacional">
  <article class="metric-cell tone-teal">
    <span class="metric-state" aria-hidden="true"></span>
    <div class="metric-copy">
      <span class="metric-label">Usuários ativos</span>
      <small>{usuarios.length} conhecidos</small>
    </div>
    <strong>{activeUsers}</strong>
  </article>

  <article class="metric-cell tone-blue">
    <span class="metric-state" aria-hidden="true"></span>
    <div class="metric-copy">
      <span class="metric-label">Planos ativos</span>
      <small>{planos.length} cadastrados</small>
    </div>
    <strong>{visiblePlans}</strong>
  </article>

  <article class="metric-cell tone-green">
    <span class="metric-state" aria-hidden="true"></span>
    <div class="metric-copy">
      <span class="metric-label">Vouchers ativos</span>
      <small>{vouchers.length} emitidos</small>
    </div>
    <strong>{activeVouchers}</strong>
  </article>

  <article class="metric-cell tone-slate">
    <span class="metric-state" aria-hidden="true"></span>
    <div class="metric-copy">
      <span class="metric-label">Banco</span>
      <small>{routerOnline}/{routerTotal} roteador online</small>
    </div>
    <strong>{health?.checks.banco_dados.status ?? 'sem dados'}</strong>
  </article>
</section>

<style>
  .metric-grid {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 1px;
    overflow: hidden;
    border: 1px solid var(--color-line);
    border-radius: var(--admin-panel-radius);
    background: var(--color-line);
    box-shadow: var(--shadow-panel);
  }

  .metric-cell {
    min-width: 0;
    display: grid;
    grid-template-columns: auto minmax(0, 1fr) auto;
    align-items: center;
    gap: 12px;
    min-height: 88px;
    padding: 16px;
    background: var(--color-surface-raised);
  }

  .metric-state {
    width: 4px;
    height: 44px;
    border-radius: 999px;
    background: var(--color-primary);
  }

  .tone-blue .metric-state {
    background: var(--color-secondary);
  }

  .tone-green .metric-state {
    background: var(--color-success);
  }

  .tone-slate .metric-state {
    background: var(--color-muted);
  }

  .metric-copy {
    min-width: 0;
    display: grid;
    gap: 4px;
  }

  .metric-label,
  .metric-copy small {
    overflow-wrap: anywhere;
  }

  .metric-label {
    color: var(--color-ink);
    font-size: 0.82rem;
    font-weight: 900;
    line-height: 1.2;
  }

  .metric-copy small {
    color: var(--color-muted);
    font-size: 0.74rem;
    font-weight: 750;
    line-height: 1.25;
  }

  .metric-cell strong {
    min-width: 0;
    color: var(--color-ink);
    font-size: clamp(1.3rem, 2vw, 1.65rem);
    font-weight: 950;
    line-height: 1;
    overflow-wrap: anywhere;
    text-align: right;
  }

  @media (max-width: 980px) {
    .metric-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }

  @media (max-width: 620px) {
    .metric-grid {
      grid-template-columns: 1fr;
    }

    .metric-cell {
      min-height: 74px;
      padding: 14px;
    }
  }
</style>

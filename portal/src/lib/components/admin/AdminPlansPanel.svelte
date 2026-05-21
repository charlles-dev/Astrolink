<script lang="ts">
  import { formatCurrency } from '../../format'
  import type { Plano } from '../../types'

  export let planos: Plano[] = []
</script>

<section class="plans-panel" aria-labelledby="planos-title">
  <div class="section-heading">
    <div>
      <h2 id="planos-title">Planos</h2>
      <p>Oferta atual do portal.</p>
    </div>
  </div>

  <div class="plan-admin-list">
    {#each planos as plano (plano.id)}
      <article>
        <div>
          <h3>{plano.nome}</h3>
          <p>{plano.duracao_formatada}</p>
        </div>
        <strong>{formatCurrency(plano.preco)}</strong>
      </article>
    {:else}
      <div class="empty-state compact">
        <h3>Nenhum plano</h3>
        <p>Cadastre planos pelo backend nas proximas etapas.</p>
      </div>
    {/each}
  </div>
</section>

<style>
  .plans-panel {
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 18px;
    background: white;
    box-shadow: 0 12px 28px rgba(15, 23, 42, 0.06);
  }

  .section-heading,
  .plan-admin-list article {
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
  .plan-admin-list p,
  .empty-state p {
    color: var(--color-muted);
  }

  .section-heading p {
    margin-top: 4px;
    font-size: 0.88rem;
  }

  .plan-admin-list {
    display: grid;
    gap: 10px;
  }

  .plan-admin-list article {
    justify-content: space-between;
    gap: 12px;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    padding: 13px;
    background: #fcfdff;
  }

  .plan-admin-list h3,
  .empty-state h3 {
    font-size: 0.95rem;
    font-weight: 900;
  }

  .plan-admin-list p {
    margin-top: 4px;
    font-size: 0.82rem;
    line-height: 1.35;
  }

  .plan-admin-list strong {
    white-space: nowrap;
    font-size: 0.95rem;
    font-weight: 920;
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

  .empty-state p {
    margin-top: 6px;
    font-size: 0.88rem;
  }
</style>

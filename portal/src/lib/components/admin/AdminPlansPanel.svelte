<script lang="ts">
  import { formatCurrency } from '../../format'
  import type { AdminPlanBody, Plano } from '../../types'
  import AdminPlanForm from './AdminPlanForm.svelte'

  export let planos: Plano[] = []
  export let loading = false
  export let onSavePlan: (input: AdminPlanBody, id?: number) => Promise<void> | void = () => {}
  export let onTogglePlanStatus: (id: number, ativo: boolean) => Promise<void> | void = () => {}

  let editingPlan: Plano | null = null

  async function savePlan(input: AdminPlanBody) {
    try {
      if (editingPlan) {
        await onSavePlan(input, editingPlan.id)
      } else {
        await onSavePlan(input)
      }
      editingPlan = null
    } catch {
      // Parent page owns the visible error message.
    }
  }

  function editPlan(plano: Plano) {
    editingPlan = plano
  }

  function cancelEdit() {
    editingPlan = null
  }

  async function toggleStatus(plano: Plano) {
    try {
      await onTogglePlanStatus(plano.id, !plano.ativo)
    } catch {
      // Parent page owns the visible error message.
    }
  }
</script>

<section class="plans-panel" aria-labelledby="planos-title">
  <div class="section-heading">
    <div>
      <h2 id="planos-title">Planos</h2>
      <p>Oferta atual do portal.</p>
    </div>
  </div>

  <AdminPlanForm plan={editingPlan} {loading} onSubmit={savePlan} onCancel={cancelEdit} />

  <div class="plan-admin-list" aria-label="Planos cadastrados">
    {#each planos as plano (plano.id)}
      <article aria-label={plano.nome}>
        <div class="plan-main">
          <div>
            <div class="title-row">
              <h3>{plano.nome}</h3>
              <span class:inactive={!plano.ativo}>{plano.ativo ? 'Ativo' : 'Inativo'}</span>
              {#if plano.recomendado}
                <span>Recomendado</span>
              {/if}
            </div>
            <p>{plano.descricao || 'Sem descricao'}</p>
            <p class="plan-meta">
              <span>{plano.duracao_formatada}</span>
              <span aria-hidden="true"> - </span>
              <span>{plano.velocidade_down}/{plano.velocidade_up} Mbps</span>
              {#if plano.dados_mb}
                <span aria-hidden="true"> - </span><span>{plano.dados_mb} MB</span>
              {/if}
            </p>
          </div>
          <strong>{formatCurrency(plano.preco)}</strong>
        </div>

        <div class="row-actions">
          <button type="button" class="ghost-button" onclick={() => editPlan(plano)} disabled={loading}>
            Editar <span class="sr-only">{plano.nome}</span>
          </button>
          <button type="button" class="ghost-button" onclick={() => void toggleStatus(plano)} disabled={loading}>
            {plano.ativo ? 'Inativar' : 'Ativar'} <span class="sr-only">{plano.nome}</span>
          </button>
        </div>
      </article>
    {:else}
      <div class="empty-state compact">
        <h3>Nenhum plano</h3>
        <p>Cadastre o primeiro plano acima.</p>
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

  .section-heading {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 14px;
    margin-bottom: 16px;
  }

  h2,
  h3,
  p {
    margin: 0;
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
    margin-top: 16px;
  }

  .plan-admin-list article {
    display: grid;
    gap: 12px;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    padding: 13px;
    background: #fcfdff;
  }

  .plan-main,
  .title-row,
  .row-actions {
    display: flex;
    align-items: center;
  }

  .plan-main {
    justify-content: space-between;
    gap: 12px;
  }

  .title-row {
    flex-wrap: wrap;
    gap: 7px;
  }

  .title-row span {
    border-radius: 999px;
    padding: 3px 7px;
    background: #dcfce7;
    color: #166534;
    font-size: 0.68rem;
    font-weight: 900;
    text-transform: uppercase;
  }

  .title-row .inactive {
    background: #fee2e2;
    color: #991b1b;
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

  .plan-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 0 5px;
  }

  .plan-admin-list strong {
    white-space: nowrap;
    font-size: 0.95rem;
    font-weight: 920;
  }

  .row-actions {
    justify-content: flex-end;
    gap: 8px;
  }

  .ghost-button {
    min-height: 36px;
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 0 11px;
    background: white;
    color: var(--color-ink);
    font-size: 0.8rem;
    font-weight: 900;
  }

  .ghost-button:disabled {
    cursor: not-allowed;
    opacity: 0.6;
  }

  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
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

  @media (max-width: 520px) {
    .plan-main,
    .row-actions {
      align-items: stretch;
      flex-direction: column;
    }

    .row-actions button {
      width: 100%;
    }
  }
</style>

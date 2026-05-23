<script lang="ts">
  import { formatCurrency } from '../../format'
  import type { AdminPlanBody, Plano } from '../../types'
  import AdminPlanForm from './AdminPlanForm.svelte'

  export let planos: Plano[] = []
  export let loading = false
  export let onSavePlan: (input: AdminPlanBody, id?: number) => Promise<void> | void = () => {}
  export let onTogglePlanStatus: (id: number, ativo: boolean) => Promise<void> | void = () => {}

  let editingPlan: Plano | null = null

  $: activePlans = planos.filter((plano) => plano.ativo).length
  $: visiblePlans = planos.filter((plano) => plano.visivel_portal).length
  $: recommendedPlans = planos.filter((plano) => plano.recomendado).length

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

<section class="plans-panel card" aria-labelledby="planos-title">
  <div class="section-heading">
    <div>
      <h2 id="planos-title">Planos</h2>
      <p>{editingPlan ? `Editando: ${editingPlan.nome}` : 'Catálogo comercial publicado no portal local.'}</p>
    </div>
    {#if editingPlan}
      <button
        type="button"
        class="btn btn-outline ghost-button"
        onclick={cancelEdit}
        disabled={loading}
        aria-label={`Cancelar edição de ${editingPlan.nome}`}
      >
        Cancelar
      </button>
    {/if}
  </div>

  <div class="plan-toolbar" aria-label="Resumo dos planos">
    <span><strong>{planos.length}</strong> cadastrados</span>
    <span><strong>{activePlans}</strong> ativos</span>
    <span><strong>{visiblePlans}</strong> no portal</span>
    <span><strong>{recommendedPlans}</strong> recomendados</span>
  </div>

  <div class="form-dock" class:editing={editingPlan}>
    <AdminPlanForm plan={editingPlan} {loading} onSubmit={savePlan} onCancel={cancelEdit} />
  </div>

  <div class="plan-admin-list" aria-label="Planos cadastrados">
    {#if planos.length > 0}
      <div class="list-head" aria-hidden="true">
        <span>Plano</span>
        <span>Oferta</span>
        <span>Estado</span>
        <span>Ações</span>
      </div>
    {/if}

    {#each planos as plano (plano.id)}
      <article aria-label={plano.nome}>
        <div class="plan-main">
          <div>
            <div class="title-row">
              <h3>{plano.nome}</h3>
              <span class="badge" class:inactive={!plano.ativo}>{plano.ativo ? 'Ativo' : 'Inativo'}</span>
              {#if plano.recomendado}
                <span class="badge">Recomendado</span>
              {/if}
            </div>
            <p>{plano.descricao || 'Sem descricao'}</p>
            <p class="plan-meta">
              <span>{plano.duracao_formatada}</span>
              {#if plano.dados_mb}
                <span aria-hidden="true"> - </span><span>{plano.dados_mb} MB</span>
              {/if}
            </p>
          </div>
        </div>

        <div class="plan-offer">
          <strong>{formatCurrency(plano.preco)}</strong>
          <span>{plano.velocidade_down}/{plano.velocidade_up} Mbps</span>
        </div>

        <div class="plan-state">
          <span class="state-dot" class:offline={!plano.ativo} aria-hidden="true"></span>
          <span>{plano.visivel_portal ? 'Visível no portal' : 'Oculto no portal'}</span>
        </div>

        <div class="row-actions">
          <button type="button" class="btn btn-outline ghost-button" onclick={() => editPlan(plano)} disabled={loading}>
            Editar <span class="sr-only">{plano.nome}</span>
          </button>
          <button type="button" class="btn btn-outline ghost-button" onclick={() => void toggleStatus(plano)} disabled={loading}>
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
    padding: var(--admin-panel-padding);
    background: var(--color-surface-raised);
    box-shadow: var(--shadow-panel);
  }

  .section-heading {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 18px;
    margin-bottom: 14px;
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

  .plan-toolbar {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 1px;
    margin-bottom: 14px;
    overflow: hidden;
    border: 1px solid var(--color-line);
    border-radius: 8px;
    background: var(--color-line);
  }

  .plan-toolbar span {
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

  .plan-toolbar strong {
    color: var(--color-ink);
    font-size: 1rem;
    line-height: 1;
  }

  .form-dock {
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 14px;
    background: var(--color-surface-subtle);
  }

  .form-dock.editing {
    border-color: var(--state-info-line);
    background: color-mix(in srgb, var(--state-info-bg) 34%, var(--color-surface-raised));
  }

  .plan-admin-list {
    display: grid;
    gap: 0;
    margin-top: 14px;
    overflow: hidden;
    border: 1px solid var(--color-line);
    border-radius: 8px;
    background: var(--color-line);
  }

  .list-head {
    display: grid;
    grid-template-columns: minmax(220px, 1.5fr) minmax(120px, 0.55fr) minmax(140px, 0.65fr) minmax(180px, 0.75fr);
    gap: 12px;
    padding: 9px 12px;
    background: var(--color-surface-muted);
    color: var(--color-muted);
    font-size: 0.68rem;
    font-weight: 950;
    letter-spacing: 0.02em;
    text-transform: uppercase;
  }

  .plan-admin-list article {
    display: grid;
    grid-template-columns: minmax(220px, 1.5fr) minmax(120px, 0.55fr) minmax(140px, 0.65fr) minmax(180px, 0.75fr);
    align-items: center;
    gap: 12px;
    border: 0;
    border-top: 1px solid var(--color-line);
    border-radius: 0;
    padding: 12px;
    background: var(--color-row);
  }

  .plan-main,
  .title-row,
  .row-actions {
    display: flex;
    align-items: center;
  }

  .plan-main {
    min-width: 0;
    gap: 12px;
  }

  .title-row {
    flex-wrap: wrap;
    gap: 8px;
  }

  .title-row span {
    background: var(--state-success-bg);
    color: var(--state-success-text);
    font-size: 0.68rem;
    font-weight: 900;
    text-transform: uppercase;
  }

  .title-row .inactive {
    background: var(--state-error-bg);
    color: var(--state-error-text);
  }

  .plan-admin-list h3,
  .empty-state h3 {
    font-size: 0.95rem;
    font-weight: 900;
  }

  .plan-admin-list p {
    margin-top: 3px;
    font-size: 0.82rem;
    line-height: 1.35;
  }

  .plan-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 0 5px;
  }

  .plan-offer {
    display: grid;
    gap: 3px;
  }

  .plan-offer strong {
    white-space: nowrap;
    color: var(--color-ink);
    font-size: 0.95rem;
    font-weight: 920;
  }

  .plan-offer span,
  .plan-state span:not(.state-dot) {
    color: var(--color-muted);
    font-size: 0.76rem;
    font-weight: 850;
  }

  .plan-state {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
  }

  .state-dot {
    flex: 0 0 auto;
    width: 9px;
    height: 9px;
    border-radius: 999px;
    background: var(--state-success-text);
    box-shadow: 0 0 0 4px var(--state-success-bg);
  }

  .state-dot.offline {
    background: var(--state-error-text);
    box-shadow: 0 0 0 4px var(--state-error-bg);
  }

  .row-actions {
    justify-content: flex-end;
    gap: 8px;
  }

  .ghost-button {
    min-height: 38px;
    border: 1px solid var(--color-line-strong);
    border-radius: 8px;
    padding: 0 11px;
    background: var(--color-surface-raised);
    color: var(--color-ink);
    font-size: 0.8rem;
    font-weight: 900;
    box-shadow: none;
  }

  .ghost-button:hover {
    border-color: color-mix(in srgb, var(--color-primary) 46%, var(--color-line-strong));
    background: var(--color-surface-muted);
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
    padding: 22px;
    background: var(--color-surface-subtle);
  }

  .empty-state.compact {
    padding: 14px;
  }

  .empty-state p {
    margin-top: 6px;
    font-size: 0.88rem;
  }

  @media (max-width: 520px) {
    .plan-toolbar,
    .list-head,
    .plan-admin-list article {
      grid-template-columns: 1fr;
    }

    .list-head {
      display: none;
    }

    .row-actions {
      align-items: stretch;
      flex-direction: column;
    }

    .row-actions button {
      width: 100%;
    }
  }
</style>

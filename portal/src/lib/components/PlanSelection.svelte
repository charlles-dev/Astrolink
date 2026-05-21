<script lang="ts">
  import type { Plano } from '../types'
  import ErrorMessage from './ErrorMessage.svelte'
  import PlanCard from './PlanCard.svelte'

  export let planos: Plano[] = []
  export let loading = false
  export let error = ''
  export let onBack: () => void = () => {}
  export let onRetry: () => void = () => {}
  export let onVoucher: () => void = () => {}
  export let onSelectPlan: (plano: Plano) => void = () => {}
</script>

<div class="light-screen">
  <header class="screen-header">
    <button class="icon-button" type="button" aria-label="Voltar" onclick={onBack}>
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <path d="m15 18-6-6 6-6" />
      </svg>
    </button>
    <div>
      <h1>Escolha seu acesso</h1>
      <p>Planos disponiveis para este ponto Wi-Fi.</p>
    </div>
  </header>

  <ErrorMessage message={error} actionLabel="Recarregar" onAction={onRetry} />

  {#if loading}
    <div class="loading-state" role="status">Carregando planos...</div>
  {:else if planos.length === 0}
    <section class="empty-state">
      <h2>Nenhum plano disponivel</h2>
      <p>Voce ainda pode liberar o acesso com um voucher ativo.</p>
      <button type="button" onclick={onVoucher}>Usar voucher</button>
    </section>
  {:else}
    <div class="plan-list">
      {#each planos as plano (plano.id)}
        <PlanCard {plano} disabled={loading} onSelect={onSelectPlan} />
      {/each}
    </div>
  {/if}

  <button class="text-action" type="button" onclick={onVoucher}>Tenho voucher</button>
</div>

<style>
  .light-screen {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 18px;
    padding: 24px;
    background: var(--color-paper);
  }

  .screen-header {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 14px;
    align-items: start;
  }

  .icon-button {
    width: 44px;
    height: 44px;
    display: grid;
    place-items: center;
    border: 1px solid var(--color-line);
    border-radius: 14px;
    background: white;
    color: var(--color-ink);
  }

  .icon-button svg {
    width: 22px;
    height: 22px;
    fill: none;
    stroke: currentColor;
    stroke-linecap: round;
    stroke-linejoin: round;
    stroke-width: 2.4;
  }

  h1,
  h2,
  p {
    margin: 0;
  }

  h1 {
    color: var(--color-ink);
    font-size: 1.65rem;
    font-weight: 900;
    line-height: 1.06;
  }

  .screen-header p {
    margin-top: 6px;
    color: var(--color-muted);
    font-size: 0.95rem;
    line-height: 1.4;
  }

  .plan-list {
    display: grid;
    gap: 12px;
  }

  .loading-state,
  .empty-state {
    border-radius: 18px;
    padding: 22px;
    background: white;
    color: var(--color-muted);
  }

  .empty-state {
    display: grid;
    gap: 10px;
    border: 1px dashed var(--color-line);
  }

  .empty-state h2 {
    color: var(--color-ink);
    font-size: 1.1rem;
    font-weight: 900;
  }

  .empty-state button,
  .text-action {
    min-height: 48px;
    border-radius: 14px;
    font-weight: 850;
  }

  .empty-state button {
    border: 0;
    background: var(--color-ink);
    color: white;
  }

  .text-action {
    margin-top: auto;
    border: 0;
    background: transparent;
    color: color-mix(in srgb, var(--color-primary) 72%, #075985);
  }
</style>

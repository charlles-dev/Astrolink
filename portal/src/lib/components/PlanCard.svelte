<script lang="ts">
  import { formatCurrency, formatDuration } from '../format'
  import type { Plano } from '../types'

  export let plano: Plano
  export let disabled = false
  export let onSelect: (plano: Plano) => void = () => {}

  $: duration = plano.duracao_formatada || formatDuration(plano.duracao_minutos)
  $: speed = `${plano.velocidade_down} Mbps`
</script>

<button
  type="button"
  class="plan-card"
  class:recommended={plano.recomendado}
  disabled={disabled}
  onclick={() => onSelect(plano)}
>
  <span class="plan-card__main">
    <span class="plan-card__title-row">
      <span class="plan-card__name">{plano.nome}</span>
      {#if plano.recomendado}
        <span class="plan-card__badge">RECOMENDADO</span>
      {/if}
    </span>
    {#if plano.descricao}
      <span class="plan-card__description">{plano.descricao}</span>
    {/if}
    <span class="plan-card__meta">
      <span>{duration}</span>
      <span>{speed}</span>
    </span>
  </span>
  <span class="plan-card__price">{formatCurrency(plano.preco)}</span>
</button>

<style>
  .plan-card {
    width: 100%;
    min-height: 116px;
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 16px;
    align-items: center;
    border: 1px solid var(--color-line);
    border-radius: 18px;
    padding: 18px;
    background: var(--color-surface);
    color: var(--color-ink);
    box-shadow: 0 10px 26px rgba(15, 23, 42, 0.08);
    text-align: left;
    transition:
      transform 160ms ease,
      border-color 160ms ease,
      box-shadow 160ms ease;
  }

  .plan-card:hover:not(:disabled),
  .plan-card:focus-visible {
    border-color: var(--color-primary);
    box-shadow: 0 18px 38px rgba(14, 165, 168, 0.18);
    transform: translateY(-1px);
  }

  .plan-card:disabled {
    cursor: wait;
    opacity: 0.68;
  }

  .plan-card.recommended {
    border-color: color-mix(in srgb, var(--color-primary) 70%, white);
  }

  .plan-card__main,
  .plan-card__title-row,
  .plan-card__meta {
    min-width: 0;
    display: flex;
  }

  .plan-card__main {
    flex-direction: column;
    gap: 8px;
  }

  .plan-card__title-row {
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
  }

  .plan-card__name {
    color: var(--color-ink);
    font-size: 1.04rem;
    font-weight: 850;
    line-height: 1.15;
  }

  .plan-card__badge {
    border-radius: 999px;
    padding: 5px 8px;
    background: color-mix(in srgb, var(--color-primary) 14%, white);
    color: color-mix(in srgb, var(--color-primary) 58%, #075985);
    font-size: 0.65rem;
    font-weight: 900;
    line-height: 1;
  }

  .plan-card__description {
    color: var(--color-muted);
    font-size: 0.92rem;
    line-height: 1.35;
  }

  .plan-card__meta {
    flex-wrap: wrap;
    gap: 8px;
    color: #475569;
    font-size: 0.78rem;
    font-weight: 750;
  }

  .plan-card__meta span {
    border-radius: 999px;
    padding: 6px 8px;
    background: #f1f5f9;
  }

  .plan-card__price {
    color: var(--color-ink);
    font-size: 1.2rem;
    font-weight: 900;
    line-height: 1;
    white-space: nowrap;
  }

  @media (max-width: 360px) {
    .plan-card {
      grid-template-columns: 1fr;
    }

    .plan-card__price {
      font-size: 1.1rem;
    }
  }
</style>

<script lang="ts">
  import { formatCountdown } from '../format'
  import type { Settings } from '../types'

  export let settings: Settings
  export let planName = 'Acesso liberado'
  export let secondsRemaining = 0
  export let onNavigate: () => void = () => {}
</script>

<div class="success-screen">
  <div class="success-mark" aria-hidden="true">
    <svg viewBox="0 0 24 24">
      <path d="m5 12 4 4L19 6" />
    </svg>
  </div>

  <section class="success-copy">
    <h1>Acesso liberado</h1>
    <p>{planName}</p>
  </section>

  <section class="session-summary" aria-label="Resumo da sessao">
    <div>
      <span>Tempo restante</span>
      <strong>{formatCountdown(secondsRemaining)}</strong>
    </div>
    <div>
      <span>Rede</span>
      <strong>{settings.hotspot_nome}</strong>
    </div>
  </section>

  <button class="primary-action" type="button" onclick={onNavigate}>Comecar a navegar</button>
</div>

<style>
  .success-screen {
    min-height: 100%;
    flex: 1;
    display: flex;
    flex-direction: column;
    justify-content: center;
    gap: 22px;
    padding: 28px;
    background: var(--color-paper);
    text-align: center;
  }

  .success-mark {
    width: 86px;
    height: 86px;
    display: grid;
    place-items: center;
    align-self: center;
    border-radius: 28px;
    background: color-mix(in srgb, var(--color-success) 16%, white);
    color: #15803d;
  }

  .success-mark svg {
    width: 44px;
    height: 44px;
    fill: none;
    stroke: currentColor;
    stroke-linecap: round;
    stroke-linejoin: round;
    stroke-width: 2.4;
  }

  h1,
  p {
    margin: 0;
  }

  h1 {
    color: var(--color-ink);
    font-size: 2.2rem;
    font-weight: 900;
    line-height: 1;
  }

  p {
    margin-top: 8px;
    color: var(--color-muted);
    font-size: 1rem;
    line-height: 1.45;
  }

  .session-summary {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
    border: 1px solid var(--color-line);
    border-radius: 18px;
    padding: 14px;
    background: white;
    text-align: left;
  }

  .session-summary div {
    min-width: 0;
    display: grid;
    gap: 4px;
  }

  .session-summary span {
    color: var(--color-muted);
    font-size: 0.74rem;
    font-weight: 850;
  }

  .session-summary strong {
    overflow: hidden;
    color: var(--color-ink);
    font-size: 1rem;
    font-weight: 900;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .primary-action {
    width: 100%;
    min-height: 56px;
    border: 0;
    border-radius: 16px;
    background: var(--color-ink);
    color: white;
    font-weight: 900;
  }

  @media (max-width: 360px) {
    .session-summary {
      grid-template-columns: 1fr;
    }
  }
</style>

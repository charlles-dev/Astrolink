<script lang="ts">
  import { formatCountdown, formatCurrency } from '../format'
  import type { PixTransaction, Plano } from '../types'
  import ErrorMessage from './ErrorMessage.svelte'

  export let plano: Plano | null = null
  export let pix: PixTransaction | null = null
  export let secondsRemaining = 0
  export let loading = false
  export let error = ''
  export let copyMessage = ''
  export let onBack: () => void = () => {}
  export let onCopy: () => void = () => {}
</script>

<div class="light-screen pix-screen">
  <header class="screen-header">
    <button class="icon-button" type="button" aria-label="Voltar" onclick={onBack}>
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <path d="m15 18-6-6 6-6" />
      </svg>
    </button>
    <div>
      <h1>Pague com PIX</h1>
      <p>{plano ? plano.nome : 'Gerando cobranca segura.'}</p>
    </div>
  </header>

  <ErrorMessage message={error} actionLabel="Voltar" onAction={onBack} />

  {#if loading || !pix}
    <section class="loading-box" role="status">
      <div class="spinner" aria-hidden="true"></div>
      <strong>Gerando PIX...</strong>
      <span>A cobranca aparece em instantes.</span>
    </section>
  {:else}
    <section class="pix-summary">
      <div>
        <span>Valor</span>
        <strong>{formatCurrency(pix.valor)}</strong>
      </div>
      <div>
        <span>Expira em</span>
        <strong>{formatCountdown(secondsRemaining)}</strong>
      </div>
    </section>

    <section class="qr-frame" aria-label="QRCode PIX">
      {#if pix.qr_code_base64}
        <img src={pix.qr_code_base64} alt="QRCode PIX" />
      {:else}
        <div class="qr-fallback">PIX</div>
      {/if}
    </section>

    <label class="pix-code">
      <span>Codigo copia e cola</span>
      <textarea readonly value={pix.pix_copia_cola}></textarea>
    </label>

    <footer class="pix-actions">
      <button class="primary-action" type="button" onclick={onCopy}>
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <rect x="9" y="9" width="11" height="11" rx="2" />
          <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
        </svg>
        Copiar codigo PIX
      </button>
      {#if copyMessage}
        <p class="copy-message" role="status">{copyMessage}</p>
      {/if}
    </footer>
  {/if}
</div>

<style>
  .light-screen {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 16px;
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

  svg {
    fill: none;
    stroke: currentColor;
    stroke-linecap: round;
    stroke-linejoin: round;
    stroke-width: 2.2;
  }

  .icon-button svg {
    width: 22px;
    height: 22px;
  }

  h1,
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

  .loading-box,
  .pix-summary,
  .qr-frame,
  .pix-code textarea {
    border: 1px solid var(--color-line);
    border-radius: 18px;
    background: white;
  }

  .loading-box {
    min-height: 250px;
    display: grid;
    place-items: center;
    align-content: center;
    gap: 10px;
    color: var(--color-muted);
    text-align: center;
  }

  .loading-box strong {
    color: var(--color-ink);
  }

  .spinner {
    width: 38px;
    height: 38px;
    border: 4px solid #e2e8f0;
    border-top-color: var(--color-primary);
    border-radius: 50%;
    animation: spin 800ms linear infinite;
  }

  .pix-summary {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
    padding: 14px;
  }

  .pix-summary div {
    display: grid;
    gap: 4px;
  }

  .pix-summary span,
  .pix-code span {
    color: var(--color-muted);
    font-size: 0.75rem;
    font-weight: 850;
  }

  .pix-summary strong {
    color: var(--color-ink);
    font-size: 1.18rem;
    font-weight: 900;
  }

  .qr-frame {
    display: grid;
    place-items: center;
    aspect-ratio: 1 / 1;
    padding: 22px;
  }

  .qr-frame img {
    width: 100%;
    height: 100%;
    object-fit: contain;
  }

  .qr-fallback {
    width: 100%;
    height: 100%;
    display: grid;
    place-items: center;
    border: 1px dashed var(--color-line);
    border-radius: 14px;
    color: var(--color-muted);
    font-weight: 900;
  }

  .pix-code {
    display: grid;
    gap: 8px;
  }

  .pix-code textarea {
    width: 100%;
    min-height: 92px;
    resize: none;
    padding: 12px;
    color: var(--color-ink);
    font-size: 0.76rem;
    line-height: 1.35;
  }

  .pix-actions {
    display: grid;
    gap: 8px;
  }

  .primary-action {
    width: 100%;
    min-height: 56px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    border: 0;
    border-radius: 16px;
    background: var(--color-primary);
    color: #082f49;
    font-weight: 900;
  }

  .primary-action svg {
    width: 20px;
    height: 20px;
  }

  .copy-message {
    color: var(--color-muted);
    font-size: 0.86rem;
    text-align: center;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .spinner {
      animation: none;
    }
  }
</style>

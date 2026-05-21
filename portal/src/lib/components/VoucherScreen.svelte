<script lang="ts">
  import { maskVoucherCode } from '../format'
  import ErrorMessage from './ErrorMessage.svelte'

  export let submitting = false
  export let error = ''
  export let onBack: () => void = () => {}
  export let onSubmit: (codigo: string) => void = () => {}

  let codigo = ''

  function handleInput(event: Event) {
    const target = event.currentTarget as HTMLInputElement
    codigo = maskVoucherCode(target.value)
  }

  function handleSubmit(event: SubmitEvent) {
    event.preventDefault()
    if (codigo.length >= 4 && !submitting) {
      onSubmit(codigo)
    }
  }
</script>

<form class="light-screen voucher-screen" onsubmit={handleSubmit}>
  <header class="screen-header">
    <button class="icon-button" type="button" aria-label="Voltar" onclick={onBack}>
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <path d="m15 18-6-6 6-6" />
      </svg>
    </button>
    <div>
      <h1>Inserir voucher</h1>
      <p>Digite o codigo recebido para liberar o acesso.</p>
    </div>
  </header>

  <label class="voucher-field">
    <span>Codigo do voucher</span>
    <input
      autocomplete="one-time-code"
      inputmode="text"
      maxlength="13"
      placeholder="TEST-1234"
      value={codigo}
      oninput={handleInput}
    />
  </label>

  <ErrorMessage message={error} />

  <footer class="form-actions">
    <button class="primary-action" type="submit" disabled={submitting || codigo.length < 4}>
      {submitting ? 'Validando...' : 'Liberar acesso'}
    </button>
  </footer>
</form>

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

  .voucher-field {
    display: grid;
    gap: 9px;
  }

  .voucher-field span {
    color: var(--color-ink);
    font-size: 0.82rem;
    font-weight: 850;
  }

  input {
    width: 100%;
    min-height: 64px;
    border: 1px solid var(--color-line);
    border-radius: 18px;
    padding: 0 18px;
    background: white;
    color: var(--color-ink);
    font-size: 1.28rem;
    font-weight: 900;
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }

  input::placeholder {
    color: #94a3b8;
  }

  .form-actions {
    margin-top: auto;
  }

  .primary-action {
    width: 100%;
    min-height: 56px;
    border: 0;
    border-radius: 16px;
    background: var(--color-primary);
    color: #082f49;
    font-weight: 900;
  }

  .primary-action:disabled {
    cursor: not-allowed;
    opacity: 0.56;
  }
</style>

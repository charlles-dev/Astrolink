<script lang="ts">
  import type { DeviceInfo, Settings } from '../types'
  import ErrorMessage from './ErrorMessage.svelte'

  export let settings: Settings
  export let device: DeviceInfo
  export let loading = false
  export let error = ''
  export let onShowPlans: () => void = () => {}
  export let onShowVoucher: () => void = () => {}
  export let onRetry: () => void = () => {}
</script>

<div class="welcome-screen">
  <header class="welcome-header">
    {#if settings.hotspot_logo_url}
      <img class="brand-logo" src={settings.hotspot_logo_url} alt={settings.hotspot_nome} />
    {:else}
      <div class="brand-logo fallback" aria-hidden="true">
        <svg viewBox="0 0 24 24" role="img">
          <path d="M4 10a12 12 0 0 1 16 0" />
          <path d="M7.5 13.5a7 7 0 0 1 9 0" />
          <path d="M11 17h2" />
        </svg>
      </div>
    {/if}
    <span class="network-state">Walled garden ativo</span>
  </header>

  <section class="welcome-copy">
    <h1>{settings.hotspot_nome}</h1>
    <p>{settings.mensagem_boas_vindas}</p>
  </section>

  <section class="connection-panel" aria-label="Dados da conexao">
    <div>
      <span>Dispositivo</span>
      <strong>{device.mac}</strong>
    </div>
    <div>
      <span>IP atual</span>
      <strong>{device.ip}</strong>
    </div>
  </section>

  <ErrorMessage message={error} actionLabel="Tentar de novo" onAction={onRetry} />

  <footer class="welcome-actions">
    <button class="primary-action" type="button" disabled={loading} onclick={onShowPlans}>
      {loading ? 'Carregando...' : 'Ver planos de acesso'}
    </button>
    <button class="secondary-action" type="button" disabled={loading} onclick={onShowVoucher}>
      Tenho voucher
    </button>
  </footer>
</div>

<style>
  .welcome-screen {
    width: 100%;
    min-height: 100%;
    min-width: 0;
    display: flex;
    flex: 1;
    flex-direction: column;
    justify-content: space-between;
    gap: 24px;
    padding: 28px;
    color: white;
    overflow: hidden;
  }

  .welcome-header,
  .welcome-actions {
    display: flex;
    align-items: center;
  }

  .welcome-header {
    min-width: 0;
    justify-content: space-between;
    gap: 16px;
  }

  .brand-logo {
    width: 58px;
    height: 58px;
    border-radius: 18px;
    object-fit: contain;
    background: rgba(255, 255, 255, 0.1);
  }

  .brand-logo.fallback {
    display: grid;
    place-items: center;
    border: 1px solid rgba(255, 255, 255, 0.18);
  }

  .brand-logo svg {
    width: 30px;
    height: 30px;
    fill: none;
    stroke: var(--color-primary);
    stroke-linecap: round;
    stroke-width: 2.2;
  }

  .network-state {
    max-width: calc(100% - 74px);
    overflow: hidden;
    border: 1px solid rgba(255, 255, 255, 0.18);
    border-radius: 999px;
    padding: 9px 12px;
    color: #cbd5e1;
    font-size: 0.78rem;
    font-weight: 800;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .welcome-copy {
    display: grid;
    gap: 12px;
  }

  h1,
  p {
    margin: 0;
  }

  h1 {
    max-width: 11ch;
    color: white;
    font-size: clamp(2.45rem, 12vw, 4.6rem);
    font-weight: 900;
    letter-spacing: 0;
    line-height: 0.95;
  }

  p {
    max-width: 30ch;
    color: #dbeafe;
    font-size: 1.05rem;
    line-height: 1.45;
  }

  .connection-panel {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
    border: 1px solid rgba(255, 255, 255, 0.16);
    border-radius: 18px;
    padding: 14px;
    background: rgba(15, 23, 42, 0.32);
  }

  .connection-panel div {
    min-width: 0;
    display: grid;
    gap: 5px;
  }

  .connection-panel span {
    color: #94a3b8;
    font-size: 0.72rem;
    font-weight: 800;
  }

  .connection-panel strong {
    overflow: hidden;
    color: white;
    font-size: 0.88rem;
    font-weight: 850;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .welcome-actions {
    flex-direction: column;
    gap: 12px;
  }

  button {
    width: 100%;
    max-width: 100%;
    min-height: 56px;
    border-radius: 16px;
    font-size: 0.98rem;
    font-weight: 850;
  }

  .primary-action {
    border: 0;
    background: var(--color-primary);
    color: #082f49;
    box-shadow: 0 16px 32px rgba(56, 189, 248, 0.22);
  }

  .secondary-action {
    border: 1px solid rgba(255, 255, 255, 0.18);
    background: rgba(255, 255, 255, 0.08);
    color: white;
  }

  button:disabled {
    cursor: wait;
    opacity: 0.65;
  }

  @media (max-width: 360px) {
    .welcome-screen {
      padding: 22px;
    }

    .network-state {
      max-width: calc(100% - 68px);
      padding: 8px 10px;
      font-size: 0.72rem;
    }

    .connection-panel {
      grid-template-columns: 1fr;
    }
  }
</style>

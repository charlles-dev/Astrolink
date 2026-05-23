<script lang="ts">
  import type { AdminRestoreBackupBody } from '../../types'

  export let loading = false
  export let backupMessage = ''
  export let onCreateBackup: () => void = () => {}
  export let onRestoreBackup: (input: AdminRestoreBackupBody) => void = () => {}

  let arquivo = ''
  let confirmacao = ''

  $: restoreReady = arquivo.trim().length > 0 && confirmacao === 'RESTAURAR'

  function submitRestore() {
    if (!restoreReady) return
    onRestoreBackup({
      arquivo: arquivo.trim(),
      confirmacao
    })
  }
</script>

<section class="backup-panel card" aria-labelledby="backup-title">
  <div class="section-heading">
    <div>
      <h2 id="backup-title">Backup</h2>
      <p>Procedimento crítico para snapshot e validação de restore local.</p>
    </div>
  </div>

  {#if backupMessage}
    <p class="backup-message" role="status">{backupMessage}</p>
  {/if}

  <div class="backup-action">
    <div>
      <strong>Snapshot manual</strong>
      <span>Gera um ponto de recuperação antes de manutenção ou alteração sensível.</span>
    </div>
    <button type="button" class="btn btn-primary ink-button" onclick={onCreateBackup} disabled={loading}>
      Gerar backup
    </button>
  </div>

  <form
    class="restore-form"
    onsubmit={(event) => {
      event.preventDefault()
      submitRestore()
    }}
  >
    <div>
      <h3>Restore protegido</h3>
      <p>Etapa de segurança: valida intenção antes de qualquer ação manual no Postgres.</p>
      <p class="restore-rule">Informe o arquivo e digite RESTAURAR exatamente para validar.</p>
    </div>

    <label>
      Arquivo do backup
      <input class="input input-bordered" bind:value={arquivo} name="arquivo" placeholder="backup.sql" autocomplete="off" />
    </label>

    <label>
      Confirmação RESTAURAR
      <input class="input input-bordered" bind:value={confirmacao} name="confirmacao" placeholder="RESTAURAR" autocomplete="off" />
    </label>

    <button type="submit" class="btn btn-error btn-outline danger-button" disabled={loading || !restoreReady}>
      Validar restore protegido
    </button>
  </form>
</section>

<style>
  .backup-panel {
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: var(--admin-panel-padding);
    background: var(--color-surface-raised);
    box-shadow: var(--shadow-panel);
  }

  h2,
  h3,
  p {
    margin: 0;
  }

  .section-heading {
    margin-bottom: 20px;
  }

  h2 {
    font-size: 1.05rem;
    font-weight: 900;
  }

  h3 {
    font-size: 0.92rem;
    font-weight: 900;
  }

  .section-heading p {
    margin-top: 4px;
    color: var(--color-muted);
    font-size: 0.88rem;
  }

  .backup-action {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 14px;
    align-items: center;
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 14px;
    background: var(--color-row);
  }

  .backup-action div {
    display: grid;
    gap: 4px;
    min-width: 0;
  }

  .backup-action strong {
    color: var(--color-ink);
    font-size: 0.9rem;
    font-weight: 900;
  }

  .backup-action span {
    color: var(--color-muted);
    font-size: 0.8rem;
    font-weight: 750;
    line-height: 1.35;
  }

  .backup-message {
    margin-bottom: 16px;
    border: 1px solid var(--state-info-line);
    border-radius: 8px;
    padding: 10px;
    background: var(--state-info-bg);
    color: var(--state-info-text);
    font-size: 0.84rem;
    font-weight: 800;
    line-height: 1.35;
  }

  .restore-form {
    display: grid;
    gap: 14px;
    margin-top: 20px;
    border: 1px solid var(--state-warning-line);
    border-radius: 8px;
    padding: 14px;
    background: var(--color-surface-subtle);
  }

  .restore-form p {
    margin-top: 3px;
    color: var(--color-muted);
    font-size: 0.8rem;
    line-height: 1.35;
  }

  .restore-form .restore-rule {
    color: var(--state-warning-text);
    font-weight: 850;
  }

  label {
    display: grid;
    gap: 5px;
    color: var(--color-muted);
    font-size: 0.76rem;
    font-weight: 850;
  }

  input {
    width: 100%;
    min-height: 38px;
    border-radius: 8px;
    padding: 0 10px;
    font: inherit;
    font-size: 0.84rem;
    box-sizing: border-box;
  }

  .ink-button,
  .danger-button {
    min-height: 42px;
    border-radius: 8px;
    padding: 0 14px;
    font-size: 0.86rem;
    font-weight: 850;
  }

  .danger-button {
    width: 100%;
  }

  .ink-button:disabled,
  .danger-button:disabled {
    cursor: not-allowed;
    opacity: 0.58;
  }

  @media (max-width: 560px) {
    .backup-action {
      grid-template-columns: 1fr;
    }

    .ink-button {
      width: 100%;
    }
  }
</style>

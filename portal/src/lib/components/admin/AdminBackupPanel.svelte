<script lang="ts">
  import type { AdminRestoreBackupBody } from '../../types'

  export let loading = false
  export let backupMessage = ''
  export let onCreateBackup: () => void = () => {}
  export let onRestoreBackup: (input: AdminRestoreBackupBody) => void = () => {}

  let arquivo = ''
  let confirmacao = ''

  function submitRestore() {
    onRestoreBackup({
      arquivo: arquivo.trim(),
      confirmacao
    })
  }
</script>

<section class="backup-panel" aria-labelledby="backup-title">
  <div class="section-heading">
    <div>
      <h2 id="backup-title">Backup</h2>
      <p>Snapshot operacional.</p>
    </div>
  </div>

  {#if backupMessage}
    <p class="backup-message" role="status">{backupMessage}</p>
  {/if}

  <button type="button" class="ink-button" onclick={onCreateBackup} disabled={loading}>
    Gerar backup
  </button>

  <form
    class="restore-form"
    onsubmit={(event) => {
      event.preventDefault()
      submitRestore()
    }}
  >
    <div>
      <h3>Restore protegido</h3>
      <p>Valida o pedido; restore real continua manual/Postgres.</p>
    </div>

    <label>
      Arquivo do backup
      <input bind:value={arquivo} name="arquivo" placeholder="backup.sql" autocomplete="off" />
    </label>

    <label>
      Confirmacao RESTAURAR
      <input bind:value={confirmacao} name="confirmacao" placeholder="RESTAURAR" autocomplete="off" />
    </label>

    <button type="submit" class="danger-button" disabled={loading || !arquivo.trim()}>
      Validar restore protegido
    </button>
  </form>
</section>

<style>
  .backup-panel {
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 18px;
    background: white;
    box-shadow: 0 12px 28px rgba(15, 23, 42, 0.06);
  }

  h2,
  h3,
  p {
    margin: 0;
  }

  .section-heading {
    margin-bottom: 14px;
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

  .backup-message {
    margin-bottom: 12px;
    border: 1px solid #bae6fd;
    border-radius: 8px;
    padding: 10px;
    background: #e0f2fe;
    color: #075985;
    font-size: 0.84rem;
    font-weight: 800;
    line-height: 1.35;
  }

  .restore-form {
    display: grid;
    gap: 10px;
    margin-top: 14px;
    border-top: 1px solid var(--color-line);
    padding-top: 14px;
  }

  .restore-form p {
    margin-top: 3px;
    color: var(--color-muted);
    font-size: 0.8rem;
    line-height: 1.35;
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
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 0 10px;
    color: var(--color-ink);
    font: inherit;
    font-size: 0.84rem;
    box-sizing: border-box;
  }

  .ink-button,
  .danger-button {
    width: 100%;
    min-height: 42px;
    border: 0;
    border-radius: 8px;
    padding: 0 14px;
    font-size: 0.86rem;
    font-weight: 850;
  }

  .ink-button {
    background: var(--color-ink);
    color: white;
  }

  .danger-button {
    border: 1px solid #fecaca;
    background: #fee2e2;
    color: #991b1b;
  }

  .ink-button:disabled,
  .danger-button:disabled {
    cursor: not-allowed;
    opacity: 0.58;
  }
</style>

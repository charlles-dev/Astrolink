<script lang="ts">
  export let loading = false
  export let backupMessage = ''
  export let onCreateBackup: () => void = () => {}
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

  .ink-button {
    width: 100%;
    min-height: 42px;
    border: 0;
    border-radius: 12px;
    padding: 0 14px;
    background: var(--color-ink);
    color: white;
    font-size: 0.86rem;
    font-weight: 850;
  }

  .ink-button:disabled {
    cursor: not-allowed;
    opacity: 0.58;
  }
</style>

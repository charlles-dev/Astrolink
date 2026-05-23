<script lang="ts">
  export let message = ''
  export let actionLabel = ''
  export let onAction: (() => void) | undefined = undefined
</script>

{#if message}
  <div class="error-message" role="alert">
    <div>
      <strong>Algo saiu do trilho</strong>
      <p>{message}</p>
    </div>
    {#if actionLabel && onAction}
      <button type="button" onclick={onAction}>{actionLabel}</button>
    {/if}
  </div>
{/if}

<style>
  .error-message {
    display: flex;
    gap: 12px;
    align-items: flex-start;
    justify-content: space-between;
    padding: 14px;
    border: 1px solid #fecaca;
    border-radius: 14px;
    background: #fef2f2;
    color: #991b1b;
  }

  strong,
  p {
    margin: 0;
  }

  strong {
    display: block;
    font-size: 0.86rem;
    line-height: 1.2;
  }

  p {
    margin-top: 4px;
    color: #b91c1c;
    font-size: 0.92rem;
    line-height: 1.35;
  }

  button {
    min-height: 38px;
    flex: 0 0 auto;
    border: 0;
    border-radius: 12px;
    padding: 0 12px;
    background: #991b1b;
    color: white;
    font-size: 0.82rem;
    font-weight: 800;
  }

  @media (max-width: 390px) {
    .error-message {
      flex-direction: column;
    }

    button {
      width: 100%;
    }
  }
</style>

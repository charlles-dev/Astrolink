<script lang="ts">
  export let step = 'welcome'
  export let dark = false
</script>

<main class="portal-shell" class:dark data-step={step}>
  <section class="portal-device" aria-live="polite">
    <slot />
  </section>
</main>

<style>
  .portal-shell {
    min-height: 100vh;
    display: grid;
    place-items: center;
    padding: 28px;
    background:
      linear-gradient(180deg, rgba(14, 165, 168, 0.08), transparent 42%),
      var(--color-paper);
  }

  .portal-shell.dark {
    background:
      linear-gradient(160deg, color-mix(in srgb, var(--color-bg) 82%, black), var(--color-bg)),
      var(--color-bg);
  }

  .portal-device {
    width: min(100%, 430px);
    min-height: min(820px, calc(100vh - 56px));
    display: flex;
    flex-direction: column;
    overflow: hidden;
    border-radius: 28px;
    background: var(--color-surface);
    box-shadow: 0 28px 70px rgba(15, 23, 42, 0.22);
  }

  .dark .portal-device {
    background: transparent;
    box-shadow: none;
  }

  @media (max-width: 560px) {
    .portal-shell {
      display: block;
      padding: 0;
      background: var(--color-surface);
    }

    .portal-shell.dark {
      background: var(--color-bg);
    }

    .portal-device {
      width: 100%;
      min-height: 100vh;
      border-radius: 0;
      box-shadow: none;
    }
  }
</style>

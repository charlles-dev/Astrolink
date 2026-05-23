<script lang="ts">
  import { browser } from '$app/environment'
  import { onMount } from 'svelte'

  import '../app.css'

  onMount(() => {
    if (!browser || !('serviceWorker' in navigator)) return
    navigator.serviceWorker.register('/service-worker.js').catch(() => {
      // PWA support should never block the captive portal flow.
    })
  })
</script>

<slot />

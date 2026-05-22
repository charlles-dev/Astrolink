<script lang="ts">
  import type { SetupField, SetupStatus } from '../../types'

  export let setupStatus: SetupStatus | null = null
  export let loading = false
  export let setupMessage = ''
  export let onSaveSetup: (values: Record<string, string>) => Promise<void> | void = () => {}

  let values: Record<string, string> = {}
  let lastStatus: SetupStatus | null = null
  let localMessage = ''

  $: if (setupStatus !== lastStatus) {
    values = initialValues(setupStatus)
    lastStatus = setupStatus
    localMessage = ''
  }

  $: groupEntries = Object.entries(setupStatus?.groups ?? {})
  $: dirtyCount = setupStatus && values ? buildPatch().size : 0

  function initialValues(status: SetupStatus | null) {
    const nextValues: Record<string, string> = {}
    Object.values(status?.groups ?? {}).forEach((group) => {
      group.fields.forEach((field) => {
        nextValues[field.key] = field.secret ? '' : field.value ?? ''
      })
    })
    return nextValues
  }

  function originalValue(field: SetupField) {
    return field.secret ? '' : field.value ?? ''
  }

  function fieldPlaceholder(field: SetupField) {
    if (field.secret && field.configured) return 'Configurado'
    if (field.secret) return 'Nao configurado'
    return ''
  }

  function updateValue(key: string, event: Event) {
    values = {
      ...values,
      [key]: (event.currentTarget as HTMLInputElement).value
    }
  }

  function buildPatch() {
    const patch = new Map<string, string>()
    Object.values(setupStatus?.groups ?? {}).forEach((group) => {
      group.fields.forEach((field) => {
        const value = values[field.key] ?? ''
        if (field.secret && value === '') return
        if (value !== originalValue(field)) patch.set(field.key, value)
      })
    })
    return patch
  }

  async function saveSetup() {
    localMessage = ''
    const patch = Object.fromEntries(buildPatch())
    if (!Object.keys(patch).length) {
      localMessage = 'Nenhuma alteracao para salvar'
      return
    }
    try {
      await onSaveSetup(patch)
    } catch {
      // Parent page owns the visible error message.
    }
  }
</script>

<section class="setup-panel" aria-labelledby="setup-title">
  <div class="section-heading">
    <div>
      <h2 id="setup-title">Setup local</h2>
      <p>Variaveis pessoais do ambiente local.</p>
    </div>
    {#if setupStatus?.requires_restart}
      <span class="restart-badge">Reiniciar</span>
    {/if}
  </div>

  {#if setupMessage || localMessage}
    <p class="setup-message" role="status">{setupMessage || localMessage}</p>
  {/if}

  {#if !setupStatus}
    <div class="empty-state">
      <h3>Setup indisponivel</h3>
      <p>Atualize o painel quando o endpoint local estiver ativo.</p>
    </div>
  {:else}
    {#if !setupStatus.writable}
      <p class="warning-message" role="status">
        Arquivo de ambiente somente leitura neste momento.
      </p>
    {/if}

    <form
      class="setup-form"
      onsubmit={(event) => {
        event.preventDefault()
        void saveSetup()
      }}
    >
      {#each groupEntries as [groupKey, group] (groupKey)}
        <fieldset>
          <legend>{group.label}</legend>
          <div class="field-grid">
            {#each group.fields as field (field.key)}
              <div class="field">
                <span>{field.label}</span>
                <input
                  id={field.key}
                  aria-label={field.label}
                  value={values[field.key] ?? ''}
                  oninput={(event) => updateValue(field.key, event)}
                  type={field.secret ? 'password' : 'text'}
                  placeholder={fieldPlaceholder(field)}
                  autocomplete="off"
                  disabled={loading || !setupStatus.writable}
                />
                <small>{field.description}</small>
              </div>
            {/each}
          </div>
        </fieldset>
      {:else}
        <div class="empty-state compact">
          <h3>Nenhum campo publicado</h3>
          <p>O backend ainda nao retornou grupos de configuracao.</p>
        </div>
      {/each}

      <div class="setup-actions">
        <span>{dirtyCount} {dirtyCount === 1 ? 'alteracao' : 'alteracoes'}</span>
        <button
          type="submit"
          class="ink-button"
          disabled={loading || !setupStatus.writable || dirtyCount === 0}
        >
          Salvar setup local
        </button>
      </div>
    </form>
  {/if}
</section>

<style>
  .setup-panel {
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 18px;
    background: white;
    box-shadow: 0 12px 28px rgba(15, 23, 42, 0.06);
  }

  .section-heading,
  .setup-actions {
    display: flex;
    align-items: center;
  }

  h2,
  h3,
  p {
    margin: 0;
  }

  .section-heading {
    justify-content: space-between;
    gap: 14px;
    margin-bottom: 14px;
  }

  h2 {
    font-size: 1.05rem;
    font-weight: 900;
  }

  .section-heading p,
  .field small,
  .setup-actions span,
  .empty-state p {
    color: var(--color-muted);
  }

  .section-heading p {
    margin-top: 4px;
    font-size: 0.88rem;
  }

  .restart-badge {
    border-radius: 999px;
    padding: 6px 9px;
    background: #fef9c3;
    color: #854d0e;
    font-size: 0.7rem;
    font-weight: 900;
    text-transform: uppercase;
  }

  .setup-message,
  .warning-message {
    margin-bottom: 12px;
    border-radius: 8px;
    padding: 10px;
    font-size: 0.84rem;
    font-weight: 800;
    line-height: 1.35;
  }

  .setup-message {
    border: 1px solid #bae6fd;
    background: #e0f2fe;
    color: #075985;
  }

  .warning-message {
    border: 1px solid #fde68a;
    background: #fffbeb;
    color: #92400e;
  }

  .setup-form {
    display: grid;
    gap: 14px;
  }

  fieldset {
    min-width: 0;
    margin: 0;
    border: 0;
    border-top: 1px solid var(--color-line);
    padding: 14px 0 0;
  }

  legend {
    padding: 0 8px 0 0;
    color: var(--color-ink);
    font-size: 0.86rem;
    font-weight: 900;
  }

  .field-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 10px;
    margin-top: 10px;
  }

  .field {
    display: grid;
    gap: 6px;
  }

  .field span {
    color: var(--color-ink);
    font-size: 0.78rem;
    font-weight: 850;
  }

  input {
    width: 100%;
    min-height: 42px;
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 0 11px;
    background: #f8fafc;
    color: var(--color-ink);
    font: inherit;
    font-size: 0.84rem;
    box-sizing: border-box;
  }

  input::placeholder {
    color: #64748b;
  }

  input:disabled {
    cursor: not-allowed;
    opacity: 0.7;
  }

  .field small {
    font-size: 0.75rem;
    line-height: 1.35;
  }

  .setup-actions {
    justify-content: flex-end;
    gap: 10px;
    border-top: 1px solid var(--color-line);
    padding-top: 14px;
  }

  .setup-actions span {
    font-size: 0.78rem;
    font-weight: 850;
  }

  .ink-button {
    min-height: 42px;
    border: 0;
    border-radius: 8px;
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

  .empty-state {
    border: 1px dashed var(--color-line);
    border-radius: 8px;
    padding: 18px;
    background: #f8fafc;
  }

  .empty-state.compact {
    padding: 14px;
  }

  .empty-state h3 {
    font-size: 0.95rem;
    font-weight: 900;
  }

  .empty-state p {
    margin-top: 6px;
    font-size: 0.88rem;
  }

  @media (max-width: 720px) {
    .field-grid {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 520px) {
    .section-heading,
    .setup-actions {
      align-items: stretch;
      flex-direction: column;
    }

    .setup-actions button {
      width: 100%;
    }
  }
</style>

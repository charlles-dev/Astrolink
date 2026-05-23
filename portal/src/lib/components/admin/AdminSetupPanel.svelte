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
    if (field.secret) return 'Não configurado'
    return ''
  }

  function fieldStatusLabel(field: SetupField) {
    return field.configured ? 'Configurado' : 'Não configurado'
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
      localMessage = 'Nenhuma alteração para salvar'
      return
    }
    try {
      await onSaveSetup(patch)
    } catch {
      // Parent page owns the visible error message.
    }
  }
</script>

<section class="setup-panel card" aria-labelledby="setup-title">
  <div class="section-heading">
    <div>
      <h2 id="setup-title">Setup local</h2>
      <p>Configuração segura do ambiente local e integrações sensíveis.</p>
    </div>
    <span class={`setup-state ${setupStatus?.writable ? 'writable' : 'locked'}`}>
      {setupStatus?.writable ? 'Editável' : 'Somente leitura'}
    </span>
  </div>

  {#if setupMessage || localMessage}
    <p class="setup-message" role="status">{setupMessage || localMessage}</p>
  {/if}

  {#if setupStatus?.requires_restart}
    <p class="restart-alert" role="status" aria-label="Reinício necessário">
      Reinicie o serviço para aplicar as alterações de setup local.
    </p>
  {/if}

  {#if !setupStatus}
    <div class="empty-state">
      <h3>Setup indisponível</h3>
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
                <span class="field-label">
                  <span>{field.label}</span>
                  <span class={`badge field-badge ${field.configured ? 'configured' : 'missing'}`}>
                    {fieldStatusLabel(field)}
                  </span>
                </span>
                <input
                  class="input input-bordered"
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
                {#if field.secret && field.configured}
                  <small class="secret-help">Deixe em branco para manter o valor atual.</small>
                {/if}
              </div>
            {/each}
          </div>
        </fieldset>
      {:else}
        <div class="empty-state compact">
          <h3>Nenhum campo publicado</h3>
          <p>O backend ainda não retornou grupos de configuração.</p>
        </div>
      {/each}

      <div class="setup-actions">
        <span>{dirtyCount} {dirtyCount === 1 ? 'alteração' : 'alterações'} pendentes</span>
        <button
          type="submit"
          class="btn btn-primary ink-button"
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
    padding: var(--admin-panel-padding);
    background: var(--color-surface-raised);
    box-shadow: var(--shadow-panel);
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
    gap: 18px;
    margin-bottom: 20px;
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

  .setup-state {
    flex: 0 0 auto;
    border: 1px solid var(--color-line);
    border-radius: 999px;
    padding: 7px 10px;
    background: var(--color-surface-subtle);
    color: var(--color-muted);
    font-size: 0.72rem;
    font-weight: 900;
    text-transform: uppercase;
  }

  .setup-state.writable {
    border-color: var(--state-success-line);
    background: var(--state-success-bg);
    color: var(--state-success-text);
  }

  .setup-state.locked {
    border-color: var(--state-warning-line);
    background: var(--state-warning-bg);
    color: var(--state-warning-text);
  }

  .setup-message,
  .restart-alert,
  .warning-message {
    margin-bottom: 16px;
    border-radius: 8px;
    padding: 10px;
    font-size: 0.84rem;
    font-weight: 800;
    line-height: 1.35;
  }

  .setup-message {
    border: 1px solid var(--state-info-line);
    background: var(--state-info-bg);
    color: var(--state-info-text);
  }

  .warning-message {
    border: 1px solid var(--state-warning-line);
    background: var(--state-warning-bg);
    color: var(--state-warning-text);
  }

  .restart-alert {
    border: 1px solid var(--state-warning-line);
    background: var(--state-warning-bg);
    color: var(--state-warning-text);
  }

  .setup-form {
    display: grid;
    gap: 20px;
  }

  fieldset {
    min-width: 0;
    margin: 0;
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 16px;
    background: var(--color-row);
  }

  legend {
    padding: 0 8px;
    color: var(--color-ink);
    font-size: 0.86rem;
    font-weight: 900;
  }

  .field-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 14px;
    margin-top: 14px;
  }

  .field {
    display: grid;
    gap: 7px;
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 12px;
    background: var(--color-surface-raised);
  }

  .field-label {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }

  .field-label > span:first-child {
    color: var(--color-ink);
    font-size: 0.78rem;
    font-weight: 850;
  }

  .field-badge {
    flex: 0 0 auto;
    font-size: 0.66rem;
    font-weight: 900;
  }

  .field-badge.configured {
    background: var(--state-success-bg);
    color: var(--state-success-text);
  }

  .field-badge.missing {
    background: var(--state-error-bg);
    color: var(--state-error-text);
  }

  input {
    width: 100%;
    min-height: 42px;
    border-radius: 8px;
    padding: 0 11px;
    font: inherit;
    font-size: 0.84rem;
    box-sizing: border-box;
  }

  input::placeholder {
    color: var(--color-muted);
  }

  input:disabled {
    cursor: not-allowed;
    opacity: 0.7;
  }

  .field small {
    font-size: 0.75rem;
    line-height: 1.35;
  }

  .field .secret-help {
    color: var(--color-muted);
    font-weight: 800;
  }

  .setup-actions {
    justify-content: flex-end;
    gap: 12px;
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 12px;
    background: var(--color-surface-subtle);
  }

  .setup-actions span {
    font-size: 0.78rem;
    font-weight: 850;
  }

  .ink-button {
    min-height: 42px;
    border-radius: 8px;
    padding: 0 14px;
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
    padding: 22px;
    background: var(--color-surface-subtle);
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

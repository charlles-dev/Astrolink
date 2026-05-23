<script lang="ts">
  import type { AdminPlanBody, Plano } from '../../types'

  export let plan: Plano | null = null
  export let loading = false
  export let onSubmit: (input: AdminPlanBody) => Promise<void> | void = () => {}
  export let onCancel: () => void = () => {}

  let nome = ''
  let descricao = ''
  let preco = ''
  let duracaoMinutos = ''
  let dadosMb = ''
  let velocidadeDown = ''
  let velocidadeUp = ''
  let ordem = '0'
  let recomendado = false
  let ativo = true
  let visivelPortal = true
  let loadedPlanId: number | null = null
  let errors: Record<string, string> = {}

  $: if ((plan?.id ?? null) !== loadedPlanId) {
    loadedPlanId = plan?.id ?? null
    nome = plan?.nome ?? ''
    descricao = plan?.descricao ?? ''
    preco = plan?.preco ?? ''
    duracaoMinutos = valueOrEmpty(plan?.duracao_minutos)
    dadosMb = valueOrEmpty(plan?.dados_mb)
    velocidadeDown = valueOrEmpty(plan?.velocidade_down)
    velocidadeUp = valueOrEmpty(plan?.velocidade_up)
    ordem = valueOrEmpty(plan?.ordem) || '0'
    recomendado = plan?.recomendado ?? false
    ativo = plan?.ativo ?? true
    visivelPortal = plan?.visivel_portal ?? true
    errors = {}
  }

  function valueOrEmpty(value: number | null | undefined) {
    return value === null || value === undefined ? '' : String(value)
  }

  function parseNumber(value: string | number) {
    const trimmed = String(value).trim()
    if (!trimmed) return { value: null, valid: true }
    const parsed = Number(trimmed)
    return { value: parsed, valid: Number.isFinite(parsed) }
  }

  function optionalNumber(value: string | number, min: number) {
    const parsed = parseNumber(value)
    if (!parsed.valid || (parsed.value !== null && parsed.value < min)) return null
    return parsed.value
  }

  function requiredNumber(value: string, min: number, field: string, message: string) {
    const trimmed = value.trim()
    const parsed = Number(trimmed)
    if (!trimmed || !Number.isFinite(parsed) || parsed < min) {
      errors = { ...errors, [field]: message }
      return null
    }
    return parsed
  }

  function validate() {
    errors = {}

    const parsedPreco = requiredNumber(preco, 0.01, 'preco', 'Informe um preço válido.')
    const parsedDownload = requiredNumber(velocidadeDown, 0, 'velocidadeDown', 'Informe o download em Mbps.')
    const parsedUpload = requiredNumber(velocidadeUp, 0, 'velocidadeUp', 'Informe o upload em Mbps.')
    const parsedDuracao = optionalNumber(duracaoMinutos, 1)
    const parsedDados = optionalNumber(dadosMb, 0)
    const parsedOrdem = optionalNumber(ordem, 0) ?? 0

    if (String(duracaoMinutos).trim() && parsedDuracao === null) {
      errors = { ...errors, duracaoMinutos: 'Informe a duração em minutos.' }
    }
    if (String(dadosMb).trim() && parsedDados === null) {
      errors = { ...errors, dadosMb: 'Informe os dados em MB.' }
    }
    if (String(ordem).trim() && optionalNumber(ordem, 0) === null) {
      errors = { ...errors, ordem: 'Informe a ordem de exibição.' }
    }

    if (Object.keys(errors).length > 0 || parsedPreco === null || parsedDownload === null || parsedUpload === null) {
      return null
    }

    return {
      preco: parsedPreco,
      duracao_minutos: parsedDuracao,
      dados_mb: parsedDados,
      velocidade_down: parsedDownload,
      velocidade_up: parsedUpload,
      ordem: parsedOrdem
    }
  }

  async function submit() {
    const validated = validate()
    if (!validated) return

    await onSubmit({
      nome: nome.trim(),
      descricao: descricao.trim(),
      preco: validated.preco,
      duracao_minutos: validated.duracao_minutos,
      dados_mb: validated.dados_mb,
      velocidade_down: validated.velocidade_down,
      velocidade_up: validated.velocidade_up,
      recomendado,
      ativo,
      visivel_portal: visivelPortal,
      ordem: validated.ordem
    })
  }
</script>

<form class="plan-form" onsubmit={(event) => { event.preventDefault(); void submit() }}>
  <div class="form-mode">
    <div>
      <strong>{plan ? 'Editar plano existente' : 'Novo plano'}</strong>
      <span>{plan ? 'Revise preço, limites e publicação antes de salvar.' : 'Cadastre uma oferta pronta para venda local.'}</span>
    </div>
    <span class="mode-pill">{plan ? 'Edição' : 'Cadastro'}</span>
  </div>

  {#if plan}
    <div class="edit-state alert">
      <strong>Plano em edição</strong>
      <span class="edit-plan-name">{plan.nome}</span>
      <button type="button" class="btn btn-outline btn-sm" onclick={onCancel} disabled={loading}>
        Cancelar edição
      </button>
    </div>
  {/if}

  <fieldset class="form-block">
    <legend>Identificação</legend>
    <div class="form-grid">
      <label class="wide">
        Nome
        <input class="input input-bordered" bind:value={nome} required autocomplete="off" disabled={loading} />
      </label>

      <label class="wide">
        Descrição
        <textarea class="textarea textarea-bordered" bind:value={descricao} rows="3" disabled={loading}></textarea>
      </label>
    </div>
  </fieldset>

  <fieldset class="form-block">
    <legend>Preço e validade</legend>
    <div class="form-grid">
      <label>
        Preço
        <input class="input input-bordered" class:input-error={errors.preco} bind:value={preco} required inputmode="decimal" autocomplete="off" disabled={loading} aria-invalid={Boolean(errors.preco)} />
        {#if errors.preco}<span class="field-error">{errors.preco}</span>{/if}
      </label>

      <label>
        Duração (min)
        <input class="input input-bordered" class:input-error={errors.duracaoMinutos} bind:value={duracaoMinutos} type="number" min="1" step="1" disabled={loading} aria-invalid={Boolean(errors.duracaoMinutos)} />
        {#if errors.duracaoMinutos}<span class="field-error">{errors.duracaoMinutos}</span>{/if}
      </label>
    </div>
  </fieldset>

  <fieldset class="form-block">
    <legend>Limites técnicos</legend>
    <div class="form-grid">
      <label>
        Dados (MB)
        <input class="input input-bordered" class:input-error={errors.dadosMb} bind:value={dadosMb} type="number" min="0" step="1" disabled={loading} aria-invalid={Boolean(errors.dadosMb)} />
        {#if errors.dadosMb}<span class="field-error">{errors.dadosMb}</span>{/if}
      </label>

      <label>
        Download (Mbps)
        <input class="input input-bordered" class:input-error={errors.velocidadeDown} bind:value={velocidadeDown} inputmode="decimal" autocomplete="off" disabled={loading} aria-invalid={Boolean(errors.velocidadeDown)} />
        {#if errors.velocidadeDown}<span class="field-error">{errors.velocidadeDown}</span>{/if}
      </label>

      <label>
        Upload (Mbps)
        <input class="input input-bordered" class:input-error={errors.velocidadeUp} bind:value={velocidadeUp} inputmode="decimal" autocomplete="off" disabled={loading} aria-invalid={Boolean(errors.velocidadeUp)} />
        {#if errors.velocidadeUp}<span class="field-error">{errors.velocidadeUp}</span>{/if}
      </label>
    </div>
  </fieldset>

  <fieldset class="form-block">
    <legend>Exibição</legend>
    <div class="form-grid">
      <label>
        Ordem
        <input class="input input-bordered" class:input-error={errors.ordem} bind:value={ordem} type="number" step="1" disabled={loading} aria-invalid={Boolean(errors.ordem)} />
        {#if errors.ordem}<span class="field-error">{errors.ordem}</span>{/if}
      </label>
    </div>

    <div class="switch-row">
      <label class="switch-option" class:enabled={recomendado}>
        <input
          class="switch-input"
          aria-label="Recomendado"
          bind:checked={recomendado}
          type="checkbox"
          disabled={loading}
        />
        <span class="switch-control" aria-hidden="true"></span>
        <span class="switch-copy">
          <strong>Recomendado</strong>
          <small>Destaque no portal</small>
        </span>
      </label>
      <label class="switch-option" class:enabled={ativo}>
        <input
          class="switch-input"
          aria-label="Ativo"
          bind:checked={ativo}
          type="checkbox"
          disabled={loading}
        />
        <span class="switch-control" aria-hidden="true"></span>
        <span class="switch-copy">
          <strong>Ativo</strong>
          <small>Disponível para uso</small>
        </span>
      </label>
      <label class="switch-option" class:enabled={visivelPortal}>
        <input
          class="switch-input"
          aria-label="Visível no portal"
          bind:checked={visivelPortal}
          type="checkbox"
          disabled={loading}
        />
        <span class="switch-control" aria-hidden="true"></span>
        <span class="switch-copy">
          <strong>Visível no portal</strong>
          <small>Publicado no portal</small>
        </span>
      </label>
    </div>
  </fieldset>

  <div class="form-actions">
    <button type="submit" class="btn btn-primary ink-button" disabled={loading}>
      {plan ? 'Atualizar plano' : 'Salvar plano'}
    </button>
  </div>
</form>

<style>
  .plan-form {
    display: grid;
    gap: 12px;
  }

  .form-mode {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    border-bottom: 1px solid var(--color-line);
    padding-bottom: 12px;
  }

  .form-mode div {
    display: grid;
    gap: 2px;
  }

  .form-mode strong {
    color: var(--color-ink);
    font-size: 0.9rem;
    font-weight: 950;
  }

  .form-mode span {
    color: var(--color-muted);
    font-size: 0.76rem;
    font-weight: 800;
  }

  .mode-pill {
    flex: 0 0 auto;
    border: 1px solid var(--color-line);
    border-radius: 999px;
    padding: 5px 9px;
    background: var(--color-surface-raised);
    text-transform: uppercase;
  }

  .edit-state {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    border: 1px solid var(--state-info-line);
    border-radius: 8px;
    padding: 10px 12px;
    background: var(--state-info-bg);
    color: var(--state-info-text);
  }

  .edit-state strong {
    font-size: 0.9rem;
    font-weight: 900;
  }

  .edit-plan-name {
    font-size: 0.9rem;
    font-weight: 900;
  }

  .form-block {
    display: grid;
    gap: 10px;
    border: 0;
    border-top: 1px solid var(--color-line);
    border-radius: 0;
    margin: 0;
    padding: 12px 0 0;
    background: transparent;
  }

  legend {
    padding: 0 8px 0 0;
    color: var(--color-ink);
    font-size: 0.74rem;
    font-weight: 900;
    letter-spacing: 0.02em;
    text-transform: uppercase;
  }

  .form-grid {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 10px;
  }

  label {
    display: grid;
    gap: 6px;
    color: var(--color-ink);
    font-size: 0.74rem;
    font-weight: 850;
  }

  .wide {
    grid-column: 1 / -1;
  }

  input,
  textarea {
    width: 100%;
    min-width: 0;
    border-radius: 8px;
    padding: 8px 10px;
    font: inherit;
    font-size: 0.82rem;
    font-weight: 650;
  }

  input {
    min-height: 38px;
  }

  textarea {
    resize: vertical;
  }

  .switch-row {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    margin-top: 10px;
    gap: 8px;
  }

  .switch-option {
    position: relative;
    min-width: 0;
    display: grid;
    grid-template-columns: auto minmax(0, 1fr);
    align-items: center;
    border: 1px solid var(--color-line);
    border-radius: 8px;
    gap: 10px;
    padding: 10px;
    background: var(--color-surface-raised);
    cursor: pointer;
    transition:
      border-color 160ms ease,
      background 160ms ease,
      color 160ms ease;
  }

  .switch-option.enabled {
    border-color: color-mix(in srgb, var(--color-primary) 44%, var(--color-line));
    background: color-mix(in srgb, var(--state-success-bg) 36%, var(--color-surface-raised));
  }

  .switch-option:has(.switch-input:focus-visible) {
    outline: 3px solid color-mix(in srgb, var(--color-primary) 28%, transparent);
    outline-offset: 2px;
  }

  .switch-option:has(.switch-input:disabled) {
    cursor: not-allowed;
    opacity: 0.65;
  }

  .switch-input {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    opacity: 0;
  }

  .switch-control {
    width: 34px;
    height: 20px;
    display: inline-flex;
    align-items: center;
    border: 1px solid var(--color-line-strong);
    border-radius: 999px;
    padding: 2px;
    background: var(--color-surface-muted);
    transition:
      border-color 160ms ease,
      background 160ms ease;
  }

  .switch-control::before {
    content: '';
    width: 14px;
    height: 14px;
    border-radius: 999px;
    background: var(--color-muted);
    transition:
      background 160ms ease,
      transform 160ms ease;
  }

  .switch-input:checked + .switch-control {
    border-color: var(--color-primary);
    background: var(--color-primary);
  }

  .switch-input:checked + .switch-control::before {
    background: var(--color-primary-content);
    transform: translateX(14px);
  }

  .switch-copy {
    min-width: 0;
    display: grid;
    gap: 2px;
  }

  .switch-copy strong {
    overflow-wrap: anywhere;
    font-size: 0.78rem;
    font-weight: 900;
    line-height: 1.2;
  }

  .switch-copy small {
    color: var(--color-muted);
    font-size: 0.7rem;
    font-weight: 750;
    line-height: 1.2;
  }

  .form-actions {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    border-top: 1px solid var(--color-line);
    padding-top: 12px;
  }

  button {
    min-height: 40px;
    border-radius: 8px;
    padding: 0 13px;
    font-size: 0.84rem;
    font-weight: 900;
  }

  .ink-button {
    border: 0;
    background: var(--color-ink);
    color: var(--color-surface);
  }

  .field-error {
    color: var(--state-error-text);
    font-size: 0.75rem;
    font-weight: 800;
  }

  button:disabled,
  input:disabled,
  textarea:disabled {
    cursor: not-allowed;
    opacity: 0.65;
  }

  @media (max-width: 520px) {
    .form-mode {
      align-items: stretch;
      flex-direction: column;
    }

    .form-grid {
      grid-template-columns: 1fr;
    }

    .form-actions {
      flex-direction: column-reverse;
    }

    .switch-row {
      grid-template-columns: 1fr;
    }

    .edit-state {
      align-items: stretch;
      flex-direction: column;
    }

    button {
      width: 100%;
    }
  }
</style>

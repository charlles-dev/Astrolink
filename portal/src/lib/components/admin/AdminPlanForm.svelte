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
  }

  function valueOrEmpty(value: number | null | undefined) {
    return value === null || value === undefined ? '' : String(value)
  }

  function numberOrNull(value: string | number) {
    const trimmed = String(value).trim()
    if (!trimmed) return null
    const parsed = Number(trimmed)
    return Number.isFinite(parsed) ? parsed : null
  }

  function numberOrZero(value: string) {
    return numberOrNull(value) ?? 0
  }

  async function submit() {
    await onSubmit({
      nome: nome.trim(),
      descricao: descricao.trim(),
      preco: numberOrZero(preco),
      duracao_minutos: numberOrNull(duracaoMinutos),
      dados_mb: numberOrNull(dadosMb),
      velocidade_down: numberOrZero(velocidadeDown),
      velocidade_up: numberOrZero(velocidadeUp),
      recomendado,
      ativo,
      visivel_portal: visivelPortal,
      ordem: numberOrZero(ordem)
    })
  }
</script>

<form class="plan-form" onsubmit={(event) => { event.preventDefault(); void submit() }}>
  <div class="form-grid">
    <label class="wide">
      Nome
      <input bind:value={nome} required autocomplete="off" disabled={loading} />
    </label>

    <label class="wide">
      Descricao
      <textarea bind:value={descricao} rows="3" disabled={loading}></textarea>
    </label>

    <label>
      Preco
      <input bind:value={preco} required inputmode="decimal" autocomplete="off" disabled={loading} />
    </label>

    <label>
      Duracao (min)
      <input bind:value={duracaoMinutos} type="number" min="1" step="1" disabled={loading} />
    </label>

    <label>
      Dados (MB)
      <input bind:value={dadosMb} type="number" min="0" step="1" disabled={loading} />
    </label>

    <label>
      Download (Mbps)
      <input bind:value={velocidadeDown} type="number" min="0" step="1" disabled={loading} />
    </label>

    <label>
      Upload (Mbps)
      <input bind:value={velocidadeUp} type="number" min="0" step="1" disabled={loading} />
    </label>

    <label>
      Ordem
      <input bind:value={ordem} type="number" step="1" disabled={loading} />
    </label>
  </div>

  <div class="check-row">
    <label>
      <input bind:checked={recomendado} type="checkbox" disabled={loading} />
      Recomendado
    </label>
    <label>
      <input bind:checked={ativo} type="checkbox" disabled={loading} />
      Ativo
    </label>
    <label>
      <input bind:checked={visivelPortal} type="checkbox" disabled={loading} />
      Visivel no portal
    </label>
  </div>

  <div class="form-actions">
    {#if plan}
      <button type="button" class="ghost-button" onclick={onCancel} disabled={loading}>
        Cancelar edicao
      </button>
    {/if}
    <button type="submit" class="ink-button" disabled={loading}>
      {plan ? 'Atualizar plano' : 'Salvar plano'}
    </button>
  </div>
</form>

<style>
  .plan-form {
    display: grid;
    gap: 14px;
  }

  .form-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
  }

  label {
    display: grid;
    gap: 6px;
    color: var(--color-ink);
    font-size: 0.78rem;
    font-weight: 850;
  }

  .wide {
    grid-column: 1 / -1;
  }

  input,
  textarea {
    width: 100%;
    min-width: 0;
    border: 1px solid var(--color-line);
    border-radius: 8px;
    padding: 10px 11px;
    background: #f8fafc;
    color: var(--color-ink);
    font: inherit;
    font-weight: 650;
  }

  textarea {
    resize: vertical;
  }

  input[type='checkbox'] {
    width: 16px;
    height: 16px;
    padding: 0;
  }

  .check-row {
    display: flex;
    flex-wrap: wrap;
    gap: 10px 14px;
  }

  .check-row label {
    display: flex;
    align-items: center;
    gap: 7px;
  }

  .form-actions {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
  }

  button {
    min-height: 40px;
    border-radius: 8px;
    padding: 0 13px;
    font-size: 0.84rem;
    font-weight: 900;
  }

  .ghost-button {
    border: 1px solid var(--color-line);
    background: white;
    color: var(--color-ink);
  }

  .ink-button {
    border: 0;
    background: var(--color-ink);
    color: white;
  }

  button:disabled,
  input:disabled,
  textarea:disabled {
    cursor: not-allowed;
    opacity: 0.65;
  }

  @media (max-width: 520px) {
    .form-grid {
      grid-template-columns: 1fr;
    }

    .form-actions {
      flex-direction: column-reverse;
    }

    button {
      width: 100%;
    }
  }
</style>

<script lang="ts">
  import type {
    AdminBlacklistBody,
    AdminBlacklistEntry,
    AdminRouter,
    AdminRouterBody,
    AdminWalledGardenBody,
    AdminWalledGardenEntry
  } from '../../types'

  export let roteadores: AdminRouter[] = []
  export let blacklist: AdminBlacklistEntry[] = []
  export let walledGarden: AdminWalledGardenEntry[] = []
  export let loading = false
  export let onSaveRouter: (input: AdminRouterBody, id?: number) => Promise<void> | void = () => {}
  export let onDeleteRouter: (id: number) => Promise<void> | void = () => {}
  export let onDiagnoseRouter: (id: number) => Promise<void> | void = () => {}
  export let onSpeedtestRouter: (id: number) => Promise<void> | void = () => {}
  export let onAddBlacklist: (input: AdminBlacklistBody) => Promise<void> | void = () => {}
  export let onDeleteBlacklist: (mac: string) => Promise<void> | void = () => {}
  export let onAddWalledGarden: (input: AdminWalledGardenBody) => Promise<void> | void = () => {}
  export let onDeleteWalledGarden: (id: number) => Promise<void> | void = () => {}

  let editingRouterId: number | null = null
  let routerNome = ''
  let routerIp = ''
  let routerPorta = 22
  let routerUsuario = 'root'
  let routerChave = ''
  let routerAtivo = true
  let blacklistMac = ''
  let blacklistMotivo = 'Bloqueio manual'
  let gardenHost = ''
  let gardenDescricao = ''
  let gardenTipo = 'dominio'

  $: routerFormTitle = editingRouterId ? 'Editar roteador' : 'Novo roteador'
  $: canSaveRouter = routerNome.trim().length > 0 && routerIp.trim().length > 0
  $: canAddBlacklist = blacklistMac.trim().length > 0
  $: canAddGarden = gardenHost.trim().length > 0

  function startRouterEdit(router: AdminRouter) {
    editingRouterId = router.id
    routerNome = router.nome
    routerIp = router.ip
    routerPorta = router.porta_ssh || 22
    routerUsuario = router.usuario_ssh || 'root'
    routerChave = router.chave_ssh_path ?? ''
    routerAtivo = router.ativo
  }

  function resetRouterForm() {
    editingRouterId = null
    routerNome = ''
    routerIp = ''
    routerPorta = 22
    routerUsuario = 'root'
    routerChave = ''
    routerAtivo = true
  }

  async function submitRouter() {
    if (!canSaveRouter) return
    await onSaveRouter(
      {
        nome: routerNome.trim(),
        ip: routerIp.trim(),
        porta_ssh: Number(routerPorta) || 22,
        usuario_ssh: routerUsuario.trim() || 'root',
        chave_ssh_path: routerChave.trim(),
        ativo: routerAtivo
      },
      editingRouterId ?? undefined
    )
    resetRouterForm()
  }

  async function submitBlacklist() {
    if (!canAddBlacklist) return
    await onAddBlacklist({
      mac: blacklistMac.trim(),
      motivo: blacklistMotivo.trim() || 'Bloqueio manual'
    })
    blacklistMac = ''
    blacklistMotivo = 'Bloqueio manual'
  }

  async function submitGarden() {
    if (!canAddGarden) return
    await onAddWalledGarden({
      host: gardenHost.trim(),
      descricao: gardenDescricao.trim(),
      tipo: gardenTipo
    })
    gardenHost = ''
    gardenDescricao = ''
    gardenTipo = 'dominio'
  }
</script>

<section class="network-grid" aria-label="Rede local">
  <article class="network-panel router-panel" aria-labelledby="routers-title">
    <div class="section-heading">
      <div>
        <h2 id="routers-title">Roteadores OpenNDS</h2>
        <p>Cadastro local dos gateways e comandos operacionais.</p>
      </div>
      {#if loading}
        <span class="loading-chip badge badge-soft">Atualizando</span>
      {/if}
    </div>

    <form
      class="router-form"
      onsubmit={(event) => {
        event.preventDefault()
        void submitRouter()
      }}
    >
      <div class="form-title">
        <strong>{routerFormTitle}</strong>
        {#if editingRouterId}
          <button type="button" class="btn btn-ghost inline-button" onclick={resetRouterForm}>
            Cancelar
          </button>
        {/if}
      </div>

      <div class="form-grid">
        <label>
          Nome
          <input class="input input-bordered" bind:value={routerNome} placeholder="Roteador principal" />
        </label>
        <label>
          IP
          <input class="input input-bordered" bind:value={routerIp} placeholder="192.168.1.1" />
        </label>
        <label>
          Porta SSH
          <input class="input input-bordered" type="number" min="1" bind:value={routerPorta} />
        </label>
        <label>
          Usuário SSH
          <input class="input input-bordered" bind:value={routerUsuario} placeholder="root" />
        </label>
        <label class="wide-field">
          Chave SSH
          <input class="input input-bordered" bind:value={routerChave} placeholder="/etc/astrolink/router_key" />
        </label>
        <label class="toggle-field">
          Ativo
          <input class="toggle toggle-primary" type="checkbox" bind:checked={routerAtivo} />
        </label>
      </div>

      <div class="form-actions">
        <button type="submit" class="btn btn-primary" disabled={loading || !canSaveRouter}>
          {editingRouterId ? 'Salvar alterações' : 'Cadastrar roteador'}
        </button>
      </div>
    </form>

    {#if roteadores.length === 0}
      <div class="empty-state">
        <h3>Nenhum roteador cadastrado</h3>
        <p>Cadastre o gateway OpenWrt principal para liberar diagnóstico e controle local.</p>
      </div>
    {:else}
      <div class="network-table" role="table" aria-label="Roteadores cadastrados">
        <div class="network-head" role="row">
          <span role="columnheader">Roteador</span>
          <span role="columnheader">SSH</span>
          <span role="columnheader">Estado</span>
          <span role="columnheader">Ações</span>
        </div>
        {#each roteadores as router (router.id)}
          <div class="network-row router-row" role="row">
            <div class="primary-cell" role="cell">
              <span class="status-dot" class:online={router.status === 'online'}></span>
              <div>
                <h3>{router.nome}</h3>
                <p>{router.ip}</p>
              </div>
            </div>
            <div class="metric-cell" role="cell">
              <span>{router.usuario_ssh}:{router.porta_ssh}</span>
              <small>{router.chave_ssh_path || 'sem chave dedicada'}</small>
            </div>
            <div class="metric-cell" role="cell">
              <span>{router.status}</span>
              <small>{router.usuarios_ativos} usuários ativos</small>
            </div>
            <div class="action-cell compact-actions" role="cell">
              <button type="button" class="btn btn-outline row-button" onclick={() => startRouterEdit(router)}>
                Editar
              </button>
              <button
                type="button"
                class="btn btn-outline row-button"
                disabled={loading}
                onclick={() => onDiagnoseRouter(router.id)}
              >
                Diagnóstico
              </button>
              <button
                type="button"
                class="btn btn-outline row-button"
                disabled={loading}
                onclick={() => onSpeedtestRouter(router.id)}
              >
                Speedtest
              </button>
              <button
                type="button"
                class="btn btn-error btn-outline row-button"
                disabled={loading}
                onclick={() => onDeleteRouter(router.id)}
              >
                Remover
              </button>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </article>

  <div class="side-stack">
    <article class="network-panel" aria-labelledby="blacklist-title">
      <div class="section-heading">
        <div>
          <h2 id="blacklist-title">Blacklist MAC</h2>
          <p>Bloqueios locais aplicados antes da liberação do acesso.</p>
        </div>
      </div>

      <form
        class="side-form"
        onsubmit={(event) => {
          event.preventDefault()
          void submitBlacklist()
        }}
      >
        <label>
          MAC
          <input class="input input-bordered" bind:value={blacklistMac} placeholder="AA:BB:CC:DD:EE:FF" />
        </label>
        <label>
          Motivo
          <input class="input input-bordered" bind:value={blacklistMotivo} />
        </label>
        <button type="submit" class="btn btn-primary" disabled={loading || !canAddBlacklist}>
          Bloquear MAC
        </button>
      </form>

      <div class="simple-list" aria-label="MACs bloqueados">
        {#if blacklist.length === 0}
          <p class="muted-row">Nenhum MAC bloqueado.</p>
        {:else}
          {#each blacklist as entry (entry.mac)}
            <div class="simple-row">
              <div>
                <strong>{entry.mac}</strong>
                <span>{entry.motivo || 'Sem motivo registrado'}</span>
              </div>
              <button
                type="button"
                class="btn btn-outline row-button"
                disabled={loading}
                onclick={() => onDeleteBlacklist(entry.mac)}
              >
                Liberar
              </button>
            </div>
          {/each}
        {/if}
      </div>
    </article>

    <article class="network-panel" aria-labelledby="garden-title">
      <div class="section-heading">
        <div>
          <h2 id="garden-title">Walled garden</h2>
          <p>Hosts liberados antes do login do captive portal.</p>
        </div>
      </div>

      <form
        class="side-form"
        onsubmit={(event) => {
          event.preventDefault()
          void submitGarden()
        }}
      >
        <label>
          Host
          <input class="input input-bordered" bind:value={gardenHost} placeholder="api.exemplo.com" />
        </label>
        <label>
          Tipo
          <select class="select select-bordered" bind:value={gardenTipo}>
            <option value="dominio">Domínio</option>
            <option value="ip">IP</option>
          </select>
        </label>
        <label>
          Descrição
          <input class="input input-bordered" bind:value={gardenDescricao} placeholder="Serviço permitido" />
        </label>
        <button type="submit" class="btn btn-primary" disabled={loading || !canAddGarden}>
          Adicionar host
        </button>
      </form>

      <div class="simple-list" aria-label="Hosts do walled garden">
        {#if walledGarden.length === 0}
          <p class="muted-row">Nenhum host liberado.</p>
        {:else}
          {#each walledGarden as entry (entry.id)}
            <div class="simple-row">
              <div>
                <strong>{entry.host}</strong>
                <span>{entry.descricao || entry.tipo}</span>
              </div>
              <div class="row-tools">
                {#if entry.sistema}
                  <span class="system-badge">Sistema</span>
                {/if}
                <button
                  type="button"
                  class="btn btn-outline row-button"
                  disabled={loading || entry.sistema}
                  onclick={() => onDeleteWalledGarden(entry.id)}
                >
                  Remover
                </button>
              </div>
            </div>
          {/each}
        {/if}
      </div>
    </article>
  </div>
</section>

<style>
  .network-grid {
    display: grid;
    grid-template-columns: minmax(0, 1.25fr) minmax(300px, 0.75fr);
    gap: var(--admin-section-gap);
  }

  .side-stack {
    display: grid;
    gap: var(--admin-section-gap);
  }

  .network-panel {
    border: 1px solid var(--color-line);
    border-radius: var(--admin-panel-radius);
    overflow: hidden;
    background: var(--color-surface-raised);
    box-shadow: var(--shadow-panel);
  }

  .section-heading,
  .form-title,
  .primary-cell,
  .simple-row,
  .row-tools {
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
    gap: 16px;
    border-bottom: 1px solid var(--color-line);
    padding: 15px 16px;
  }

  h2 {
    font-size: 1rem;
    font-weight: 900;
    line-height: 1.2;
  }

  .section-heading p,
  .metric-cell small,
  .primary-cell p,
  .simple-row span,
  .muted-row,
  .network-head {
    color: var(--color-muted);
  }

  .section-heading p {
    margin-top: 3px;
    font-size: 0.8rem;
    line-height: 1.3;
  }

  .loading-chip,
  .system-badge {
    border-radius: 999px;
    padding: 7px 10px;
    background: var(--state-neutral-bg);
    color: var(--state-neutral-text);
    font-size: 0.72rem;
    font-weight: 850;
  }

  .system-badge {
    padding: 5px 8px;
  }

  .router-form,
  .side-form {
    display: grid;
    gap: 14px;
    border-bottom: 1px solid var(--color-line);
    padding: 16px;
    background: var(--color-surface-subtle);
  }

  .form-title {
    justify-content: space-between;
    gap: 12px;
  }

  .form-title strong {
    font-size: 0.9rem;
    font-weight: 900;
  }

  .form-grid {
    display: grid;
    grid-template-columns: minmax(150px, 1fr) minmax(140px, 1fr) 104px;
    gap: 12px;
  }

  label {
    display: grid;
    gap: 6px;
    color: var(--color-ink);
    font-size: 0.76rem;
    font-weight: 850;
  }

  input,
  select {
    min-height: 40px;
    border-radius: 8px;
  }

  .wide-field {
    grid-column: span 2;
  }

  .toggle-field {
    align-content: end;
    justify-items: start;
  }

  .form-actions {
    display: flex;
    justify-content: flex-end;
  }

  .form-actions .btn,
  .side-form .btn {
    min-height: 40px;
    border-radius: 8px;
    font-size: 0.82rem;
    font-weight: 850;
  }

  .inline-button {
    min-height: 32px;
    border-radius: 8px;
    padding-inline: 10px;
    font-size: 0.78rem;
    font-weight: 850;
  }

  .network-table {
    display: grid;
  }

  .network-head,
  .network-row {
    display: grid;
    grid-template-columns: minmax(190px, 1.1fr) minmax(150px, 0.75fr) minmax(130px, 0.7fr) minmax(280px, auto);
    align-items: center;
    column-gap: 14px;
  }

  .network-head {
    border-bottom: 1px solid var(--color-line);
    padding: 9px 16px;
    background: var(--color-surface-subtle);
    font-size: 0.7rem;
    font-weight: 900;
    text-transform: uppercase;
  }

  .network-head span:last-child {
    text-align: right;
  }

  .network-row {
    min-height: 74px;
    border-bottom: 1px solid var(--color-line);
    padding: 12px 16px;
    background: var(--color-row);
  }

  .network-row:last-child {
    border-bottom: 0;
  }

  .primary-cell {
    min-width: 0;
    gap: 10px;
  }

  .primary-cell h3,
  .empty-state h3 {
    overflow-wrap: anywhere;
    font-size: 0.88rem;
    font-weight: 900;
    line-height: 1.25;
  }

  .primary-cell p,
  .metric-cell small,
  .simple-row span {
    margin-top: 2px;
    overflow-wrap: anywhere;
    font-size: 0.74rem;
    font-weight: 750;
    line-height: 1.25;
  }

  .metric-cell {
    min-width: 0;
    display: grid;
    gap: 2px;
  }

  .metric-cell span,
  .simple-row strong {
    overflow-wrap: anywhere;
    font-size: 0.82rem;
    font-weight: 850;
    line-height: 1.25;
  }

  .compact-actions {
    min-width: 0;
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 7px;
  }

  .row-button {
    min-height: 32px;
    border-radius: 6px;
    padding: 0 10px;
    font-size: 0.76rem;
    font-weight: 850;
    white-space: nowrap;
  }

  .status-dot {
    width: 8px;
    height: 8px;
    flex: 0 0 auto;
    border-radius: 999px;
    background: var(--color-muted);
  }

  .status-dot.online {
    background: var(--color-success);
    box-shadow: 0 0 0 4px var(--state-success-bg);
  }

  .simple-list {
    display: grid;
  }

  .simple-row {
    justify-content: space-between;
    gap: 12px;
    border-bottom: 1px solid var(--color-line);
    padding: 12px 16px;
    background: var(--color-row);
  }

  .simple-row:last-child {
    border-bottom: 0;
  }

  .simple-row > div:first-child {
    min-width: 0;
    display: grid;
    gap: 2px;
  }

  .row-tools {
    flex: 0 0 auto;
    gap: 8px;
  }

  .muted-row,
  .empty-state {
    padding: 18px 16px;
    background: var(--color-surface-subtle);
  }

  .muted-row {
    font-size: 0.84rem;
    font-weight: 800;
  }

  .empty-state p {
    margin-top: 6px;
    color: var(--color-muted);
    font-size: 0.84rem;
    line-height: 1.4;
  }

  @media (max-width: 1120px) {
    .network-grid {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 820px) {
    .network-head {
      display: none;
    }

    .network-row {
      grid-template-columns: 1fr;
      row-gap: 10px;
    }

    .compact-actions {
      justify-content: flex-start;
    }
  }

  @media (max-width: 620px) {
    .section-heading,
    .form-title,
    .simple-row {
      align-items: flex-start;
      flex-direction: column;
    }

    .form-grid {
      grid-template-columns: 1fr;
    }

    .wide-field {
      grid-column: auto;
    }

    .form-actions,
    .form-actions .btn,
    .side-form .btn,
    .row-tools,
    .row-button {
      width: 100%;
    }

    .compact-actions {
      display: grid;
      grid-template-columns: 1fr;
    }
  }
</style>

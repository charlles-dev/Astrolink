# Admin Panel Routes Design

## Objetivo

Transformar o painel local em uma area administrativa com paginas reais, sem perder a fluidez do shell visual aprovado. Esta fase combina as tres opcoes aprovadas pelo usuario:

- A: rotas reais para cada area do painel.
- B: estado ativo do shell e fallback fluido para manter a experiencia de abas.
- C: refatoracao do painel em componentes menores e mais legiveis.

## Escopo

Implementar no frontend SvelteKit:

- `/painel` como visao geral.
- `/painel/usuarios` para usuarios conectados.
- `/painel/planos` para planos.
- `/painel/vouchers` para vouchers.
- `/painel/pagamentos` para pagamentos.
- `/painel/setup` para configuracoes locais de `.env`.
- `/painel/logs` para eventos ao vivo, logs e backup.

O backend, contratos da API, autenticacao e endpoints existentes ficam iguais.

## Arquitetura

O painel passa a ter um componente de rota compartilhado, responsavel por login, sessao, carregamento de dados, SSE, mensagens e handlers. Cada rota SvelteKit instancia esse componente com uma pagina ativa.

A apresentacao fica separada em tres camadas:

- `AdminPanelRoute.svelte`: orquestra estado, login, chamadas API e escolhe a pagina ativa.
- `AdminShell.svelte`: renderiza sidebar, header, botoes, mensagem global e container do workspace.
- `AdminPanelContent.svelte`: renderiza somente o conteudo da pagina ativa usando os paineis existentes.

Essa abordagem entrega rotas reais agora sem duplicar toda a logica de dados em cada arquivo de rota. A troca de URL recarrega o componente compartilhado, reaproveita o token salvo em `sessionStorage` e carrega a area correta.

## Navegacao

A sidebar e a barra de atalhos deixam de usar anchors internos e passam a usar links reais:

- Visao geral: `/painel`
- Usuarios: `/painel/usuarios`
- Planos: `/painel/planos`
- Vouchers: `/painel/vouchers`
- Pagamentos: `/painel/pagamentos`
- Setup: `/painel/setup`
- Logs: `/painel/logs`

O item ativo usa `aria-current="page"` e a mesma linguagem visual da fase anterior. Em mobile, a navegacao continua horizontal e sem overflow da pagina.

## Conteudo Por Pagina

### Visao Geral

Mostra `AdminMetrics`, eventos ao vivo e um resumo operacional leve. Nao deve concentrar todos os formularios.

### Usuarios

Mostra `AdminUsersPanel` com a acao de desconectar MACs.

### Planos

Mostra `AdminPlansPanel` com criar, editar e alternar status.

### Vouchers

Mostra `AdminVouchersPanel` com geracao, filtros, desativacao e exportacao.

### Pagamentos

Mostra `AdminPaymentsPanel` com filtros e exportacao.

### Setup

Mostra `AdminSetupPanel`, preservando modo somente leitura quando escrita de `.env` estiver desabilitada.

### Logs

Mostra `AdminLiveEventsPanel`, `AdminLogsPanel` e `AdminBackupPanel`.

## Compatibilidade

- Login simples e login com 2FA continuam iguais.
- `Atualizar` continua carregando o dashboard completo.
- `Sair` limpa sessao e volta para a tela de login.
- Filtros, exportacoes, backup, restore protegido, setup, vouchers e planos continuam chamando os mesmos handlers.
- `AdminDashboard.svelte` pode virar um wrapper de compatibilidade ou ser substituido nos testes, desde que nenhum uso real fique preso ao layout antigo.

## Testes

Atualizar a cobertura para:

- Login ainda chama `api.loginAdmin` com os mesmos campos.
- Login carrega health, planos, usuarios, vouchers, pagamentos, logs e setup.
- `/painel` renderiza visao geral.
- `/painel/usuarios` renderiza `Usuarios conectados`.
- `/painel/planos` renderiza `Planos`.
- `/painel/vouchers` renderiza `Vouchers`.
- `/painel/pagamentos` renderiza `Pagamentos`.
- `/painel/setup` salva setup local.
- `/painel/logs` renderiza eventos/logs/backup.

## Fora De Escopo

- Criar endpoints novos.
- Alterar backend.
- Mudar modelo de auth.
- Implementar cache compartilhado entre rotas.
- Criar admin cloud.

## Verificacao

- `npm test` no portal.
- `npm run check`.
- `npm run build`.
- Browser em `/painel` e pelo menos duas subrotas.
- Browser mobile para garantir que sidebar/top rail e conteudo nao gerem overflow horizontal.

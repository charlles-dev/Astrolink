# Admin Panel App Shell Redesign

## Objetivo

Dar ao painel local uma cara mais profissional de sistema operacional, reduzindo a sensacao de pagina unica empilhada. Esta etapa muda a casca visual e a hierarquia do painel atual. A divisao real em varias rotas fica para a etapa seguinte.

## Direcao Aprovada

Opcao A: app shell operacional.

- Menu lateral fixo com marca Astrolink, navegacao por areas e estado ativo.
- Topo do workspace com titulo, contexto do node, acoes principais e sinalizacao de carregamento.
- Area de abas/atalhos para preparar a futura divisao em paginas.
- Conteudo reorganizado em secoes de leitura mais clara: resumo, operacao, configuracao e observabilidade.
- Aparencia de produto SaaS/admin: mais densa que landing page, menos simplista que cards soltos.

## Escopo Desta Etapa

Implementar dentro do painel existente:

- Redesenhar `AdminDashboard.svelte` como app shell com sidebar e workspace.
- Reorganizar a composicao visual sem mudar contratos de dados nem endpoints.
- Melhorar metric cards, spacing, bordas, tipografia de chrome, botoes e mensagens.
- Manter todos os paineis existentes funcionando: usuarios, planos, setup local, vouchers, pagamentos, eventos, logs e backup.
- Manter responsivo: sidebar vira navegacao horizontal/compacta em telas menores.

Fora de escopo por enquanto:

- Criar rotas separadas de verdade.
- Remover funcionalidades existentes.
- Alterar backend/API.
- Recriar todos os paineis internos em profundidade. Ajustes finos podem acontecer onde forem necessarios para o novo shell.

## Layout

Desktop:

- `admin-shell` ocupa 100vh+ com fundo cinza claro controlado.
- `sidebar` esquerda com largura fixa, marca, links por area e bloco inferior de sessao.
- `workspace` a direita com largura fluida.
- `workspace-header` contem titulo "Painel local", subtitulo operacional, status do node e botoes `Atualizar`/`Sair`.
- `section-tabs` exibe atalhos visuais: Visao geral, Planos, Usuarios, Pagamentos, Setup.
- Conteudo principal usa grid:
  - faixa de metricas no topo;
  - coluna principal para usuarios e pagamentos;
  - coluna secundaria para planos, setup e vouchers;
  - area inferior para eventos, logs e backup.

Mobile/tablet:

- Sidebar deixa de ser uma coluna fixa e vira barra compacta no topo.
- Grids colapsam para uma coluna.
- Botoes mantem altura estavel e textos sem overflow.

## Visual System

- Fundo: cinza frio claro, sem gradientes decorativos grandes.
- Superficies: branco ou quase branco, borda sutil, raio 8px.
- Texto: alto contraste, labels menores em uppercase apenas para metadados.
- Acentos: teal/verde para estado ativo e sucesso; azul para informacao; amarelo apenas para avisos.
- Botoes: primario escuro, secundario branco com borda, sem excesso de arredondamento.
- Densidade: operacional, escaneavel e consistente.

## Interacoes

- Links de sidebar e tabs sao visuais nesta etapa; devem usar `button`/`a` sem navegar ainda ou apontar para anchors internos se isso ajudar.
- `Atualizar`, `Sair` e handlers existentes continuam funcionando.
- Estados de carregamento, mensagens de erro/sucesso e disabled continuam preservados.
- Setup local continua somente leitura quando `ASTROLINK_ALLOW_ENV_WRITE=false`.

## Verificacao

- `npm test` no portal.
- `npm run check`.
- Browser em `/painel` desktop.
- Conferir responsividade em largura mobile.
- Conferir que nenhuma funcao atual desapareceu da tela.

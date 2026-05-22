# Arquitetura de Software

## Estado Atual

Astrolink e um sistema local para vender e liberar acesso Wi-Fi em redes com
OpenWrt/OpenNDS. A implementacao atual tem dois servicos principais:

- Backend local em Go (`node/`), expondo APIs HTTP e falando com OpenNDS.
- Portal cativo em SvelteKit (`portal/`), consumindo a API local.

O painel cloud fica fora desta fase. A prioridade e consolidar o no local:
portal, vouchers, PIX demonstrativo, admin local e integracao OpenNDS.

## Componentes

### Roteador OpenWrt/OpenNDS

O roteador intercepta clientes ainda nao autenticados, injeta `mac`, `ip` e
`token` na URL do portal e controla a liberacao da internet. O backend autoriza
e desconecta clientes usando comandos `ndsctl` via SSH.

### Backend Go (`node/`)

Responsabilidades atuais:

- Servir health check em `GET /api/saude`.
- Expor configuracoes de white-label em `GET /api/settings`.
- Listar planos em `GET /api/planos`.
- Criar cobrancas PIX demonstrativas em `POST /api/pix/gerar`.
- Consultar status de PIX em `GET /api/pix/status/:txid`.
- Expor SSE simples em `GET /api/pix/aguardar/:txid`.
- Resgatar vouchers em `POST /api/voucher/resgatar`.
- Expor endpoints iniciais de admin local.
- Autorizar/desautorizar MACs no OpenNDS quando habilitado.

O backend usa store em memoria quando `DATABASE_URL` nao esta configurado e
Postgres quando `DATABASE_URL` aponta para o banco local.

### Portal SvelteKit (`portal/`)

Responsabilidades atuais:

- Ler `mac`, `ip` e `token` da URL.
- Exibir experiencia visual do portal cativo.
- Carregar settings e planos da API.
- Permitir fluxo de voucher.
- Permitir fluxo PIX demonstrativo.
- Exibir estado de sucesso com tempo de acesso.

### Banco Local Postgres

O schema atual vive em `node/migrations/000001_initial_schema.up.sql`.
As tabelas principais sao:

- `planos`
- `usuarios_mac`
- `transacoes_pix`
- `voucher_lotes`
- `vouchers`
- `voucher_usos`
- `roteadores`
- `blacklist_mac`
- `walled_garden`
- `system_settings`
- `logs`
- `sessoes_admin`

## Fluxo de Voucher

1. Cliente conecta ao Wi-Fi e abre o portal.
2. OpenNDS redireciona para o portal com `mac`, `ip` e `token`.
3. Cliente informa o voucher.
4. Portal chama `POST /api/voucher/resgatar`.
5. Backend valida o voucher e cria/atualiza `usuarios_mac`.
6. Backend chama `ndsctl auth <mac> <duracao>` via SSH quando OpenNDS esta
   habilitado.
7. Portal mostra acesso liberado.

## Fluxo PIX

O fluxo PIX atual e demonstrativo:

1. Cliente escolhe um plano.
2. Portal chama `POST /api/pix/gerar`.
3. Backend gera `txid`, copia-e-cola e QR placeholder.
4. Portal acompanha status por polling/SSE.

A integracao real com Mercado Pago ainda e backlog. O backend ja possui uma
abstracao `internal/payments.Provider` com provider demo para manter o portal
offline por padrao enquanto o Mercado Pago real nao e configurado.

## Expiracao de Sessao

O schema e as APIs ja representam `fim_acesso`, e o pacote `internal/jobs`
expoe o hook `ExpireSessions` para stores que implementem expiracao ativa. A
ativacao recorrente e a chamada de `ndsctl deauth` permanecem como proximo
passo operacional.

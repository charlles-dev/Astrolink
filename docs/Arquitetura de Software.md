# Arquitetura de Software

## Estado Atual

Astrolink e um sistema local para vender e liberar acesso Wi-Fi em redes com
OpenWrt/OpenNDS. A implementacao atual tem dois servicos principais:

- Backend local em Go (`node/`), expondo APIs HTTP e falando com OpenNDS.
- Portal cativo em SvelteKit (`portal/`), consumindo a API local.

O painel cloud fica fora desta fase. A prioridade e consolidar o no local:
portal, vouchers, PIX por provider, admin local e integracao OpenNDS.

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
- Criar cobrancas PIX por provider demo ou Mercado Pago em `POST /api/pix/gerar`.
- Consultar status de PIX em `GET /api/pix/status/:txid`.
- Expor SSE simples em `GET /api/pix/aguardar/:txid`.
- Aprovar PIX local em desenvolvimento por `POST /api/pix/dev/aprovar/:txid`.
- Receber Webhooks Mercado Pago em `POST /api/webhooks/mercadopago` com
  validacao HMAC quando o segredo esta configurado.
- Resgatar vouchers em `POST /api/voucher/resgatar`.
- Expor endpoints iniciais de admin local.
- Expor stream SSE protegido do admin em `GET /admin/eventos`.
- Autorizar/desautorizar MACs no OpenNDS quando habilitado.

O backend usa store em memoria quando `DATABASE_URL` nao esta configurado e
Postgres quando `DATABASE_URL` aponta para o banco local.

### Portal SvelteKit (`portal/`)

Responsabilidades atuais:

- Ler `mac`, `ip` e `token` da URL.
- Exibir experiencia visual do portal cativo.
- Carregar settings e planos da API.
- Permitir fluxo de voucher.
- Permitir fluxo PIX por provider demo ou Mercado Pago.
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

O fluxo PIX continua offline por padrao com `PAYMENTS_PROVIDER=demo`, mas pode
criar cobrancas reais quando Mercado Pago esta configurado:

1. Cliente escolhe um plano.
2. Portal chama `POST /api/pix/gerar`.
3. Backend gera `txid` e chama o provider configurado para obter copia-e-cola e QR.
4. Portal acompanha status por polling/SSE.
5. Em desenvolvimento, `POST /api/pix/dev/aprovar/:txid` simula pagamento
   aprovado sem depender de webhook publico.
6. O endpoint `POST /api/webhooks/mercadopago` valida a assinatura recebida,
   consulta o provider de pagamentos e atualiza a transacao local quando o
   status externo for aprovado.

A integracao HTTP real com Mercado Pago usa `POST /v1/payments`, bearer token,
idempotencia por `txid`, `payer.email` configurado por env e QR retornado por
`point_of_interaction.transaction_data`.

## Operacao Local

O painel local em `/painel` concentra as acoes de operacao do no:

- CRUD de planos.
- Usuarios conectados e desconexao via OpenNDS.
- Vouchers com geracao, filtros, CSV, desativacao e folha impressa.
- Historico de pagamentos com CSV.
- Eventos ao vivo com snapshot operacional.
- Logs operacionais/auditoria com CSV.
- Backup manual quando Postgres esta configurado.
- Validacao protegida de restore, sem executar restore destrutivo pela API.

As acoes mutaveis do admin local registram auditoria em modo best-effort:
falha ao gravar log nao derruba a acao principal. No store em memoria os logs
vivem durante o processo; no Postgres usam a tabela local `logs`.

## Expiracao de Sessao

O schema e as APIs ja representam `fim_acesso`, e o pacote `internal/jobs`
expoe o hook `ExpireSessions` para stores que implementem expiracao ativa. A
ativacao recorrente e a chamada de `ndsctl deauth` permanecem como proximo
passo operacional.

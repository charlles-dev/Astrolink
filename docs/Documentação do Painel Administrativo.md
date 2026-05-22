# Documentacao do Painel Administrativo Local

## Escopo

O admin cloud esta pausado. Esta documentacao trata apenas do admin local do no
Astrolink.

Nesta fase existe uma primeira interface visual do admin local no app SvelteKit:

```text
http://127.0.0.1:5173/painel
```

Ela usa as credenciais configuradas em `ADMIN_USUARIO` e `ADMIN_SENHA`.
As rotas admin, exceto login e refresh, exigem `Authorization: Bearer <access_token>`.

## Endpoints Implementados

### Autenticacao

`POST /admin/auth/login`

```json
{
  "usuario": "admin",
  "senha": "admin123"
}
```

Resposta:

```json
{
  "access_token": "...",
  "refresh_token": "...",
  "expires_in": 28800,
  "token_type": "Bearer"
}
```

O access token e um JWT HS256 com validade de 8 horas. O refresh token e opaco,
armazenado como hash no no local, com validade de 30 dias.

`POST /admin/auth/refresh`

Renova access token e refresh token.

`POST /admin/auth/logout`

Revoga o refresh token informado. Requer Bearer token.

`GET /admin/auth/me`

Retorna o usuario autenticado.

### Saude do Sistema

`GET /admin/sistema/saude`

Retorna status do banco e placeholders de Redis, RabbitMQ, Mercado Pago e
roteador.

### Planos

`GET /admin/planos`

Lista todos os planos cadastrados.

`POST /admin/planos`

Cria plano local.

`PUT /admin/planos/:id`

Atualiza plano local.

`PATCH /admin/planos/:id/status`

Ativa ou desativa plano local.

### Usuarios

`GET /admin/usuarios`

Lista ate 200 usuarios conhecidos pelo no local.

### Desconectar Usuario

`POST /admin/usuarios/:mac/desconectar`

Chama o gateway OpenNDS para executar `ndsctl deauth <mac>` quando habilitado.

### Vouchers

`GET /admin/vouchers`

Lista os vouchers emitidos, com filtros por status, plano, codigo e lote.

`POST /admin/vouchers/gerar`

Gera um lote de vouchers para venda presencial.

```json
{
  "plano_id": 2,
  "quantidade": 10,
  "tipo": "single_use",
  "validade_dias": 30,
  "prefixo": "VIP"
}
```

`PATCH /admin/vouchers/:id/desativar`

Desativa voucher ainda ativo.

`GET /admin/vouchers/export.csv`

Exporta os vouchers filtrados em CSV. A tela tambem oferece impressao de folha
de vouchers a partir da lista atual.

### Pagamentos

`GET /admin/pagamentos`

Lista historico local de cobrancas PIX e totais por status.

`GET /admin/pagamentos/export.csv`

Exporta pagamentos filtrados em CSV.

### Logs

`GET /admin/logs`

Lista eventos operacionais locais e auditoria best-effort das acoes mutaveis do
admin local, como planos, vouchers, backup, restore e desconexao de usuario.

`GET /admin/logs/export.csv`

Exporta logs filtrados em CSV.

### Eventos ao Vivo

`GET /admin/eventos`

Stream SSE protegido usado pela tela de eventos ao vivo. O painel consome essa
rota com `fetch` autenticado por Bearer token e exibe snapshots de usuarios,
vouchers, PIX e logs.

### Backup e Restore

`POST /admin/backup`

Solicita backup manual. No store em memoria retorna `501 backup_indisponivel`.

`POST /admin/backup/restaurar`

Valida `arquivo` e confirmacao literal `RESTAURAR`. O restore destrutivo nao e
executado pela API nesta fase; com confirmacao correta retorna
`501 restore_indisponivel`.

## Tela Implementada

O painel local cobre:

- login com access token e refresh token
- dashboard de saude
- tabela de usuarios
- botao de desconectar usuario
- CRUD de planos
- geracao, filtros, CSV, desativacao e impressao de vouchers
- historico de pagamentos e exportacao CSV
- eventos ao vivo com snapshot operacional
- logs operacionais/auditoria e exportacao CSV
- backup manual e validacao protegida de restore

## Proxima Etapa Recomendada

Evoluir o admin local com:

- provider real do Mercado Pago
- 2FA opcional para o admin local
- template PDF desenhado para vouchers
- agendamento automatico de jobs operacionais

O admin cloud continua fora desta fase.

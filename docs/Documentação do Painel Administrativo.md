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

### Usuarios

`GET /admin/usuarios`

Lista ate 200 usuarios conhecidos pelo no local.

### Desconectar Usuario

`POST /admin/usuarios/:mac/desconectar`

Chama o gateway OpenNDS para executar `ndsctl deauth <mac>` quando habilitado.

### Vouchers

`GET /admin/vouchers`

Lista os vouchers emitidos, com plano, uso atual, validade, status e lote.

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

## Tela Implementada

O painel local inicial cobre:

- login simples
- dashboard de saude
- tabela de usuarios
- botao de desconectar usuario
- lista de planos
- geracao e listagem de vouchers, incluindo validade e tipo universal

## Proxima Etapa Recomendada

Evoluir o admin local com:

- CRUD de planos
- exportacao/impressao de vouchers
- status real do roteador
- logs de auditoria

O admin cloud continua fora desta fase.

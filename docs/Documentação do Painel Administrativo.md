# Documentacao do Painel Administrativo Local

## Escopo

O admin cloud esta pausado. Esta documentacao trata apenas do admin local do no
Astrolink.

Nesta fase existe uma primeira interface visual do admin local no app SvelteKit:

```text
http://127.0.0.1:5173/painel
```

Ela usa as credenciais configuradas em `ADMIN_USUARIO` e `ADMIN_SENHA`.

## Endpoints Implementados

### Login

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

O token atual e simples e temporario. JWT real ainda e backlog.

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

## Tela Implementada

O painel local inicial cobre:

- login simples
- dashboard de saude
- tabela de usuarios
- botao de desconectar usuario
- lista de planos

## Proxima Etapa Recomendada

Evoluir o admin local com:

- geracao/listagem de vouchers
- CRUD de planos
- status real do roteador
- autenticacao JWT real
- logs de auditoria

O admin cloud continua fora desta fase.

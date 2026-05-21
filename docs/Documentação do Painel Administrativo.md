# Documentacao do Painel Administrativo Local

## Escopo

O admin cloud esta pausado. Esta documentacao trata apenas do admin local do no
Astrolink.

Nesta fase existe API inicial de admin no backend Go, mas ainda nao existe app
visual dedicada para o painel local.

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

## Proxima Etapa Recomendada

Implementar o painel local visual, provavelmente em uma nova pasta `admin/`,
com estes blocos:

- login simples
- dashboard de saude
- tabela de usuarios
- botao de desconectar usuario
- lista de planos
- geracao/listagem de vouchers

O admin cloud continua fora desta fase.

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
  "senha": "admin123",
  "totp_codigo": "123456"
}
```

`totp_codigo` e enviado apenas quando o 2FA opcional esta habilitado.

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

O login aplica bloqueio local por usuario/IP depois de 5 falhas em 15 minutos.
Enquanto bloqueado, retorna `429 login_bloqueado` e nao cria sessao.
Quando `ADMIN_TOTP_SECRET` esta configurado, o backend exige `totp_codigo`;
sem codigo retorna `428 totp_obrigatorio`, e codigo invalido conta como falha
de login.

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

### Setup Local

`GET /admin/setup/status`

Retorna o status redigido da configuracao local usada pelo assistente de setup.
A rota exige Bearer token. Campos secretos, como tokens Mercado Pago, segredo de
webhook, senha admin, TOTP e chave SSH, nunca retornam em texto; eles aparecem
somente como `configured: true` ou `configured: false`.

`PUT /admin/setup/env`

Atualiza chaves permitidas do `.env` local. A rota exige Bearer token e so grava
quando `ASTROLINK_ALLOW_ENV_WRITE=true`; por padrao a escrita fica desabilitada.
O arquivo alvo tem default `.env`; para usar outro arquivo, defina
`ASTROLINK_ENV_FILE` no processo antes de iniciar o node. O backend le esse
arquivo no startup, preservando prioridade para variaveis ja definidas no
processo.

Use o painel como alternativa local controlada. O fluxo recomendado continua
sendo executar o CLI dentro de `node/`:

```powershell
go run ./cmd/setup
```

Toda alteracao feita pelo painel ou CLI exige reiniciar o node para valer. O
admin cloud continua fora de escopo nesta fase.

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
PDF/impressao de vouchers a partir da lista atual, com resumo de lote, plano,
usos, validade e instrucoes para recorte.

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
- campo 2FA sob demanda quando o backend exige TOTP
- dashboard de saude
- tabela de usuarios
- botao de desconectar usuario
- CRUD de planos
- geracao, filtros, CSV, desativacao e impressao de vouchers
- folha PDF/impressao de vouchers com tickets de recorte
- historico de pagamentos e exportacao CSV
- eventos ao vivo com snapshot operacional
- logs operacionais/auditoria e exportacao CSV
- backup manual e validacao protegida de restore
- status de setup local e escrita opcional do `.env` quando liberada por env

## Proxima Etapa Recomendada

Evoluir o admin local com:

- agendamento automatico de jobs operacionais

O admin cloud continua fora desta fase.

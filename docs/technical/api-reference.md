# Referencia da API

Base local:

```text
http://localhost:5000
```

Formato padrao: JSON.

## Health

### `GET /api/saude`

Resposta:

```json
{
  "status": "healthy",
  "versao": "0.1.0",
  "node": "dev-node-01",
  "uptime_segundos": 0,
  "database": "memory"
}
```

`database` pode ser `memory`, `ok` ou `error`.

## Portal

### `GET /api/settings`

Retorna configuracoes de white-label usadas pelo portal.

Resposta:

```json
{
  "hotspot_nome": "Astrolink Wi-Fi",
  "hotspot_logo_url": "",
  "cor_primaria": "#06B6D4",
  "cor_secundaria": "#0E7490",
  "cor_fundo": "#0F172A",
  "mensagem_boas_vindas": "Bem-vindo! Conecte-se e aproveite.",
  "url_pos_conexao": "https://google.com",
  "coleta_nome": false,
  "mostrar_velocidade": true
}
```

### `GET /api/planos`

Lista planos ativos e visiveis no portal.

Resposta:

```json
{
  "planos": [
    {
      "id": 2,
      "nome": "Acesso 24 Horas",
      "descricao": "Um dia completo de internet.",
      "preco": 15,
      "preco_formatado": "15.00",
      "duracao_minutos": 1440,
      "duracao_formatada": "24 horas",
      "velocidade_down": 10,
      "velocidade_up": 5,
      "recomendado": true,
      "ativo": true,
      "visivel_portal": true,
      "ordem": 1
    }
  ]
}
```

### `GET /api/sessao/status?mac=AA:BB:CC:DD:EE:FF`

Retorna se o dispositivo tem sessao ativa.

Sem sessao:

```json
{
  "ativa": false
}
```

Com sessao:

```json
{
  "ativa": true,
  "plano": "Acesso 24 Horas",
  "fim_acesso": "2026-05-21T15:00:00Z",
  "tempo_restante_segundos": 3600,
  "dados_consumidos_mb": 0
}
```

### `POST /api/pix/gerar`

Cria cobranca PIX. Usa o provider demo por padrao; com
`PAYMENTS_PROVIDER=mercadopago`, `MERCADOPAGO_ACCESS_TOKEN` e
`MERCADOPAGO_PAYER_EMAIL` configurados, cria a cobranca pela API de pagamentos
do Mercado Pago.

Body:

```json
{
  "plano_id": 2,
  "mac": "AA:BB:CC:DD:EE:FF",
  "ip": "192.168.1.50",
  "nome": "Cliente"
}
```

Resposta `201`:

```json
{
  "txid": "ast_123",
  "valor": "15.00",
  "descricao": "Astrolink Wi-Fi - Acesso 24 Horas",
  "pix_copia_cola": "000201...",
  "qr_code_base64": "data:image/svg+xml;base64,...",
  "expira_em": "2026-05-21T15:15:00Z",
  "expira_em_segundos": 900
}
```

### `GET /api/pix/status/:txid`

Retorna status da cobranca PIX.

Resposta:

```json
{
  "txid": "ast_123",
  "status": "pendente",
  "expira_em": "2026-05-21T15:15:00Z"
}
```

### `POST /api/pix/dev/aprovar/:txid`

Aprova uma cobranca PIX localmente apenas quando `GO_ENV=development`. Em
producao, a rota retorna `404`.

Resposta:

```json
{
  "txid": "ast_123",
  "status": "aprovado"
}
```

### `POST /api/webhooks/mercadopago`

Recebe notificacoes Webhook do Mercado Pago. Quando
`MERCADOPAGO_WEBHOOK_SECRET` esta configurado, valida `x-signature` com
`x-request-id`, consulta o provider de pagamentos e atualiza a transacao local
somente quando o provider retorna status `aprovado`.

Headers esperados:

```text
x-request-id: bb56a2f1-6aae-46ac-982e-9dcd3581d08e
x-signature: ts=1742505638683,v1=<hmac>
```

Body minimo:

```json
{
  "data": {
    "id": "123456"
  },
  "external_reference": "ast_123"
}
```

Em desenvolvimento sem segredo configurado, retorna `202` com status
`ignored` e nao altera transacoes locais.

### `GET /api/pix/aguardar/:txid`

Stream SSE simples para status do PIX.

Eventos:

```text
event: status
data: {"status":"pendente","txid":"ast_123"}
```

### `POST /api/voucher/resgatar`

Body:

```json
{
  "codigo": "TEST-1234",
  "mac": "AA:BB:CC:DD:EE:FF",
  "ip": "192.168.1.50"
}
```

Resposta:

```json
{
  "sucesso": true,
  "plano": "Acesso 24 Horas",
  "tempo_adicionado_minutos": 1440,
  "fim_acesso": "2026-05-22T15:00:00Z",
  "tempo_restante_segundos": 86400,
  "acesso_anterior": false,
  "roteador_autorizado": true
}
```

Erros comuns:

| Status | `erro` | Caso |
|---|---|---|
| 400 | `validacao_falhou` | JSON invalido |
| 404 | `nao_encontrado` | voucher nao encontrado |
| 410 | `recurso_esgotado` | voucher ja utilizado |
| 422 | `regra_negocio` | voucher expirado/inativo |
| 500 | `erro_interno` | falha inesperada |

## Admin Local

Todas as rotas abaixo exigem `Authorization: Bearer <access_token>`, exceto
`POST /admin/auth/login` e `POST /admin/auth/refresh`.

### `POST /admin/auth/login`

Body:

```json
{
  "usuario": "admin",
  "senha": "admin123",
  "totp_codigo": "123456"
}
```

`totp_codigo` e opcional e so precisa ser enviado quando `ADMIN_TOTP_SECRET`
esta configurado.

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
armazenado como hash, com validade de 30 dias.

Depois de 5 falhas de senha recentes para o mesmo usuario/IP em uma janela de
15 minutos, a rota retorna `429 login_bloqueado` ate a janela expirar.
Quando 2FA esta habilitado e o codigo nao foi enviado, retorna
`428 totp_obrigatorio`; codigo invalido retorna `401 nao_autenticado` e tambem
conta como falha de login.

### `POST /admin/auth/refresh`

Body:

```json
{
  "refresh_token": "..."
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

### `POST /admin/auth/logout`

Body:

```json
{
  "refresh_token": "..."
}
```

Resposta: `204 No Content`.

### `GET /admin/auth/me`

Resposta:

```json
{
  "usuario": "admin"
}
```

### `GET /admin/sistema/saude`

Retorna health detalhado do no local.

### `GET /admin/setup/status`

Retorna o estado seguro da configuracao local. Exige Bearer token.

Resposta:

```json
{
  "requires_restart": false,
  "writable": false,
  "groups": {
    "payments": {
      "label": "Mercado Pago",
      "fields": [
        {
          "key": "MERCADOPAGO_ACCESS_TOKEN",
          "label": "Access token",
          "description": "Token privado da conta Mercado Pago.",
          "secret": true,
          "configured": true
        },
        {
          "key": "MERCADOPAGO_PAYER_EMAIL",
          "label": "E-mail pagador",
          "description": "E-mail padrao para simulacoes locais.",
          "secret": false,
          "configured": true,
          "value": "cliente@example.com"
        }
      ]
    }
  }
}
```

Campos marcados como `secret: true` nunca retornam `value`; a API informa apenas
se estao configurados. O arquivo consultado tem default `.env`; para usar outro
arquivo, defina `ASTROLINK_ENV_FILE` no processo antes de iniciar o node. O mesmo
arquivo tambem e lido pelo backend no startup.

### `PUT /admin/setup/env`

Atualiza chaves permitidas do `.env` local. Exige Bearer token e
`ASTROLINK_ALLOW_ENV_WRITE=true`; com o default `false`, retorna erro de escrita
desabilitada.

Body:

```json
{
  "values": {
    "PAYMENTS_PROVIDER": "mercadopago",
    "MERCADOPAGO_ACCESS_TOKEN": "APP_USR-...",
    "MERCADOPAGO_PAYER_EMAIL": "cliente@example.com",
    "OPENNDS_ENABLED": "false"
  }
}
```

Resposta: mesmo formato de `GET /admin/setup/status`, com
`requires_restart: true` quando alguma variavel foi gravada. Exemplo abreviado:

```json
{
  "requires_restart": true,
  "writable": true,
  "groups": {
    "payments": {
      "label": "Mercado Pago",
      "fields": []
    }
  }
}
```

A rota nao e um editor generico de `.env`: somente chaves de setup local devem
ser aceitas. Alteracoes feitas por essa rota ou pelo CLI `go run ./cmd/setup`
exigem reiniciar o node para entrar em vigor.

### `GET /admin/planos`

Lista todos os planos.

### `GET /admin/usuarios`

Lista usuarios conhecidos.

### `GET /admin/pagamentos`

Lista historico local de transacoes PIX.

Query params opcionais:

- `status`: `pendente`, `aprovado`, `cancelado`, `expirado` ou `todos`.
- `inicio`: data `YYYY-MM-DD` ou timestamp RFC3339.
- `fim`: data `YYYY-MM-DD` ou timestamp RFC3339.

Resposta:

```json
{
  "total": 1,
  "totais": {
    "pendente": 1,
    "aprovado": 0,
    "cancelado": 0,
    "expirado": 0,
    "valor_total": "15.00"
  },
  "pagamentos": [
    {
      "txid": "ast_123",
      "status": "pendente",
      "valor": "15.00",
      "descricao": "Astrolink Wi-Fi - Acesso 24 Horas",
      "mac": "AA:BB:CC:DD:EE:FF",
      "plano_id": 2,
      "plano": { "id": 2, "nome": "Acesso 24 Horas" },
      "created_at": "2026-05-21T21:24:35Z",
      "expira_em": "2026-05-21T21:39:35Z"
    }
  ]
}
```

### `GET /admin/pagamentos/export.csv`

Exporta pagamentos em CSV usando os mesmos filtros de `/admin/pagamentos`.

### `GET /admin/vouchers`

Lista vouchers emitidos no no local.

Resposta:

```json
{
  "total": 2,
  "vouchers": [
    {
      "id": 2,
      "codigo": "UNIV-0000",
      "plano": { "id": 1, "nome": "Acesso 1 Hora" },
      "tipo": "universal",
      "usos_maximos": 25,
      "usos_atuais": 0,
      "validade_em": null,
      "ativo": true
    }
  ]
}
```

### `POST /admin/vouchers/gerar`

Gera um lote de vouchers.

Body:

```json
{
  "plano_id": 2,
  "quantidade": 10,
  "tipo": "single_use",
  "validade_dias": 30,
  "prefixo": "VIP"
}
```

Resposta `201`:

```json
{
  "lote_id": 1,
  "quantidade": 10,
  "vouchers": [
    {
      "id": 3,
      "codigo": "VIP-1234",
      "plano": { "id": 2, "nome": "Acesso 24 Horas" },
      "tipo": "single_use",
      "usos_atuais": 0,
      "ativo": true,
      "prefixo": "VIP",
      "lote_id": 1
    }
  ]
}
```

### `POST /admin/usuarios/:mac/desconectar`

Desconecta o MAC no OpenNDS quando gateway real esta habilitado.

Resposta:

```json
{
  "sucesso": true
}
```

### `GET /admin/logs`

Lista registros operacionais locais. Quando nao ha store persistente de logs,
retorna eventos de estado do ambiente local/dev.

Query params opcionais:

- `nivel`: `info`, `aviso` ou `erro`.
- `tipo`: categoria do evento.
- `texto`: busca textual simples.

Resposta:

```json
{
  "total": 1,
  "logs": [
    {
      "timestamp": "2026-05-21T21:25:00Z",
      "nivel": "info",
      "tipo": "sistema",
      "mensagem": "ambiente local/dev ativo sem log persistente configurado"
    }
  ]
}
```

### `GET /admin/logs/export.csv`

Exporta logs em CSV usando os mesmos filtros de `/admin/logs`.

### `GET /admin/eventos`

Stream SSE protegido para o painel local. Exige Bearer token e emite snapshots
operacionais periodicos.

Evento:

```text
event: snapshot
data: {"timestamp":"2026-05-22T13:00:00Z","database":"memory","usuarios_total":1,"usuarios_ativos":0,"vouchers_total":2,"vouchers_ativos":2,"pagamentos_total":1,"pagamentos_pendentes":1,"pagamentos_aprovados":0,"logs_total":3}
```

Para smoke tests, `GET /admin/eventos?once=1` emite um snapshot e encerra a
resposta.

### `POST /admin/backup`

Solicita backup manual. No store em memoria retorna `501 backup_indisponivel`,
porque backup manual depende de Postgres configurado.

### `POST /admin/backup/restaurar`

Valida uma solicitacao de restore com confirmacao explicita. Por seguranca,
nenhum restore destrutivo e executado pela API nesta fase; mesmo com confirmacao
correta a rota retorna `501 restore_indisponivel`.

Body:

```json
{
  "arquivo": "backup.sql",
  "confirmacao": "RESTAURAR"
}
```

Erros esperados:

| Status | `erro` | Caso |
|---|---|---|
| 400 | `confirmacao_invalida` | arquivo vazio ou confirmacao diferente de `RESTAURAR` |
| 501 | `restore_indisponivel` | pedido validado, mas restore real permanece manual/Postgres |

## Backlog da API

- Ampliar cobertura de auditoria para fluxos futuros.
- Agendamento automatico de jobs operacionais.

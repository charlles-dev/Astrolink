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

Cria cobranca PIX demonstrativa.

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
armazenado como hash, com validade de 30 dias.

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

### `GET /admin/planos`

Lista todos os planos.

### `GET /admin/usuarios`

Lista usuarios conhecidos.

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

## Backlog da API

- Logs de auditoria para acoes admin.
- CRUD de planos.
- Exportacao e impressao de vouchers.
- Webhook real do Mercado Pago.
- Relatorios de pagamento.
- Backup/restore.
- WebSocket ou SSE para admin local.
- Jobs de expiracao e sincronizacao futura.

# Referência da API — Backend Go (Nó Local)

## Informações Gerais

**Base URL (local):** `http://[ip-do-no]:5000`
**Formato:** JSON
**Autenticação:** Bearer JWT (rotas `/admin/*`) ou nenhuma (rotas públicas do portal)
**Versão da API:** v1 (prefixo `/api/v1` planejado para futura versão)

---

## Autenticação

### `POST /admin/auth/login`

```http
POST /admin/auth/login
Content-Type: application/json

{
  "usuario": "admin",
  "senha": "minha_senha_segura",
  "totp_code": "123456"   // opcional, se 2FA ativado
}
```

**Resposta 200:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 28800,
  "token_type": "Bearer"
}
```

**Erros:**
- `401` Credenciais inválidas
- `429` Muitas tentativas (bloqueado por 10 min)

---

### `POST /admin/auth/refresh`

```http
POST /admin/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJ..."
}
```

**Resposta 200:** Mesmo formato do login.

---

### `POST /admin/auth/logout`

```http
POST /admin/auth/logout
Authorization: Bearer eyJ...
```

Invalida o refresh token atual. **Resposta 204.**

---

## Portal Cativo (público)

### `GET /api/settings`

Retorna configurações de white-label para o portal.

**Cache:** 5 minutos (header `Cache-Control: public, max-age=300`)

**Resposta 200:**
```json
{
  "hotspot_nome": "Wi-Fi Pousada Recanto Verde",
  "hotspot_logo_url": "/uploads/logo.png",
  "cor_primaria": "#2ECC71",
  "cor_fundo": "#0D1117",
  "mensagem_boas_vindas": "Bem-vindo!",
  "url_pos_conexao": "https://meusite.com.br",
  "coleta_nome": false
}
```

---

### `GET /api/planos`

Lista planos disponíveis para o portal.

**Resposta 200:**
```json
{
  "planos": [
    {
      "id": 1,
      "nome": "Acesso 24 Horas",
      "descricao": "Um dia completo de internet",
      "preco": "15.00",
      "duracao_minutos": 1440,
      "duracao_formatada": "24 horas",
      "dados_mb": null,
      "recomendado": true,
      "velocidade_down": 10,
      "velocidade_up": 5
    },
    {
      "id": 2,
      "nome": "Acesso 1 Hora",
      "preco": "5.00",
      "duracao_minutos": 60,
      "duracao_formatada": "1 hora",
      "recomendado": false,
      "velocidade_down": 5,
      "velocidade_up": 2
    }
  ]
}
```

---

### `GET /api/sessao/status`

Verifica se o dispositivo já tem uma sessão ativa.

**Query params:** `mac=AA:BB:CC:DD:EE:FF`

**Resposta 200 — com sessão:**
```json
{
  "ativa": true,
  "plano": "Acesso 24 Horas",
  "fim_acesso": "2025-05-20T14:30:00Z",
  "tempo_restante_segundos": 66600,
  "dados_consumidos_mb": 1240
}
```

**Resposta 200 — sem sessão:**
```json
{
  "ativa": false
}
```

---

### `POST /api/pix/gerar`

Inicia o fluxo de pagamento PIX.

**Body:**
```json
{
  "plano_id": 1,
  "mac": "AA:BB:CC:DD:EE:FF",
  "ip": "192.168.1.50",
  "nome": "João Silva"   // opcional
}
```

**Resposta 201:**
```json
{
  "txid": "ast_xxxxxxxxxxxxxxxxxxx",
  "valor": "15.00",
  "descricao": "Astrolink Wi-Fi — Acesso 24 Horas",
  "pix_copia_cola": "00020126580014br.gov.bcb.pix...",
  "qr_code_base64": "data:image/png;base64,iVBORw0KGgo...",
  "expira_em": "2025-05-19T15:02:00Z",
  "expira_em_segundos": 900
}
```

**Erros:**
- `400` MAC inválido ou plano não encontrado
- `402` Usuário está na blacklist
- `429` Rate limit: máximo 3 cobranças pendentes por MAC

---

### `GET /api/pix/aguardar/:txid`

**Server-Sent Events** para aguardar confirmação de pagamento em tempo real.

```http
GET /api/pix/aguardar/ast_xxxx
Accept: text/event-stream
```

**Eventos SSE:**
```
event: heartbeat
data: {"timestamp": "2025-05-19T14:32:00Z"}

event: status
data: {"status": "pendente", "txid": "ast_xxxx"}

event: aprovado
data: {"status": "aprovado", "fim_acesso": "2025-05-20T14:32:00Z"}

event: expirado
data: {"status": "expirado"}
```

Conexão fechada pelo servidor após aprovação ou expiração.

**Fallback (polling):** `GET /api/pix/status/:txid` retorna status atual sem SSE.

---

### `POST /api/webhooks/mercadopago`

Endpoint interno chamado pelo Mercado Pago ao confirmar pagamento.

**Autenticação:** Verificação de assinatura HMAC com secret configurado.

```json
{
  "action": "payment.updated",
  "data": { "id": "1234567890" }
}
```

**Fluxo interno após confirmação:**
1. Consultar status do pagamento na API do MP
2. Atualizar `transacoes_pix.status = 'aprovado'`
3. Criar/atualizar `usuarios_mac` com status `ativo` e `fim_acesso`
4. Executar `ndsctl auth <mac> <duration>` via SSH
5. Aplicar regras de velocidade (`tc qdisc`)
6. Publicar evento no RabbitMQ (sync com Cloud)
7. Responder 200 para o Mercado Pago

**Resposta:** `200 OK` (mesmo em caso de erro interno — para evitar reenvio do webhook)

---

### `POST /api/voucher/resgatar`

```json
{
  "codigo": "ABCD-1234",
  "mac": "AA:BB:CC:DD:EE:FF",
  "ip": "192.168.1.50"
}
```

**Resposta 200 — sucesso:**
```json
{
  "sucesso": true,
  "plano": "Acesso 24 Horas",
  "tempo_adicionado_minutos": 1440,
  "fim_acesso": "2025-05-20T14:32:00Z",
  "tempo_restante_segundos": 86400,
  "acesso_anterior": true   // true se já tinha sessão ativa (tempo adicionado)
}
```

**Erros:**
- `404` Voucher não encontrado
- `410` Voucher já utilizado (single use)
- `422` Voucher expirado
- `403` MAC na blacklist

---

## Admin — Usuários

Todas as rotas `/admin/*` requerem `Authorization: Bearer {access_token}`.

### `GET /admin/usuarios`

```
Query params:
  status    = ativo | expirado | bloqueado | walled_garden
  page      = 1 (default)
  limit     = 50 (default, max 200)
  sort      = fim_acesso | created_at | dados_consumidos
  order     = asc | desc
  busca     = (MAC ou nome parcial)
```

**Resposta 200:**
```json
{
  "total": 142,
  "page": 1,
  "limit": 50,
  "usuarios": [
    {
      "id": 1,
      "mac": "AA:BB:CC:DD:EE:FF",
      "ip_atual": "192.168.1.50",
      "nome": "João Silva",
      "status": "ativo",
      "plano": { "id": 1, "nome": "Acesso 24 Horas" },
      "inicio_acesso": "2025-05-19T14:30:00Z",
      "fim_acesso": "2025-05-20T14:30:00Z",
      "tempo_restante_segundos": 66600,
      "dados_consumidos_mb": 1240,
      "roteador": { "id": 1, "nome": "Roteador Principal" }
    }
  ]
}
```

---

### `GET /admin/usuarios/:mac`

Detalhes completos incluindo histórico de sessões.

---

### `POST /admin/usuarios/:mac/desconectar`

```json
{}  // body vazio
```

Executa `ndsctl deauth <mac>` via SSH. **Resposta 200.**

---

### `POST /admin/usuarios/:mac/estender`

```json
{
  "minutos": 120
}
```

---

### `POST /admin/usuarios/:mac/banir`

```json
{
  "motivo": "Uso abusivo de banda"
}
```

Adiciona à blacklist e executa deauth. **Resposta 200.**

---

## Admin — Planos

### `GET /admin/planos`
### `POST /admin/planos`
### `PUT /admin/planos/:id`
### `DELETE /admin/planos/:id`

**Body (criar/editar):**
```json
{
  "nome": "Acesso 48 Horas",
  "descricao": "Dois dias de internet",
  "preco": 25.00,
  "duracao_minutos": 2880,
  "dados_mb": null,
  "velocidade_down": 10,
  "velocidade_up": 5,
  "recomendado": false,
  "ativo": true,
  "visivel_portal": true,
  "ordem": 3
}
```

---

## Admin — Vouchers

### `GET /admin/vouchers`
### `POST /admin/vouchers/gerar`

```json
{
  "plano_id": 1,
  "quantidade": 50,
  "tipo": "single_use",
  "usos_maximos": null,
  "validade_dias": 30,
  "prefixo": "VIP"
}
```

**Resposta 201:**
```json
{
  "lote_id": 12,
  "quantidade": 50,
  "codigos": ["VIPABCD-1234", "VIPEFGH-5678", "..."]
}
```

### `GET /admin/vouchers/exportar`

```
Query params: lote_id=12&formato=pdf|csv
```

Retorna arquivo para download.

### `DELETE /admin/vouchers/:codigo`

Desativa o voucher.

---

## Admin — Pagamentos

### `GET /admin/pagamentos`

```
Query params:
  de = 2025-05-01 (data ISO)
  ate = 2025-05-31
  status = aprovado | pendente | cancelado | expirado
  page = 1
  limit = 50
```

### `GET /admin/pagamentos/relatorio`

```
Query params: de=2025-05-01&ate=2025-05-31&formato=json|csv|pdf
```

---

## Admin — Rede

### `GET /admin/rede/roteadores`
### `POST /admin/rede/roteadores`
### `PUT /admin/rede/roteadores/:id`
### `DELETE /admin/rede/roteadores/:id`

### `POST /admin/rede/roteadores/:id/diagnostico`

Retorna resultado de ping, uptime, versão OpenWrt/OpenNDS.

### `POST /admin/rede/roteadores/:id/speedtest`

Executa teste de velocidade no uplink do roteador.

### `GET /admin/rede/blacklist`
### `POST /admin/rede/blacklist`
```json
{ "mac": "AA:BB:CC:DD:EE:FF", "motivo": "Spam" }
```
### `DELETE /admin/rede/blacklist/:mac`

### `GET /admin/rede/walled-garden`
### `POST /admin/rede/walled-garden`
```json
{ "host": "meusite.com.br", "descricao": "Site do estabelecimento" }
```
### `DELETE /admin/rede/walled-garden/:id`

---

## Admin — Sistema

### `GET /admin/sistema/settings`
### `PUT /admin/sistema/settings`

```json
{
  "hotspot_nome": "Wi-Fi Recanto Verde",
  "cor_primaria": "#2ECC71",
  "mp_access_token": "APP_USR-XXX"
}
```

### `GET /admin/sistema/backup`

Retorna arquivo `.sql.gz` para download.

### `POST /admin/sistema/restore`

Upload de arquivo `.sql.gz` para restauração.
`Content-Type: multipart/form-data`

### `GET /admin/sistema/logs`

```
Query params:
  nivel = INFO | WARN | ERROR
  categoria = auth | payment | network | system | admin
  de = (ISO datetime)
  ate = (ISO datetime)
  page = 1
  limit = 100
```

### `GET /admin/sistema/saude`

Health check detalhado.

**Resposta 200:**
```json
{
  "status": "healthy",
  "versao": "1.2.0",
  "uptime_segundos": 1209600,
  "checks": {
    "banco_dados": { "status": "ok", "latencia_ms": 2 },
    "redis": { "status": "ok", "latencia_ms": 1 },
    "rabbitmq": { "status": "ok" },
    "mercadopago": { "status": "ok" },
    "roteadores": {
      "total": 4,
      "online": 3,
      "offline": 1
    }
  }
}
```

---

## WebSocket — Admin Real-time

```
ws://[host]:5000/admin/ws?token=[jwt]
```

### Eventos enviados pelo servidor

```typescript
// Usuário conectou
{ event: "user.connected", data: { mac, plano, ip, roteador_id } }

// Usuário desconectou/expirou
{ event: "user.expired", data: { mac, motivo: "expired" | "manual" | "banned" } }

// Pagamento aprovado
{ event: "payment.approved", data: { mac, valor, plano, txid } }

// Métricas periódicas (a cada 10s)
{ event: "metrics.update", data: {
    usuarios_ativos: 23,
    receita_hoje: 345.00,
    banda_down_mbps: 45.2,
    banda_up_mbps: 8.1
  }
}

// Status de roteador mudou
{ event: "router.status", data: { id, nome, status: "online" | "offline", latencia_ms } }

// Voucher resgatado
{ event: "voucher.redeemed", data: { codigo, mac, plano } }
```

---

## Erros Padrão

```json
{
  "erro": "Código de erro snake_case",
  "mensagem": "Descrição amigável do erro",
  "detalhes": { }   // opcional, para validação de campos
}
```

| HTTP | erro | Situação |
|---|---|---|
| 400 | `validacao_falhou` | Dados inválidos no body |
| 401 | `nao_autenticado` | Token ausente ou inválido |
| 403 | `acesso_negado` | Permissão insuficiente |
| 404 | `nao_encontrado` | Recurso não existe |
| 409 | `conflito` | Recurso já existe (ex: MAC duplicado) |
| 410 | `recurso_esgotado` | Voucher já usado |
| 422 | `regra_negocio` | Violação de regra de negócio |
| 429 | `rate_limit` | Muitas requisições |
| 500 | `erro_interno` | Erro inesperado no servidor |

---

## Rate Limiting

| Endpoint | Limite |
|---|---|
| `POST /api/pix/gerar` | 5/minuto por MAC |
| `POST /api/voucher/resgatar` | 10/minuto por IP |
| `POST /admin/auth/login` | 5/minuto por IP |
| Admin geral | 300/minuto por token |
| Webhooks | Sem limite |

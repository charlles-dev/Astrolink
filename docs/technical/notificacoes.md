# Sistema de Notificações

## Visão Geral

O sistema de notificações opera em duas camadas:

1. **Notificações para o operador** — alertas sobre seus nós (push, email, WhatsApp)
2. **Notificações para o usuário final** — confirmação de pagamento, aviso de expiração

---

## Canais Disponíveis

| Canal | Para quem | Quando |
|---|---|---|
| Push (FCM) | Operador (app mobile) | Nó offline, pagamento, alertas |
| Email (SMTP) | Operador | Alertas críticos, relatórios |
| WhatsApp (Evolution API) | Usuário final | PIX confirmado, sessão expirando |
| SMS (Zenvia/Twilio) | Usuário final | Alternativa ao WhatsApp |
| In-app (WebSocket) | Operador (web/mobile) | Todos os eventos em tempo real |
| SSE (portal cativo) | Usuário final | Pagamento PIX confirmado |

---

## Notificações para o Operador

### Push via Firebase Cloud Messaging (FCM)

**Configuração no Cloud:**
```typescript
// cloud/supabase/functions/send-push/index.ts
import { initializeApp } from 'firebase-admin/app'
import { getMessaging } from 'firebase-admin/messaging'

const app = initializeApp({
    credential: applicationDefault()
})

export async function sendPushNotification(
    fcmTokens: string[],
    notification: {
        title: string
        body: string
        data?: Record<string, string>
        imageUrl?: string
    }
) {
    const message = {
        tokens: fcmTokens,
        notification: {
            title: notification.title,
            body: notification.body,
            imageUrl: notification.imageUrl,
        },
        data: notification.data,
        android: {
            priority: 'high' as const,
            notification: {
                sound: 'default',
                channelId: 'astrolink-alerts',
            },
        },
        apns: {
            payload: {
                aps: {
                    sound: 'default',
                    badge: 1,
                },
            },
        },
    }

    return getMessaging(app).sendEachForMulticast(message)
}
```

**Templates de notificação:**
```typescript
const NotificationTemplates = {
    'node.offline': (data: { nome: string, minutos: number }) => ({
        title: `⚠️ ${data.nome} offline`,
        body: `Seu nó ficou offline há ${data.minutos} minuto${data.minutos > 1 ? 's' : ''}`,
        data: { tipo: 'node.offline', screen: 'NodeDetails' }
    }),

    'node.online': (data: { nome: string }) => ({
        title: `✅ ${data.nome} voltou`,
        body: 'Seu nó está online novamente',
        data: { tipo: 'node.online', screen: 'NodeDetails' }
    }),

    'payment.approved': (data: { valor: string, plano: string, no: string }) => ({
        title: `💰 R$ ${data.valor} recebido`,
        body: `Pagamento aprovado em ${data.no} — ${data.plano}`,
        data: { tipo: 'payment', screen: 'Financeiro' }
    }),

    'vouchers.low': (data: { no: string, quantidade: number, plano: string }) => ({
        title: `🎟️ Estoque baixo`,
        body: `${data.quantidade} vouchers de ${data.plano} em ${data.no}`,
        data: { tipo: 'vouchers', screen: 'Vouchers' }
    }),

    'revenue.goal': (data: { valor: string, no: string }) => ({
        title: `🎉 Meta atingida!`,
        body: `R$ ${data.valor} de receita hoje em ${data.no}`,
        data: { tipo: 'revenue', screen: 'Financeiro' }
    }),
}
```

---

### Email (SMTP)

**Configuração:**
```go
// Suporte a qualquer SMTP: Resend, Mailgun, Brevo, Gmail, servidor próprio
type EmailConfig struct {
    Host     string // smtp.resend.com
    Port     int    // 587
    Username string
    Password string
    From     string // nao-responder@astrolink.app
    FromName string // Astrolink
}
```

**Template — Nó Offline (HTML):**
```html
<!-- Estilo simples, compatível com todos os clientes de email -->
<div style="font-family: sans-serif; max-width: 600px; margin: 0 auto;">
  <div style="background: #0F172A; padding: 24px; text-align: center;">
    <img src="https://astrolink.app/logo.png" width="120" alt="Astrolink">
  </div>

  <div style="padding: 32px; background: #fff;">
    <h2 style="color: #EF4444;">⚠️ Nó Offline Detectado</h2>

    <p>Seu nó <strong>{{nome_no}}</strong> ficou offline às <strong>{{hora}}</strong>.</p>

    <p style="background: #FEF2F2; padding: 16px; border-radius: 8px; border-left: 4px solid #EF4444;">
      <strong>Último heartbeat:</strong> {{ultimo_heartbeat}}<br>
      <strong>Usuários afetados:</strong> {{usuarios_ativos}}
    </p>

    <p>Possíveis causas:</p>
    <ul>
      <li>Queda de energia no local</li>
      <li>Problema na conexão Starlink</li>
      <li>Servidor reiniciando</li>
    </ul>

    <a href="{{link_no}}" style="display: inline-block; background: #06B6D4; color: #fff;
       padding: 12px 24px; border-radius: 8px; text-decoration: none; margin-top: 16px;">
      Ver detalhes do nó →
    </a>
  </div>

  <div style="padding: 16px; text-align: center; color: #94A3B8; font-size: 12px;">
    <a href="{{link_configuracoes}}">Gerenciar alertas</a> ·
    <a href="{{link_cancelar}}">Cancelar notificações</a>
  </div>
</div>
```

---

## Notificações para o Usuário Final

### WhatsApp via Evolution API

O Evolution API é uma API open source que conecta ao WhatsApp Web. Self-hosted, gratuito.

```go
// internal/notifications/whatsapp.go
type EvolutionAPIClient struct {
    baseURL    string
    apiKey     string
    instanceID string
    httpClient *http.Client
}

func (c *EvolutionAPIClient) SendTextMessage(phone, message string) error {
    body := map[string]any{
        "number":  formatPhone(phone), // "5511999999999"
        "options": map[string]any{"delay": 1200},
        "textMessage": map[string]any{
            "text": message,
        },
    }

    data, _ := json.Marshal(body)
    req, _ := http.NewRequest("POST",
        fmt.Sprintf("%s/message/sendText/%s", c.baseURL, c.instanceID),
        bytes.NewBuffer(data))
    req.Header.Set("apikey", c.apiKey)
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(req)
    if err != nil || resp.StatusCode >= 400 {
        return fmt.Errorf("enviar WhatsApp: status %d", resp.StatusCode)
    }
    return nil
}
```

**Templates WhatsApp:**

```go
var WhatsAppTemplates = map[string]string{
    "payment.confirmed": `✅ *Pagamento confirmado!*

Seu acesso à internet foi liberado.

📦 *Plano:* {{plano}}
⏰ *Válido por:* {{duracao}}
📅 *Expira em:* {{expira_em}}

Aproveite a navegação! 🚀
_Astrolink Wi-Fi_`,

    "session.expiring": `⏰ *Seu acesso está quase expirando*

Restam apenas *{{tempo_restante}}* de internet.

Quer continuar navegando?
Acesse: {{url_portal}}

_Astrolink Wi-Fi_`,

    "session.expired": `❌ *Seu acesso expirou*

Para continuar navegando, adquira um novo plano:
{{url_portal}}

_Astrolink Wi-Fi_`,
}
```

**Instalação do Evolution API:**
```bash
# docker-compose.yml (adicionar ao nó)
evolution-api:
  image: atendai/evolution-api:latest
  ports:
    - "8080:8080"
  environment:
    - AUTHENTICATION_API_KEY=minha_api_key_secreta
    - DATABASE_CONNECTION_URI=postgresql://...
  volumes:
    - ./data/evolution:/evolution/instances
```

---

### SMS via Zenvia (alternativa ao WhatsApp)

```go
type ZenviaClient struct {
    apiToken string
}

func (c *ZenviaClient) SendSMS(to, message string) error {
    body := map[string]any{
        "from": "Astrolink",
        "to":   to,
        "contents": []map[string]string{
            {"type": "text", "text": message},
        },
    }
    // POST https://api.zenvia.com/v2/channels/sms/messages
    // Authorization: Token {apiToken}
}
```

---

## Configuração de Alertas por Operador

### No Cloud Panel (Settings → Alertas)

```typescript
// Schema no Supabase
interface AlertConfig {
    tenant_id: string
    node_id: string | null  // null = todos os nós

    // Canais habilitados
    push_enabled: boolean
    email_enabled: boolean
    whatsapp_enabled: boolean  // futuramente

    // Tipos de alerta
    node_offline: {
        enabled: boolean
        delay_minutes: number  // avisar só depois de X min offline
    }
    vouchers_low: {
        enabled: boolean
        threshold: number  // quantidade mínima
        per_plan: boolean
    }
    revenue_goal: {
        enabled: boolean
        valor: number
    }
    suspicious_activity: boolean

    // Silenciar em horários
    quiet_hours: {
        enabled: boolean
        start: string  // "22:00"
        end: string    // "07:00"
        timezone: string
    }
}
```

---

## Pipeline de Notificações

```
1. Evento ocorre no Nó (ex: router.offline)
         ↓
2. Agente publica no RabbitMQ
         ↓
3. Consumer no Cloud recebe
         ↓
4. Edge Function: NotificationDispatcher
         ↓
5. Consulta configurações do tenant (alert_configs)
         ↓
6. Filtra: deve notificar? (tipo, quiet hours, threshold)
         ↓
7. Para cada canal habilitado:
   - FCM: sendPushNotification()
   - Email: sendEmail() via SMTP
   - In-app: INSERT em notification_feed → Supabase Realtime → WebSocket
         ↓
8. Log da notificação enviada (para histórico e deduplicação)
```

---

## Deduplicação

```go
// Evitar notificar múltiplas vezes para o mesmo evento
func (d *NotificationDispatcher) ShouldNotify(key, tipo string, cooldown time.Duration) bool {
    cacheKey := fmt.Sprintf("notif:%s:%s", key, tipo)
    set, _ := d.redis.SetNX(ctx, cacheKey, "1", cooldown).Result()
    return set // true = ainda não notificou neste período
}

// Uso:
// Nó offline: notificar só uma vez a cada 30 minutos
if d.ShouldNotify(nodeID, "node.offline", 30*time.Minute) {
    d.SendPush(tokens, templates["node.offline"](data))
}
```

---

## Notificação do Portal (SSE)

Para o usuário final aguardando o PIX no portal cativo:

```go
// GET /api/pix/aguardar/:txid
func HandleAwaitPIX(c *fiber.Ctx) error {
    txid := c.Params("txid")

    c.Set("Content-Type", "text/event-stream")
    c.Set("Cache-Control", "no-cache")
    c.Set("Connection", "keep-alive")

    // Canal Redis pub/sub para este txid específico
    pubsub := redisClient.Subscribe(ctx, fmt.Sprintf("pix:%s", txid))
    defer pubsub.Close()

    // Heartbeat a cada 15s para manter conexão viva
    ticker := time.NewTicker(15 * time.Second)
    defer ticker.Stop()

    // Timeout: 20 minutos (expiração máxima do PIX)
    timeout := time.After(20 * time.Minute)

    for {
        select {
        case msg := <-pubsub.Channel():
            fmt.Fprintf(c, "event: aprovado\ndata: %s\n\n", msg.Payload)
            return nil

        case <-ticker.C:
            fmt.Fprintf(c, "event: heartbeat\ndata: {}\n\n")

        case <-timeout:
            fmt.Fprintf(c, "event: expirado\ndata: {}\n\n")
            return nil

        case <-c.Context().Done():
            return nil
        }
    }
}

// Quando webhook do MP chega:
func NotifyPIXApproved(txid string, payload any) {
    data, _ := json.Marshal(payload)
    redisClient.Publish(ctx, fmt.Sprintf("pix:%s", txid), data)
}
```

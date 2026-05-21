# Sincronização em Tempo Real (Nó ↔ Cloud)

## Visão Geral

A sincronização entre os Nós locais e o Cloud Panel acontece em duas direções:

- **Nó → Cloud:** eventos (pagamentos, conexões, status) publicados via RabbitMQ
- **Cloud → Nó:** comandos (banir MAC, atualizar config, reiniciar) consumidos pelo agente Go

O sistema é **offline-first**: o Nó funciona completamente sem conexão com a Cloud. Quando a conexão cai, eventos são bufferizados localmente e sincronizados quando a conexão retorna.

---

## Arquitetura do Sistema de Mensagens

```
┌─────────────────────────────────────────────────────────────┐
│                         NÓ LOCAL                            │
│                                                             │
│  ┌──────────────────┐    ┌───────────────────────────────┐ │
│  │  Backend Go      │    │   Agente de Sync (goroutine)  │ │
│  │                  │    │                               │ │
│  │  Eventos:        │───►│  Publisher (AMQP)             │ │
│  │  • payment.approved   │  Consumer (comandos)         │ │
│  │  • user.connected│    │  Buffer local (Redis)         │ │
│  │  • node.heartbeat│    │  Retry com backoff            │ │
│  └──────────────────┘    └──────────────┬────────────────┘ │
└───────────────────────────────────────── │ ──────────────────┘
                                           │ TLS AMQPS
┌──────────────────────────────────────────▼──────────────────┐
│                      RABBITMQ CLOUD                          │
│                                                             │
│  Exchange: astrolink.events  (topic, durable)               │
│  Exchange: astrolink.commands (direct, durable)             │
│                                                             │
│  Queues:                                                    │
│  • events.cloud-panel  → consumer no Cloud                  │
│  • commands.{node-id}  → consumer no Nó específico          │
│  • dlx.failed          → Dead Letter (falhas)               │
└──────────────────────────────────────────┬──────────────────┘
                                           │
┌──────────────────────────────────────────▼──────────────────┐
│                     CLOUD (Supabase + Edge)                  │
│                                                             │
│  Consumer Go/TS: recebe eventos dos nós                     │
│  Salva em: node_events, node_metrics                        │
│  Supabase Realtime: distribui para WebSocket dos clientes   │
│  Publisher: envia comandos de volta para nós                │
└─────────────────────────────────────────────────────────────┘
```

---

## Topologia RabbitMQ

### Exchanges

```go
// Declaração da topologia (feita na inicialização do agente)
exchanges := []ExchangeDeclaration{
    {
        Name:    "astrolink.events",
        Type:    "topic",      // routing por tipo de evento
        Durable: true,
    },
    {
        Name:    "astrolink.commands",
        Type:    "direct",     // routing direto por node-id
        Durable: true,
    },
    {
        Name:    "astrolink.dlx",   // Dead Letter Exchange
        Type:    "fanout",
        Durable: true,
    },
}
```

### Queues e Bindings

```go
queues := []QueueDeclaration{
    // Eventos de todos os nós → Cloud Panel
    {
        Name:    "events.cloud-panel",
        Durable: true,
        Args: amqp.Table{
            "x-dead-letter-exchange": "astrolink.dlx",
            "x-message-ttl":         86400000, // 24h
        },
        Binding: QueueBinding{
            Exchange:   "astrolink.events",
            RoutingKey: "#",  // todos os eventos
        },
    },
    // Comandos para este nó específico
    {
        Name:    fmt.Sprintf("commands.%s", nodeID),
        Durable: true,
        Args: amqp.Table{
            "x-dead-letter-exchange": "astrolink.dlx",
            "x-message-ttl":         3600000, // 1h
        },
        Binding: QueueBinding{
            Exchange:   "astrolink.commands",
            RoutingKey: nodeID,
        },
    },
    // Dead Letter Queue
    {
        Name:    "dlx.failed",
        Durable: true,
    },
}
```

---

## Eventos Publicados pelo Nó

### Formato Padrão

```json
{
  "id": "uuid-v4",
  "node_id": "node-uuid",
  "tenant_id": "tenant-uuid",
  "tipo": "payment.approved",
  "timestamp": "2025-05-19T14:32:00.123Z",
  "versao": "1",
  "payload": { ... }
}
```

### Catálogo de Eventos

```
payment.approved
  payload: { txid, mac, valor, plano_id, plano_nome }

payment.failed
  payload: { txid, mac, motivo }

user.connected
  payload: { mac, ip, plano_id, plano_nome, duracao_minutos, roteador_id }

user.expired
  payload: { mac, motivo: "time_limit" | "data_limit" | "manual" | "banned" }

user.extended
  payload: { mac, minutos_adicionados, novo_fim_acesso, por: "admin" | "voucher" }

voucher.redeemed
  payload: { codigo, mac, plano_id, plano_nome }

router.offline
  payload: { roteador_id, nome, ip, ultimo_ping_ms }

router.online
  payload: { roteador_id, nome, ip, latencia_ms }

mac.banned
  payload: { mac, motivo, por: "admin" | "system" }

mac.unbanned
  payload: { mac, por: "admin" }

node.heartbeat
  payload: {
    versao_software: "1.2.0",
    usuarios_ativos: 23,
    receita_hoje: 345.00,
    banda_down_mbps: 45.2,
    banda_up_mbps: 8.1,
    roteadores_online: 3,
    roteadores_offline: 1
  }

node.config_changed
  payload: { campo: "hotspot_nome", novo_valor: "Wi-Fi Recanto" }
```

---

## Comandos Recebidos pelo Nó

```
mac.ban
  payload: { mac, motivo, por_usuario_id }
  resposta: { sucesso, erro? }

mac.unban
  payload: { mac, por_usuario_id }

session.disconnect
  payload: { mac, por_usuario_id }

session.extend
  payload: { mac, minutos, por_usuario_id }

config.update
  payload: { chave, valor }

nds.restart
  payload: { roteador_id? }  // null = todos

node.diagnose
  payload: {}
  resposta: { backend_ok, db_ok, redis_ok, roteadores: [...] }
```

---

## Implementação do Agente de Sync

```go
// internal/sync/agent.go
package sync

type SyncAgent struct {
    nodeID   string
    tenantID string
    amqp     *amqp.Connection
    redis    *redis.Client
    db       *queries.Queries
    nds      *network.OpenNDSManager
    log      *slog.Logger

    // Buffer para eventos quando offline
    buffer []Event
    mu     sync.Mutex
}

func (a *SyncAgent) Start(ctx context.Context) {
    go a.heartbeatLoop(ctx)
    go a.publishLoop(ctx)
    go a.consumeCommands(ctx)
}

// heartbeatLoop envia heartbeat a cada 30 segundos
func (a *SyncAgent) heartbeatLoop(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    for {
        select {
        case <-ticker.C:
            metrics := a.collectMetrics()
            a.PublishEvent("node.heartbeat", metrics)
        case <-ctx.Done():
            return
        }
    }
}

// PublishEvent publica com buffer automático se offline
func (a *SyncAgent) PublishEvent(tipo string, payload any) {
    event := Event{
        ID:       uuid.New().String(),
        NodeID:   a.nodeID,
        TenantID: a.tenantID,
        Tipo:     tipo,
        Timestamp: time.Now().UTC(),
        Payload:  payload,
    }

    if err := a.publishToAMQP(event); err != nil {
        // Falhou: bufferizar no Redis
        a.log.Warn("falha ao publicar evento, bufferizando", "tipo", tipo, "err", err)
        a.bufferEvent(event)
    }
}

// publishLoop drena o buffer quando a conexão retorna
func (a *SyncAgent) publishLoop(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Second)
    for {
        select {
        case <-ticker.C:
            a.drainBuffer()
        case <-ctx.Done():
            return
        }
    }
}

// Reconexão com backoff exponencial
func (a *SyncAgent) connectWithBackoff() {
    backoff := []time.Duration{1, 2, 4, 8, 16, 30}
    for i := 0; ; i++ {
        conn, err := amqp.DialTLS(a.amqpURL, &tls.Config{})
        if err == nil {
            a.amqp = conn
            a.log.Info("reconectado ao RabbitMQ")
            return
        }

        delay := backoff[min(i, len(backoff)-1)] * time.Second
        a.log.Warn("falha ao conectar RabbitMQ, tentando em", "delay", delay, "err", err)
        time.Sleep(delay)
    }
}

// consumeCommands escuta comandos do Cloud
func (a *SyncAgent) consumeCommands(ctx context.Context) {
    msgs, err := a.channel.Consume(
        fmt.Sprintf("commands.%s", a.nodeID),
        "astrolink-node",
        false, // manual ack
        false, false, false, nil,
    )

    for msg := range msgs {
        if err := a.handleCommand(msg.Body); err != nil {
            a.log.Error("erro ao executar comando", "err", err)
            msg.Nack(false, true) // requeue
        } else {
            msg.Ack(false)
        }
    }
}
```

---

## Realtime no Cloud (Supabase)

```typescript
// cloud/src/lib/realtime.ts

// Escutar eventos de nós em tempo real no painel web
export function subscribeToNodeEvents(tenantId: string, nodeId: string) {
    return supabase
        .channel(`node-events-${nodeId}`)
        .on(
            'postgres_changes',
            {
                event: 'INSERT',
                schema: 'public',
                table: 'node_events',
                filter: `node_id=eq.${nodeId}`,
            },
            (payload) => {
                handleNodeEvent(payload.new as NodeEvent)
            }
        )
        .subscribe()
}

// Escutar métricas em tempo real
export function subscribeToNodeMetrics(nodeId: string, callback: (m: NodeMetrics) => void) {
    return supabase
        .channel(`node-metrics-${nodeId}`)
        .on(
            'postgres_changes',
            {
                event: 'INSERT',
                schema: 'public',
                table: 'node_metrics',
                filter: `node_id=eq.${nodeId}`,
            },
            (payload) => callback(payload.new as NodeMetrics)
        )
        .subscribe()
}

// Enviar comando para um nó (via Edge Function)
export async function sendNodeCommand(nodeId: string, tipo: string, payload: object) {
    const { data, error } = await supabase.functions.invoke('node-command', {
        body: { node_id: nodeId, tipo, payload }
    })
    if (error) throw error
    return data
}
```

---

## Buffer Local (Redis)

```go
// Eventos bufferizados quando offline
const bufferKey = "sync:event_buffer"
const maxBufferSize = 10000

func (a *SyncAgent) bufferEvent(event Event) {
    data, _ := json.Marshal(event)

    // LPUSH com limite de tamanho
    pipe := a.redis.Pipeline()
    pipe.LPush(ctx, bufferKey, data)
    pipe.LTrim(ctx, bufferKey, 0, maxBufferSize-1)  // limitar buffer
    pipe.Exec(ctx)
}

func (a *SyncAgent) drainBuffer() {
    for {
        data, err := a.redis.RPop(ctx, bufferKey).Bytes()
        if err == redis.Nil {
            break // buffer vazio
        }

        var event Event
        json.Unmarshal(data, &event)

        if err := a.publishToAMQP(event); err != nil {
            // Falhou de novo: devolver ao buffer e parar
            a.redis.LPush(ctx, bufferKey, data)
            break
        }
    }
}
```

---

## Monitoramento da Sync

### Métricas Prometheus

```go
var (
    eventsPublished = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "astrolink_sync_events_published_total",
        Help: "Total de eventos publicados",
    }, []string{"tipo", "status"})

    commandsReceived = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "astrolink_sync_commands_received_total",
    }, []string{"tipo", "status"})

    bufferSize = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "astrolink_sync_buffer_size",
        Help: "Eventos no buffer (aguardando sync)",
    })

    lastHeartbeat = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "astrolink_sync_last_heartbeat_timestamp",
    })
)
```

### Alertas (Grafana)

- Buffer > 1000 eventos → warning (conexão instável)
- Buffer > 5000 eventos → critical (offline há muito tempo)
- Último heartbeat > 2 min → nó potencialmente offline
- Taxa de erros de publicação > 10% → problema de conexão

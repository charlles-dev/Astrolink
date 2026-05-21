# Schema do Banco de Dados

## Visão Geral

O sistema usa dois bancos PostgreSQL distintos:

1. **Banco Local** — roda no servidor de cada Nó, gerenciado pelo backend Go
2. **Banco Cloud** — Supabase Cloud, para o painel multi-tenant

---

## Banco Local (por Nó)

### Diagrama de Relacionamentos

```
planos ──────────────── transacoes_pix
  │                           │
  │                     usuarios_mac
  │                           │
  └──── vouchers ─────────────┘
             │
         voucher_usos

system_settings (KV store)
logs (auditoria)
roteadores
blacklist_mac
walled_garden
```

---

### Tabela: `planos`

```sql
CREATE TABLE planos (
    id              SERIAL PRIMARY KEY,
    nome            VARCHAR(100) NOT NULL,
    descricao       TEXT,
    preco           NUMERIC(10,2) NOT NULL CHECK (preco >= 0),
    duracao_minutos INTEGER,          -- NULL se plano por dados
    dados_mb        INTEGER,          -- NULL se plano por tempo
    velocidade_down INTEGER DEFAULT 0, -- Mbps, 0 = ilimitado
    velocidade_up   INTEGER DEFAULT 0, -- Mbps, 0 = ilimitado
    recomendado     BOOLEAN DEFAULT FALSE,
    ativo           BOOLEAN DEFAULT TRUE,
    visivel_portal  BOOLEAN DEFAULT TRUE,
    ordem           INTEGER DEFAULT 0, -- ordenação no portal
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_planos_ativo ON planos(ativo);
```

---

### Tabela: `usuarios_mac`

Representa um dispositivo (identificado pelo MAC) e seu estado de acesso.

```sql
CREATE TABLE usuarios_mac (
    id              SERIAL PRIMARY KEY,
    mac             MACADDR NOT NULL UNIQUE,
    ip_atual        INET,
    nome            VARCHAR(100),       -- coletado no portal (opcional)
    status          VARCHAR(20) NOT NULL DEFAULT 'walled_garden'
                    CHECK (status IN ('ativo', 'expirado', 'bloqueado', 'walled_garden')),
    plano_id        INTEGER REFERENCES planos(id),
    inicio_acesso   TIMESTAMPTZ,
    fim_acesso      TIMESTAMPTZ,        -- NULL se plano por dados
    dados_consumidos_mb INTEGER DEFAULT 0,
    dados_limite_mb INTEGER,            -- NULL se plano por tempo
    roteador_id     INTEGER REFERENCES roteadores(id),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_usuarios_mac_mac ON usuarios_mac(mac);
CREATE INDEX idx_usuarios_mac_status ON usuarios_mac(status);
CREATE INDEX idx_usuarios_mac_fim_acesso ON usuarios_mac(fim_acesso)
    WHERE status = 'ativo';
```

---

### Tabela: `transacoes_pix`

```sql
CREATE TABLE transacoes_pix (
    id              SERIAL PRIMARY KEY,
    txid            VARCHAR(50) UNIQUE NOT NULL, -- ID do Mercado Pago
    mac             MACADDR NOT NULL,
    plano_id        INTEGER REFERENCES planos(id),
    valor           NUMERIC(10,2) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pendente'
                    CHECK (status IN ('pendente', 'aprovado', 'cancelado', 'expirado')),
    pix_copia_cola  TEXT,              -- código PIX para o usuário copiar
    qr_code_base64  TEXT,             -- imagem QR como base64
    mp_payment_id   BIGINT,           -- ID do pagamento no MP
    webhook_at      TIMESTAMPTZ,      -- quando chegou o webhook de confirmação
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_transacoes_txid ON transacoes_pix(txid);
CREATE INDEX idx_transacoes_mac ON transacoes_pix(mac);
CREATE INDEX idx_transacoes_status ON transacoes_pix(status);
CREATE INDEX idx_transacoes_created ON transacoes_pix(created_at DESC);
```

---

### Tabela: `vouchers`

```sql
CREATE TABLE vouchers (
    id              SERIAL PRIMARY KEY,
    codigo          VARCHAR(20) NOT NULL UNIQUE,
    plano_id        INTEGER NOT NULL REFERENCES planos(id),
    tipo            VARCHAR(20) NOT NULL DEFAULT 'single_use'
                    CHECK (tipo IN ('single_use', 'universal')),
    usos_maximos    INTEGER,          -- NULL = 1 (single_use), N (universal)
    usos_atuais     INTEGER DEFAULT 0,
    validade_em     TIMESTAMPTZ,      -- NULL = sem validade
    ativo           BOOLEAN DEFAULT TRUE,
    prefixo         VARCHAR(10),
    lote_id         INTEGER REFERENCES voucher_lotes(id),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_vouchers_codigo ON vouchers(codigo);
CREATE INDEX idx_vouchers_ativo ON vouchers(ativo);
CREATE INDEX idx_vouchers_lote ON vouchers(lote_id);
```

---

### Tabela: `voucher_lotes`

Agrupa vouchers gerados juntos.

```sql
CREATE TABLE voucher_lotes (
    id              SERIAL PRIMARY KEY,
    quantidade      INTEGER NOT NULL,
    plano_id        INTEGER REFERENCES planos(id),
    criado_por      VARCHAR(50) DEFAULT 'admin',
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
```

---

### Tabela: `voucher_usos`

```sql
CREATE TABLE voucher_usos (
    id              SERIAL PRIMARY KEY,
    voucher_id      INTEGER NOT NULL REFERENCES vouchers(id),
    mac             MACADDR NOT NULL,
    ip              INET,
    tempo_adicionado_min INTEGER,     -- tempo que foi creditado
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_voucher_usos_voucher ON voucher_usos(voucher_id);
CREATE INDEX idx_voucher_usos_mac ON voucher_usos(mac);
```

---

### Tabela: `roteadores`

```sql
CREATE TABLE roteadores (
    id              SERIAL PRIMARY KEY,
    nome            VARCHAR(100) NOT NULL,
    ip              INET NOT NULL,
    porta_ssh       INTEGER DEFAULT 22,
    usuario_ssh     VARCHAR(50) DEFAULT 'root',
    chave_ssh_path  TEXT,             -- path para a chave privada no servidor
    status          VARCHAR(20) DEFAULT 'unknown'
                    CHECK (status IN ('online', 'offline', 'unknown')),
    ultimo_ping_ms  INTEGER,
    ultimo_check_at TIMESTAMPTZ,
    versao_openwrt  VARCHAR(50),
    versao_opennds  VARCHAR(50),
    ativo           BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
```

---

### Tabela: `blacklist_mac`

```sql
CREATE TABLE blacklist_mac (
    id              SERIAL PRIMARY KEY,
    mac             MACADDR NOT NULL UNIQUE,
    motivo          TEXT,
    criado_por      VARCHAR(50) DEFAULT 'admin',
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_blacklist_mac ON blacklist_mac(mac);
```

---

### Tabela: `walled_garden`

```sql
CREATE TABLE walled_garden (
    id              SERIAL PRIMARY KEY,
    host            VARCHAR(255) NOT NULL UNIQUE, -- ex: pagamentos.mercadopago.com
    descricao       TEXT,
    tipo            VARCHAR(20) DEFAULT 'dominio'
                    CHECK (tipo IN ('dominio', 'ip', 'subnet')),
    sistema         BOOLEAN DEFAULT FALSE, -- TRUE = não pode ser removido pelo admin
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
```

---

### Tabela: `system_settings`

KV store para configurações dinâmicas (white-label, integrações, etc.)

```sql
CREATE TABLE system_settings (
    chave           VARCHAR(100) PRIMARY KEY,
    valor           TEXT NOT NULL,
    tipo            VARCHAR(20) DEFAULT 'string'
                    CHECK (tipo IN ('string', 'integer', 'boolean', 'json')),
    descricao       TEXT,
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Valores padrão
INSERT INTO system_settings (chave, valor, tipo, descricao) VALUES
('hotspot_nome',          'Astrolink Wi-Fi',   'string',  'Nome exibido no portal'),
('hotspot_logo_url',      '',                  'string',  'URL ou path da logo'),
('cor_primaria',          '#06B6D4',           'string',  'Cor primária (hex)'),
('cor_fundo',             '#0F172A',           'string',  'Cor de fundo (hex)'),
('url_pos_conexao',       'https://google.com','string',  'Redirecionar após login'),
('coleta_nome',           'false',             'boolean', 'Pedir nome no portal'),
('mp_access_token',       '',                  'string',  'Mercado Pago Access Token'),
('mp_public_key',         '',                  'string',  'Mercado Pago Public Key'),
('mp_modo',               'sandbox',           'string',  'sandbox ou producao'),
('cloud_token',           '',                  'string',  'Token de vinculação Cloud'),
('cloud_enabled',         'false',             'boolean', 'Sync com Cloud Panel'),
('backup_frequencia',     'daily',             'string',  'daily, weekly, disabled'),
('backup_hora',           '03:00',             'string',  'Hora do backup automático'),
('backup_reter',          '7',                 'integer', 'Qtd de backups a manter');
```

---

### Tabela: `logs`

```sql
CREATE TABLE logs (
    id              BIGSERIAL PRIMARY KEY,
    nivel           VARCHAR(10) NOT NULL CHECK (nivel IN ('DEBUG','INFO','WARN','ERROR')),
    categoria       VARCHAR(50) NOT NULL, -- 'auth', 'payment', 'network', 'system', 'admin'
    mensagem        TEXT NOT NULL,
    dados           JSONB,               -- dados extras estruturados
    usuario         VARCHAR(50),         -- quem executou (admin, system, etc.)
    ip_origem       INET,
    mac_relacionado MACADDR,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_logs_nivel ON logs(nivel);
CREATE INDEX idx_logs_categoria ON logs(categoria);
CREATE INDEX idx_logs_created ON logs(created_at DESC);
CREATE INDEX idx_logs_mac ON logs(mac_relacionado);

-- Limpar logs com mais de 90 dias (job agendado)
-- DELETE FROM logs WHERE created_at < NOW() - INTERVAL '90 days';
```

---

### Tabela: `sessoes_admin`

```sql
CREATE TABLE sessoes_admin (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario         VARCHAR(50) NOT NULL,
    refresh_token   TEXT UNIQUE NOT NULL,
    ip              INET,
    user_agent      TEXT,
    expira_em       TIMESTAMPTZ NOT NULL,
    revogado        BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_sessoes_refresh ON sessoes_admin(refresh_token) WHERE NOT revogado;
```

---

### Jobs Agendados (pg_cron ou go-cron)

```sql
-- Verificar sessões expiradas a cada 1 minuto
-- Implementado no Go scheduler, não no PostgreSQL

-- Limpar logs antigos (>90 dias) toda madrugada
-- SELECT cron.schedule('cleanup-logs', '0 2 * * *',
--   'DELETE FROM logs WHERE created_at < NOW() - INTERVAL ''90 days''');

-- Verificar status dos roteadores a cada 2 minutos
-- Implementado no Go scheduler
```

---

## Banco Cloud (Supabase)

### Extensões necessárias

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS postgis;       -- para queries de proximidade
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
```

---

### Tabela: `tenants`

```sql
CREATE TABLE tenants (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome            VARCHAR(200) NOT NULL,
    slug            VARCHAR(100) UNIQUE NOT NULL, -- provedor-xyz
    plano_billing   VARCHAR(20) DEFAULT 'free'
                    CHECK (plano_billing IN ('free', 'pro', 'business')),
    status          VARCHAR(20) DEFAULT 'active'
                    CHECK (status IN ('active', 'suspended', 'cancelled')),
    max_nos         INTEGER DEFAULT 1,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
```

---

### Tabela: `tenant_members`

```sql
CREATE TABLE tenant_members (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    role            VARCHAR(20) NOT NULL DEFAULT 'viewer'
                    CHECK (role IN ('owner', 'admin', 'viewer')),
    invited_at      TIMESTAMPTZ DEFAULT NOW(),
    accepted_at     TIMESTAMPTZ,
    UNIQUE(tenant_id, user_id)
);

-- RLS
ALTER TABLE tenant_members ENABLE ROW LEVEL SECURITY;
CREATE POLICY "tenant_member_access" ON tenant_members
    FOR ALL USING (
        tenant_id IN (
            SELECT tenant_id FROM tenant_members
            WHERE user_id = auth.uid()
        )
    );
```

---

### Tabela: `nodes`

```sql
CREATE TABLE nodes (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    nome            VARCHAR(100) NOT NULL,
    slug            VARCHAR(100) NOT NULL,
    descricao       TEXT,
    -- Localização geográfica
    lat             DOUBLE PRECISION,
    lng             DOUBLE PRECISION,
    localizacao     GEOGRAPHY(POINT, 4326),  -- PostGIS
    endereco        TEXT,
    cidade          VARCHAR(100),
    estado          VARCHAR(2),
    -- Status
    status          VARCHAR(20) DEFAULT 'pending'
                    CHECK (status IN ('online', 'offline', 'degraded', 'pending')),
    versao          VARCHAR(20),
    last_heartbeat_at TIMESTAMPTZ,
    -- Vinculação
    token_hash      TEXT UNIQUE NOT NULL,   -- hash do token de vinculação
    -- Configurações públicas
    public_listing  BOOLEAN DEFAULT FALSE,
    public_nome     VARCHAR(200),
    public_descricao TEXT,
    public_foto_url TEXT,
    public_site     VARCHAR(255),
    public_horario  VARCHAR(200),
    -- Timestamps
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, slug)
);

-- Índice geoespacial
CREATE INDEX idx_nodes_localizacao ON nodes USING GIST(localizacao);
CREATE INDEX idx_nodes_tenant ON nodes(tenant_id);
CREATE INDEX idx_nodes_public ON nodes(public_listing, status) WHERE public_listing = TRUE;

-- RLS
ALTER TABLE nodes ENABLE ROW LEVEL SECURITY;
CREATE POLICY "node_tenant_access" ON nodes
    FOR ALL USING (
        tenant_id IN (
            SELECT tenant_id FROM tenant_members WHERE user_id = auth.uid()
        )
    );

-- Política pública: qualquer um pode ler nós listados publicamente
CREATE POLICY "node_public_read" ON nodes
    FOR SELECT USING (public_listing = TRUE AND status = 'online');
```

---

### Tabela: `node_metrics`

Snapshots de métricas enviados pelos nós a cada N segundos.

```sql
CREATE TABLE node_metrics (
    id              BIGSERIAL PRIMARY KEY,
    node_id         UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    timestamp       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    usuarios_ativos INTEGER DEFAULT 0,
    receita_hoje    NUMERIC(10,2) DEFAULT 0,
    banda_down_mbps FLOAT DEFAULT 0,
    banda_up_mbps   FLOAT DEFAULT 0,
    latencia_ms     INTEGER,
    dados_down_gb   FLOAT DEFAULT 0,
    dados_up_gb     FLOAT DEFAULT 0
);

-- Particionamento por mês (escala)
CREATE INDEX idx_metrics_node_time ON node_metrics(node_id, timestamp DESC);

-- Manter apenas 90 dias de métricas granulares
-- Dados agregados por dia podem ser mantidos por mais tempo
```

---

### Tabela: `node_events`

```sql
CREATE TABLE node_events (
    id              BIGSERIAL PRIMARY KEY,
    node_id         UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    tipo            VARCHAR(50) NOT NULL,
    -- Tipos: user.connected, user.expired, payment.approved, payment.failed,
    --        voucher.redeemed, node.offline, node.online, mac.banned,
    --        router.offline, router.online, admin.action
    payload         JSONB,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_events_node_time ON node_events(node_id, created_at DESC);
CREATE INDEX idx_events_tenant_time ON node_events(tenant_id, created_at DESC);
CREATE INDEX idx_events_tipo ON node_events(tipo);
```

---

### Tabela: `node_commands`

Fila de comandos enviados do Cloud para o Nó.

```sql
CREATE TABLE node_commands (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    node_id         UUID NOT NULL REFERENCES nodes(id),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    tipo            VARCHAR(50) NOT NULL,
    -- Tipos: mac.ban, mac.unban, session.extend, session.disconnect,
    --        config.reload, nds.restart, node.diagnose
    payload         JSONB NOT NULL,
    status          VARCHAR(20) DEFAULT 'pending'
                    CHECK (status IN ('pending', 'sent', 'ack', 'failed')),
    enviado_por     UUID REFERENCES auth.users(id),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    sent_at         TIMESTAMPTZ,
    ack_at          TIMESTAMPTZ,
    error_msg       TEXT
);

CREATE INDEX idx_commands_node_status ON node_commands(node_id, status);
```

---

### Tabela: `subscriptions` (Billing)

```sql
CREATE TABLE subscriptions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    plano           VARCHAR(20) NOT NULL CHECK (plano IN ('free', 'pro', 'business')),
    status          VARCHAR(20) NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active', 'past_due', 'cancelled', 'trial')),
    abacatepay_id   VARCHAR(100) UNIQUE,    -- ID do cliente no AbacatePay
    periodo_inicio  TIMESTAMPTZ NOT NULL,
    periodo_fim     TIMESTAMPTZ NOT NULL,
    trial_fim       TIMESTAMPTZ,
    cancelado_em    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
```

---

### Tabela: `public_reviews` (Mapa Público)

```sql
CREATE TABLE public_reviews (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    node_id         UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    rating          SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comentario      TEXT CHECK (char_length(comentario) <= 280),
    autor_nome      VARCHAR(50),
    ip_hash         TEXT NOT NULL,          -- SHA256 do IP (privacidade)
    aprovado        BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Sem RLS para leitura pública (aprovados)
-- Inserção: anônima, com rate limit por ip_hash
```

---

### Função: Nós próximos (PostGIS)

```sql
CREATE OR REPLACE FUNCTION nodes_nearby(
    lat DOUBLE PRECISION,
    lng DOUBLE PRECISION,
    radius_km FLOAT DEFAULT 50
)
RETURNS TABLE (
    id UUID, public_nome VARCHAR, cidade VARCHAR, estado VARCHAR,
    lat DOUBLE PRECISION, lng DOUBLE PRECISION,
    distancia_km FLOAT, rating_medio FLOAT, total_reviews BIGINT
) AS $$
    SELECT
        n.id, n.public_nome, n.cidade, n.estado, n.lat, n.lng,
        ST_Distance(n.localizacao, ST_MakePoint(lng, lat)::geography) / 1000 as distancia_km,
        COALESCE(AVG(r.rating), 0) as rating_medio,
        COUNT(r.id) as total_reviews
    FROM nodes n
    LEFT JOIN public_reviews r ON r.node_id = n.id AND r.aprovado = true
    WHERE
        n.public_listing = true
        AND n.status = 'online'
        AND ST_DWithin(n.localizacao, ST_MakePoint(lng, lat)::geography, radius_km * 1000)
    GROUP BY n.id
    ORDER BY distancia_km;
$$ LANGUAGE SQL STABLE;
```

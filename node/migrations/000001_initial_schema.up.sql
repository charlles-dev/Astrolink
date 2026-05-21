CREATE TABLE IF NOT EXISTS planos (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(100) NOT NULL,
    descricao TEXT,
    preco NUMERIC(10,2) NOT NULL CHECK (preco >= 0),
    duracao_minutos INTEGER,
    dados_mb INTEGER,
    velocidade_down INTEGER DEFAULT 0,
    velocidade_up INTEGER DEFAULT 0,
    recomendado BOOLEAN DEFAULT FALSE,
    ativo BOOLEAN DEFAULT TRUE,
    visivel_portal BOOLEAN DEFAULT TRUE,
    ordem INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_planos_ativo ON planos(ativo);

CREATE TABLE IF NOT EXISTS roteadores (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(100) NOT NULL,
    ip INET NOT NULL,
    porta_ssh INTEGER DEFAULT 22,
    usuario_ssh VARCHAR(50) DEFAULT 'root',
    chave_ssh_path TEXT,
    status VARCHAR(20) DEFAULT 'unknown' CHECK (status IN ('online', 'offline', 'unknown')),
    ultimo_ping_ms INTEGER,
    ultimo_check_at TIMESTAMPTZ,
    versao_openwrt VARCHAR(50),
    versao_opennds VARCHAR(50),
    ativo BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS usuarios_mac (
    id SERIAL PRIMARY KEY,
    mac MACADDR NOT NULL UNIQUE,
    ip_atual INET,
    nome VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'walled_garden'
        CHECK (status IN ('ativo', 'expirado', 'bloqueado', 'walled_garden')),
    plano_id INTEGER REFERENCES planos(id),
    inicio_acesso TIMESTAMPTZ,
    fim_acesso TIMESTAMPTZ,
    dados_consumidos_mb INTEGER DEFAULT 0,
    dados_limite_mb INTEGER,
    roteador_id INTEGER REFERENCES roteadores(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_usuarios_mac_mac ON usuarios_mac(mac);
CREATE INDEX IF NOT EXISTS idx_usuarios_mac_status ON usuarios_mac(status);
CREATE INDEX IF NOT EXISTS idx_usuarios_mac_fim_acesso ON usuarios_mac(fim_acesso) WHERE status = 'ativo';

CREATE TABLE IF NOT EXISTS transacoes_pix (
    id SERIAL PRIMARY KEY,
    txid VARCHAR(50) UNIQUE NOT NULL,
    mac MACADDR NOT NULL,
    plano_id INTEGER REFERENCES planos(id),
    valor NUMERIC(10,2) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pendente'
        CHECK (status IN ('pendente', 'aprovado', 'cancelado', 'expirado')),
    pix_copia_cola TEXT,
    qr_code_base64 TEXT,
    mp_payment_id BIGINT,
    webhook_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transacoes_txid ON transacoes_pix(txid);
CREATE INDEX IF NOT EXISTS idx_transacoes_mac ON transacoes_pix(mac);
CREATE INDEX IF NOT EXISTS idx_transacoes_status ON transacoes_pix(status);
CREATE INDEX IF NOT EXISTS idx_transacoes_created ON transacoes_pix(created_at DESC);

CREATE TABLE IF NOT EXISTS voucher_lotes (
    id SERIAL PRIMARY KEY,
    quantidade INTEGER NOT NULL,
    plano_id INTEGER REFERENCES planos(id),
    criado_por VARCHAR(50) DEFAULT 'admin',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS vouchers (
    id SERIAL PRIMARY KEY,
    codigo VARCHAR(20) NOT NULL UNIQUE,
    plano_id INTEGER NOT NULL REFERENCES planos(id),
    tipo VARCHAR(20) NOT NULL DEFAULT 'single_use'
        CHECK (tipo IN ('single_use', 'universal')),
    usos_maximos INTEGER,
    usos_atuais INTEGER DEFAULT 0,
    validade_em TIMESTAMPTZ,
    ativo BOOLEAN DEFAULT TRUE,
    prefixo VARCHAR(10),
    lote_id INTEGER REFERENCES voucher_lotes(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_vouchers_codigo ON vouchers(codigo);
CREATE INDEX IF NOT EXISTS idx_vouchers_ativo ON vouchers(ativo);
CREATE INDEX IF NOT EXISTS idx_vouchers_lote ON vouchers(lote_id);

CREATE TABLE IF NOT EXISTS voucher_usos (
    id SERIAL PRIMARY KEY,
    voucher_id INTEGER NOT NULL REFERENCES vouchers(id),
    mac MACADDR NOT NULL,
    ip INET,
    tempo_adicionado_min INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_voucher_usos_voucher ON voucher_usos(voucher_id);
CREATE INDEX IF NOT EXISTS idx_voucher_usos_mac ON voucher_usos(mac);

CREATE TABLE IF NOT EXISTS blacklist_mac (
    id SERIAL PRIMARY KEY,
    mac MACADDR NOT NULL UNIQUE,
    motivo TEXT,
    criado_por VARCHAR(50) DEFAULT 'admin',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_blacklist_mac ON blacklist_mac(mac);

CREATE TABLE IF NOT EXISTS walled_garden (
    id SERIAL PRIMARY KEY,
    host VARCHAR(255) NOT NULL UNIQUE,
    descricao TEXT,
    tipo VARCHAR(20) DEFAULT 'dominio' CHECK (tipo IN ('dominio', 'ip', 'subnet')),
    sistema BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS system_settings (
    chave VARCHAR(100) PRIMARY KEY,
    valor TEXT NOT NULL,
    tipo VARCHAR(20) DEFAULT 'string' CHECK (tipo IN ('string', 'integer', 'boolean', 'json')),
    descricao TEXT,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS logs (
    id BIGSERIAL PRIMARY KEY,
    nivel VARCHAR(10) NOT NULL CHECK (nivel IN ('DEBUG','INFO','WARN','ERROR')),
    categoria VARCHAR(50) NOT NULL,
    mensagem TEXT NOT NULL,
    dados JSONB,
    usuario VARCHAR(50),
    ip_origem INET,
    mac_relacionado MACADDR,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_logs_nivel ON logs(nivel);
CREATE INDEX IF NOT EXISTS idx_logs_categoria ON logs(categoria);
CREATE INDEX IF NOT EXISTS idx_logs_created ON logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_logs_mac ON logs(mac_relacionado);

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS sessoes_admin (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario VARCHAR(50) NOT NULL,
    refresh_token TEXT UNIQUE NOT NULL,
    ip INET,
    user_agent TEXT,
    expira_em TIMESTAMPTZ NOT NULL,
    revogado BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sessoes_refresh ON sessoes_admin(refresh_token) WHERE NOT revogado;

INSERT INTO planos (nome, descricao, preco, duracao_minutos, velocidade_down, velocidade_up, recomendado, ordem)
VALUES
    ('Acesso 24 Horas', 'Um dia completo de internet', 15.00, 1440, 10, 5, TRUE, 1),
    ('Acesso 1 Hora', 'Internet rapida para o essencial', 5.00, 60, 5, 2, FALSE, 2)
ON CONFLICT DO NOTHING;

INSERT INTO system_settings (chave, valor, tipo, descricao) VALUES
    ('hotspot_nome', 'Astrolink Wi-Fi', 'string', 'Nome exibido no portal'),
    ('hotspot_logo_url', '', 'string', 'URL ou path da logo'),
    ('cor_primaria', '#06B6D4', 'string', 'Cor primaria'),
    ('cor_fundo', '#0F172A', 'string', 'Cor de fundo'),
    ('url_pos_conexao', 'https://google.com', 'string', 'Redirecionar apos login'),
    ('coleta_nome', 'false', 'boolean', 'Pedir nome no portal'),
    ('mp_access_token', '', 'string', 'Mercado Pago Access Token'),
    ('mp_public_key', '', 'string', 'Mercado Pago Public Key')
ON CONFLICT (chave) DO NOTHING;

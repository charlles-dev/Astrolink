CREATE TABLE IF NOT EXISTS admin_login_failures (
    id BIGSERIAL PRIMARY KEY,
    usuario TEXT NOT NULL,
    ip INET,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_admin_login_failures_identity_created
    ON admin_login_failures (usuario, ip, created_at DESC);

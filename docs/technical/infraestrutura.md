# Infraestrutura — Deploy e Operações

## Arquitetura de Deploy

### Nó Local (Self-Hosted)

```
┌─────────────────────────────────────────────────────┐
│              Servidor do Nó (mini PC / VPS)          │
│                                                     │
│  ┌────────────────────────────────────────────────┐ │
│  │                  Docker Compose                 │ │
│  │                                                 │ │
│  │  ┌──────────────┐  ┌──────────────────────┐   │ │
│  │  │ astrolink-   │  │      PostgreSQL 15    │   │ │
│  │  │ node (Go)    │  │      + pgBouncer      │   │ │
│  │  │ :5000        │  │      :5432/:6432      │   │ │
│  │  └──────┬───────┘  └──────────────────────┘   │ │
│  │         │          ┌──────────────────────┐   │ │
│  │         │          │       Redis 7         │   │ │
│  │         │          │       :6379           │   │ │
│  │         │          └──────────────────────┘   │ │
│  │         │          ┌──────────────────────┐   │ │
│  │         └─────────►│   Nginx (opcional)   │   │ │
│  │                    │   :80/:443           │   │ │
│  │                    └──────────────────────┘   │ │
│  └────────────────────────────────────────────────┘ │
│                                                     │
│  Volume: /data/astrolink/ (banco, logs, backups)    │
└─────────────────────────────────────────────────────┘
```

### Docker Compose — Produção

```yaml
# docker-compose.yml
version: '3.9'

services:
  node:
    image: ghcr.io/astrolink/node:latest
    container_name: astrolink-node
    restart: unless-stopped
    ports:
      - "5000:5000"
    environment:
      - DATABASE_URL=postgres://astrolink:${DB_PASSWORD}@pgbouncer:6432/astrolink
      - REDIS_URL=redis://:${REDIS_PASSWORD}@redis:6379
      - AMQP_URL=${AMQP_URL}
      - JWT_SECRET=${JWT_SECRET}
      - MP_ACCESS_TOKEN=${MP_ACCESS_TOKEN}
      - NODE_NAME=${NODE_NAME}
      - CLOUD_TOKEN=${CLOUD_TOKEN}
      - TZ=${TIMEZONE:-America/Sao_Paulo}
    volumes:
      - ./data/uploads:/app/uploads
      - ./data/backups:/app/backups
      - ./data/ssh-keys:/app/ssh-keys:ro
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:5000/api/saude"]
      interval: 30s
      timeout: 10s
      retries: 3

  postgres:
    image: postgres:15-alpine
    container_name: astrolink-postgres
    restart: unless-stopped
    environment:
      - POSTGRES_DB=astrolink
      - POSTGRES_USER=astrolink
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - ./data/postgres:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d:ro
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U astrolink"]
      interval: 10s
      timeout: 5s
      retries: 5

  pgbouncer:
    image: bitnami/pgbouncer:latest
    container_name: astrolink-pgbouncer
    restart: unless-stopped
    environment:
      - POSTGRESQL_HOST=postgres
      - POSTGRESQL_PORT=5432
      - POSTGRESQL_DATABASE=astrolink
      - POSTGRESQL_USERNAME=astrolink
      - POSTGRESQL_PASSWORD=${DB_PASSWORD}
      - PGBOUNCER_POOL_MODE=transaction
      - PGBOUNCER_MAX_CLIENT_CONN=200
      - PGBOUNCER_DEFAULT_POOL_SIZE=20
    depends_on:
      - postgres

  redis:
    image: redis:7-alpine
    container_name: astrolink-redis
    restart: unless-stopped
    command: redis-server --requirepass ${REDIS_PASSWORD} --maxmemory 256mb --maxmemory-policy allkeys-lru
    volumes:
      - ./data/redis:/data
    healthcheck:
      test: ["CMD", "redis-cli", "--no-auth-warning", "-a", "${REDIS_PASSWORD}", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Opcional: apenas se RabbitMQ for self-hosted no nó
  # rabbitmq:
  #   image: rabbitmq:3-management-alpine
  #   ...
```

---

### Arquivo `.env` de Produção

```bash
# .env — Nó Local
NODE_NAME=parauapebas-01
TIMEZONE=America/Belem

# Banco de dados
DB_PASSWORD=senha_muito_segura_aqui

# Redis
REDIS_PASSWORD=outra_senha_segura

# JWT (gerar com: openssl rand -hex 32)
JWT_SECRET=hex_aleatorio_de_64_chars

# Mercado Pago
MP_ACCESS_TOKEN=APP_USR-XXXX-XXXX-XXXX
MP_PUBLIC_KEY=APP_USR-XXXX-XXXX-XXXX
MP_WEBHOOK_SECRET=secret_para_validar_webhooks

# Cloud Panel (opcional)
CLOUD_TOKEN=ASTRO-XXXXXXXX-XXXXXXXXXXXX
AMQP_URL=amqps://user:pass@broker.cloudamqp.com/vhost

# Admin
ADMIN_USUARIO=admin
ADMIN_SENHA_HASH=bcrypt_hash_aqui

# Ambiente
GO_ENV=production
LOG_LEVEL=info
```

---

## Requisitos de Hardware

### Configuração Mínima (para desenvolvimento e testes)
- CPU: 1 core (ARM Cortex-A53 ou superior)
- RAM: 1GB
- Disco: 8GB
- Exemplos: Raspberry Pi 3, Orange Pi Zero

### Configuração Recomendada (produção — até 100 usuários simultâneos)
- CPU: 2 cores (x86_64 ou ARM Cortex-A55+)
- RAM: 2GB
- Disco: 32GB SSD
- Exemplos: Raspberry Pi 4 2GB, Orange Pi 5, mini PC Intel N100

### Configuração Para Alto Volume (100–500 usuários simultâneos)
- CPU: 4+ cores
- RAM: 4–8GB
- Disco: 64GB+ SSD NVMe
- Exemplos: mini PC Intel N5105, VPS 4 vCPU

### Modo Embedded (no próprio roteador)
- Requer roteador com ≥ 128MB RAM e ≥ 256MB flash
- Exemplos: GL.iNet GL-MT3000, GL.iNet GL-AX1800, alguns TP-Link AX

---

## Instalação Rápida (5 minutos)

```bash
# 1. Instalar Docker (Ubuntu/Debian)
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# 2. Baixar Astrolink
curl -sSL https://install.astrolink.app | sh
# Ou manualmente:
git clone https://github.com/astrolink/node.git astrolink
cd astrolink

# 3. Configurar
cp .env.example .env
nano .env  # editar variáveis

# 4. Iniciar
docker compose up -d

# 5. Verificar
docker compose ps
curl http://localhost:5000/api/saude

# 6. Aplicar migrations
docker compose exec node ./astrolink migrate up

# Acessar painel admin
# http://[ip-do-servidor]:5000/admin
```

---

## Nginx (SSL + Reverse Proxy)

Se quiser HTTPS no Nó local (recomendado para produção em VPS):

```nginx
# /etc/nginx/sites-available/astrolink
server {
    listen 80;
    server_name hotspot.meudominio.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name hotspot.meudominio.com;

    ssl_certificate /etc/letsencrypt/live/hotspot.meudominio.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/hotspot.meudominio.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;

    # Segurança
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header Referrer-Policy strict-origin-when-cross-origin;

    location / {
        proxy_pass http://127.0.0.1:5000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        # SSE
        proxy_buffering off;
        proxy_cache off;
    }

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=portal:10m rate=30r/m;
    location /api/pix/gerar {
        limit_req zone=portal burst=5 nodelay;
        proxy_pass http://127.0.0.1:5000;
    }
}
```

```bash
# Instalar Certbot e gerar certificado
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d hotspot.meudominio.com
```

---

## Systemd (alternativa ao Docker)

Para quem prefere rodar o binário Go diretamente:

```ini
# /etc/systemd/system/astrolink.service
[Unit]
Description=Astrolink Node
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=astrolink
WorkingDirectory=/opt/astrolink
ExecStart=/opt/astrolink/astrolink-node
Restart=always
RestartSec=5
EnvironmentFile=/opt/astrolink/.env

# Segurança
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ReadWritePaths=/opt/astrolink/data

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable astrolink
sudo systemctl start astrolink
sudo systemctl status astrolink
```

---

## Monitoramento

### Prometheus + Grafana (opcional, via docker-compose)

```yaml
# Adicionar ao docker-compose.yml
prometheus:
  image: prom/prometheus:latest
  volumes:
    - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
  ports:
    - "9090:9090"

grafana:
  image: grafana/grafana:latest
  ports:
    - "3000:3000"
  volumes:
    - ./monitoring/dashboards:/var/lib/grafana/dashboards
```

**Métricas expostas pelo backend Go em `/metrics`:**
- `astrolink_usuarios_ativos` — gauge
- `astrolink_requisicoes_total{endpoint, status}` — counter
- `astrolink_pagamentos_total{status}` — counter
- `astrolink_pagamentos_valor_total` — gauge (receita)
- `astrolink_roteadores_online` — gauge
- `astrolink_jobs_agendados_duration_seconds` — histogram

---

## Backup Automático

```bash
#!/bin/bash
# /opt/astrolink/scripts/backup.sh
# Agendado no cron: 0 3 * * * /opt/astrolink/scripts/backup.sh

BACKUP_DIR="/data/astrolink/backups"
DATE=$(date +%Y%m%d_%H%M%S)
FILENAME="astrolink_backup_${DATE}.sql.gz"
RETER=7 # dias

# Criar backup
docker exec astrolink-postgres pg_dump -U astrolink astrolink | \
    gzip > "${BACKUP_DIR}/${FILENAME}"

# Upload para S3/Rclone (opcional)
# rclone copy "${BACKUP_DIR}/${FILENAME}" remote:astrolink-backups/

# Limpar backups antigos
find "${BACKUP_DIR}" -name "*.sql.gz" -mtime +${RETER} -delete

echo "Backup concluído: ${FILENAME}"
```

---

## Atualização de Versão

```bash
# Atualização com zero downtime (usando Docker)
docker compose pull node
docker compose up -d --no-deps --build node
docker compose exec node ./astrolink migrate up

# Verificar
curl http://localhost:5000/api/saude
```

---

## CI/CD (GitHub Actions)

```yaml
# .github/workflows/deploy.yml
name: Deploy Node

on:
  push:
    tags: ['v*']

jobs:
  build-push:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Login to GHCR
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: ./node
          platforms: linux/amd64,linux/arm64,linux/arm/v7
          push: true
          tags: |
            ghcr.io/astrolink/node:latest
            ghcr.io/astrolink/node:${{ github.ref_name }}

  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: test
          POSTGRES_DB: astrolink_test

    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: cd node && go test ./...
```

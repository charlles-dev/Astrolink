# Astrolink Rebuild Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild Astrolink around the documentation in `docs/`, replacing the half-finished FastAPI/React shape with the documented Go node, local portal/admin APIs, migrations, and development workflow.

**Architecture:** The new core lives in `node/` as a Go service with `internal/api`, `internal/domain`, `internal/infra`, `internal/network`, `internal/scheduler`, and `internal/sync`. The first pass uses an in-memory store so API contracts, tests, and frontend integration can stabilize before Postgres repositories are wired in.

**Tech Stack:** Go 1.22+, Fiber, PostgreSQL migrations, Docker Compose, pnpm/SvelteKit for the next frontend phase.

---

### Task 1: Documentation And Repo Shape

**Files:**
- Modify: `docs/**`
- Create: `docs/superpowers/plans/2026-05-21-astrolink-rebuild.md`

- [x] **Step 1: Copy the provided documentation into the repository**

Run:
```powershell
Copy-Item -Path C:\Users\charl\Downloads\docs\docs\* -Destination docs -Recurse -Force
```

Expected: `docs/specs`, `docs/technical`, `docs/dev`, and `docs/business` exist in the repository.

### Task 2: Go Node Foundation

**Files:**
- Create: `node/go.mod`
- Create: `node/cmd/server/main.go`
- Create: `node/internal/config/config.go`
- Create: `node/internal/api/server.go`
- Create: `node/internal/api/portal/handlers.go`
- Create: `node/internal/api/admin/handlers.go`
- Create: `node/internal/domain/planos/plano.go`
- Create: `node/internal/domain/vouchers/voucher.go`
- Create: `node/internal/domain/vouchers/voucher_test.go`
- Create: `node/internal/infra/memory/store.go`

- [x] **Step 1: Add the module and HTTP server**

Use module `github.com/astrolink/node`, Fiber for HTTP, and a single `api.NewServer` entry point that registers portal and admin routes.

- [x] **Step 2: Add domain behavior first**

Implement voucher validation and code generation with tests for valid, used, expired, inactive, and duplicate-resistant generated codes.

- [x] **Step 3: Add MVP API routes**

Implement `GET /api/saude`, `GET /api/settings`, `GET /api/planos`, `GET /api/sessao/status`, `POST /api/pix/gerar`, `GET /api/pix/status/:txid`, `POST /api/voucher/resgatar`, `POST /admin/auth/login`, `GET /admin/sistema/saude`, `GET /admin/planos`, and `GET /admin/usuarios`.

### Task 3: Database And Dev Workflow

**Files:**
- Create: `node/migrations/000001_initial_schema.up.sql`
- Create: `node/migrations/000001_initial_schema.down.sql`
- Create: `.env.example`
- Create: `docker-compose.dev.yml`
- Modify: `docker-compose.yml`
- Create: `Makefile`
- Modify: `.gitignore`
- Modify: `README.md`

- [x] **Step 1: Add local schema migration**

Mirror the documented local tables: `planos`, `usuarios_mac`, `transacoes_pix`, `voucher_lotes`, `vouchers`, `voucher_usos`, `roteadores`, `blacklist_mac`, `walled_garden`, `system_settings`, `logs`, and `sessoes_admin`.

- [x] **Step 2: Add dev commands**

Add `make install`, `make dev-infra`, `make dev`, `make migrate`, `make test`, `make build`, and `make clean`.

- [x] **Step 3: Update README**

Replace the old Python-first instructions with the documented Go-first setup and clearly call out legacy folders that still need migration.

### Task 4: Verification

**Files:**
- Verify: `node/**`
- Verify: root workflow files

- [x] **Step 1: Run Go tests**

Run:
```powershell
cd node
go test ./...
```

Expected: all tests pass.

- [x] **Step 2: Build the server**

Run:
```powershell
cd node
go build ./cmd/server
```

Expected: the Go server builds without errors.

### Task 5: Local Postgres Store

**Files:**
- Create: `node/internal/store/store.go`
- Create: `node/internal/infra/postgres/open.go`
- Create: `node/internal/infra/postgres/store.go`
- Create: `node/internal/infra/postgres/store_test.go`
- Create: `node/internal/api/portal/handlers_test.go`
- Modify: `node/internal/infra/memory/store.go`
- Modify: `node/internal/api/server.go`
- Modify: `node/internal/api/portal/handlers.go`
- Modify: `node/internal/api/admin/handlers.go`
- Modify: `node/cmd/server/main.go`
- Modify: `node/internal/config/config.go`
- Modify: `node/go.mod`
- Modify: `node/go.sum`

- [x] **Step 1: Write failing tests for store decoupling and Postgres behavior**

Run:
```powershell
cd node
go test ./...
```

Expected before implementation: FAIL because `internal/store`, `internal/infra/postgres`, and `sqlmock` are missing.

- [x] **Step 2: Extract shared store contracts**

Create `internal/store` with DTOs, errors, and the `Store` interface used by API handlers.

- [x] **Step 3: Implement Postgres store**

Implement settings, planos, usuarios, sessao status, PIX creation/status, voucher redemption, and database health methods using `database/sql` + pgx.

- [x] **Step 4: Wire runtime selection**

Use Postgres when `DATABASE_URL` is configured and reachable; fall back to memory with a warning if unavailable.

- [x] **Step 5: Verify**

Run:
```powershell
cd node
go test ./...
go build ./cmd/server
```

Expected: PASS and build exit 0.

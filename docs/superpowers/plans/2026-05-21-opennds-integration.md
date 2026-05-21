# OpenNDS Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Wire the local Go node to OpenWrt/OpenNDS so successful access grants can run `ndsctl auth` and admin disconnects can run `ndsctl deauth`.

**Architecture:** Add a focused `internal/gateway` package with a small controller interface, a no-op implementation for local development, an OpenNDS command controller, and an SSH command runner. The API layer depends only on the controller interface; store/database code remains the source of truth for sessions and vouchers.

**Tech Stack:** Go 1.22, Fiber, `golang.org/x/crypto/ssh`, existing in-memory/Postgres stores, table-driven Go tests.

---

### Task 1: Gateway Command Controller

**Files:**
- Create: `node/internal/gateway/gateway.go`
- Create: `node/internal/gateway/opennds.go`
- Create: `node/internal/gateway/opennds_test.go`

- [x] **Step 1: Write failing command-format tests**

Create tests that call `Authorize`, `Deauthorize`, and invalid MAC paths:

```go
func TestOpenNDSController_Authorize_FormatsNDSAuthCommand(t *testing.T) {
	runner := &recordingRunner{}
	controller := gateway.NewOpenNDSController(runner, gateway.OpenNDSOptions{Retries: 1})
	err := controller.Authorize(context.Background(), gateway.Authorization{
		MAC: "aa:bb:cc:dd:ee:ff", Duration: 24 * time.Hour, DownloadMB: 100, UploadMB: 50,
	})
	if err != nil { t.Fatal(err) }
	want := "ndsctl auth AA:BB:CC:DD:EE:FF 86400 104857600 52428800"
	if runner.commands[0] != want { t.Fatalf("command = %q, want %q", runner.commands[0], want) }
}
```

- [x] **Step 2: Run test to verify RED**

Run: `cd node && go test ./internal/gateway`

Expected: FAIL because `internal/gateway` does not exist.

- [x] **Step 3: Implement gateway controller**

Define:

```go
type Controller interface {
	Authorize(context.Context, Authorization) error
	Deauthorize(context.Context, string) error
	Ping(context.Context) (time.Duration, error)
}
type CommandRunner interface { Run(context.Context, string) (string, error) }
```

Add `NoopController`, `OpenNDSController`, MAC normalization, MB-to-byte conversion, and bounded retries.

- [x] **Step 4: Run gateway tests**

Run: `cd node && go test ./internal/gateway`

Expected: PASS.

### Task 2: SSH Runner And Config

**Files:**
- Create: `node/internal/gateway/ssh.go`
- Modify: `node/internal/config/config.go`
- Modify: `docs/dev/setup-local.md`

- [x] **Step 1: Add configuration fields**

Add env-backed fields:

```go
OpenNDSEnabled bool
OpenNDSHost string
OpenNDSPort int
OpenNDSUser string
OpenNDSKeyPath string
OpenNDSTimeout time.Duration
OpenNDSRetries int
```

- [x] **Step 2: Implement SSH runner**

Implement `SSHRunner.Run(ctx, command)` using `ssh.Dial`, private key auth, context timeout, and combined output for better diagnostics.

- [x] **Step 3: Document local env**

Document that OpenNDS is disabled by default and can be enabled with:

```env
OPENNDS_ENABLED=true
OPENNDS_SSH_HOST=192.168.1.1
OPENNDS_SSH_PORT=22
OPENNDS_SSH_USER=root
OPENNDS_SSH_KEY_PATH=C:\Users\charl\.ssh\id_ed25519
OPENNDS_AUTH_RETRIES=3
```

### Task 3: API Wiring

**Files:**
- Modify: `node/cmd/server/main.go`
- Modify: `node/internal/api/server.go`
- Modify: `node/internal/api/portal/handlers.go`
- Modify: `node/internal/api/portal/handlers_test.go`
- Modify: `node/internal/api/admin/handlers.go`
- Create: `node/internal/api/admin/handlers_test.go`

- [x] **Step 1: Write failing portal test for router authorization**

Extend voucher success test with a fake gateway and assert it receives MAC, duration, and quota after `RedeemVoucher`.

- [x] **Step 2: Wire gateway into server dependencies**

Add `Gateway gateway.Controller` to API dependencies. If nil, use `gateway.NoopController{}`.

- [x] **Step 3: Authorize after voucher success**

After voucher redemption succeeds, call:

```go
duration := time.Duration(minutes) * time.Minute
err := deps.Gateway.Authorize(ctx, gateway.Authorization{MAC: result.Usuario.MAC, Duration: duration})
```

Do not consume admin cloud scope. If authorization fails after the store committed the voucher, log the gateway error and preserve the successful portal response with `"roteador_autorizado": false`.

- [x] **Step 4: Add admin disconnect endpoint**

Add `POST /admin/usuarios/:mac/desconectar` to call `Gateway.Deauthorize(ctx, mac)` and return `{ "sucesso": true }` on success.

- [x] **Step 5: Run API tests**

Run: `cd node && go test ./internal/api/...`

Expected: PASS.

### Task 4: Verification

**Files:**
- Verify: `node/**`
- Verify: `portal/**`

- [x] **Step 1: Run full Go tests**

Run: `cd node && go test ./...`

Expected: PASS.

- [x] **Step 2: Run Go build**

Run: `cd node && go build ./cmd/server`

Expected: PASS. Remove generated `server.exe` afterwards on Windows.

- [x] **Step 3: Run portal smoke checks**

Run:

```powershell
cd portal
npm test
npm run check
npm run build
```

Expected: PASS.


# Local Setup Wizard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a local-first setup flow so non-technical operators can configure personal secrets and local integrations without manually editing `.env` or depending on Supabase.

**Architecture:** Use a small Go env-file layer with an allowlist of configurable keys, then expose two setup surfaces: a local CLI wizard that safely writes `.env`, and an authenticated admin setup panel that shows redacted status and can write env changes only when explicitly enabled for local installs. The running node still reads configuration from env at startup; changes that affect runtime require a restart message instead of live mutation.

**Tech Stack:** Go 1.22, Fiber, local filesystem, SvelteKit, Vitest, svelte-check.

---

## Boundaries And Security Rules

- Do not store Mercado Pago, JWT, TOTP, or SSH secrets in browser localStorage/sessionStorage.
- Do not return full secret values from any HTTP endpoint after saving.
- Only allow writes for an explicit allowlist of env keys.
- Do not expose a generic `.env` editor.
- Web-based writes must be protected by admin auth and disabled unless `ASTROLINK_ALLOW_ENV_WRITE=true`.
- CLI setup is the preferred setup path for first install because it runs locally and does not expose secrets over HTTP.
- After changing `.env`, return a clear `requires_restart: true` status.

## Config Groups

### Required Local Node

- `NODE_NAME`
- `HTTP_ADDR`
- `JWT_SECRET`
- `ADMIN_USUARIO`
- `ADMIN_SENHA`
- `ADMIN_TOTP_SECRET`

### Database

- `DATABASE_URL`
- `DB_PASSWORD`
- `POSTGRES_PORT`

### Payments

- `PAYMENTS_PROVIDER`
- `MERCADOPAGO_ACCESS_TOKEN`
- `MERCADOPAGO_API_BASE_URL`
- `MERCADOPAGO_PAYER_EMAIL`
- `MERCADOPAGO_WEBHOOK_SECRET`

### Router

- `OPENNDS_ENABLED`
- `OPENNDS_SSH_HOST`
- `OPENNDS_SSH_PORT`
- `OPENNDS_SSH_USER`
- `OPENNDS_SSH_KEY_PATH`
- `OPENNDS_SSH_TIMEOUT`
- `OPENNDS_AUTH_RETRIES`

## File Structure

- Create `node/internal/config/envfile.go`: parse, update, and atomically write `.env` files while preserving comments and unknown lines.
- Create `node/internal/config/envfile_test.go`: parser/writer tests for comments, quoted values, blank values, and append behavior.
- Create `node/internal/config/setup.go`: allowlisted setup schema, redaction, status computation, and env patch generation.
- Create `node/internal/config/setup_test.go`: setup status, redaction, validation, and allowlist tests.
- Create `node/cmd/setup/main.go`: local terminal setup wizard for creating/updating `.env`.
- Create `node/internal/api/admin/setup.go`: authenticated setup status/update endpoints.
- Modify `node/internal/api/admin/handlers.go`: register setup routes behind existing admin middleware.
- Modify `node/internal/config/config.go`: add `ASTROLINK_ENV_FILE` and `ASTROLINK_ALLOW_ENV_WRITE`.
- Modify `.env.example`: add setup env toggles with safe defaults.
- Modify `portal/src/lib/types.ts`: add setup status/update types.
- Modify `portal/src/lib/api.ts`: add setup API methods.
- Create `portal/src/lib/components/admin/AdminSetupPanel.svelte`: local setup panel with masked fields and restart notice.
- Modify `portal/src/routes/painel/+page.svelte`: include the setup panel in the admin route.
- Create `portal/src/lib/components/admin/AdminSetupPanel.test.ts`: form/status tests.
- Modify docs: `docs/dev/setup-local.md`, `docs/Documentação do Painel Administrativo.md`, and this roadmap.

---

## Task 1: Env File Parser And Writer

**Files:**
- Create: `node/internal/config/envfile.go`
- Create: `node/internal/config/envfile_test.go`

- [ ] **Step 1: Write parser/writer tests**

Create `node/internal/config/envfile_test.go`:

```go
package config

import (
	"strings"
	"testing"
)

func TestParseEnvFilePreservesCommentsAndUpdatesValues(t *testing.T) {
	input := []byte("# Astrolink\nPAYMENTS_PROVIDER=demo\nMERCADOPAGO_ACCESS_TOKEN=\"old-token\"\n\nUNKNOWN=value\n")

	file, err := ParseEnvFile(input)
	if err != nil {
		t.Fatalf("ParseEnvFile() error = %v", err)
	}
	file.Set("MERCADOPAGO_ACCESS_TOKEN", "new token")
	file.Set("MERCADOPAGO_PAYER_EMAIL", "cliente@example.com")

	got := string(file.Bytes())
	for _, want := range []string{
		"# Astrolink",
		"PAYMENTS_PROVIDER=demo",
		"MERCADOPAGO_ACCESS_TOKEN=\"new token\"",
		"UNKNOWN=value",
		"MERCADOPAGO_PAYER_EMAIL=cliente@example.com",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q:\n%s", want, got)
		}
	}
}

func TestParseEnvFileReadsBlankAndUnquotedValues(t *testing.T) {
	file, err := ParseEnvFile([]byte("ADMIN_TOTP_SECRET=\nOPENNDS_ENABLED=false\n"))
	if err != nil {
		t.Fatalf("ParseEnvFile() error = %v", err)
	}
	if got := file.Get("ADMIN_TOTP_SECRET"); got != "" {
		t.Fatalf("ADMIN_TOTP_SECRET = %q, want empty", got)
	}
	if got := file.Get("OPENNDS_ENABLED"); got != "false" {
		t.Fatalf("OPENNDS_ENABLED = %q, want false", got)
	}
}
```

- [ ] **Step 2: Run failing tests**

Run:

```powershell
go test ./internal/config -run EnvFile -count=1
```

Expected: FAIL because `ParseEnvFile` does not exist.

- [ ] **Step 3: Implement env file support**

Create `node/internal/config/envfile.go`:

```go
package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type EnvFile struct {
	lines []envLine
	index map[string]int
}

type envLine struct {
	key   string
	value string
	raw   string
	kind  envLineKind
}

type envLineKind int

const (
	envLineRaw envLineKind = iota
	envLineKeyValue
)

func ParseEnvFile(data []byte) (*EnvFile, error) {
	result := &EnvFile{index: map[string]int{}}
	for _, raw := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		line := envLine{raw: raw, kind: envLineRaw}
		trimmed := strings.TrimSpace(raw)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			key, value, ok := strings.Cut(raw, "=")
			if ok {
				line.kind = envLineKeyValue
				line.key = strings.TrimSpace(key)
				line.value = unquoteEnvValue(strings.TrimSpace(value))
				result.index[line.key] = len(result.lines)
			}
		}
		result.lines = append(result.lines, line)
	}
	return result, nil
}

func LoadEnvFile(path string) (*EnvFile, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return ParseEnvFile(nil)
	}
	if err != nil {
		return nil, err
	}
	return ParseEnvFile(data)
}

func SaveEnvFileAtomic(path string, file *EnvFile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, file.Bytes(), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (f *EnvFile) Get(key string) string {
	if f == nil || f.index == nil {
		return ""
	}
	if idx, ok := f.index[key]; ok {
		return f.lines[idx].value
	}
	return ""
}

func (f *EnvFile) Set(key, value string) {
	if f.index == nil {
		f.index = map[string]int{}
	}
	if idx, ok := f.index[key]; ok {
		f.lines[idx] = envLine{kind: envLineKeyValue, key: key, value: value}
		return
	}
	f.index[key] = len(f.lines)
	f.lines = append(f.lines, envLine{kind: envLineKeyValue, key: key, value: value})
}

func (f *EnvFile) Bytes() []byte {
	var out bytes.Buffer
	for _, line := range f.lines {
		switch line.kind {
		case envLineKeyValue:
			_, _ = fmt.Fprintf(&out, "%s=%s\n", line.key, quoteEnvValue(line.value))
		default:
			if line.raw != "" {
				out.WriteString(line.raw)
			}
			out.WriteByte('\n')
		}
	}
	return out.Bytes()
}

func quoteEnvValue(value string) string {
	if value == "" {
		return ""
	}
	if strings.ContainsAny(value, " \t#\"") {
		return `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
	}
	return value
}

func unquoteEnvValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 && strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
		return strings.ReplaceAll(value[1:len(value)-1], `\"`, `"`)
	}
	return value
}

func sortedEnvKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
```

- [ ] **Step 4: Run tests**

Run:

```powershell
go test ./internal/config -run EnvFile -count=1
```

Expected: PASS.

---

## Task 2: Setup Schema, Redaction, And Env Patch

**Files:**
- Create: `node/internal/config/setup.go`
- Create/modify: `node/internal/config/setup_test.go`
- Modify: `node/internal/config/config.go`
- Modify: `.env.example`

- [ ] **Step 1: Add config flags**

Modify `node/internal/config/config.go` with:

```go
const (
	EnvAstrolinkEnvFile       = "ASTROLINK_ENV_FILE"
	EnvAstrolinkAllowEnvWrite = "ASTROLINK_ALLOW_ENV_WRITE"
)

type Config struct {
	// existing fields...
	AstrolinkEnvFile       string
	AstrolinkAllowEnvWrite bool
}
```

Inside `FromEnv()`:

```go
AstrolinkEnvFile:       env(EnvAstrolinkEnvFile, ".env"),
AstrolinkAllowEnvWrite: parseBool(env(EnvAstrolinkAllowEnvWrite, "false")),
```

Modify `.env.example`:

```env
ASTROLINK_ALLOW_ENV_WRITE=false
```

- [ ] **Step 2: Write setup tests**

Create `node/internal/config/setup_test.go`:

```go
package config

import "testing"

func TestSetupStatusRedactsSecrets(t *testing.T) {
	file, _ := ParseEnvFile([]byte("PAYMENTS_PROVIDER=mercadopago\nMERCADOPAGO_ACCESS_TOKEN=secret-token\nMERCADOPAGO_PAYER_EMAIL=cliente@example.com\n"))

	status := BuildSetupStatus(file)

	if status.Groups["payments"].Fields["MERCADOPAGO_ACCESS_TOKEN"].Configured != true {
		t.Fatal("access token should be configured")
	}
	if status.Groups["payments"].Fields["MERCADOPAGO_ACCESS_TOKEN"].Value != "" {
		t.Fatal("secret value should not be exposed")
	}
	if status.Groups["payments"].Fields["MERCADOPAGO_PAYER_EMAIL"].Value != "cliente@example.com" {
		t.Fatalf("payer email value = %q", status.Groups["payments"].Fields["MERCADOPAGO_PAYER_EMAIL"].Value)
	}
}

func TestApplySetupPatchRejectsUnknownKeys(t *testing.T) {
	file, _ := ParseEnvFile(nil)

	err := ApplySetupPatch(file, map[string]string{"SHELL": "powershell"})

	if err == nil {
		t.Fatal("ApplySetupPatch() error = nil, want error")
	}
}
```

- [ ] **Step 3: Implement setup schema**

Create `node/internal/config/setup.go`:

```go
package config

import (
	"fmt"
	"strings"
)

type SetupStatus struct {
	RequiresRestart bool                  `json:"requires_restart"`
	Groups          map[string]SetupGroup `json:"groups"`
}

type SetupGroup struct {
	Label  string                `json:"label"`
	Fields map[string]SetupField `json:"fields"`
}

type SetupField struct {
	Key        string `json:"key"`
	Label      string `json:"label"`
	Value      string `json:"value,omitempty"`
	Configured bool   `json:"configured"`
	Secret     bool   `json:"secret"`
	Required   bool   `json:"required"`
}

type setupDefinition struct {
	Group    string
	Label    string
	Key      string
	Secret   bool
	Required bool
}

var setupDefinitions = []setupDefinition{
	{Group: "node", Label: "Nome do no", Key: "NODE_NAME", Required: true},
	{Group: "node", Label: "Endereco HTTP", Key: "HTTP_ADDR", Required: true},
	{Group: "node", Label: "JWT secret", Key: "JWT_SECRET", Secret: true, Required: true},
	{Group: "admin", Label: "Usuario admin", Key: "ADMIN_USUARIO", Required: true},
	{Group: "admin", Label: "Senha admin", Key: "ADMIN_SENHA", Secret: true, Required: true},
	{Group: "admin", Label: "Secret TOTP", Key: "ADMIN_TOTP_SECRET", Secret: true},
	{Group: "payments", Label: "Provider", Key: "PAYMENTS_PROVIDER", Required: true},
	{Group: "payments", Label: "Access token Mercado Pago", Key: "MERCADOPAGO_ACCESS_TOKEN", Secret: true},
	{Group: "payments", Label: "API base Mercado Pago", Key: "MERCADOPAGO_API_BASE_URL"},
	{Group: "payments", Label: "Email pagador Mercado Pago", Key: "MERCADOPAGO_PAYER_EMAIL"},
	{Group: "payments", Label: "Webhook secret Mercado Pago", Key: "MERCADOPAGO_WEBHOOK_SECRET", Secret: true},
	{Group: "router", Label: "OpenNDS habilitado", Key: "OPENNDS_ENABLED"},
	{Group: "router", Label: "Host SSH OpenNDS", Key: "OPENNDS_SSH_HOST"},
	{Group: "router", Label: "Porta SSH OpenNDS", Key: "OPENNDS_SSH_PORT"},
	{Group: "router", Label: "Usuario SSH OpenNDS", Key: "OPENNDS_SSH_USER"},
	{Group: "router", Label: "Chave SSH OpenNDS", Key: "OPENNDS_SSH_KEY_PATH", Secret: true},
	{Group: "router", Label: "Timeout SSH OpenNDS", Key: "OPENNDS_SSH_TIMEOUT"},
	{Group: "router", Label: "Tentativas auth OpenNDS", Key: "OPENNDS_AUTH_RETRIES"},
}

var setupGroupLabels = map[string]string{
	"node":     "No local",
	"admin":    "Admin local",
	"payments": "Pagamentos",
	"router":   "Roteador",
}

func BuildSetupStatus(file *EnvFile) SetupStatus {
	status := SetupStatus{Groups: map[string]SetupGroup{}}
	for _, def := range setupDefinitions {
		group := status.Groups[def.Group]
		if group.Fields == nil {
			group = SetupGroup{Label: setupGroupLabels[def.Group], Fields: map[string]SetupField{}}
		}
		value := strings.TrimSpace(file.Get(def.Key))
		field := SetupField{
			Key:        def.Key,
			Label:      def.Label,
			Configured: value != "",
			Secret:     def.Secret,
			Required:   def.Required,
		}
		if !def.Secret {
			field.Value = value
		}
		group.Fields[def.Key] = field
		status.Groups[def.Group] = group
	}
	return status
}

func ApplySetupPatch(file *EnvFile, values map[string]string) error {
	allowed := map[string]bool{}
	for _, def := range setupDefinitions {
		allowed[def.Key] = true
	}
	for _, key := range sortedEnvKeys(values) {
		if !allowed[key] {
			return fmt.Errorf("campo de setup nao permitido: %s", key)
		}
		file.Set(key, values[key])
	}
	return nil
}
```

- [ ] **Step 4: Run setup tests**

Run:

```powershell
go test ./internal/config -run "Setup|FromEnv" -count=1
```

Expected: PASS.

---

## Task 3: Local Setup CLI

**Files:**
- Create: `node/cmd/setup/main.go`
- Modify: `docs/dev/setup-local.md`

- [ ] **Step 1: Create CLI command**

Create `node/cmd/setup/main.go`:

```go
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/astrolink/node/internal/config"
)

func main() {
	envPath := flag.String("env", ".env", "caminho do arquivo .env")
	flag.Parse()

	file, err := config.LoadEnvFile(*envPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro ao ler %s: %v\n", *envPath, err)
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	values := map[string]string{}
	prompt(reader, values, "NODE_NAME", file.Get("NODE_NAME"), "Nome do no")
	prompt(reader, values, "ADMIN_USUARIO", defaultValue(file.Get("ADMIN_USUARIO"), "admin"), "Usuario admin")
	promptSecret(reader, values, "ADMIN_SENHA", file.Get("ADMIN_SENHA"), "Senha admin")
	prompt(reader, values, "PAYMENTS_PROVIDER", defaultValue(file.Get("PAYMENTS_PROVIDER"), "demo"), "Provider de pagamentos (demo/mercadopago)")
	promptSecret(reader, values, "MERCADOPAGO_ACCESS_TOKEN", file.Get("MERCADOPAGO_ACCESS_TOKEN"), "Access token Mercado Pago")
	prompt(reader, values, "MERCADOPAGO_PAYER_EMAIL", file.Get("MERCADOPAGO_PAYER_EMAIL"), "Email pagador Mercado Pago")
	promptSecret(reader, values, "MERCADOPAGO_WEBHOOK_SECRET", file.Get("MERCADOPAGO_WEBHOOK_SECRET"), "Webhook secret Mercado Pago")
	prompt(reader, values, "OPENNDS_ENABLED", defaultValue(file.Get("OPENNDS_ENABLED"), "false"), "OpenNDS habilitado (true/false)")
	prompt(reader, values, "OPENNDS_SSH_HOST", file.Get("OPENNDS_SSH_HOST"), "Host SSH OpenNDS")
	prompt(reader, values, "OPENNDS_SSH_KEY_PATH", file.Get("OPENNDS_SSH_KEY_PATH"), "Caminho da chave SSH OpenNDS")

	if err := config.ApplySetupPatch(file, values); err != nil {
		fmt.Fprintf(os.Stderr, "erro de validacao: %v\n", err)
		os.Exit(1)
	}
	if err := config.SaveEnvFileAtomic(*envPath, file); err != nil {
		fmt.Fprintf(os.Stderr, "erro ao salvar %s: %v\n", *envPath, err)
		os.Exit(1)
	}
	fmt.Printf("Configuracao salva em %s. Reinicie o node para aplicar alteracoes.\n", *envPath)
}

func prompt(reader *bufio.Reader, values map[string]string, key, current, label string) {
	fmt.Printf("%s [%s]: ", label, current)
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if text == "" {
		text = current
	}
	values[key] = text
}

func promptSecret(reader *bufio.Reader, values map[string]string, key, current, label string) {
	display := ""
	if strings.TrimSpace(current) != "" {
		display = "configurado"
	}
	fmt.Printf("%s [%s]: ", label, display)
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if text == "" {
		text = current
	}
	values[key] = text
}

func defaultValue(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
```

- [ ] **Step 2: Run CLI manually against a temp file**

Run:

```powershell
Copy-Item .env.example .env.setup-test
go run ./cmd/setup -env ../.env.setup-test
Remove-Item ../.env.setup-test
```

Expected: command prompts fields, writes `.env.setup-test`, and prints restart notice.

- [ ] **Step 3: Document CLI**

Add to `docs/dev/setup-local.md`:

```md
## Setup Guiado

Para configurar segredos locais sem editar `.env` manualmente:

```powershell
cd node
go run ./cmd/setup -env ../.env
```

O wizard mascara segredos existentes e salva apenas chaves permitidas. Reinicie
o backend depois de alterar `.env`.
```

---

## Task 4: Protected Setup API

**Files:**
- Create: `node/internal/api/admin/setup.go`
- Modify: `node/internal/api/admin/handlers.go`
- Test: `node/internal/api/admin/handlers_test.go`

- [ ] **Step 1: Write admin API tests**

Add tests to `node/internal/api/admin/handlers_test.go`:

```go
func TestSetupStatus_RedactsSecrets(t *testing.T) {
	app := fiber.New()
	cfg := testConfig()
	cfg.AstrolinkEnvFile = tempEnvFile(t, "MERCADOPAGO_ACCESS_TOKEN=secret-token\n")
	admin.Register(app, admin.Dependencies{Config: cfg, Store: &fakeStore{}, Gateway: &fakeGateway{}})
	token := loginAndGetToken(t, app)

	req := httptest.NewRequest("GET", "/admin/setup/status", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if strings.Contains(string(body), "secret-token") {
		t.Fatalf("setup status leaked secret: %s", string(body))
	}
}
```

- [ ] **Step 2: Implement setup API**

Create `node/internal/api/admin/setup.go`:

```go
package admin

import (
	"github.com/astrolink/node/internal/config"
	"github.com/gofiber/fiber/v2"
)

func setupStatusHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		file, err := config.LoadEnvFile(deps.Config.AstrolinkEnvFile)
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "setup_indisponivel", "erro ao ler configuracao local")
		}
		return c.JSON(config.BuildSetupStatus(file))
	}
}

func setupUpdateHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !deps.Config.AstrolinkAllowEnvWrite {
			return adminError(c, fiber.StatusForbidden, "setup_escrita_desabilitada", "escrita de .env desabilitada")
		}
		var body struct {
			Values map[string]string `json:"values"`
		}
		if err := c.BodyParser(&body); err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", "JSON invalido")
		}
		file, err := config.LoadEnvFile(deps.Config.AstrolinkEnvFile)
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "setup_indisponivel", "erro ao ler configuracao local")
		}
		if err := config.ApplySetupPatch(file, body.Values); err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", err.Error())
		}
		if err := config.SaveEnvFileAtomic(deps.Config.AstrolinkEnvFile, file); err != nil {
			return adminError(c, fiber.StatusInternalServerError, "setup_indisponivel", "erro ao salvar configuracao local")
		}
		status := config.BuildSetupStatus(file)
		status.RequiresRestart = true
		return c.JSON(status)
	}
}
```

Modify admin route registration:

```go
admin.Get("/setup/status", setupStatusHandler(deps))
admin.Put("/setup/env", setupUpdateHandler(deps))
```

- [ ] **Step 3: Run admin tests**

Run:

```powershell
go test ./internal/api/admin -run Setup -count=1
```

Expected: PASS.

---

## Task 5: Admin Setup Panel

**Files:**
- Modify: `portal/src/lib/types.ts`
- Modify: `portal/src/lib/api.ts`
- Create: `portal/src/lib/components/admin/AdminSetupPanel.svelte`
- Create: `portal/src/lib/components/admin/AdminSetupPanel.test.ts`
- Modify: `portal/src/routes/painel/+page.svelte`

- [ ] **Step 1: Add types and API client**

Add to `portal/src/lib/types.ts`:

```ts
export interface SetupField {
  key: string
  label: string
  value?: string
  configured: boolean
  secret: boolean
  required: boolean
}

export interface SetupGroup {
  label: string
  fields: Record<string, SetupField>
}

export interface SetupStatus {
  requires_restart: boolean
  groups: Record<string, SetupGroup>
}
```

Add to `portal/src/lib/api.ts`:

```ts
getSetupStatus: (token: string) =>
  request<SetupStatus>('GET', '/admin/setup/status', undefined, { token }),
updateSetupEnv: (token: string, values: Record<string, string>) =>
  request<SetupStatus>('PUT', '/admin/setup/env', { values }, { token }),
```

- [ ] **Step 2: Create setup panel**

Create `portal/src/lib/components/admin/AdminSetupPanel.svelte`:

```svelte
<script lang="ts">
  import type { SetupStatus } from '$lib/types'

  export let status: SetupStatus | null = null
  export let saving = false
  export let error = ''
  export let message = ''
  export let onSave: (values: Record<string, string>) => Promise<void>

  let values: Record<string, string> = {}

  $: groups = status ? Object.entries(status.groups) : []

  async function submit() {
    await onSave(values)
    values = {}
  }
</script>

<section class="admin-panel setup-panel">
  <div class="panel-heading">
    <div>
      <p class="eyebrow">Setup local</p>
      <h2>Configuracoes do no</h2>
    </div>
    <button type="button" class="secondary" disabled={saving} on:click={submit}>
      {saving ? 'Salvando...' : 'Salvar configuracoes'}
    </button>
  </div>

  {#if message}
    <p class="success-message">{message}</p>
  {/if}
  {#if error}
    <p class="error-message">{error}</p>
  {/if}
  {#if status?.requires_restart}
    <p class="warning-message">Reinicie o node para aplicar as alteracoes.</p>
  {/if}

  {#each groups as [groupKey, group]}
    <fieldset>
      <legend>{group.label}</legend>
      {#each Object.values(group.fields) as field}
        <label>
          <span>{field.label}</span>
          <input
            type={field.secret ? 'password' : 'text'}
            placeholder={field.secret && field.configured ? 'Configurado' : field.key}
            value={values[field.key] ?? field.value ?? ''}
            on:input={(event) => (values[field.key] = event.currentTarget.value)}
          />
        </label>
      {/each}
    </fieldset>
  {/each}
</section>
```

- [ ] **Step 3: Add focused component tests**

Create `portal/src/lib/components/admin/AdminSetupPanel.test.ts`:

```ts
import { fireEvent, render, screen } from '@testing-library/svelte'
import { describe, expect, it, vi } from 'vitest'
import AdminSetupPanel from './AdminSetupPanel.svelte'

describe('AdminSetupPanel', () => {
  it('shows configured secret without exposing its value', () => {
    render(AdminSetupPanel, {
      status: {
        requires_restart: false,
        groups: {
          payments: {
            label: 'Pagamentos',
            fields: {
              MERCADOPAGO_ACCESS_TOKEN: {
                key: 'MERCADOPAGO_ACCESS_TOKEN',
                label: 'Access token Mercado Pago',
                configured: true,
                secret: true,
                required: false
              }
            }
          }
        }
      },
      onSave: vi.fn()
    })

    expect(screen.getByPlaceholderText('Configurado')).toBeInTheDocument()
    expect(screen.queryByDisplayValue('secret-token')).not.toBeInTheDocument()
  })

  it('sends changed values on save', async () => {
    const onSave = vi.fn()
    render(AdminSetupPanel, {
      status: {
        requires_restart: false,
        groups: {
          payments: {
            label: 'Pagamentos',
            fields: {
              MERCADOPAGO_PAYER_EMAIL: {
                key: 'MERCADOPAGO_PAYER_EMAIL',
                label: 'Email pagador Mercado Pago',
                value: '',
                configured: false,
                secret: false,
                required: false
              }
            }
          }
        }
      },
      onSave
    })

    await fireEvent.input(screen.getByLabelText('Email pagador Mercado Pago'), {
      target: { value: 'cliente@example.com' }
    })
    await fireEvent.click(screen.getByRole('button', { name: 'Salvar configuracoes' }))

    expect(onSave).toHaveBeenCalledWith({ MERCADOPAGO_PAYER_EMAIL: 'cliente@example.com' })
  })
})
```

- [ ] **Step 4: Wire panel into `/painel`**

Modify `portal/src/routes/painel/+page.svelte` to load `api.getSetupStatus(token)` after login and pass `api.updateSetupEnv(token, values)` to `AdminSetupPanel`.

- [ ] **Step 5: Run frontend checks**

Run:

```powershell
npm test -- src/lib/components/admin/AdminSetupPanel.test.ts
npm run check
```

Expected: PASS.

---

## Task 6: Full Verification And Docs

**Files:**
- Modify: `docs/dev/setup-local.md`
- Modify: `docs/Documentação do Painel Administrativo.md`
- Modify: `docs/technical/api-reference.md`
- Modify: `docs/superpowers/plans/2026-05-21-admin-local-parallel-roadmap.md`

- [ ] **Step 1: Document setup surfaces**

Document:

- CLI command: `go run ./cmd/setup -env ../.env`
- Admin endpoints: `GET /admin/setup/status`, `PUT /admin/setup/env`
- Security guard: `ASTROLINK_ALLOW_ENV_WRITE=false` by default
- Restart requirement after `.env` changes

- [ ] **Step 2: Run full verification**

Run:

```powershell
cd node
go test ./... -count=1
cd ..\portal
npm test
npm run check
npm run build
cd ..
git diff --check
```

Expected:

- Go tests pass.
- Vitest passes.
- `svelte-check` reports 0 errors and 0 warnings.
- Vite build succeeds.
- `git diff --check` reports no whitespace errors.

- [ ] **Step 3: Browser verification**

Open:

```text
http://127.0.0.1:5173/painel
```

Verify:

- Login still works.
- Setup panel renders.
- Secret fields show configured state without revealing secrets.
- Saving while `ASTROLINK_ALLOW_ENV_WRITE=false` shows a clear disabled-write error.
- Saving while enabled returns restart notice.

- [ ] **Step 4: Commit**

Run:

```powershell
git add -A
git commit -m "feat: add local setup wizard"
```

---

## Execution Recommendation

Use subagents in parallel:

- Agent A owns envfile/setup schema and CLI (`node/internal/config`, `node/cmd/setup`).
- Agent B owns backend admin setup API and tests (`node/internal/api/admin`).
- Agent C owns frontend setup panel and tests (`portal/src/lib/*`, `/painel` route).
- Coordinator integrates docs, resolves shared type/API mismatches, runs full verification, and commits.

# Estratégia de Testes

## Filosofia

**Testar comportamento, não implementação.** Testes devem verificar o que o sistema faz, não como faz. Isso torna os testes resilientes a refatorações.

**Pirâmide de testes:**
```
        /\
       /  \    E2E (poucos, lentos, caros)
      /────\
     /      \  Integração (quantidade média)
    /────────\
   /          \ Unitários (muitos, rápidos, baratos)
  /────────────\
```

---

## Testes Go (Backend)

### Estrutura de Arquivos

```
node/
├── internal/
│   ├── domain/
│   │   ├── vouchers/
│   │   │   ├── voucher.go
│   │   │   └── voucher_test.go     ← testes unitários
│   │   └── sessoes/
│   │       ├── sessao.go
│   │       └── sessao_test.go
│   ├── api/
│   │   ├── portal/
│   │   │   ├── handlers.go
│   │   │   └── handlers_test.go    ← testes de integração da API
│   └── infra/
│       └── db/
│           └── voucher_repo_test.go ← testes com banco real (testcontainers)
└── testdata/                        ← fixtures e seeds de teste
    └── fixtures.sql
```

---

### Testes Unitários (domain/)

```go
// internal/domain/vouchers/voucher_test.go
package vouchers_test

import (
    "testing"
    "time"
    "github.com/astrolink/node/internal/domain/vouchers"
)

func TestValidarVoucher(t *testing.T) {
    t.Run("voucher_valido", func(t *testing.T) {
        v := &vouchers.Voucher{
            Codigo:     "ABCD-1234",
            Ativo:      true,
            Tipo:       "single_use",
            UsosAtuais: 0,
            UsosMaximos: ptr(1),
            ValidadeEm: nil,
        }

        err := v.Validar()
        if err != nil {
            t.Errorf("esperava nil, got %v", err)
        }
    })

    t.Run("voucher_ja_usado", func(t *testing.T) {
        v := &vouchers.Voucher{
            Ativo:       true,
            Tipo:        "single_use",
            UsosAtuais:  1,
            UsosMaximos: ptr(1),
        }

        err := v.Validar()
        if !errors.Is(err, vouchers.ErrJaUtilizado) {
            t.Errorf("esperava ErrJaUtilizado, got %v", err)
        }
    })

    t.Run("voucher_expirado", func(t *testing.T) {
        v := &vouchers.Voucher{
            Ativo:      true,
            ValidadeEm: ptr(time.Now().Add(-24 * time.Hour)), // ontem
        }

        err := v.Validar()
        if !errors.Is(err, vouchers.ErrExpirado) {
            t.Errorf("esperava ErrExpirado, got %v", err)
        }
    })

    t.Run("voucher_inativo", func(t *testing.T) {
        v := &vouchers.Voucher{Ativo: false}
        err := v.Validar()
        if !errors.Is(err, vouchers.ErrInativo) {
            t.Errorf("esperava ErrInativo, got %v", err)
        }
    })
}

func TestGerarCodigo(t *testing.T) {
    codigo := vouchers.GerarCodigo("VIP")

    // Verificar formato: VIPXXX-XXXX
    if !strings.HasPrefix(codigo, "VIP") {
        t.Errorf("código deve começar com 'VIP', got %q", codigo)
    }

    parts := strings.Split(codigo, "-")
    if len(parts) != 2 {
        t.Errorf("código deve ter 2 partes separadas por '-', got %q", codigo)
    }

    // Verificar unicidade (gerar 1000 e verificar duplicatas)
    codigos := make(map[string]bool)
    for i := 0; i < 1000; i++ {
        c := vouchers.GerarCodigo("")
        if codigos[c] {
            t.Errorf("código duplicado gerado: %q", c)
        }
        codigos[c] = true
    }
}

// Helper
func ptr[T any](v T) *T { return &v }
```

---

### Testes de Integração da API (com banco real)

Usando `testcontainers-go` para subir PostgreSQL em Docker durante os testes:

```go
// internal/api/portal/handlers_test.go
package portal_test

import (
    "context"
    "encoding/json"
    "net/http/httptest"
    "testing"

    "github.com/testcontainers/testcontainers-go/modules/postgres"
    "github.com/gofiber/fiber/v2"
)

// TestMain: setup e teardown de toda a suite
func TestMain(m *testing.M) {
    ctx := context.Background()

    // Subir PostgreSQL em container
    pgContainer, err := postgres.Run(ctx,
        "postgres:15-alpine",
        postgres.WithDatabase("astrolink_test"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        postgres.WithInitScripts("../../../testdata/fixtures.sql"),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer pgContainer.Terminate(ctx)

    // Configurar variável de ambiente para os testes
    connStr, _ := pgContainer.ConnectionString(ctx, "sslmode=disable")
    os.Setenv("DATABASE_URL", connStr)

    os.Exit(m.Run())
}

func TestGetPlanos(t *testing.T) {
    app := setupTestApp(t)

    req := httptest.NewRequest("GET", "/api/planos", nil)
    resp, err := app.Test(req, -1)

    if err != nil {
        t.Fatal(err)
    }

    if resp.StatusCode != 200 {
        t.Errorf("esperava 200, got %d", resp.StatusCode)
    }

    var body struct {
        Planos []map[string]any `json:"planos"`
    }
    json.NewDecoder(resp.Body).Decode(&body)

    if len(body.Planos) == 0 {
        t.Error("esperava ao menos 1 plano")
    }
}

func TestResgatarVoucher_Sucesso(t *testing.T) {
    app := setupTestApp(t)

    body := strings.NewReader(`{
        "codigo": "TEST-1234",
        "mac": "AA:BB:CC:DD:EE:FF",
        "ip": "192.168.1.50"
    }`)

    req := httptest.NewRequest("POST", "/api/voucher/resgatar", body)
    req.Header.Set("Content-Type", "application/json")

    resp, _ := app.Test(req, -1)

    if resp.StatusCode != 200 {
        t.Errorf("esperava 200, got %d", resp.StatusCode)
    }
}

func TestResgatarVoucher_CodigoInexistente_Retorna404(t *testing.T) {
    app := setupTestApp(t)

    body := strings.NewReader(`{"codigo": "XXXX-9999", "mac": "AA:BB:CC:DD:EE:FF", "ip": "192.168.1.50"}`)
    req := httptest.NewRequest("POST", "/api/voucher/resgatar", body)
    req.Header.Set("Content-Type", "application/json")

    resp, _ := app.Test(req, -1)

    if resp.StatusCode != 404 {
        t.Errorf("esperava 404, got %d", resp.StatusCode)
    }
}

// Helper
func setupTestApp(t *testing.T) *fiber.App {
    t.Helper()
    // Inicializar app com deps de teste
    db := setupTestDB(t)
    redisCli := setupTestRedis(t)
    return buildApp(db, redisCli, MockNDSManager{})
}
```

---

### Fixtures SQL (testdata/fixtures.sql)

```sql
-- Dados de teste inseridos antes dos testes
INSERT INTO planos (id, nome, preco, duracao_minutos, ativo) VALUES
  (1, 'Acesso 24 Horas', 15.00, 1440, true),
  (2, 'Acesso 1 Hora',    5.00,   60, true),
  (3, 'Plano Inativo',   10.00,  120, false);  -- para testar filtragem

INSERT INTO vouchers (id, codigo, plano_id, tipo, usos_maximos, usos_atuais, ativo) VALUES
  (1, 'TEST-1234', 1, 'single_use', 1, 0, true),   -- válido
  (2, 'USED-5678', 1, 'single_use', 1, 1, true),   -- já usado
  (3, 'INAC-9012', 1, 'single_use', 1, 0, false),  -- inativo
  (4, 'UNIV-0000', 1, 'universal',  10, 3, true);  -- universal
```

---

### Mocks

```go
// internal/network/mock_nds.go
// (apenas para testes — não incluso no build de produção via build tag)
//go:build !integration

package network

type MockNDSManager struct {
    AuthorizedMACs map[string]bool
    Errors         map[string]error
}

func (m *MockNDSManager) AuthorizeMAC(mac string, duration time.Duration, down, up int64) error {
    if err, ok := m.Errors["auth:"+mac]; ok {
        return err
    }
    m.AuthorizedMACs[mac] = true
    return nil
}

func (m *MockNDSManager) DeauthMAC(mac string) error {
    delete(m.AuthorizedMACs, mac)
    return nil
}
```

---

## Testes SvelteKit (Frontend)

### Ferramentas

- **Vitest** — testes unitários de utilities e stores
- **Testing Library (@testing-library/svelte)** — testes de componentes
- **Playwright** — testes E2E

### Testes de Componentes

```typescript
// src/lib/components/PlanCard.test.ts
import { render, fireEvent } from '@testing-library/svelte'
import PlanCard from './PlanCard.svelte'

const mockPlano = {
  id: 1,
  nome: 'Acesso 24 Horas',
  preco: 15.00,
  duracaoMinutos: 1440,
  recomendado: true,
}

test('exibe nome e preço do plano', () => {
  const { getByText } = render(PlanCard, { props: { plano: mockPlano } })

  expect(getByText('Acesso 24 Horas')).toBeInTheDocument()
  expect(getByText('R$ 15,00')).toBeInTheDocument()
})

test('exibe badge recomendado quando marcado', () => {
  const { getByText } = render(PlanCard, { props: { plano: mockPlano } })
  expect(getByText('RECOMENDADO')).toBeInTheDocument()
})

test('chama onSelect ao clicar', async () => {
  const handleSelect = vi.fn()
  const { getByRole } = render(PlanCard, {
    props: { plano: mockPlano, onSelect: handleSelect }
  })

  await fireEvent.click(getByRole('button'))
  expect(handleSelect).toHaveBeenCalledWith(1)
})
```

### Testes E2E (Playwright)

```typescript
// e2e/portal-flow.spec.ts
import { test, expect } from '@playwright/test'

test.describe('Fluxo do portal cativo', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/?mac=AA:BB:CC:DD:EE:FF&ip=192.168.1.50&token=test')
  })

  test('exibe tela de boas-vindas', async ({ page }) => {
    await expect(page.getByText('Bem-vindo')).toBeVisible()
    await expect(page.getByRole('button', { name: /ver planos/i })).toBeVisible()
  })

  test('navega para seleção de planos', async ({ page }) => {
    await page.getByRole('button', { name: /ver planos/i }).click()
    await expect(page.getByText('Acesso 24 Horas')).toBeVisible()
    await expect(page.getByText('R$ 15,00')).toBeVisible()
  })

  test('fluxo completo de voucher', async ({ page }) => {
    await page.getByText(/tem voucher/i).click()
    await page.getByPlaceholder(/código/i).fill('TEST-1234')
    await page.getByRole('button', { name: /resgatar/i }).click()
    await expect(page.getByText(/acesso liberado/i)).toBeVisible({ timeout: 5000 })
  })
})
```

---

## Coverage

### Metas de Cobertura

| Componente | Meta |
|---|---|
| `internal/domain/` | ≥ 85% |
| `internal/api/` | ≥ 70% |
| `internal/infra/` | ≥ 60% |
| `internal/network/` | ≥ 50% (SSH é difícil de mockar) |
| Frontend components | ≥ 60% |

### Verificar Coverage

```bash
# Go
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Verificar threshold (fail se < 70%)
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total | awk '{print $3}' | \
  awk -F'%' '{if ($1 < 70) exit 1}'

# Frontend (Vitest)
pnpm vitest run --coverage
```

---

## CI Integration

```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  go-test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_DB: astrolink_test
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s

    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true

      - name: Run tests
        run: cd node && go test ./... -race -coverprofile=coverage.out
        env:
          DATABASE_URL: postgres://test:test@localhost:5432/astrolink_test?sslmode=disable

      - name: Check coverage threshold
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
          echo "Coverage: ${COVERAGE}%"
          if (( $(echo "$COVERAGE < 70" | bc -l) )); then
            echo "❌ Coverage abaixo de 70%!"
            exit 1
          fi

  frontend-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v3
      - run: cd portal && pnpm install && pnpm test
      - run: cd admin && pnpm install && pnpm test
```

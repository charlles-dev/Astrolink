# Padrões de Código

## Go (Backend)

### Estrutura de Packages

```
internal/
├── api/            # HTTP handlers — só trata request/response
├── domain/         # Lógica de negócio pura — sem dependências externas
├── infra/          # Implementações de interfaces (DB, cache, queue)
├── network/        # SSH, ndsctl, tc
├── scheduler/      # Jobs agendados
└── sync/           # Agente Cloud

# Regra: domain/ não importa nada de api/, infra/ ou network/
# domain/ define interfaces; infra/ implementa
```

### Naming Conventions

```go
// Pacotes: lowercase, sem underscore
package vouchers       // ✅
package VoucherService // ❌
package voucher_svc    // ❌

// Interfaces: substantivo + "er" ou apenas substantivo
type VoucherRepository interface { ... }  // ✅
type IVoucher interface { ... }           // ❌ (prefixo I é Java, não Go)

// Structs: PascalCase
type Voucher struct { ... }        // ✅
type voucherModel struct { ... }   // ❌ (exported types PascalCase)

// Variáveis e funções: camelCase
func generateCode() string { ... }     // ✅
func Generate_Code() string { ... }    // ❌

// Constantes: PascalCase ou SCREAMING_SNAKE para sentinel values
const MaxVoucherLength = 12            // ✅
const MAX_VOUCHER_LENGTH = 12          // Aceitável, mas prefira PascalCase em Go

// Erros: prefixo Err
var ErrVoucherNotFound = errors.New("voucher não encontrado")   // ✅
var VoucherNotFoundError = errors.New(...)                      // ❌

// Contexto: sempre primeiro parâmetro
func (r *VoucherRepo) GetByCode(ctx context.Context, code string) (*Voucher, error) // ✅
func (r *VoucherRepo) GetByCode(code string, ctx context.Context) (*Voucher, error) // ❌
```

### Error Handling

```go
// ✅ Sempre wrappear erros com contexto
v, err := repo.GetByCode(ctx, code)
if err != nil {
    return nil, fmt.Errorf("buscar voucher %s: %w", code, err)
}

// ✅ Erros de domínio como tipos
type DomainError struct {
    Code    string
    Message string
}

func (e *DomainError) Error() string { return e.Message }

var (
    ErrVoucherJaUsado  = &DomainError{Code: "voucher_ja_usado", Message: "voucher já utilizado"}
    ErrVoucherExpirado = &DomainError{Code: "voucher_expirado", Message: "voucher expirado"}
)

// ✅ Checar tipos de erro com errors.Is / errors.As
if errors.Is(err, ErrVoucherJaUsado) {
    return c.Status(410).JSON(fiber.Map{"erro": "voucher_ja_usado"})
}

// ❌ Nunca ignorar erros
result, _ := someFunc() // NUNCA faça isso (exceto em defer close)
```

### Logging

```go
// Usar slog (stdlib Go 1.21+) com campos estruturados
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: logLevel,
}))

// ✅ Logging com contexto
logger.InfoContext(ctx, "voucher resgatado",
    "codigo", voucher.Codigo,
    "mac", mac,
    "plano", voucher.PlanoNome,
    "duracao_min", voucher.PlanoMinutos,
)

// ❌ Nunca logar dados sensíveis
logger.Info("login", "senha", senha) // NUNCA
logger.Info("mp_token", "token", token) // NUNCA

// ❌ Nunca usar fmt.Println em produção
fmt.Println("pagamento aprovado") // apenas em scripts locais
```

### Testes

```go
// Nomear testes: Test<Func>_<Cenário>_<Resultado>
func TestResgataVoucher_CodigoInvalido_RetornaErro(t *testing.T) { ... }
func TestResgataVoucher_VoucherJaUsado_Retorna410(t *testing.T) { ... }
func TestResgataVoucher_Sucesso_LiberaAcesso(t *testing.T) { ... }

// Usar table-driven tests para múltiplos cenários
func TestValidarMAC(t *testing.T) {
    tests := []struct {
        name    string
        mac     string
        wantErr bool
    }{
        {"valido",          "AA:BB:CC:DD:EE:FF", false},
        {"sem_colons",      "AABBCCDDEEFF",       true},
        {"muito_curto",     "AA:BB:CC",           true},
        {"broadcast",       "FF:FF:FF:FF:FF:FF",  true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateMAC(tt.mac)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateMAC(%q) error = %v, wantErr %v", tt.mac, err, tt.wantErr)
            }
        })
    }
}
```

### Formatação e Lint

```bash
# Formatar código
gofmt -w .

# ou com goimports (recomendado)
goimports -w .

# Lint (golangci-lint)
golangci-lint run
```

**`.golangci.yaml`:**
```yaml
linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - gosec
    - unused
    - misspell

linters-settings:
  govet:
    check-shadowing: true
  errcheck:
    check-type-assertions: true

issues:
  exclude-rules:
    - path: _test\.go
      linters: [gosec]  # testes podem ter dados hardcoded
```

---

## TypeScript / SvelteKit (Frontend)

### Configuração Base

```json
// tsconfig.json
{
  "extends": "./.svelte-kit/tsconfig.json",
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "exactOptionalPropertyTypes": true
  }
}
```

### Conventions

```typescript
// Componentes: PascalCase
// src/lib/components/PlanCard.svelte ✅

// Utilities: camelCase
// src/lib/utils/formatCurrency.ts ✅

// Tipos: PascalCase, sem prefixo I
interface Plano {              // ✅
  id: number
  nome: string
  preco: number
  duracaoMinutos: number | null
}
type PlanoID = number          // ✅
interface IPlano { ... }       // ❌

// Props de componentes: definir com $props()
// PlanCard.svelte
<script lang="ts">
  interface Props {
    plano: Plano
    selecionado?: boolean
    onSelect: (id: number) => void
  }
  const { plano, selecionado = false, onSelect }: Props = $props()
</script>

// Evitar 'any' — use 'unknown' quando necessário
function processData(data: unknown): Plano {  // ✅
  if (!isPlano(data)) throw new Error('dados inválidos')
  return data
}
```

### API Client (portal/src/lib/api.ts)

```typescript
// Centralizar todas as chamadas de API
const BASE_URL = '/api'

class APIError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string
  ) {
    super(message)
  }
}

async function request<T>(
  method: string,
  path: string,
  body?: unknown
): Promise<T> {
  const response = await fetch(`${BASE_URL}${path}`, {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: body ? JSON.stringify(body) : undefined,
  })

  const data = await response.json()

  if (!response.ok) {
    throw new APIError(response.status, data.erro, data.mensagem)
  }

  return data as T
}

export const api = {
  getPlanos: () => request<{ planos: Plano[] }>('GET', '/planos'),
  gerarPIX: (body: GerarPIXBody) => request<GerarPIXResponse>('POST', '/pix/gerar', body),
  resgatarVoucher: (body: ResgatarVoucherBody) => request<ResgatarVoucherResponse>('POST', '/voucher/resgatar', body),
}
```

### Svelte Stores

```typescript
// src/lib/store.ts
import { writable, derived } from 'svelte/store'

// Estado global do portal
export const mac = writable<string>('')
export const ip = writable<string>('')
export const planoSelecionado = writable<Plano | null>(null)
export const step = writable<'boas-vindas' | 'planos' | 'pix' | 'sucesso'>('boas-vindas')
export const sessaoAtiva = writable<SessaoStatus | null>(null)

// Derived: calculado automaticamente
export const tempoRestante = derived(sessaoAtiva, ($sessao) => {
  if (!$sessao?.ativa || !$sessao.fim_acesso) return null
  return Math.max(0, new Date($sessao.fim_acesso).getTime() - Date.now())
})
```

### TailwindCSS

```svelte
<!-- ✅ Classes atômicas, sem estilos inline -->
<div class="flex flex-col gap-4 p-6 rounded-xl bg-slate-800 border border-slate-700">

<!-- ✅ Variantes condicionais com clsx/tailwind-merge -->
<button class={clsx(
  'px-6 py-3 rounded-lg font-semibold transition-colors',
  selecionado
    ? 'bg-cyan-500 text-slate-900'
    : 'bg-slate-700 text-white hover:bg-slate-600'
)}>

<!-- ❌ Evitar style inline quando existir classe Tailwind -->
<div style="color: #06B6D4">  <!-- usar text-cyan-500 -->
```

---

## SQL (Migrations)

```sql
-- ✅ Migrations são sempre up+down
-- migrations/000010_add_planos_ordem.up.sql
ALTER TABLE planos ADD COLUMN ordem INTEGER DEFAULT 0;
CREATE INDEX idx_planos_ordem ON planos(ordem);

-- migrations/000010_add_planos_ordem.down.sql
DROP INDEX IF EXISTS idx_planos_ordem;
ALTER TABLE planos DROP COLUMN IF EXISTS ordem;

-- ✅ Nomes de migration: NNNNNN_descricao_snake_case
-- migrations/000001_create_planos.up.sql
-- migrations/000002_create_usuarios_mac.up.sql
-- migrations/000010_add_planos_ordem.up.sql

-- ✅ Constraints explícitas com nome
ALTER TABLE vouchers
  ADD CONSTRAINT chk_vouchers_tipo
  CHECK (tipo IN ('single_use', 'universal'));

-- ✅ Índices para colunas usadas em WHERE, JOIN, ORDER BY
CREATE INDEX idx_transacoes_mac_status ON transacoes_pix(mac, status);

-- ❌ Nunca DROP sem IF EXISTS em migrations de rollback
DROP TABLE planos;           -- ❌ (vai falhar se não existir)
DROP TABLE IF EXISTS planos; -- ✅
```

---

## Git

### .gitignore

```
# Nunca commitar:
.env
*.env.local
*.env.production
*.pem
*.key
node_modules/
.svelte-kit/
build/
dist/
tmp/
*.sql.gz
*.log
coverage.out
```

### Commit Messages

Sempre seguir Conventional Commits (veja `docs/dev/contribuindo.md`).

### Branch Protection (configuração GitHub)

```yaml
# Branches protegidas: main
rules:
  - require_pull_request_reviews: true
    required_approving_review_count: 1
  - require_status_checks_to_pass:
      - go-test
      - go-lint
      - frontend-build
  - require_linear_history: true
  - restrict_pushes: true  # apenas via PR
```

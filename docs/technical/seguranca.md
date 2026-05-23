# Segurança

## Princípios

1. **Defense in Depth:** múltiplas camadas de proteção
2. **Principle of Least Privilege:** cada componente acessa apenas o que precisa
3. **Zero Trust:** nunca confiar implicitamente, sempre verificar
4. **Audit Everything:** todo acesso e ação é logado

---

## Autenticação e Autorização

### JWT (Admin Local)

```go
// Configuração JWT
const (
    AccessTokenExpiry  = 8 * time.Hour
    RefreshTokenExpiry = 30 * 24 * time.Hour
)

// Claims customizados
type AdminClaims struct {
    UserID   string `json:"uid"`
    Role     string `json:"role"`
    NodeName string `json:"node"`
    jwt.RegisteredClaims
}
```

**Boas práticas implementadas:**
- Algoritmo: HS256 (chave de 256 bits, gerada com `openssl rand -hex 32`)
- Access token: expira em 8 horas, armazenado apenas em memória no cliente
- Refresh token: expira em 30 dias, rotação automática a cada uso
- Revogação: refresh tokens são invalidados no banco ao fazer logout
- Proteção CSRF: não necessária (API JSON sem cookies de sessão)

### Rotação de Refresh Tokens

```go
func (s *AuthService) RefreshTokens(oldRefresh string) (*TokenPair, error) {
    session, err := s.db.GetValidSession(ctx, oldRefresh)
    if err != nil {
        return nil, ErrInvalidRefreshToken
    }

    // Revogar token antigo (rotação — previne replay attacks)
    s.db.RevokeSession(ctx, session.ID)

    // Emitir novo par de tokens
    return s.generateTokenPair(session.UserID)
}
```

### 2FA com TOTP

```go
// Implementação com github.com/pquerna/otp
secret, err := totp.Generate(totp.GenerateOpts{
    Issuer:      "Astrolink",
    AccountName: "admin@" + nodeName,
})

// Verificação
valid := totp.Validate(userInputCode, secret.Secret())
```

---

## Rate Limiting

### Camadas de Rate Limiting

```go
// 1. Por IP global (middleware mais externo)
// Usando redis-based token bucket
limits := map[string]RateLimit{
    "/api/pix/gerar":         {Requests: 5,   Window: time.Minute},
    "/api/voucher/resgatar":  {Requests: 10,  Window: time.Minute},
    "/admin/auth/login":      {Requests: 5,   Window: time.Minute},
    "/admin/*":               {Requests: 300, Window: time.Minute},
    "global":                 {Requests: 100, Window: time.Second},
}

// Implementação Redis
func (r *RateLimiter) Allow(key string, limit RateLimit) bool {
    pipe := r.redis.Pipeline()
    now := time.Now().UnixMilli()
    windowStart := now - limit.Window.Milliseconds()

    // Sliding window com sorted set
    pipe.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(windowStart, 10))
    pipe.ZCard(ctx, key)
    pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
    pipe.Expire(ctx, key, limit.Window)

    results, _ := pipe.Exec(ctx)
    count := results[1].(*redis.IntCmd).Val()

    return count < int64(limit.Requests)
}
```

### Bloqueio por Tentativas de Login

```go
const maxLoginAttempts = 5
const lockoutDuration = 10 * time.Minute

func (s *AuthService) CheckLoginAttempts(ip string) error {
    key := fmt.Sprintf("login_attempts:%s", ip)
    attempts, _ := s.redis.Get(ctx, key).Int()

    if attempts >= maxLoginAttempts {
        ttl, _ := s.redis.TTL(ctx, key).Result()
        return fmt.Errorf("IP bloqueado por %v", ttl.Round(time.Minute))
    }
    return nil
}

func (s *AuthService) RecordFailedLogin(ip string) {
    key := fmt.Sprintf("login_attempts:%s", ip)
    s.redis.Incr(ctx, key)
    s.redis.Expire(ctx, key, lockoutDuration)
}
```

---

## Validação de Webhooks Mercado Pago

```go
func ValidateMercadoPagoWebhook(r *http.Request, secret string) error {
    signature := r.Header.Get("X-Signature")
    requestID := r.Header.Get("X-Request-Id")
    dataID := r.URL.Query().Get("data.id")

    // Construir string para validar
    manifest := fmt.Sprintf("id:%s;request-id:%s;ts:%s;",
        dataID, requestID, extractTimestamp(signature))

    // Calcular HMAC-SHA256
    h := hmac.New(sha256.New, []byte(secret))
    h.Write([]byte(manifest))
    expected := hex.EncodeToString(h.Sum(nil))

    // Extrair hash da assinatura
    received := extractV1Hash(signature)

    if !hmac.Equal([]byte(expected), []byte(received)) {
        return errors.New("assinatura webhook inválida")
    }
    return nil
}
```

---

## Segurança do Banco de Dados

### Usuário PostgreSQL com Permissões Mínimas

```sql
-- Criar usuário da aplicação com permissões mínimas
CREATE USER astrolink_app WITH PASSWORD 'senha_segura';
GRANT CONNECT ON DATABASE astrolink TO astrolink_app;
GRANT USAGE ON SCHEMA public TO astrolink_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO astrolink_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO astrolink_app;

-- Usuário de backup (somente leitura)
CREATE USER astrolink_backup WITH PASSWORD 'outra_senha';
GRANT CONNECT ON DATABASE astrolink TO astrolink_backup;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO astrolink_backup;
```

### Queries Parametrizadas (SQLC)

```go
// NUNCA fazer isso:
db.Query("SELECT * FROM usuarios_mac WHERE mac = '" + mac + "'")

// SEMPRE usar SQLC ou queries parametrizadas:
// query.sql
-- name: GetUsuarioByMAC :one
SELECT * FROM usuarios_mac WHERE mac = $1;

// Código Go gerado automaticamente pelo SQLC:
func (q *Queries) GetUsuarioByMAC(ctx context.Context, mac string) (UsuarioMac, error) {
    row := q.db.QueryRowContext(ctx, getUsuarioByMAC, mac)
    // ...
}
```

---

## Segurança de Rede

### Firewall (iptables) no Servidor do Nó

```bash
#!/bin/bash
# /opt/astrolink/scripts/firewall.sh

# Limpar regras existentes
iptables -F INPUT
iptables -F FORWARD

# Política padrão: DROP
iptables -P INPUT DROP
iptables -P FORWARD DROP

# Permitir tráfego estabelecido
iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT

# Loopback
iptables -A INPUT -i lo -j ACCEPT

# SSH (limitado ao IP de gerência se possível)
iptables -A INPUT -p tcp --dport 22 -j ACCEPT

# Astrolink (porta 5000 — apenas rede local)
iptables -A INPUT -p tcp --dport 5000 -s 192.168.0.0/16 -j ACCEPT

# PostgreSQL e Redis — NUNCA expor externamente
# Acessíveis apenas via Docker network interno

# ICMP (ping)
iptables -A INPUT -p icmp --icmp-type echo-request -j ACCEPT

# Rate limiting anti-DDoS
iptables -A INPUT -p tcp --dport 5000 -m state --state NEW \
    -m recent --set --name portal
iptables -A INPUT -p tcp --dport 5000 -m state --state NEW \
    -m recent --update --seconds 60 --hitcount 100 --name portal -j DROP
```

### SSH Hardening no Roteador

```bash
# /etc/ssh/sshd_config no servidor do nó
PermitRootLogin no          # nunca root via SSH
PasswordAuthentication no   # apenas chaves
MaxAuthTries 3
LoginGraceTime 30
AllowUsers astrolink        # usuário dedicado
```

---

## Segurança do Portal Cativo

### Content Security Policy

```go
// middleware no Go
func CSPMiddleware(c *fiber.Ctx) error {
    c.Set("Content-Security-Policy", strings.Join([]string{
        "default-src 'self'",
        "script-src 'self' 'nonce-"+generateNonce()+"'",
        "style-src 'self' 'unsafe-inline'",  // TailwindCSS inline
        "img-src 'self' data: blob:",
        "connect-src 'self'",
        "frame-ancestors 'none'",
    }, "; "))
    return c.Next()
}
```

### Headers de Segurança

```go
func SecurityHeaders(c *fiber.Ctx) error {
    c.Set("X-Content-Type-Options", "nosniff")
    c.Set("X-Frame-Options", "DENY")
    c.Set("X-XSS-Protection", "1; mode=block")
    c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
    c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
    return c.Next()
}
```

---

## Dados Sensíveis — LGPD

### O que coletamos e por quê

| Dado | Finalidade | Retenção |
|---|---|---|
| MAC address | Autenticação de rede | 90 dias após última sessão |
| IP address | Diagnóstico de rede | 90 dias |
| Nome (opcional) | Personalização | Até solicitação de exclusão |
| Histórico de pagamentos | Obrigação fiscal | 5 anos |
| Logs de acesso | Segurança | 90 dias |

### O que NUNCA coletamos
- Senhas (apenas hash bcrypt)
- Dados de navegação
- Conteúdo de comunicações
- Dados de geolocalização do dispositivo

### Anonimização de IPs nos Logs Públicos

```go
func AnonymizeIP(ip string) string {
    parts := strings.Split(ip, ".")
    if len(parts) == 4 {
        return parts[0] + "." + parts[1] + ".x.x"
    }
    return "x:x:x:x"
}
```

### Endpoint de Exclusão de Dados

```go
// Para conformidade com LGPD, direito ao esquecimento
// POST /api/privacidade/excluir-dados
// Body: { "mac": "AA:BB:CC:DD:EE:FF" }
func HandleDeleteUserData(c *fiber.Ctx) error {
    mac := c.Locals("validated_mac").(string)

    // Anonimizar em vez de deletar (manter integridade referencial)
    db.AnonymizeUserData(ctx, mac)
    // Substitui nome por "Usuário Removido" e trunca dados identificáveis

    return c.JSON(fiber.Map{"mensagem": "Dados removidos com sucesso"})
}
```

---

## Auditoria de Segurança

### Checklist OWASP Top 10 (a verificar antes de cada release)

- [ ] **A01 Broken Access Control:** RLS no Cloud, JWT no Local validado em cada request
- [ ] **A02 Cryptographic Failures:** HTTPS obrigatório, senhas com bcrypt, secrets em env vars
- [ ] **A03 Injection:** SQLC com queries parametrizadas, validação de input com `go-playground/validator`
- [ ] **A04 Insecure Design:** Modelo de ameaças documentado, threat modeling feito
- [ ] **A05 Security Misconfiguration:** docker-compose não expõe DB externamente, `.env` no `.gitignore`
- [ ] **A06 Vulnerable Components:** `govulncheck ./...` no CI, dependabot ativo
- [ ] **A07 Auth Failures:** Rate limiting em login, 2FA disponível, rotação de tokens
- [ ] **A08 Data Integrity:** Verificação de assinatura em webhooks, integridade de backups
- [ ] **A09 Logging Failures:** Todos os acessos logados, sem dados sensíveis nos logs
- [ ] **A10 SSRF:** URLs de webhook validadas contra lista de CIDRs permitidos

### Ferramentas de Segurança no CI

```yaml
# .github/workflows/security.yml
- name: Go Vulnerability Check
  run: |
    go install golang.org/x/vuln/cmd/govulncheck@latest
    govulncheck ./...

- name: Static Analysis (gosec)
  uses: securego/gosec@master
  with:
    args: ./...

- name: Dependency Audit
  run: |
    go list -json -m all | nancy sleuth
```

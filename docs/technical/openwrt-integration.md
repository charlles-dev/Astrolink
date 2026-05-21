# Integração com OpenWrt e OpenNDS

## Visão Geral

O Astrolink usa o **OpenNDS** (fork moderno do NoDogSplash) como engine de captive portal nos roteadores OpenWrt. O backend Go se comunica com o roteador via **SSH**, executando comandos `ndsctl` para autorizar, revogar e consultar sessões.

---

## Pré-requisitos

### Roteador
- **OpenWrt** 21.02+ (recomendado 22.03 ou 23.05)
- **OpenNDS** 9.0+ (`opkg install nodogsplash` ou `opkg install opennds`)
- **SSH habilitado** com autenticação por chave
- **RAM mínima:** 64MB (recomendado 128MB+)
- **Flash:** sem requisito extra (OpenNDS < 1MB)

### Roteadores Testados
| Modelo | RAM | Flash | Compatível |
|---|---|---|---|
| GL.iNet GL-MT3000 | 512MB | 8GB | ✅ Recomendado |
| GL.iNet GL-AX1800 | 512MB | 8GB | ✅ Recomendado |
| GL.iNet GL-MT300N-V2 | 128MB | 16MB | ✅ Funciona |
| TP-Link Archer C7 | 128MB | 16MB | ✅ Funciona |
| TP-Link TL-WR940N | 64MB | 8MB | ⚠️ Limitado |
| Raspberry Pi (OpenWrt) | 1GB+ | — | ✅ Excelente |

---

## Instalação do OpenNDS

```bash
# Conectar ao roteador via SSH
ssh root@192.168.1.1

# Atualizar pacotes
opkg update

# Instalar OpenNDS
opkg install nodogsplash
# ou (versões mais recentes)
opkg install opennds

# Verificar instalação
ndsctl status
```

---

## Configuração do OpenNDS

O Astrolink configura o OpenNDS automaticamente via SSH ao adicionar um roteador. A configuração manual é:

### `/etc/config/nodogsplash` (OpenWrt UCI format)

```
config nodogsplash
    option enabled 1
    option fwhook_enabled 1

    # Interface do Wi-Fi (ajustar conforme roteador)
    option gatewayinterface 'br-lan'

    # IP do servidor Astrolink (backend Go)
    option remoteauthenticator '192.168.1.100'
    option remoteauthenticatorport '5000'

    # Caminho do endpoint de auth (OpenNDS chama para validar)
    option remoteauthenticatorpath '/api/opennds/auth'

    # Página de redirecionamento (URL do portal cativo)
    # OpenNDS injeta parâmetros: ?mac=AA:BB&ip=192.168&token=xxx
    option splashpage 'http://192.168.1.100:5000/?mac=%h&ip=%i&token=%t'

    # Timeouts
    option preauthidletimeout 30      # minutos no walled garden sem auth
    option authidletimeout 120        # minutos sem atividade após auth (0 = desabilitado)
    option checkinterval 60           # segundos entre verificações

    # Limites de banda global (0 = sem limite)
    option uploadrate 0
    option downloadrate 0

    # Walled Garden — domínios acessíveis sem autenticação
    list walledgarden 'pagamentos.mercadopago.com'
    list walledgarden 'api.mercadopago.com'
    list walledgarden 'secure.mlstatic.com'

    # Firewall — portas abertas para walled garden
    list walledgardenport 'tcp:80'
    list walledgardenport 'tcp:443'
```

### Aplicar configuração

```bash
# Via UCI
uci set nodogsplash.@nodogsplash[0].remoteauthenticator='192.168.1.100'
uci commit nodogsplash

# Reiniciar OpenNDS
/etc/init.d/nodogsplash restart

# Verificar status
ndsctl status
```

---

## Comandos `ndsctl` Usados pelo Astrolink

### Autorizar usuário

```bash
# Liberar acesso por duração (segundos)
ndsctl auth AA:BB:CC:DD:EE:FF 86400 1073741824 1073741824

# Parâmetros:
#   AA:BB:CC:DD:EE:FF   = MAC do dispositivo
#   86400               = duração em segundos (86400 = 24h)
#   1073741824          = download máximo em bytes (1GB, 0 = ilimitado)
#   1073741824          = upload máximo em bytes (1GB, 0 = ilimitado)

# Liberar sem limite de dados
ndsctl auth AA:BB:CC:DD:EE:FF 86400 0 0
```

### Revogar acesso

```bash
ndsctl deauth AA:BB:CC:DD:EE:FF
```

### Listar clientes conectados

```bash
ndsctl clients
# Retorna JSON com todos os clientes autorizados e no walled garden
```

**Exemplo de output:**
```json
{
  "client_count": 3,
  "clients": {
    "AA:BB:CC:DD:EE:FF": {
      "mac": "AA:BB:CC:DD:EE:FF",
      "ip": "192.168.1.50",
      "added": "2025-05-19T14:30:00Z",
      "auth_time": "2025-05-19T14:30:05Z",
      "session_start": "2025-05-19T14:30:05Z",
      "session_end": "2025-05-20T14:30:05Z",
      "state": "Authenticated",
      "download_quota_remaining": 0,
      "upload_quota_remaining": 0,
      "download_this_session": 1267891200,
      "upload_this_session": 245760000
    }
  }
}
```

### Status do daemon

```bash
ndsctl status
```

---

## Implementação no Backend Go

### Cliente SSH

```go
// internal/network/ssh.go
package network

import (
    "fmt"
    "time"
    "golang.org/x/crypto/ssh"
)

type SSHClient struct {
    host    string
    port    int
    user    string
    keyPath string
}

func (c *SSHClient) ExecCommand(cmd string) (string, error) {
    key, err := os.ReadFile(c.keyPath)
    if err != nil {
        return "", fmt.Errorf("ler chave SSH: %w", err)
    }

    signer, err := ssh.ParsePrivateKey(key)
    if err != nil {
        return "", fmt.Errorf("parse chave SSH: %w", err)
    }

    config := &ssh.ClientConfig{
        User:            c.user,
        Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
        HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: usar known_hosts em produção
        Timeout:         10 * time.Second,
    }

    conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", c.host, c.port), config)
    if err != nil {
        return "", fmt.Errorf("conectar SSH: %w", err)
    }
    defer conn.Close()

    session, err := conn.NewSession()
    if err != nil {
        return "", fmt.Errorf("criar sessão SSH: %w", err)
    }
    defer session.Close()

    output, err := session.Output(cmd)
    return string(output), err
}
```

### Gerenciador OpenNDS

```go
// internal/network/opennds.go
package network

import (
    "fmt"
    "time"
)

type OpenNDSManager struct {
    ssh *SSHClient
}

// AuthorizeMAC libera acesso para um MAC address
func (m *OpenNDSManager) AuthorizeMAC(mac string, duration time.Duration, downloadMB, uploadMB int64) error {
    seconds := int64(duration.Seconds())

    var downloadBytes, uploadBytes int64
    if downloadMB > 0 {
        downloadBytes = downloadMB * 1024 * 1024
    }
    if uploadMB > 0 {
        uploadBytes = uploadMB * 1024 * 1024
    }

    cmd := fmt.Sprintf("ndsctl auth %s %d %d %d",
        mac, seconds, downloadBytes, uploadBytes)

    _, err := m.ssh.ExecCommand(cmd)
    if err != nil {
        return fmt.Errorf("ndsctl auth %s: %w", mac, err)
    }

    return nil
}

// DeauthMAC revoga o acesso de um MAC address
func (m *OpenNDSManager) DeauthMAC(mac string) error {
    cmd := fmt.Sprintf("ndsctl deauth %s", mac)
    _, err := m.ssh.ExecCommand(cmd)
    return err
}

// GetClients retorna todos os clientes conectados
func (m *OpenNDSManager) GetClients() ([]NDSClient, error) {
    output, err := m.ssh.ExecCommand("ndsctl clients")
    if err != nil {
        return nil, err
    }
    return parseNDSClients(output)
}

// Ping verifica conectividade com o roteador
func (m *OpenNDSManager) Ping() (time.Duration, error) {
    start := time.Now()
    _, err := m.ssh.ExecCommand("echo pong")
    return time.Since(start), err
}
```

### Limitação de Velocidade (tc/iptables)

```go
// internal/network/traffic_control.go
package network

import "fmt"

// ApplySpeedLimit aplica limite de velocidade por MAC usando tc
func (m *OpenNDSManager) ApplySpeedLimit(mac string, downloadMbps, uploadMbps int) error {
    if downloadMbps == 0 && uploadMbps == 0 {
        return nil // sem limite
    }

    // Converter MAC para formato sem dois-pontos (para nome do qdisc)
    handle := strings.ReplaceAll(mac, ":", "")[:8]

    // Download (ingress no roteador = saída para o cliente)
    if downloadMbps > 0 {
        cmds := []string{
            fmt.Sprintf("tc qdisc add dev br-lan root handle 1: htb default 10"),
            fmt.Sprintf("tc class add dev br-lan parent 1: classid 1:%s htb rate %dmbit ceil %dmbit",
                handle, downloadMbps, downloadMbps),
            fmt.Sprintf("tc filter add dev br-lan parent 1: protocol ip u32 match ether dst %s classid 1:%s",
                mac, handle),
        }
        for _, cmd := range cmds {
            if _, err := m.ssh.ExecCommand(cmd); err != nil {
                return fmt.Errorf("tc download: %w", err)
            }
        }
    }

    return nil
}

// RemoveSpeedLimit remove as regras de tc para o MAC
func (m *OpenNDSManager) RemoveSpeedLimit(mac string) error {
    handle := strings.ReplaceAll(mac, ":", "")[:8]
    cmd := fmt.Sprintf("tc class del dev br-lan classid 1:%s 2>/dev/null || true", handle)
    _, err := m.ssh.ExecCommand(cmd)
    return err
}
```

---

## Retry e Resiliência

```go
// internal/network/retry.go
func WithRetry(attempts int, delay time.Duration, fn func() error) error {
    var lastErr error
    for i := 0; i < attempts; i++ {
        if err := fn(); err != nil {
            lastErr = err
            if i < attempts-1 {
                time.Sleep(delay * time.Duration(i+1)) // backoff linear
            }
            continue
        }
        return nil
    }
    return fmt.Errorf("após %d tentativas: %w", attempts, lastErr)
}

// Uso:
err := WithRetry(3, 2*time.Second, func() error {
    return nds.AuthorizeMAC(mac, duration, 0, 0)
})
```

---

## Scheduler — Verificação de Sessões Expiradas

```go
// internal/scheduler/sessions.go
func (s *Scheduler) checkExpiredSessions() {
    ctx := context.Background()

    // Buscar usuários que expiraram nos últimos 2 minutos
    // (janela de 2min para garantir que não pula nenhum)
    expired, err := s.db.GetExpiredSessions(ctx, 2*time.Minute)
    if err != nil {
        s.log.Error("buscar sessões expiradas", "err", err)
        return
    }

    for _, usuario := range expired {
        // Deauth em todos os roteadores (não sabemos em qual está)
        for _, router := range s.routers {
            router.nds.DeauthMAC(usuario.MAC)
            router.nds.RemoveSpeedLimit(usuario.MAC)
        }

        // Atualizar status no banco
        s.db.UpdateUserStatus(ctx, usuario.MAC, "expirado")

        // Log
        s.log.Info("sessão expirada", "mac", usuario.MAC, "plano", usuario.PlanoNome)

        // Publicar evento no RabbitMQ
        s.amqp.Publish("user.expired", map[string]any{
            "mac":   usuario.MAC,
            "plano": usuario.PlanoNome,
        })
    }
}
```

---

## Modo Embedded (opkg)

Para roteadores com recursos suficientes, o backend Go pode rodar diretamente no roteador:

```bash
# No roteador OpenWrt (ARM)
opkg update
opkg install astrolink-node

# Configurar
uci set astrolink.@astrolink[0].mp_token='APP_USR-XXXX'
uci set astrolink.@astrolink[0].admin_password='senha123'
uci commit astrolink

# Iniciar
/etc/init.d/astrolink start
/etc/init.d/astrolink enable
```

**Processo de build para OpenWrt:**
```bash
# Cross-compile para MIPS (TP-Link antigos)
GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -o astrolink-node ./cmd/server

# Cross-compile para ARM (GL.iNet, Raspberry Pi)
GOOS=linux GOARCH=arm64 go build -o astrolink-node ./cmd/server
GOOS=linux GOARCH=arm GOARM=7 go build -o astrolink-node ./cmd/server
```

O binário resultante tem ~8–12MB, sem dependências externas, roda com 64MB de RAM.

---

## Suporte a Outros Fabricantes

### Ubiquiti UniFi (futura integração)

```go
// A UniFi tem API REST própria para gerenciar clientes
// Autenticação: POST /api/login
// Autorizar: POST /api/s/default/cmd/stamgr (action: authorize-guest)
// Desconectar: POST /api/s/default/cmd/stamgr (action: kick-sta)
```

### TP-Link Omada (futura integração)

```go
// API REST similar à UniFi
// Documentação: https://www.tp-link.com/en/support/download/omada-software-controller/
```

### MikroTik RouterOS (futura integração)

```go
// API binária ou REST (RouterOS 7+)
// Hotspot API nativa do RouterOS
// github.com/go-routeros/routeros
```

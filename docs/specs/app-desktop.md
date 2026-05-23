# Spec: App Desktop Installer (Tauri)

## Visão Geral

Aplicativo desktop para **técnicos e instaladores** realizarem o setup inicial de um novo Nó Astrolink em campo. Elimina a necessidade de abrir terminal, editar arquivos de configuração ou saber usar linha de comando.

**Stack:** Tauri 2 (Rust backend) + SvelteKit (frontend) + TypeScript
**Plataformas:** Windows 10+, macOS 12+, Ubuntu 20.04+
**Bundle:** ~8MB (vs ~150MB do Electron)
**Distribuição:** GitHub Releases, Winget, Homebrew Cask

---

## Casos de Uso

1. **Novo Nó:** Técnico instala Astrolink do zero num mini PC no local do cliente
2. **Atualização:** Atualizar versão do Astrolink mantendo configurações e dados
3. **Diagnóstico:** Técnico acessa remotamente para verificar problema
4. **Migração:** Mover Nó de um servidor para outro preservando banco de dados

---

## Fluxo Principal — Instalação de Novo Nó

```
[Tela de Boas-vindas]
    ↓
[Pré-requisitos]
    ↓
[Conexão com o servidor]
    ↓
[Instalação de dependências]
    ↓
[Configuração do banco]
    ↓
[Configuração de pagamentos]
    ↓
[Adicionar roteadores]
    ↓
[Vincular ao Cloud Panel]
    ↓
[Verificação final]
    ↓
[Concluído!]
```

---

## Telas em Detalhe

### Tela 1 — Boas-vindas

```
┌────────────────────────────────────────────────────┐
│                                                    │
│              🌐 Astrolink Installer                │
│                    v1.2.0                          │
│                                                    │
│  O que você quer fazer?                            │
│                                                    │
│  ┌──────────────────────────────────────────────┐ │
│  │  🆕 Instalar novo Nó                         │ │
│  │  Configurar um servidor do zero              │ │
│  └──────────────────────────────────────────────┘ │
│                                                    │
│  ┌──────────────────────────────────────────────┐ │
│  │  🔄 Atualizar instalação existente           │ │
│  │  Atualizar versão mantendo dados             │ │
│  └──────────────────────────────────────────────┘ │
│                                                    │
│  ┌──────────────────────────────────────────────┐ │
│  │  🔧 Diagnóstico e manutenção                 │ │
│  │  Verificar problemas e corrigir              │ │
│  └──────────────────────────────────────────────┘ │
│                                                    │
└────────────────────────────────────────────────────┘
```

---

### Tela 2 — Pré-requisitos

Checklist automático verificado em tempo real:

```
┌────────────────────────────────────────────────────┐
│ ← Voltar                                           │
│                                                    │
│ Verificando pré-requisitos...                      │
│                                                    │
│ ✅ Servidor Linux detectado (Ubuntu 22.04)         │
│ ✅ Arquitetura: x86_64                             │
│ ✅ RAM: 4GB (mínimo 2GB)                           │
│ ✅ Disco: 28GB livres (mínimo 10GB)                │
│ ✅ Acesso à internet: OK                           │
│ ✅ Porta 5000 livre                                │
│ ⚠️  Docker não instalado (será instalado agora)   │
│ ✅ SSH habilitado                                  │
│                                                    │
│ 7 de 8 requisitos OK                               │
│ Docker será instalado automaticamente.             │
│                                                    │
│                          [Continuar →]             │
└────────────────────────────────────────────────────┘
```

---

### Tela 3 — Conexão com o Servidor

**Se rodando localmente (no próprio servidor):**
```
Modo de instalação:
  ● Instalar neste computador
  ○ Instalar em servidor remoto (SSH)
```

**Se servidor remoto:**
```
┌────────────────────────────────────────────────────┐
│ Conexão com o servidor                             │
│                                                    │
│ Host / IP:  [192.168.1.100              ]          │
│ Porta SSH:  [22    ]                               │
│ Usuário:    [root                       ]          │
│                                                    │
│ Autenticação:                                      │
│   ● Chave SSH  ○ Senha                             │
│   Arquivo: [/Users/joao/.ssh/id_rsa] [Selecionar] │
│                                                    │
│ [Testar conexão]                                   │
│                                                    │
│ ✅ Conectado! Ubuntu 22.04 LTS, 4GB RAM           │
│                                                    │
│                          [Continuar →]             │
└────────────────────────────────────────────────────┘
```

---

### Tela 4 — Instalação de Dependências

```
┌────────────────────────────────────────────────────┐
│ Instalando dependências                            │
│                                                    │
│ ████████████████████░░░░░░  68%                   │
│                                                    │
│ ✅ Docker instalado                                │
│ ✅ Docker Compose instalado                        │
│ ✅ Imagem astrolink/node:1.2.0 baixada            │
│ ⏳ Configurando PostgreSQL...                      │
│ ○ Configurando Redis                               │
│ ○ Configurando serviço systemd                     │
│                                                    │
│ Log:                                               │
│ ┌──────────────────────────────────────────────┐  │
│ │ $ docker pull astrolink/node:1.2.0           │  │
│ │ 1.2.0: Pulling from astrolink/node           │  │
│ │ digest: sha256:abc123...                     │  │
│ │ Status: Image is up to date                  │  │
│ │ $ docker-compose up -d postgres              │  │
│ │ Creating network "astrolink_default"...      │  │
│ └──────────────────────────────────────────────┘  │
│                                                    │
└────────────────────────────────────────────────────┘
```

---

### Tela 5 — Configuração do Nó

```
┌────────────────────────────────────────────────────┐
│ Configurar seu Nó                                  │
│                                                    │
│ Nome do Nó (para identificação):                   │
│ [Parauapebas-01                          ]         │
│                                                    │
│ Cidade / localização:                              │
│ [Parauapebas, PA                         ]         │
│                                                    │
│ Fuso horário:                                      │
│ [America/Belem ▼]                                  │
│                                                    │
│ Senha do painel admin:                             │
│ [••••••••••••••••]  [Gerar senha segura]           │
│ [••••••••••••••••]  (confirmar)                    │
│                                                    │
│ ┌──────────────────────────────────────────────┐  │
│ │ ✅ Senha forte (12 chars, maiúsc, núm, simb) │  │
│ └──────────────────────────────────────────────┘  │
│                                                    │
│                          [Continuar →]             │
└────────────────────────────────────────────────────┘
```

---

### Tela 6 — Configuração de Pagamentos

```
┌────────────────────────────────────────────────────┐
│ Configurar Mercado Pago                            │
│                                                    │
│ Acesse sua conta Mercado Pago e gere um           │
│ Access Token de Produção.                          │
│                                                    │
│ [📖 Como gerar meu Access Token]                  │
│                                                    │
│ Access Token:                                      │
│ [APP_USR-XXXX-XXXX-XXXX-XXXX            ]  [👁]  │
│                                                    │
│ [Verificar credenciais]                            │
│                                                    │
│ ✅ Conectado!                                      │
│    Conta: João Silva (CPF •••.234.•••-56)         │
│    Modo: Produção                                  │
│                                                    │
│ Não tenho conta Mercado Pago ainda:                │
│ [Criar conta gratuita →]                           │
│                                                    │
│                 [Pular por agora]  [Continuar →]   │
└────────────────────────────────────────────────────┘
```

---

### Tela 7 — Adicionar Roteadores OpenWrt

```
┌────────────────────────────────────────────────────┐
│ Configurar Roteadores                              │
│                                                    │
│ Roteadores OpenWrt detectados na rede:             │
│                                                    │
│ ┌──────────────────────────────────────────────┐  │
│ │ ☑ 192.168.1.1 — OpenWrt 22.03, TP-Link     │  │
│ │ ☑ 192.168.1.2 — OpenWrt 22.03, GL.iNet     │  │
│ │ ○ 192.168.1.3 — Outro dispositivo          │  │
│ └──────────────────────────────────────────────┘  │
│                                                    │
│ [+ Adicionar manualmente]                         │
│                                                    │
│ Credenciais SSH para os roteadores:                │
│ Usuário: [root      ]  Senha: [•••••••••••] [👁]  │
│ (ou usar chave SSH: [Selecionar arquivo])          │
│                                                    │
│ [Configurar OpenNDS automaticamente ☑]            │
│                                                    │
│ [Testar e configurar selecionados]                 │
│                                                    │
│ ✅ 192.168.1.1: OpenNDS configurado com sucesso   │
│ ✅ 192.168.1.2: OpenNDS configurado com sucesso   │
│                                                    │
│                          [Continuar →]             │
└────────────────────────────────────────────────────┘
```

O que a configuração automática faz via SSH:
```bash
# Instala OpenNDS se não presente
opkg update && opkg install nodogsplash

# Configura /etc/config/nodogsplash
uci set nodogsplash.@nodogsplash[0].gatewayaddress='192.168.1.1'
uci set nodogsplash.@nodogsplash[0].remoteauthenticator='192.168.1.100:5000'
uci commit nodogsplash
/etc/init.d/nodogsplash restart
```

---

### Tela 8 — Vincular ao Cloud Panel (opcional)

```
┌────────────────────────────────────────────────────┐
│ Vincular ao Cloud Panel                            │
│                                                    │
│ Gerencie este nó remotamente junto com outros     │
│ de qualquer lugar.                                 │
│                                                    │
│ Cole o token gerado no Cloud Panel:                │
│ [ASTRO-XXXXXXXX-XXXXXXXXXXXX              ]        │
│                                                    │
│ [Vincular]                                         │
│                                                    │
│ ✅ Nó vinculado!                                  │
│    Workspace: Provedor XYZ                         │
│    Nó registrado como: Parauapebas-01             │
│    Próxima sync em: 30 segundos                   │
│                                                    │
│                 [Pular por agora]  [Continuar →]   │
└────────────────────────────────────────────────────┘
```

---

### Tela 9 — Verificação Final

```
┌────────────────────────────────────────────────────┐
│ Verificação de saúde do sistema                    │
│                                                    │
│ ✅ Backend Go: rodando (porta 5000)               │
│ ✅ PostgreSQL: rodando e acessível                 │
│ ✅ Redis: rodando                                  │
│ ✅ Portal cativo: acessível                        │
│ ✅ Painel admin: acessível                         │
│ ✅ Roteador 192.168.1.1: respondendo              │
│ ✅ Roteador 192.168.1.2: respondendo              │
│ ✅ Mercado Pago: credenciais válidas              │
│ ✅ Cloud Panel: sincronizado                       │
│                                                    │
│ Teste de pagamento PIX:                            │
│ [Gerar PIX de teste R$ 0,01] → ✅ Aprovado!      │
│                                                    │
│ Tudo pronto! 🎉                                    │
│                                                    │
│                        [Abrir painel admin]        │
└────────────────────────────────────────────────────┘
```

---

## Modo Diagnóstico

Para técnicos verificando um nó com problema:

```
Diagnóstico do Nó — 192.168.1.100

Conectividade:
  ✅ SSH: OK
  ✅ Backend API: OK (230ms)
  ⚠️  Redis: Lento (890ms — esperado < 100ms)
  ✅ PostgreSQL: OK
  ✅ RabbitMQ: Conectado

Processos:
  ✅ astrolink-node (PID 1234) — rodando há 14 dias
  ✅ postgres (PID 567) — rodando
  ✅ redis (PID 890) — rodando
  ⚠️  rabbitmq — Alto uso de memória (78%)

Espaço em disco: 8.2GB / 32GB (26% usado) ✅
Memória: 3.1GB / 4GB (77% usado) ⚠️

Logs recentes (erros):
  [ERRO] 14:15:23 — Timeout SSH roteador 192.168.1.4
  [WARN] 14:10:11 — Redis latência alta

Ações sugeridas:
  → Reiniciar Redis (resolver latência)
  → Verificar roteador 192.168.1.4 (offline)

[Aplicar sugestões automaticamente]
[Baixar relatório completo (.txt)]
```

---

## Atualização de Versão

```
Versão instalada: v1.1.0
Nova versão: v1.3.0

Novidades:
  • Suporte a planos por dados (GB)
  • WhatsApp via Evolution API
  • Dashboard em tempo real melhorado

Tempo estimado: ~3 minutos
O serviço ficará fora do ar por ~30 segundos.

[Fazer backup antes de atualizar ☑]

[Atualizar agora]

Progresso:
  ✅ Backup criado (hotspot_backup_20250519.sql.gz)
  ✅ Nova imagem Docker baixada
  ✅ Banco migrado (5 migrations aplicadas)
  ✅ Serviço reiniciado
  ✅ Verificação de saúde: OK

Atualização concluída! 🎉
```

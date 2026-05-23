# Spec: CLI — astrolink

## Visão Geral

Ferramenta de linha de comando para SysAdmins e usuários técnicos gerenciarem o Astrolink sem abrir o browser. Distribuída como binário estático único, sem dependências.

**Linguagem:** Go (Cobra + Viper)
**Distribuição:** GitHub Releases, Homebrew (macOS/Linux), Scoop (Windows), APT (Debian/Ubuntu)

---

## Instalação

```bash
# macOS / Linux via curl
curl -sSL https://get.astrolink.app/cli | sh

# Homebrew
brew install astrolink/tap/astrolink

# Scoop (Windows)
scoop bucket add astrolink https://github.com/astrolink/scoop-bucket
scoop install astrolink

# Debian/Ubuntu
curl -sSL https://packages.astrolink.app/gpg | sudo apt-key add -
echo "deb https://packages.astrolink.app/apt stable main" | sudo tee /etc/apt/sources.list.d/astrolink.list
sudo apt update && sudo apt install astrolink-cli

# Verificar instalação
astrolink version
# → Astrolink CLI v1.2.0 (linux/amd64, go1.22)
```

---

## Autenticação

```bash
# Login (salva token em ~/.config/astrolink/credentials.json)
astrolink login
# → Email: joao@exemplo.com
# → Senha: ••••••••••
# → ✅ Logado como João Silva (Provedor XYZ)

# Login com token direto
astrolink login --token sk_prod_XXXXXXXXXXXXXX

# Ver conta atual
astrolink whoami
# → João Silva <joao@exemplo.com>
# → Workspace: Provedor XYZ (provedor-xyz)
# → Plano: Pro

# Logout
astrolink logout
```

---

## Referência de Comandos

### `astrolink node` — Gerenciar Nós

```bash
# Listar todos os nós
astrolink node list
# ┌─────────────────┬────────┬───────────┬───────────────┬──────────────────────┐
# │ NOME            │ STATUS │ USUÁRIOS  │ RECEITA HOJE  │ ÚLTIMO HEARTBEAT     │
# ├─────────────────┼────────┼───────────┼───────────────┼──────────────────────┤
# │ Parauapebas-01  │ online │ 23        │ R$ 345,00     │ 5s atrás             │
# │ Marabá-Centro   │ online │ 47        │ R$ 705,00     │ 12s atrás            │
# │ Açailândia-01   │ offline│ 0         │ R$ 0,00       │ 18min atrás ⚠️       │
# └─────────────────┴────────┴───────────┴───────────────┴──────────────────────┘

# Output JSON
astrolink node list --json | jq '.[] | select(.status == "offline")'

# Status detalhado de um nó
astrolink node status parauapebas-01
# Nome:            Parauapebas-01
# Status:          online
# Uptime:          14 dias, 6 horas
# Usuários ativos: 23
# Receita hoje:    R$ 345,00
# Banda:           ↓ 45.2 Mbps / ↑ 8.1 Mbps
# Heartbeat:       5 segundos atrás
# Roteadores:      3 online, 1 offline
# Versão:          v1.2.0
# Cloud sync:      ✅ ativo

# Reiniciar serviço no nó
astrolink node restart parauapebas-01
# → ⏳ Reiniciando Parauapebas-01...
# → ✅ Reiniciado em 4.2s. Verificação de saúde: OK

# Logs em tempo real de um nó
astrolink node logs parauapebas-01
astrolink node logs parauapebas-01 --follow        # stream contínuo
astrolink node logs parauapebas-01 --lines 200
astrolink node logs parauapebas-01 --level error   # apenas erros

# Diagnóstico completo
astrolink node diagnose parauapebas-01
# → Verifica: backend, DB, Redis, SSH roteadores, Mercado Pago
# → Gera relatório em /tmp/astrolink-diag-XXXX.txt
```

---

### `astrolink user` — Gerenciar Usuários

```bash
# Listar usuários conectados agora
astrolink user list --node parauapebas-01
# ┌──────────────────────┬───────────┬──────────────┬────────────────────┐
# │ MAC                  │ PLANO     │ TEMPO REST.  │ BANDA ATUAL        │
# ├──────────────────────┼───────────┼──────────────┼────────────────────┤
# │ AA:BB:CC:DD:EE:FF    │ 24h       │ 18h 32m      │ ↓2.1 ↑0.3 Mbps    │
# │ GG:HH:II:JJ:KK:LL    │ 1h        │ 23m          │ ↓0.8 ↑0.1 Mbps    │
# └──────────────────────┴───────────┴──────────────┴────────────────────┘

# Listar todos (incluindo histórico)
astrolink user list --node parauapebas-01 --all
astrolink user list --node parauapebas-01 --status expired

# Detalhes de um usuário
astrolink user get AA:BB:CC:DD:EE:FF --node parauapebas-01

# Banir MAC address
astrolink user ban AA:BB:CC:DD:EE:FF --node parauapebas-01 --reason "Uso abusivo"
# → ✅ AA:BB:CC:DD:EE:FF banido em Parauapebas-01

# Desbanir
astrolink user unban AA:BB:CC:DD:EE:FF --node parauapebas-01

# Desconectar sessão ativa
astrolink user disconnect AA:BB:CC:DD:EE:FF --node parauapebas-01

# Estender tempo de acesso
astrolink user extend AA:BB:CC:DD:EE:FF --hours 2 --node parauapebas-01
# → ✅ Sessão estendida em 2 horas

# Buscar por todos os nós
astrolink user find AA:BB:CC:DD:EE:FF
```

---

### `astrolink voucher` — Gerenciar Vouchers

```bash
# Listar vouchers
astrolink voucher list --node parauapebas-01
astrolink voucher list --node parauapebas-01 --status available
astrolink voucher list --node parauapebas-01 --plan "24h"

# Gerar vouchers
astrolink voucher generate \
  --node parauapebas-01 \
  --plan "Acesso 24 Horas" \
  --count 50 \
  --expiry 30d \
  --prefix "VIP"
# → ✅ 50 vouchers gerados
# → Códigos: VIPABCD-1234, VIPEFGH-5678, ...

# Exportar para CSV
astrolink voucher generate ... --output vouchers.csv

# Exportar para PDF imprimível
astrolink voucher generate ... --output vouchers.pdf

# Ver detalhes de um voucher
astrolink voucher get ABCD-1234 --node parauapebas-01
# Código:      ABCD-1234
# Plano:       Acesso 24 Horas
# Status:      Disponível
# Criado:      19/05/2025 10:00
# Expira:      18/06/2025 10:00
# Usado por:   —

# Desativar voucher
astrolink voucher disable ABCD-1234 --node parauapebas-01
```

---

### `astrolink payment` — Pagamentos

```bash
# Histórico de pagamentos
astrolink payment list --node parauapebas-01
astrolink payment list --node parauapebas-01 --from 2025-05-01 --to 2025-05-31
astrolink payment list --status approved --json

# Relatório de receita
astrolink payment report --node parauapebas-01 --month 2025-05
# ════════════════════════════════════════
#  Relatório de Receita — Maio 2025
#  Nó: Parauapebas-01
# ════════════════════════════════════════
#  Total:              R$ 6.750,00
#  Transações:               450
#  Ticket médio:           R$ 15,00
#  Melhor dia:     19/05 (R$ 345,00)
# ════════════════════════════════════════
#  Por plano:
#    Acesso 24 Horas:  R$ 5.400,00 (360 transações)
#    Acesso 1 Hora:    R$ 1.350,00 ( 90 transações)
# ════════════════════════════════════════

# Exportar para CSV
astrolink payment report --month 2025-05 --output relatorio.csv

# Relatório consolidado de todos os nós
astrolink payment report --all-nodes --month 2025-05
```

---

### `astrolink plan` — Gerenciar Planos

```bash
# Listar planos
astrolink plan list --node parauapebas-01

# Criar plano
astrolink plan create \
  --node parauapebas-01 \
  --name "Acesso 48 Horas" \
  --price 25.00 \
  --duration 48h \
  --speed-down 10 \
  --speed-up 5
# → ✅ Plano criado (ID: 123)

# Atualizar preço
astrolink plan update 123 --price 22.00 --node parauapebas-01

# Desativar plano
astrolink plan disable 123 --node parauapebas-01
```

---

### `astrolink backup` — Backup e Restauração

```bash
# Baixar backup do banco
astrolink backup download --node parauapebas-01
# → Salvando em: astrolink_parauapebas-01_20250519_143200.sql.gz
# → ✅ 2.3MB baixado

# Backup de todos os nós
astrolink backup download --all-nodes --output-dir ./backups/

# Restaurar backup
astrolink backup restore astrolink_parauapebas-01_20250519_143200.sql.gz \
  --node parauapebas-01
# → ⚠️  ATENÇÃO: Isso sobrescreve TODOS os dados de Parauapebas-01!
# → Confirmar? (sim/não): sim
# → ✅ Restaurado com sucesso

# Listar backups automáticos disponíveis no servidor
astrolink backup list --node parauapebas-01
```

---

### `astrolink network` — Gerenciar Rede

```bash
# Status dos roteadores
astrolink network routers --node parauapebas-01

# Ping de diagnóstico
astrolink network ping parauapebas-01 --router 192.168.1.1

# Blacklist de MACs
astrolink network blacklist list --node parauapebas-01
astrolink network blacklist add AA:BB:CC:DD:EE:FF --node parauapebas-01 --reason "Spam"
astrolink network blacklist remove AA:BB:CC:DD:EE:FF --node parauapebas-01

# Walled Garden
astrolink network walled-garden list --node parauapebas-01
astrolink network walled-garden add "meusite.com.br" --node parauapebas-01

# Teste de velocidade do uplink
astrolink network speedtest --node parauapebas-01
# → ↓ Testando download... 45.2 Mbps
# → ↑ Testando upload... 8.7 Mbps
```

---

### `astrolink upgrade` — Atualizar Software

```bash
# Ver versão atual e disponível
astrolink upgrade check
# Versão atual:       v1.2.0
# Última versão:      v1.3.0 🆕
# Lançado em:         2025-05-15
# Changelog: https://github.com/astrolink/node/releases/tag/v1.3.0

# Atualizar todos os nós
astrolink upgrade --all-nodes
# → [Parauapebas-01] Fazendo backup...
# → [Parauapebas-01] Baixando v1.3.0...
# → [Parauapebas-01] Aplicando migrations...
# → [Parauapebas-01] ✅ Atualizado (30s de downtime)
# → [Marabá-Centro] ...

# Atualizar nó específico
astrolink upgrade --node parauapebas-01 --version v1.3.0

# Rollback
astrolink upgrade rollback --node parauapebas-01
```

---

## Flags Globais

```bash
--json              Saída em JSON (para scripts e pipelines)
--csv               Saída em CSV (para relatórios)
--quiet, -q         Sem output (apenas exit codes)
--verbose, -v       Output detalhado (debug)
--node, -n          Nome do nó alvo
--all-nodes         Aplicar a todos os nós do workspace
--config            Arquivo de configuração customizado
```

---

## Configuração

```toml
# ~/.config/astrolink/config.toml

[default]
node = "parauapebas-01"    # Nó padrão (evita digitar --node sempre)
format = "table"           # table | json | csv
timezone = "America/Belem"

[cloud]
api_url = "https://api.astrolink.app"
```

---

## Exit Codes

| Código | Significado |
|---|---|
| 0 | Sucesso |
| 1 | Erro genérico |
| 2 | Argumento inválido |
| 3 | Não autenticado |
| 4 | Recurso não encontrado |
| 5 | Nó offline / inacessível |
| 6 | Permissão negada |

---

## Uso em Scripts e Automações

```bash
#!/bin/bash
# Checar nós offline e enviar alerta

OFFLINE_NODES=$(astrolink node list --json | jq -r '.[] | select(.status == "offline") | .nome')

if [ -n "$OFFLINE_NODES" ]; then
  echo "Nós offline: $OFFLINE_NODES"
  # Enviar alerta por Telegram, email, etc.
fi

# Backup diário agendado (cron)
# 0 3 * * * /usr/local/bin/astrolink backup download --all-nodes --output-dir /backups/astrolink/

# Gerar relatório mensal
astrolink payment report --all-nodes --month $(date +%Y-%m) --output "/relatorios/$(date +%Y-%m).csv"
```

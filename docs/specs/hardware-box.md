# Spec: Astrolink Box (Produto de Hardware)

## Visão Geral

O Astrolink Box é um mini PC pré-configurado com todo o software Astrolink instalado, pronto para uso. O operador recebe, pluga na rede, liga, e em 5 minutos está coletando pagamentos PIX. Sem terminal, sem configuração manual, sem necessidade de conhecimento técnico.

**Posicionamento:** "Plug and play para monetizar seu Wi-Fi"

---

## Hardware Selecionado

### Versão Padrão — Orange Pi 5 (recomendado)

| Especificação | Valor |
|---|---|
| Processador | Rockchip RK3588S (8 cores, ARM) |
| RAM | 8GB LPDDR4X |
| Armazenamento | 64GB eMMC (sem cartão SD) |
| Ethernet | 2x Gigabit (1 para internet, 1 para rede local) |
| USB | 2x USB 3.0, 2x USB 2.0 |
| Consumo | ~5–15W |
| Dimensões | 100 x 72 x 25mm |
| Preço de custo | ~R$ 200–250 |

### Versão Lite — Orange Pi 3B (mais acessível)

| Especificação | Valor |
|---|---|
| Processador | Rockchip RK3566 (4 cores, ARM) |
| RAM | 4GB |
| Armazenamento | 32GB eMMC |
| Ethernet | 1x Gigabit |
| Consumo | ~3–8W |
| Dimensões | 85 x 56mm |
| Preço de custo | ~R$ 120–150 |

### Por que não Raspberry Pi?
- Orange Pi tem Wi-Fi embutido melhor, 2 portas Ethernet nativas no modelo 5
- Preço significativamente menor (especialmente no Brasil)
- Disponibilidade local (distribuidores brasileiros)

---

## Imagem de Sistema (Armbian + Astrolink)

### Base

```
Sistema operacional: Armbian Ubuntu Jammy (22.04 LTS) minimal
Kernel: 5.10 LTS (suporte longo prazo)
Usuário padrão: astrolink (senha definida no primeiro boot)
SSH: habilitado, somente por chave após configuração
```

### O que está pré-instalado

```bash
# Infraestrutura
docker-ce              # Docker Engine
docker-compose-plugin  # Docker Compose v2

# Astrolink (via Docker Compose)
astrolink-node         # Backend Go (versão mais recente estável)
postgres:15            # Banco de dados
redis:7                # Cache
pgbouncer              # Connection pool

# Utilitários
astrolink-cli          # CLI tool
htop                   # Monitor de recursos
fail2ban               # Proteção SSH
ufw                    # Firewall
```

### Particionamento do Disco

```
Partição 1: /boot (512MB, ext4) — sistema
Partição 2: /     (16GB, ext4) — SO e aplicação
Partição 3: /data (restante, ext4) — banco de dados, uploads, backups
```

A partição `/data` é montada em `/data/astrolink/` e é o único lugar onde dados do usuário são armazenados. Facilita backup e migração.

---

## Processo de Fabricação

### Build da Imagem

```bash
#!/bin/bash
# scripts/build-image.sh
# Executado para cada lote de produção

VERSION=$(git describe --tags --abbrev=0)
BOARD="orangepi5"

# Baixar imagem base Armbian
wget -q "https://dl.armbian.com/${BOARD}/Ubuntu_jammy_minimal.img.xz"
xz -d Ubuntu_jammy_minimal.img.xz

# Montar imagem e customizar
./scripts/customize-image.sh Ubuntu_jammy_minimal.img

# Incluir versão do Astrolink
docker save ghcr.io/astrolink/node:${VERSION} | gzip > /mnt/astrolink-images/node.tar.gz
docker save postgres:15-alpine | gzip > /mnt/astrolink-images/postgres.tar.gz
docker save redis:7-alpine | gzip > /mnt/astrolink-images/redis.tar.gz

# Comprimir imagem final
gzip -9 Ubuntu_jammy_minimal.img
mv Ubuntu_jammy_minimal.img.gz "astrolink-box-${VERSION}-${BOARD}.img.gz"

echo "Imagem pronta: astrolink-box-${VERSION}-${BOARD}.img.gz"
echo "SHA256: $(sha256sum astrolink-box-${VERSION}-${BOARD}.img.gz)"
```

### Flash em Produção

```bash
#!/bin/bash
# Gravação em série — executar para cada unidade
DEVICE=$1  # ex: /dev/mmcblk0

echo "Gravando imagem em ${DEVICE}..."
dd if=astrolink-box-latest.img of=${DEVICE} bs=4M status=progress conv=fsync

echo "Expandindo sistema de arquivos..."
parted ${DEVICE} resizepart 3 100%
resize2fs ${DEVICE}p3

echo "Verificando integridade..."
md5sum -c checksums.txt

echo "✅ Unidade pronta!"
```

---

## Experiência de Configuração (Primeiro Boot)

### Passo 1: Ligar o Box

O cliente liga o Box na energia e conecta o cabo de rede (porta WAN → seu Starlink/modem).

### Passo 2: Acessar via celular

O Box cria uma rede Wi-Fi de configuração temporária:
```
SSID: Astrolink-Setup-XXXXX
Senha: (impressa na lateral do box)
```

O celular conecta e é redirecionado automaticamente para:
```
http://astrolink.local/setup
```

### Passo 3: Wizard de configuração (5 minutos)

**Tela 1 — Idioma e Fuso Horário**
```
Idioma: Português (BR) ▼
Fuso horário: America/Belem ▼
[Continuar →]
```

**Tela 2 — Nome do Hotspot**
```
Como se chama seu Wi-Fi?
[Pousada Recanto Verde          ]

Essa é a mensagem de boas-vindas:
"Bem-vindo ao Wi-Fi Pousada Recanto Verde"
[Continuar →]
```

**Tela 3 — Senha do Admin**
```
Crie uma senha para o painel admin:
[•••••••••••••] [👁]
[•••••••••••••] (confirmar)

⚠️ Guarde bem essa senha!
[Continuar →]
```

**Tela 4 — Mercado Pago**
```
Para receber pagamentos PIX, conecte sua
conta do Mercado Pago:

[Conectar Mercado Pago →]
(abre OAuth do MP)

✅ Conta conectada: João Silva
[Continuar →]
```

**Tela 5 — Conectar Roteadores**
```
Roteadores OpenWrt detectados na rede:

☑ GL.iNet GL-MT3000 (192.168.2.1)
☑ TP-Link Archer (192.168.2.2)

Senha SSH dos roteadores: [root     ] [••••••••]

[Configurar automaticamente]

✅ 2 roteadores configurados!
[Continuar →]
```

**Tela 6 — Plano Cloud (opcional)**
```
Gerencie este Box pelo app ou pelo
painel online de qualquer lugar.

[Conectar ao Cloud Panel]
(linka ao workspace Astrolink)

[Pular por agora]
```

**Tela 7 — Pronto!**
```
🎉 Tudo configurado!

Seu Wi-Fi está pronto para receber
pagamentos via PIX!

Acesse o painel admin:
http://astrolink.local/admin

Baixe o app para gerenciar pelo celular:
[Google Play] [App Store]

[Começar a usar →]
```

---

## Embalagem

### Conteúdo da Caixa
- 1x Orange Pi 5 (com case alumínio pré-montado)
- 1x Fonte 5V/5A USB-C
- 1x Cabo Ethernet Cat6 (1m)
- 1x Cartão com senha Wi-Fi de setup e URL do painel
- 1x Guia de início rápido (A5, 8 páginas, PT-BR)
- 1x Adesivo Astrolink (para colocar no local)

### Design da Embalagem

Caixa branca com logo Astrolink em azul ciano. Tagline: "Seu Wi-Fi, sua renda."

Lateral: lista de conteúdo, requisitos (conexão internet ativa)

Verso: QR code para documentação online + tutorial em vídeo

---

## Precificação

```
Custo de produção (Orange Pi 5 + acessórios + embalagem): R$ 280
Custo de software/licença: R$ 0 (open source)
Custo de envio (Sedex médio BR): R$ 50

Preço de venda: R$ 499
Margem bruta: ~34%

Incluído no preço:
  • 3 meses de plano Pro (valor: R$ 147)
  • Suporte prioritário por WhatsApp nos 3 primeiros meses

Custo real do hardware: R$ 352 (após plano incluso)
Margem real: ~30%
```

### Versão Lite (Orange Pi 3B)

```
Custo de produção: R$ 200
Preço de venda: R$ 349
Inclui: 1 mês de plano Pro
```

---

## Suporte e Garantia

- **Garantia:** 12 meses contra defeitos de fabricação
- **Suporte:** WhatsApp dedicado nos 3 primeiros meses (plano Pro incluso)
- **Troca:** Troca gratuita em até 7 dias após recebimento
- **Assistência técnica:** rede de parceiros certificados (futuramente)

---

## Distribuição

### Fase 1 (v1.0): Venda direta
- Site astrolink.app/hardware
- Pagamento: PIX (claro!) e cartão
- Envio: Correios Sedex, todo o Brasil

### Fase 2 (v1.5): Revendedores
- Programa de revendedores para integradores de rede
- Margem do revendedor: 20%
- Estoque mínimo: 5 unidades
- Treinamento via vídeo incluso

### Fase 3 (v2.0): Distribuição em escala
- Parceria com distribuidores de equipamentos de rede
- Ponto de venda em lojas de eletrônicos / informática
- Potencialmente: parceria com operadoras Starlink para bundle

---

## Roadmap do Hardware

| Versão | Hardware | Diferencial |
|---|---|---|
| v1 | Orange Pi 5 | Lançamento |
| v2 | Custom PCB | Mais compacto, PoE, mais barato de fabricar |
| v3 | Integrado com roteador | All-in-one: server + AP num único dispositivo |

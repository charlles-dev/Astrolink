# Spec: Marketplace de Temas

## Visão Geral

Plataforma onde desenvolvedores e designers podem vender (ou distribuir gratuitamente) temas para o portal cativo do Astrolink. Cada tema customiza completamente a aparência do portal que o usuário final vê ao se conectar.

**URL:** `themes.astrolink.app`
**Modelo:** Comissão de 30% em vendas pagas. Temas gratuitos sempre permitidos.

---

## O que é um Tema?

Um tema é um pacote `.astrolink-theme` (zip com estrutura específica) que substitui os componentes visuais do portal cativo, mantendo toda a lógica de negócio intacta.

### Estrutura do Pacote

```
meu-tema.astrolink-theme
├── theme.json          # Metadados (nome, autor, versão, preview)
├── styles/
│   └── theme.css       # CSS customizado (import no portal)
├── components/
│   ├── PlanCard.svelte      # Card de plano (obrigatório)
│   ├── WelcomeScreen.svelte # Tela de boas-vindas (obrigatório)
│   ├── PIXScreen.svelte     # Tela de pagamento PIX (obrigatório)
│   ├── SuccessScreen.svelte # Tela de sucesso (obrigatório)
│   └── VoucherInput.svelte  # Input de voucher (obrigatório)
├── assets/
│   ├── preview.png          # Screenshot 390x844 (mobile)
│   └── preview-desktop.png  # Screenshot 1280x720 (opcional)
└── README.md           # Instruções e customizações do tema
```

### `theme.json`

```json
{
  "id": "pousada-tropical",
  "nome": "Pousada Tropical",
  "versao": "1.0.0",
  "descricao": "Tema com visual de natureza e cores vibrantes para pousadas e resorts",
  "autor": {
    "nome": "João Designer",
    "email": "joao@designer.com",
    "site": "https://joaodesigner.com"
  },
  "astrolink_versao_minima": "1.0.0",
  "preco": 4900,        // em centavos. 0 = gratuito
  "categoria": "pousada",
  "tags": ["natureza", "tropical", "colorido", "pousada", "resort"],
  "variaveis_css": {
    "--color-primary": { "tipo": "color", "label": "Cor principal", "padrao": "#2ECC71" },
    "--color-background": { "tipo": "color", "label": "Fundo", "padrao": "#0D2818" },
    "--border-radius": { "tipo": "select", "opcoes": ["4px", "8px", "16px", "24px"], "padrao": "16px" }
  }
}
```

---

## Categorias de Temas

| Categoria | Exemplos |
|---|---|
| `pousada` | natureza, tropical, praia, montanha |
| `lan-house` | gamer, neon, dark, cyberpunk |
| `corporativo` | profissional, minimalista, clean |
| `evento` | festa, casamento, conferência |
| `saude` | clínica, hospital, farmácia |
| `educacao` | escola, faculdade, biblioteca |
| `minimalista` | ultra-clean, sem imagens |
| `gratuito` | todos os temas gratuitos |

---

## Site do Marketplace

### Página Principal

```
┌──────────────────────────────────────────────────────┐
│ 🎨 Astrolink Themes                    [Enviar tema] │
├──────────────────────────────────────────────────────┤
│                                                      │
│  Personalize seu portal cativo                       │
│  Mais de 50 temas para qualquer estilo               │
│                                                      │
│  [Todos ▼] [Pousada] [LAN House] [Gratuito] [Novo]  │
│  [🔍 Buscar temas...]                                │
│                                                      │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ │
│  │ [PREVIEW]    │ │ [PREVIEW]    │ │ [PREVIEW]    │ │
│  │              │ │              │ │              │ │
│  │ Pousada      │ │ Gamer Dark   │ │ Minimal Clean│ │
│  │ Tropical     │ │              │ │              │ │
│  │ ⭐ 4.8 (12)  │ │ ⭐ 4.6 (8)   │ │ ⭐ 4.9 (23)  │ │
│  │ João Design  │ │ TechThemes   │ │ Astrolink    │ │
│  │ R$ 49,00     │ │ R$ 29,00     │ │ GRATUITO     │ │
│  │ [Detalhes]   │ │ [Detalhes]   │ │ [Instalar]   │ │
│  └──────────────┘ └──────────────┘ └──────────────┘ │
│                                                      │
└──────────────────────────────────────────────────────┘
```

### Página do Tema

```
┌──────────────────────────────────────────────────────┐
│ ← Marketplace                                        │
│                                                      │
│ Pousada Tropical                     por João Design │
│ ⭐ 4.8 (12 avaliações)               R$ 49,00        │
│                                                      │
│ [Preview Mobile] [Preview Desktop] [Demo ao vivo]   │
│                                                      │
│ [SCREENSHOT DO TEMA — Cada tela do portal]          │
│                                                      │
│ Descrição:                                           │
│   Tema com visual de natureza e cores vibrantes...  │
│                                                      │
│ Compatível com: Astrolink v1.0+                     │
│ Categoria: Pousada · Tags: tropical, natureza       │
│                                                      │
│ Customizações incluídas:                            │
│   • Cor principal                                    │
│   • Cor de fundo                                     │
│   • Border radius                                    │
│                                                      │
│ [Comprar por R$ 49,00]                              │
│   Pagamento seguro via PIX                          │
│                                                      │
│ Avaliações:                                         │
│   ⭐⭐⭐⭐⭐ "Ficou lindo na minha pousada!" — Maria  │
│   ⭐⭐⭐⭐○ "Muito bom, fácil de instalar" — Carlos │
└──────────────────────────────────────────────────────┘
```

---

## Instalação de Tema (Admin Local)

```
Admin Local → Configurações → Aparência → Temas

Tema atual: Padrão Astrolink [Alterar tema]

┌──────────────────────────────────────────────────────┐
│ 🎨 Escolher tema                                     │
│                                                      │
│ [Meus temas comprados (2)]  [Marketplace]           │
│                                                      │
│ ┌──────────────┐  ┌──────────────┐                  │
│ │ Pousada      │  │ Minimal Clean│                  │
│ │ Tropical     │  │ (padrão)     │                  │
│ │              │  │              │                  │
│ │ [Ativar]     │  │ ✅ Ativo     │                  │
│ └──────────────┘  └──────────────┘                  │
│                                                      │
│ [+ Instalar do Marketplace]                          │
│ [+ Instalar arquivo .astrolink-theme]                │
└──────────────────────────────────────────────────────┘

Após ativar: portal cativo usa o novo tema imediatamente
```

---

## Processo de Submissão de Tema

### 1. Criar conta de desenvolvedor
- Cadastro em `themes.astrolink.app/dev`
- Verificação de email
- Aceitar termos de uso e política de comissão (30%)

### 2. Desenvolver o tema
```bash
# Instalar CLI de temas (parte do astrolink-cli)
astrolink theme create meu-tema
# Cria estrutura básica do pacote

# Testar localmente (sobe preview no browser)
astrolink theme dev meu-tema/
# → Preview em http://localhost:5050

# Validar pacote (verifica estrutura, acessibilidade, performance)
astrolink theme validate meu-tema.astrolink-theme
# ✅ Estrutura válida
# ✅ Todos os componentes obrigatórios presentes
# ✅ Preview images presentes
# ✅ Acessibilidade: contraste WCAG AA
# ✅ Performance: bundle < 100KB
# ⚠️  README.md muito curto
```

### 3. Submeter para review
```bash
astrolink theme submit meu-tema.astrolink-theme \
  --preco 4900 \
  --categoria pousada \
  --tags "tropical,natureza,pousada"
# → Submetido! ID: theme_XXXXXXXX
# → Review em até 5 dias úteis
```

### 4. Review pela equipe Astrolink
- Verificação de malware/código malicioso
- Teste em browser real (Chrome, Safari, Firefox)
- Verificação de acessibilidade
- Verificação de responsividade
- Teste no simulador de portal

### 5. Publicação
- Notificação por email
- Tema aparece no marketplace
- Link para compartilhar

---

## Critérios de Aprovação

### Obrigatório
- ✅ Todos os componentes obrigatórios implementados
- ✅ Funciona em Chrome, Safari e Firefox (últimas 2 versões)
- ✅ Mobile-first: perfeito em 390px de largura
- ✅ Contraste WCAG AA em todos os textos
- ✅ Navegável por teclado
- ✅ Bundle total < 100KB gzipped
- ✅ Sem JavaScript externo (segurança)
- ✅ Sem rastreadores externos
- ✅ Preview de alta qualidade (390x844px)

### Recomendado
- Animações suaves (prefers-reduced-motion respeitado)
- Dark mode suportado
- README com instruções de customização

### Vetado
- ❌ Código ofuscado ou malicioso
- ❌ Logos de outras empresas sem permissão
- ❌ Conteúdo adulto, violento ou discriminatório
- ❌ Requisições para servidores externos
- ❌ Modificação da lógica de pagamento

---

## Royalties e Pagamentos

```
Venda de tema R$ 49,00:
  Comissão Astrolink (30%): R$ 14,70
  Repasse ao autor (70%):   R$ 34,30

Repasse: mensalmente, via PIX, mínimo R$ 50
Relatório: dashboard do desenvolvedor com vendas por tema
```

---

## Temas Gratuitos Incluídos (pela Astrolink)

| Nome | Estilo | Indicado para |
|---|---|---|
| **Padrão** | Escuro, minimalista | Qualquer tipo |
| **Pousada** | Verde, natureza | Pousadas, eco-turismo |
| **LAN House** | Dark, neon cyan | LAN houses, games |
| **Clínica** | Branco, limpo, azul | Saúde, consultórios |
| **Evento** | Festivo, gradiente | Eventos, festas |
| **Corporativo** | Cinza, profissional | Empresas, coworkings |

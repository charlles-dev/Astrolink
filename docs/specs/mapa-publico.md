# Spec: Mapa Público de Hotspots

## Visão Geral

Site público onde qualquer pessoa pode descobrir pontos de acesso Wi-Fi Astrolink próximos a ela. Funciona como um diretório geográfico de hotspots, com foco em SEO para cidades e comunidades remotas.

**URL:** `astrolink.app/mapa`
**Stack:** Astro 4 (SSG + SSR) + Mapbox GL JS + Supabase (dados)
**Deploy:** Cloudflare Pages (CDN global, gratuito)

---

## Objetivos

1. **Descoberta:** usuário final encontra Wi-Fi disponível na região
2. **SEO:** ranquear para "Wi-Fi em [cidade]", "internet [bairro/comunidade]"
3. **Marketing para operadores:** aparecer no mapa = mais clientes
4. **Flywheel:** mais nós → mapa mais útil → mais usuários → mais operadores

---

## Estrutura de URLs (SEO-Friendly)

```
/mapa                               → Mapa geral (geolocalização do usuário)
/mapa/br                            → Brasil
/mapa/br/pa                         → Pará
/mapa/br/pa/parauapebas             → Parauapebas, PA
/hotspot/pousada-recanto-verde      → Página do hotspot
/hotspot/lan-house-gamer-maraba     → Página do hotspot
/blog/wifi-em-comunidades-remotas   → Conteúdo SEO
```

---

## Tela Principal — Mapa

```
┌─────────────────────────────────────────────────────┐
│ 🌐 Astrolink   [Encontre Wi-Fi]  [Seja um Provedor] │
├─────────────────────────────────────────────────────┤
│                                                     │
│  Encontre Wi-Fi perto de você                       │
│  ┌─────────────────────────────────────────────┐   │
│  │ 📍 Digite sua cidade ou bairro...            │   │
│  └─────────────────────────────────────────────┘   │
│  [Usar minha localização atual 📍]                  │
│                                                     │
│  Filtros: [Aberto agora] [Até R$10] [Velocidade]   │
│                                                     │
├─────────────────────────────────────────────────────┤
│                                                     │
│  [                                                ] │
│  [                  MAPA                         ] │
│  [            (Mapbox GL JS)                     ] │
│  [  🟢   🟢                                       ] │
│  [          🟢  🟢                                ] │
│  [                    🟢                          ] │
│  [                                                ] │
│                                                     │
├─────────────────────────────────────────────────────┤
│  Lista de resultados (scroll)                       │
│  ─────────────────────────────────────────────────  │
│  🟢 Wi-Fi Pousada Recanto Verde          500m      │
│     R$ 5/hora · R$ 15/dia · ⭐ 4.8 (23)           │
│     Aberto: Seg–Dom 06:00–22:00                    │
│  ─────────────────────────────────────────────────  │
│  🟢 LAN House Gamer Marabá              1.2km      │
│     R$ 3/hora · R$ 8/dia · ⭐ 4.5 (11)            │
│     Aberto agora · Abre às 09:00                   │
│  ─────────────────────────────────────────────────  │
└─────────────────────────────────────────────────────┘
```

### Mapa Interativo

**Pins:**
- Verde: hotspot online, aberto agora
- Cinza: hotspot online mas fora do horário
- Cluster: quando múltiplos pontos próximos (número dentro do cluster)

**Popup ao clicar no pin:**
```
┌────────────────────────────────┐
│ 🟢 Wi-Fi Pousada Recanto Verde │
│                                │
│ ⭐ 4.8  (23 avaliações)        │
│ 📍 Rua das Flores, 123         │
│ ⏰ Aberto agora (até 22:00)    │
│                                │
│ Planos:                        │
│  • R$ 5,00 → 1 hora            │
│  • R$ 15,00 → 24 horas         │
│                                │
│ [Ver detalhes →]               │
└────────────────────────────────┘
```

---

## Página do Hotspot (SEO core)

URL: `/hotspot/pousada-recanto-verde`
Gerada estaticamente (Astro SSG) uma vez por dia.

```html
<!-- Meta tags SEO -->
<title>Wi-Fi Pousada Recanto Verde — Parauapebas, PA | Astrolink</title>
<meta name="description" content="Wi-Fi pago disponível na Pousada Recanto Verde em Parauapebas, PA. Planos a partir de R$ 5,00/hora. Avaliação 4.8 ⭐">

<!-- JSON-LD Structured Data -->
<script type="application/ld+json">
{
  "@type": "LocalBusiness",
  "name": "Wi-Fi Pousada Recanto Verde",
  "address": { "@type": "PostalAddress", "addressLocality": "Parauapebas", "addressRegion": "PA" },
  "geo": { "@type": "GeoCoordinates", "latitude": -6.0666, "longitude": -49.8877 },
  "openingHours": "Mo-Su 06:00-22:00",
  "priceRange": "R$5–R$15",
  "aggregateRating": { "@type": "AggregateRating", "ratingValue": "4.8", "reviewCount": "23" }
}
</script>
```

### Layout da Página

```
┌─────────────────────────────────────────────────┐
│ ← Voltar ao mapa                                │
│                                                 │
│ [FOTO DO LOCAL]                                 │
│                                                 │
│ 🟢 Wi-Fi Pousada Recanto Verde                  │
│ ⭐ 4.8  (23 avaliações)                         │
│ 📍 Rua das Flores, 123 — Parauapebas, PA       │
│                                                 │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━   │
│ Planos disponíveis:                             │
│   🕐 1 hora ─────────────────── R$ 5,00        │
│   🌙 24 horas ────────────────── R$ 15,00      │
│   📅 Semana ──────────────────── R$ 50,00      │
│                                                 │
│ [🗺️ Como chegar] [📞 Contato] [🌐 Site]        │
│                                                 │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━   │
│ Informações:                                    │
│   ⏰ Seg–Dom, 06:00–22:00                       │
│   📶 Velocidade estimada: ~40 Mbps (Starlink)  │
│   📡 Cobertura: área interna + varanda          │
│                                                 │
│ Descrição:                                      │
│   Wi-Fi de qualidade para hóspedes e           │
│   visitantes. Conexão via Starlink.             │
│                                                 │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━   │
│ Avaliações (23)              [Avaliar ✍️]       │
│                                                 │
│ ⭐⭐⭐⭐⭐ João M. — "Ótima velocidade!"        │
│ ⭐⭐⭐⭐○ Maria L. — "Preço justo"             │
│ [Ver todas as avaliações]                       │
│                                                 │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━   │
│ Mapa:                                           │
│ [mini mapa Mapbox com pin]                      │
│                                                 │
│ Outros hotspots em Parauapebas:                 │
│   • LAN House Speed (2.3km) ⭐4.5              │
│   • Wi-Fi Mercadão Central (3.1km) ⭐4.2       │
└─────────────────────────────────────────────────┘
```

---

## Avaliações de Usuários

### Submeter avaliação
- Sem cadastro necessário (anônimo)
- 1–5 estrelas + comentário opcional (max 280 chars)
- Rate limiting: 1 avaliação por IP por hotspot por 30 dias
- Moderação: filtro automático de palavrões + review manual para textos negativos
- Notificação ao operador quando nova avaliação chega

### Schema no Supabase

```sql
public_reviews (
  id UUID,
  node_id UUID REFERENCES nodes(id),
  rating SMALLINT CHECK (rating BETWEEN 1 AND 5),
  comment TEXT,
  author_name TEXT,         -- opcional
  ip_hash TEXT,             -- para rate limiting (não armazenar IP direto)
  approved BOOLEAN DEFAULT false,
  created_at TIMESTAMPTZ
)
```

---

## Páginas de Cidade (SEO em escala)

Geradas automaticamente para cada cidade/estado com hotspots:

**URL:** `/mapa/br/pa/parauapebas`
**Título:** "Wi-Fi disponível em Parauapebas, PA — Astrolink"

Conteúdo:
- Contagem de hotspots na cidade
- Lista dos hotspots com cards
- Mini mapa centrado na cidade
- Texto gerado: "Em Parauapebas existem X pontos de Wi-Fi Astrolink..."

Isso gera centenas de páginas indexáveis com conteúdo relevante por localidade.

---

## Sitemap Dinâmico

Gerado uma vez por dia via Edge Function:

```xml
<url>
  <loc>https://astrolink.app/hotspot/pousada-recanto-verde</loc>
  <lastmod>2025-05-19</lastmod>
  <changefreq>weekly</changefreq>
  <priority>0.8</priority>
</url>
```

Inclui todas as páginas de hotspot + cidades com pelo menos 1 hotspot ativo.

---

## Seção "Seja um Provedor"

Landing page de conversão para operadores:

```
Tem uma conexão Starlink ou fibra?
Monetize compartilhando com sua comunidade.

Como funciona:
  1. Instale o Astrolink em um mini PC ou roteador
  2. Configure seus planos de preço
  3. Seus vizinhos pagam via PIX e navegam

Grátis para começar. Gerencie múltiplos pontos por R$ 49/mês.

[Começar agora gratuitamente]
[Ver documentação de instalação]
```

---

## Analytics e Métricas do Mapa

Dados coletados (respeitando LGPD, sem dados pessoais):
- Buscas por cidade (para priorizar expansão)
- Visualizações de página por hotspot
- Cliques em "Como chegar"
- Regiões sem cobertura mais buscadas

Painel interno para equipe Astrolink:
- Mapa de calor de buscas sem resultado → onde priorizar crescimento
- Funil: busca → visualização hotspot → visita ao local

---

## Performance e SEO

| Critério | Meta |
|---|---|
| LCP | < 1.5s (páginas estáticas pré-geradas) |
| CLS | < 0.1 |
| Core Web Vitals | Tudo verde |
| Cache | páginas estáticas: 24h CDN, dados do mapa: 5min |
| Imagens | WebP, lazy loading, tamanhos responsivos |
| Robots.txt | Permitir tudo, bloquear `/api/*` |
| Hreflang | PT-BR por padrão, EN futuro |

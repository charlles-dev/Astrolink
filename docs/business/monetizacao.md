# Estratégia de Monetização

## Modelo: Open Core

O Astrolink segue o modelo **Open Core** — o núcleo é open source (MIT), e funcionalidades avançadas de gestão multi-nó são pagas via SaaS.

Este modelo foi provado por empresas como GitLab, Elastic, MongoDB, PostHog e Supabase. A vantagem é que a comunidade constrói confiança no produto, e usuários avançados (com mais nós, mais necessidades) naturalmente evoluem para o plano pago.

---

## O que é gratuito (para sempre)

- Instalação self-hosted em 1 nó
- Usuários ilimitados no nó
- Planos ilimitados
- Vouchers ilimitados
- Integração com Mercado Pago (PIX)
- Portal cativo completo (SvelteKit)
- Admin local completo
- Integração real com OpenNDS
- CLI (`astrolink`)
- Todas as futuras atualizações do core

**Princípio:** qualquer provedor com 1 local físico pode usar 100% do Astrolink sem pagar nada.

---

## Fluxo de Monetização por Camada

### Camada 1: Nó Local (gratuito)

```
Operador instala → Usa gratuitamente → Ganha receita com seus clientes
                                              ↓
                              [natural growth: abre mais locais]
                                              ↓
                              Precisa gerenciar múltiplos nós remotamente
                                              ↓
                              Assina o Cloud Panel → $$$ para a Astrolink
```

### Camada 2: Cloud Panel (pago)

O Cloud Panel é o produto de receita principal. Sem ele, o operador com 3+ locais fica cego — não sabe se um nó caiu, não consolida receita, não consegue gerenciar remotamente.

---

## Planos e Preços

### Free
- **Preço:** Gratuito
- **Nós:** 1
- **Acesso:** Apenas painel local (sem cloud)
- **Suporte:** GitHub Issues

### Pro — R$ 49/mês
- **Nós:** até 10
- **Acesso:** Cloud Panel completo
- **App mobile:** ✅
- **Alertas (email + push):** ✅
- **Relatórios avançados:** ✅
- **Suporte:** Email (resposta em 48h)

### Business — R$ 149/mês
- **Nós:** Ilimitado
- **Acesso:** Cloud Panel completo + API pública
- **App mobile:** ✅
- **Alertas completos:** ✅ (WhatsApp incluso)
- **Relatórios + previsão:** ✅
- **White-label por nó:** ✅ (domínio customizado)
- **Webhooks e integrações:** ✅
- **Suporte:** Email prioritário (resposta em 12h) + chat

### Enterprise — Sob consulta
- Nós ilimitados
- SLA garantido (99.9% uptime)
- Suporte dedicado (WhatsApp direto)
- Integração customizada (ERP, WHMCS)
- Treinamento da equipe
- Onboarding presencial

---

## Cobrança

**Canal:** AbacatePay (PIX recorrente)
- PIX gerado no vencimento, enviado por email
- 7 dias de graça se não pago
- Suspensão automática após período de graça
- Reativação imediata ao pagar

**Vantagens sobre cartão de crédito:**
- Sem taxa de processamento de cartão
- 100% dos provedores alvo têm conta com PIX
- Não precisa de cartão de crédito (barreira de entrada menor)
- Chargeback zero

---

## Receitas Complementares

### Astrolink Box (Hardware)

```
Custo de produção:  ~R$ 150 (Orange Pi 5 + acessórios + embalagem)
Preço de venda:      R$ 450
Margem bruta:        ~67%

Incluído:
  • Hardware pré-configurado
  • 3 meses de plano Pro (R$ 147 de valor)
  • Manual impresso em PT-BR
  • Suporte prioritário nos 3 primeiros meses

Meta: vender 50 unidades/mês = R$ 22.500 de receita hardware
```

**Vantagem estratégica:** quem compra o Box provavelmente assina o Pro após os 3 meses (retenção alta).

### Marketplace de Temas

```
Comissão Astrolink: 30% de cada venda
Preço médio dos temas: R$ 49

Para viabilizar:
  • 100 temas disponíveis
  • Média 10 vendas/tema/mês
  • 30% = R$ 14,70 por venda
  • 1.000 vendas/mês = R$ 14.700

Crescimento: cresce automaticamente com a base de nós
```

### Serviço de Instalação Gerenciada

Para provedores não-técnicos que querem ter tudo configurado:

| Serviço | Preço |
|---|---|
| Instalação básica (1 nó) | R$ 350 |
| Instalação + configuração OpenWrt | R$ 550 |
| Pacote 5 nós | R$ 1.500 |
| Manutenção mensal | R$ 150/nó |

Escalado com rede de parceiros/revendedores treinados.

### Revendedores / Parceiros

Programa de parceiros para integradores de rede locais:

- 20% de comissão recorrente nas assinaturas indicadas
- Painel de parceiro para acompanhar comissões
- Badge "Parceiro Astrolink Certificado"
- Material de marketing co-branded

---

## Projeções de Receita

### Cenário Conservador (12 meses)

| Mês | Nós ativos | Pago Pro | Pago Business | MRR |
|---|---|---|---|---|
| 1 | 50 | 5 | 1 | R$ 394 |
| 3 | 150 | 20 | 5 | R$ 1.725 |
| 6 | 400 | 60 | 15 | R$ 5.175 |
| 9 | 800 | 120 | 35 | R$ 11.105 |
| 12 | 1.500 | 250 | 70 | R$ 22.750 |

### Cenário Moderado (12 meses)

| Mês | MRR | ARR |
|---|---|---|
| 6 | R$ 12.000 | R$ 144.000 |
| 12 | R$ 55.000 | R$ 660.000 |

### Cenário Otimista (24 meses)

Considerando crescimento viral via mapa público e comunidade:

| Ano | MRR | ARR |
|---|---|---|
| 1 | R$ 80.000 | R$ 960.000 |
| 2 | R$ 250.000 | R$ 3.000.000 |

---

## Churn e Retenção

**Por que o churn é naturalmente baixo:**
- O operador tem dados históricos de seus clientes no Astrolink
- Migrar para outra solução significa perder histórico e reconfigurar tudo
- A cada nó adicional que ele adiciona, mais preso ao ecossistema fica (positivo)
- A comunidade e o mapa público criam dependência de rede

**Estratégias de retenção:**
- Relatório mensal automático por email ("seu mês em números")
- Destaques de conquistas ("você processou R$ X este mês!")
- Notificações de novas features relevantes
- Webinars mensais gratuitos de boas práticas
- Grupo de operadores no WhatsApp/Discord

---

## Unit Economics (estimativas)

```
CAC (Custo de Aquisição):
  Orgânico (SEO, GitHub, boca a boca): ~R$ 0
  Ads (futuramente):                    ~R$ 150

LTV (Life Time Value):
  Churn estimado: 3% ao mês
  LTV Pro:        R$ 49 / 0.03 = R$ 1.633
  LTV Business:   R$ 149 / 0.02 = R$ 7.450

LTV/CAC ratio:
  Orgânico:  ∞ (custo zero)
  Pago:      ~11x (excelente)

Payback period: 3 meses (orgânico), 4 meses (pago)
```

---

## GitHub Sponsors / OpenCollective

Para usuários que querem apoiar o projeto sem precisar do plano pago:

- **Individual:** $5/mês — agradecimento público no README
- **Supporter:** $15/mês — acesso antecipado a features beta
- **Sponsor:** $50/mês — logo no README e site

Meta de sponsors: 100 sponsors = $500–5.000/mês
Serve como sinal de saúde do projeto para potenciais contratantes Enterprise.

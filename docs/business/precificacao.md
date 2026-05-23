# Estratégia de Preços Detalhada

## Princípios da Precificação

1. **Value-based:** o preço é baseado no valor entregue, não no custo
2. **Progressiva:** mais valor = mais nós = mais receita para o operador → mais justo pagar mais
3. **Sem surpresas:** preço fixo por nó, sem taxas por transação
4. **PIX first:** cobrança via PIX (sem taxas de cartão, sem chargeback)

---

## Tabela de Planos

### Free — R$ 0/mês (para sempre)

**Para:** Provedores iniciantes, testes, uso pessoal

| Recurso | Limite |
|---|---|
| Nós gerenciados | 1 |
| Usuários por nó | Ilimitado |
| Transações PIX | Ilimitadas (você fica com 100%) |
| Vouchers | Ilimitados |
| Painel admin local | ✅ Completo |
| Portal cativo | ✅ Completo |
| CLI | ✅ Completo |
| App mobile | ❌ |
| Cloud Panel | ❌ |
| Alertas push/email | ❌ |
| Relatórios avançados | ❌ |
| Suporte | GitHub Issues |

---

### Pro — R$ 49/mês

**Para:** Provedores com 2–10 locais

**Faturamento:** Mensal via PIX (AbacatePay)
**Equivalente anual:** R$ 490/ano (se pagar anual — 2 meses grátis)

| Recurso | Limite |
|---|---|
| Nós gerenciados | Até 10 |
| Cloud Panel | ✅ Completo |
| App mobile (iOS + Android) | ✅ |
| Alertas (push + email) | ✅ |
| Relatórios financeiros | ✅ (últimos 12 meses) |
| White-label por nó | ✅ (logo + cores) |
| Convites de equipe | Até 3 membros |
| API pública | ❌ |
| Domínio customizado por nó | ❌ |
| Suporte | Email (48h úteis) |
| SLA | Sem SLA |

---

### Business — R$ 149/mês

**Para:** Provedores com 10+ locais, ISPs locais

**Faturamento:** Mensal ou anual (R$ 1.490/ano — 2 meses grátis)

| Recurso | Limite |
|---|---|
| Nós gerenciados | Ilimitado |
| Cloud Panel | ✅ Completo |
| App mobile | ✅ |
| Alertas completos | ✅ + WhatsApp |
| Relatórios financeiros | ✅ (histórico completo + previsão) |
| White-label por nó | ✅ completo (logo + cores + domínio customizado) |
| Convites de equipe | Ilimitado |
| API pública REST | ✅ |
| Webhooks | ✅ |
| Integrações (WHMCS, Evolution API) | ✅ |
| Marketplace de temas | ✅ (compras e instalação) |
| Suporte | Email prioritário (12h úteis) + chat |
| SLA | 99.5% uptime Cloud Panel |

---

### Enterprise — Sob consulta (a partir de R$ 500/mês)

**Para:** ISPs com 50+ nós, redes municipais, franqueadoras

| Recurso | Detalhes |
|---|---|
| Nós gerenciados | Ilimitado |
| SLA | 99.9% uptime com crédito por falha |
| Suporte | WhatsApp dedicado + reunião mensal |
| Onboarding | Presencial ou remoto com equipe |
| Customizações | Features exclusivas mediante contrato |
| Billing customizado | Nota fiscal, pagamento por transferência |
| Treinamento | Equipe técnica e comercial |
| Relatórios customizados | Dashboard BI personalizado |

---

## Comparativo de Valor

### Cenário: Provedor com 5 locais

```
Receita típica por nó: R$ 800–3.000/mês
Receita total (5 nós): R$ 4.000–15.000/mês

Custo do plano Pro: R$ 49/mês

Percentual do faturamento: 0.3%–1.2%
ROI: mínimo 81x (R$ 4.000 / R$ 49)
```

**Conclusão:** pagar R$ 49 para gerir R$ 4.000+/mês é uma decisão óbvia.

### Cenário: Provedor com 20 locais (Business)

```
Receita típica: R$ 16.000–60.000/mês
Custo Business: R$ 149/mês

Percentual: 0.25%–0.93%
ROI: mínimo 107x
```

---

## Política de Trials e Descontos

### Trial gratuito
- Cloud Panel: 14 dias grátis ao se cadastrar (sem cartão)
- Hardware Box: 3 meses de Pro incluso

### Desconto anual
- Pro anual: R$ 490/ano (R$ 40,83/mês = 16% off)
- Business anual: R$ 1.490/ano (R$ 124,17/mês = 17% off)

### Desconto para ONGs e cooperativas
- 50% de desconto mediante comprovação
- Foco em comunidades de baixa renda e projetos sociais

### Programa de afiliados
- Revendedor recebe 20% recorrente
- Pago mensalmente via PIX

---

## Política de Upgrade/Downgrade

**Upgrade imediato:** ao fazer upgrade, o novo plano entra em vigor instantaneamente. O valor proporcional do mês atual é calculado e cobrado/creditado.

**Downgrade:** entra em vigor no próximo ciclo de faturamento. Não há reembolso do período atual.

**Cancelamento:** pode cancelar a qualquer momento. Acesso ao Cloud Panel termina no fim do período pago. Dados do nó local são preservados.

---

## Política de Suspensão

```
Pagamento vence → 7 dias de graça
              → PIX de cobrança enviado por email no vencimento
              → Lembrete no dia 3 de atraso
              → Suspensão automática no dia 7 de atraso
              → Cloud Panel mostra aviso de suspensão
              → Nós locais continuam funcionando (apenas cloud é suspenso)
              → Dados preservados por 30 dias após suspensão
              → Reativação imediata ao regularizar pagamento
              → Após 30 dias sem pagamento: tenant arquivado (dados podem ser exportados)
              → Após 90 dias: dados deletados permanentemente
```

---

## Comparativo com Alternativas

| Solução | Preço | Nós | Suporte BR | Open Source | Multi-nó |
|---|---|---|---|---|---|
| **Astrolink Free** | Grátis | 1 | ✅ | ✅ | — |
| **Astrolink Pro** | R$ 49/mês | 10 | ✅ | ✅ | ✅ |
| Mikrotik CHR | USD 45/mês | 1 servidor | ❌ | ❌ | Manual |
| CoovaChilli hosted | USD 30/mês | 1 | ❌ | Parcial | ❌ |
| Solução customizada | R$ 5.000+ | Ilimitado | Depende | ❌ | ❌ |
| Planilha + manual | R$ 0 | 1 | — | — | ❌ |

---

## Comunicação de Preços

### O que NÃO fazer
- ❌ Esconder o preço (sempre exibir na landing page)
- ❌ "Entre em contato para saber o preço" para planos padrão
- ❌ Cobrança surpresa por features básicas
- ❌ Taxas por transação (o operador fica com 100% do que recebe)

### O que fazer
- ✅ Preço exibido com destaque na landing page
- ✅ Comparativo claro de planos lado a lado
- ✅ Calculator de ROI interativo ("com X nós e receita Y, você paga Z% em assinatura")
- ✅ FAQ com perguntas comuns sobre preços
- ✅ "Você pode cancelar quando quiser" em destaque

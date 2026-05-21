# Go-to-Market Strategy

## Perfil do Cliente Ideal (ICP)

### Segmento Primário: Provedores Comunitários

**Quem é:**
- Dono de 1–5 pontos de acesso Starlink em comunidades rurais
- Mora no interior do Norte, Nordeste ou Centro-Oeste
- Renda mensal com Wi-Fi: R$ 500–5.000
- Não é técnico: usa o celular para tudo
- Já tem ou considera ter um roteador OpenWrt

**Dor principal:** não sabe ao certo quanto está ganhando, perde dinheiro com vouchers que não controla, não tem como saber quando o sinal cai quando está fora de casa.

**Canal para alcançá-lo:** grupos de WhatsApp de provedores rurais, YouTube, Facebook, indicação.

---

### Segmento Secundário: LAN Houses e Estabelecimentos

**Quem é:**
- Dono de LAN house, pousada, coworking, clínica, academia
- Quer monetizar o Wi-Fi ou controlar o acesso
- Tem 1 ponto mas pode escalar para vários

**Dor principal:** dar senha do Wi-Fi para todo mundo, não tem controle de quem usa, não monetiza.

**Canal:** Google (busca "controle de acesso wifi"), Instagram, indicação.

---

### Segmento Terciário: ISPs Locais (Internet Service Providers)

**Quem é:**
- ISP com 50–500 clientes
- Tem infraestrutura de fibra ou rádio
- Quer adicionar hotspot público como serviço adicional

**Dor principal:** plataformas existentes são caras ou complexas demais.

**Canal:** Grupos de ISPs no WhatsApp/Telegram, eventos de telecom, LinkedIn.

---

## Canais de Aquisição por Fase

### Fase 1 — Alpha/Beta (0–6 meses): Comunidade

**GitHub/Open Source:**
- README impecável com GIF demo, logo, badges
- `CONTRIBUTING.md` claro para atrair contribuidores
- Primeira versão funcional publicada
- Publicar no Hacker News (Show HN), Dev.to, TabNews

**Reddit e Fóruns:**
- r/selfhosted — público certo, adora soluções open source
- r/openwrt — usuários técnicos que vão contribuir
- Fórum Brasileiro de OpenWrt (grupos Facebook)

**YouTube:**
- Tutorial: "Monte seu hotspot Wi-Fi pago em 30 minutos"
- Tutorial: "Ganhe dinheiro com seu Starlink"
- Canal próprio do projeto

**Meta:** 500 estrelas no GitHub, 50 instalações ativas

---

### Fase 2 — v1.0 (6–12 meses): Crescimento Orgânico

**SEO (Site + Mapa Público):**
- Conteúdo: "Como monetizar sua conexão Starlink"
- Conteúdo: "Wi-Fi pago: guia completo para provedores"
- Páginas de cidade no mapa (centenas de URLs indexáveis)
- Backlinks: mencionar em fóruns, grupos, comunidades

**Grupos de WhatsApp:**
- Infiltrar grupos de "Provedores de Internet Interior"
- Grupos de "Starlink Brasil"
- Grupos de "LAN House Owners"
- Criar grupo oficial "Astrolink Provedores"

**Parceiros de hardware:**
- Parceria com GL.iNet Brasil para incluir Astrolink como opção
- Parceria com distribuidores de OpenWrt para mencionar no material

**Meta:** 2.000 instalações, 200 pagantes, R$ 10.000 MRR

---

### Fase 3 — Growth (12–24 meses): Escala

**Programa de afiliados:**
- Criadores de conteúdo tech BR: comissão por instalação que vira pago
- Técnicos de redes: comissão por indicação
- Rastreamento via link único

**Paid ads (baixo volume, alta segmentação):**
- Google Ads: "controle de acesso wifi hotspot", "software hotspot brasil"
- Facebook/Instagram: lookalike de instalações existentes
- Budget inicial: R$ 2.000/mês, escalar conforme ROI

**PR e Imprensa:**
- TechTudo, Canaltech, B9 (tech BR)
- Ângulo: "startup BR cria alternativa ao Mikrotik focada em comunidades rurais"

**Eventos:**
- FUTURECOM (principal evento de telecom BR)
- Eventos de ISPs regionais

**Meta:** 10.000 instalações, 1.000 pagantes, R$ 55.000 MRR

---

## Estratégia de Lançamento

### Semana -4 até -1: Pré-lançamento

- Landing page no ar com "Lista de espera"
- Posts no LinkedIn, Twitter, Threads sobre o problema que resolve
- Alcançar 500 pessoas na lista de espera
- Preparar 5 cases de uso iniciais (beta testers selecionados)

### Semana 0: Lançamento

- **Product Hunt:** lançar na terça-feira, às 00:01 PST
  - Hunter com muitos seguidores (pedir para influencer tech indicar)
  - Video demo de 2 minutos
  - Primeiro comentário preparado explicando o produto
- **Hacker News:** "Show HN: Astrolink — Open source hotspot management for Starlink communities"
- **GitHub:** release v1.0 com notas detalhadas
- **TabNews:** post em português sobre o problema e a solução
- **Dev.to e Medium:** artigo técnico sobre a arquitetura

### Pós-lançamento: semanas 1–4

- Responder todo feedback publicamente
- Publicar roadmap atualizado com base no feedback
- Primeira atualização de patch (mostra que o projeto está ativo)
- Publicar métricas de adoção (transparência atrai confiança)

---

## Mensagem de Marketing

### Para provedores comunitários:
> "Transforme sua conexão Starlink em renda. Configure em 30 minutos, sem ser técnico. Seus vizinhos pagam pelo PIX e você acompanha tudo pelo celular."

### Para LAN houses e estabelecimentos:
> "Chega de dar senha do Wi-Fi pra todo mundo. Cobre pelo acesso, controle quem usa e quanto tempo, tudo automático."

### Para a comunidade tech:
> "Alternativa open source ao Mikrotik Hotspot para gerenciamento de captive portals. Backend em Go, frontend SvelteKit, Supabase, sem vendor lock-in."

---

## Métricas de Sucesso

| Métrica | 3 meses | 6 meses | 12 meses |
|---|---|---|---|
| GitHub Stars | 200 | 800 | 3.000 |
| Instalações ativas | 100 | 500 | 3.000 |
| Pagantes | 10 | 100 | 800 |
| MRR | R$ 500 | R$ 5.000 | R$ 40.000 |
| NPS | — | > 50 | > 60 |
| Nós no mapa público | 30 | 200 | 1.500 |

---

## Defensibilidade a Longo Prazo

**Dados:** quanto mais nós na plataforma, mais difícil migrar (histórico, configurações, relatórios).

**Rede:** o mapa público só tem valor se tiver muitos nós — barreira de entrada enorme para concorrentes.

**Comunidade:** contribuidores open source melhoram o produto gratuitamente; concorrentes fechados não têm isso.

**Marca:** ser o primeiro a posicionar como "o hotspot software do provedor local brasileiro" é um moat de percepção.

**Distribuição:** parceria com fabricantes de hardware (pré-instalado) cria canal exclusivo.

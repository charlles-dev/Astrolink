# **Documentação de Infraestrutura e Redes (Hotspot Starlink)**

## **1\. Visão Geral da Topologia**

Este documento descreve o funcionamento em tempo de execução da infraestrutura física e lógica do provedor de Wi-Fi remoto. O sistema é composto por 4 elementos de hardware principais ligados em cascata.

### **1.1. O Caminho da Internet (Cabeamento)**

A regra de ligação dos cabos físicos deve ser rigorosamente seguida para que a rede funcione:

1. **Starlink (Fonte):** Recebe o sinal via satélite. O cabo de rede sai da porta nativa do roteador da Starlink Gen 5\.  
2. **TP-Link Archer C6 (O Guardião):** O cabo vindo da Starlink **entra obrigatoriamente na porta WAN (Azul)** deste roteador.  
3. **Notebook (O Cérebro):** Um cabo de rede sai da **porta LAN 1 (Amarela)** do TP-Link Archer e liga-se diretamente à porta de rede do Notebook.  
4. **Fonte PoE (Energia da Antena):** Um cabo sai da **porta LAN 2 (Amarela)** do TP-Link Archer e liga-se na porta "LAN" da Fonte PoE (aquela caixinha preta que vai na tomada).  
5. **Omada EAP225 (Transmissor):** Um cabo de rede blindado longo sai da porta **"POE"** da Fonte PoE, sobe o mastro e liga-se ao Access Point Omada lá no alto.

## **2\. Mapa de Endereços IP (Endereçamento Lógico)**

Para evitar conflitos de rede, os IPs essenciais foram fixados. Se precisar de aceder a qualquer painel de controlo, utilize os IPs abaixo no navegador do Notebook:

| Equipamento | Função | Endereço IP | Como Aceder |
| :---- | :---- | :---- | :---- |
| **Starlink** | Modem / Internet | 192.168.1.1 | App Starlink (via Wi-Fi original se ativo) |
| **TP-Link Archer** | Roteador / OpenWrt | 10.0.0.1 | Navegador: http://10.0.0.1 (User: root) |
| **Notebook** | Servidor / Python | 10.0.0.10 | IP Fixo via MAC no painel do OpenWrt |
| **Portal Captivo** | Tela de Login do Cliente | 10.0.0.10:8000 | Acedido automaticamente pelos clientes |
| **Omada EAP225** | Access Point Outdoor | 10.0.0.2 (Sugerido) | Configurar IP Estático na primeira instalação |

*Nota: A rede Wi-Fi original gerada pelo roteador Starlink e pelo TP-Link Archer (dentro da base) devem estar com senhas fortes e ocultas. Apenas a rede do Omada EAP225 (no alto do mastro) ficará Aberta (Sem Senha) para os clientes conectarem.*

## **3\. Gestão de Energia e Proteção**

Ambientes remotos e margens de rios sofrem com instabilidade de energia e raios.

* **Nobreak / UPS (Altamente Recomendado):** O Notebook tem bateria própria, mas a Starlink e o Roteador TP-Link não. Recomenda-se ligar todos os equipamentos (incluindo a Fonte PoE do Omada) num Nobreak para evitar que a rede caia durante pequenas oscilações de energia (piscas).  
* **Proteção contra Surtos (DPS):** Utilize um filtro de linha de boa qualidade (com DPS interno) entre a tomada da parede e os equipamentos para proteger contra picos de tensão.  
* **Isolamento de Humidade:** A Fonte PoE **não** é à prova de água. Ela deve ficar estritamente dentro da base segura (caixa hermética ou quarto), junto com o roteador e o Notebook. Apenas o cabo blindado sai para o tempo.

## **4\. Guia de Resolução de Problemas (Troubleshooting)**

Se o sistema parar, siga a árvore de diagnóstico abaixo antes de alterar qualquer código:

### **Problema A: "A rede Wi-Fi sumiu" (Ninguém vê o sinal)**

1. **Verifique a luz do Omada:** O EAP225 no mastro tem um LED verde. Se estiver apagado, o cabo que sobe o mastro rompeu ou a Fonte PoE queimou/está fora da tomada.  
2. **Teste o cabo:** Traga o Omada para baixo e ligue-o com um cabo curto direto na Fonte PoE para descartar problema no cabo longo.

### **Problema B: "Conecta no Wi-Fi, mas a tela de login não abre"**

1. **O Notebook está ligado?** Se o Windows/Linux suspender ou hibernar, o servidor Python para. O Notebook deve estar configurado para "Nunca suspender" nas definições de energia.  
2. **O IP está correto?** Abra o terminal do Notebook e digite ipconfig (Windows) ou ifconfig (Linux). Verifique se ele está com o IP 10.0.0.10. Se o roteador deu outro IP, o redirecionamento (FAS) do OpenNDS vai falhar.  
3. **O Roteador bloqueou?** Reinicie o TP-Link Archer tirando-o da tomada por 10 segundos.

### **Problema C: "O PIX não gera ou não confirma o pagamento"**

1. **Falta de Internet na Base:** O Notebook não está conseguindo falar com o MercadoPago. Verifique no app da Starlink se a antena principal tem sinal (pode haver nuvens muito densas ou obstrução física).  
2. **Erro de API:** Verifique o log (ecrã preto do terminal) do seu programa em Python no Notebook. O MercadoPago pode ter expirado o token de acesso.

### **Problema D: "Internet lenta para quem pagou"**

1. **Saturação de Banda:** Pode haver muitos utilizadores a assistir a vídeos ao mesmo tempo. Aceda ao painel do OpenWrt (10.0.0.1) e verifique o consumo de banda, ou ative regras de QoS (Quality of Service) para limitar cada cliente a, por exemplo, 5 Megas.
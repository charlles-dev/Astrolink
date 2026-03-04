# **Documentação de Arquitetura de Software: Sistema Hotspot Híbrido (PIX \+ Vouchers)**

## **1\. Visão Geral da Stack Tecnológica (Simplificada)**

A arquitetura foi otimizada para correr de forma leve num computador portátil comum, eliminando a necessidade de servidores RADIUS complexos e pesados.

1. **Roteador (O Guardião):** Corre **OpenWrt** com o pacote **OpenNDS**. Ele é o responsável por intercetar a navegação, redirecionar o utilizador para o Captive Portal e aplicar o "Walled Garden" (Jardim Murado) para os bancos.  
2. **Backend (O Cérebro):** Escrito em **Python** usando o framework **FastAPI**. Ele recebe os pedidos do OpenNDS, gera os códigos PIX, valida os Vouchers (PINs), gere o tempo dos utilizadores e manda o roteador libertar ou bloquear o acesso.  
3. **Base de Dados:** **SQLite** (vem embutido no Python). É leve, não requer instalação de servidores e é mais que suficiente para milhares de acessos numa operação local.  
4. **Frontend (A Cara do Negócio):** HTML/CSS simples (usando TailwindCSS) focado 100% em telemóveis (Mobile-First).

## **2\. A Base de Dados (Esquema Resumido SQLite)**

O sistema centraliza as regras de negócio em quatro tabelas principais:

* **Tabela usuarios:** Regista quem se está a ligar, usando o **MAC Address** (endereço físico) do telemóvel do cliente. Controla o momento exato do fim\_acesso.  
* **Tabela planos:** Os pacotes que vende de forma automatizada (ex: 1 Hora, 1 Dia, 1 Semana) e os seus preços.  
* **Tabela transacoes\_pix:** Rastreia as transações geradas via API bancária para saber se o cliente já pagou (necessário para o Polling).  
* **Tabela vouchers (Nova):** Armazena os códigos PIN alfanuméricos gerados previamente pelo administrador para serem vendidos a dinheiro físico.

## **3\. O Backend em Python (Endpoints / API)**

O seu script em Python terá de correr continuamente no portátil e responder às seguintes rotas (chamadas web):

* **Rotas de Frontend:**  
  * GET /login: Retorna a página HTML inicial do portal.  
* **Rotas de Transação (PIX):**  
  * POST /api/gerar-pix: Comunica com a API de pagamentos (ex: MercadoPago), gera o QR Code e liberta o Jardim Murado por 10 minutos no OpenNDS.  
  * GET /api/status-pix: O telemóvel do cliente chama esta rota a cada 5 segundos para perguntar *"O meu PIX já foi pago?"*.  
* **Rotas Híbridas (Dinheiro / Vouchers):**  
  * POST /api/resgatar-voucher: Recebe o PIN digitado pelo cliente. Se for válido, queima (inutiliza) o código, vincula o MAC Address e liberta a internet imediatamente.  
  * POST /api/admin/gerar-voucher: Rota oculta para o administrador gerar um lote de novos códigos para anotar em papel.

## **4\. O "Walled Garden" no OpenNDS (Bancos e PIX)**

Para que o cliente consiga pagar o PIX sem ter internet, o OpenNDS no roteador é configurado com uma lista de exceções temporárias (o *Jardim Murado*).

1. O IP do seu Portátil é sempre livre (para a página de login abrir).  
2. O OpenNDS liberta temporariamente os IPs/Domínios das APIs do banco.  
3. **Estratégia de Bloqueio:** Para evitar que o cliente tente aceder a outras coisas durante os 10 minutos, o Python aplica uma restrição de velocidade muito severa (ex: 128 kbps) e um bloqueio de DNS para redes sociais. A app do banco funciona, mas o YouTube não.

## **5\. Fluxos da Vida Real (Passo a Passo)**

O sistema agora atende a dois perfis de clientes simultaneamente.

### **Fluxo A: Cliente Autónomo (Paga com PIX)**

1. O cliente entra no barco, liga-se ao Wi-Fi e abre o navegador (Captive Portal).  
2. Escolhe o plano "24 Horas \- R$ 15".  
3. O Python gera o PIX e ativa o "Walled Garden" para aquele telemóvel no roteador por 10 minutos.  
4. O cliente abre a app do seu banco, cola o PIX e paga.  
5. O cliente aguarda no portal. O Python (via polling) deteta o pagamento.  
6. O Python atualiza a base de dados e envia um comando SSH para o roteador (ndsctl auth \<MAC\>) a libertar a velocidade máxima e o acesso livre.

### **Fluxo B: Cliente Tradicional (Paga em Dinheiro Vivo)**

1. O cliente chega até si com uma nota de R$ 50\.  
2. Recebe o dinheiro, abre o seu painel de admin e clica em "Gerar Voucher 7 Dias".  
3. O sistema devolve o código B4R-C0Z.  
4. O cliente liga-se ao Wi-Fi, mas em vez de clicar nos planos, insere B4R-C0Z no campo de PIN.  
5. O Python valida que o código existe e está disponível. Muda o estado do código para usado, calcula a hora de expiração (daqui a 7 dias) e liberta o acesso via SSH no roteador (ndsctl auth \<MAC\>).

### **Ponto de Convergência (A Expiração)**

Independentemente de como o cliente pagou (Fluxo A ou B), o Python tem um *Job* (tarefa de fundo) a correr a cada 1 minuto. Quando o relógio atingir o fim\_acesso daquele MAC Address, o Python executa ndsctl deauth \<MAC\> e a internet do cliente é cortada imediatamente, forçando-o a ver o portal de login de novo.
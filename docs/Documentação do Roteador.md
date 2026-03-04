# **Documentação de Configuração: Roteador (OpenWrt \+ OpenNDS)**

## **1\. Visão Geral**

O roteador TP-Link Archer C6/C7 funcionará com o firmware customizado **OpenWrt**. A sua única função é gerir o tráfego de rede: receber a internet da Starlink, distribuir os IPs locais, bloquear utilizadores não autorizados e forçá-los a ver o Captive Portal hospedado no seu Notebook.

### **Premissas de Rede (Exemplo Padrão)**

* **IP do Roteador (Gateway):** 10.0.0.1  
* **IP do Notebook (Servidor Python):** 10.0.0.10 (Este IP tem de ser fixo\!)  
* **Porta do Servidor Python:** 8000  
* **Interface da Rede Local (LAN):** br-lan

## **2\. Passo a Passo: Configuração Base da Rede**

Após instalar o OpenWrt no TP-Link Archer (processo feito enviando o ficheiro .bin na página de atualização original da TP-Link), deve configurar a rede básica:

1. **Ligar os Cabos:**  
   * Cabo da Starlink na porta **WAN** (azul).  
   * Cabo do Notebook na porta **LAN 1** (amarela).  
   * Cabo do Omada (via Fonte PoE) na porta **LAN 2** (amarela).  
2. **Fixar o IP do Notebook:**  
   No painel do OpenWrt (LuCI), vá a **Network \> DHCP and DNS \> Static Leases**.  
   * Adicione uma regra vinculando o *MAC Address* do seu Notebook ao IP 10.0.0.10. Isso garante que, mesmo que reinicie tudo, o roteador saberá sempre onde o portal está hospedado.

## **3\. Instalação do OpenNDS**

O OpenNDS é o software que faz a interceção do portal. Como o roteador já tem internet da Starlink, a instalação é feita pelo terminal do roteador (via SSH, usando o PuTTY ou o terminal do Windows).

1. Aceda ao roteador: ssh root@10.0.0.1  
2. Atualize a lista de pacotes e instale:  
   opkg update  
   opkg install opennds

## **4\. Configuração do OpenNDS (/etc/config/opennds)**

Este é o "coração" da configuração. Vai editar o ficheiro de configuração do OpenNDS para o ligar ao seu sistema em Python.

Pode editar via terminal com o comando vi /etc/config/opennds ou através da interface web do OpenWrt (se tiver o pacote luci-app-opennds instalado).

### **4.1. Configuração Básica e FAS (Redirecionamento)**

Altere/adicione as seguintes linhas para ativar o FAS (Forwarding Authentication Service), que diz ao roteador para não usar a página de login padrão dele, mas sim a página que está no seu Python.

config opennds  
    option enabled '1'  
    option gatewayinterface 'br-lan'  
    option gatewayname 'Starlink Rio Wi-Fi'  
      
    \# \--- CONFIGURAÇÃO DO FAS (APONTANDO PARA O PYTHON) \---  
    option fasport '8000'  
    option faspath '/login'  
    option fasremoteip '10.0.0.10'  
    option fassl '0'  
      
    \# Tempo em minutos que o cliente pode ficar inativo antes de ser desconectado  
    option sessiontimeout '1440' 

*O que isto faz:* Quando o cliente liga-se, o OpenNDS constrói automaticamente a seguinte URL e envia o telemóvel para lá: http://10.0.0.10:8000/login?mac=XX:XX...\&tok=123456.

### **4.2. O Jardim Murado (Walled Garden)**

Aqui definimos o que os clientes podem aceder *antes* de estarem autenticados (ou durante os 10 minutos de cortesia do PIX).

Ainda no mesmo ficheiro /etc/config/opennds, procure pelas secções preauthenticated\_users e authenticated\_users.

    \# \--- REGRAS PARA QUEM JÁ PAGOU (ACESSO TOTAL) \---  
    list authenticated\_users 'allow all'

    \# \--- REGRAS DO JARDIM MURADO (ANTES DE PAGAR) \---  
    \# 1\. Obrigatoriamente permitir que o cliente aceda à página do seu Notebook\!  
    list preauthenticated\_users 'allow tcp port 8000 to 10.0.0.10'  
      
    \# 2\. Permitir DNS (necessário para os telemóveis funcionarem)  
    list preauthenticated\_users 'allow tcp port 53'  
    list preauthenticated\_users 'allow udp port 53'  
      
    \# 3\. Liberação dos Bancos/PIX (Exemplo)  
    \# Dependendo da versão do OpenNDS, usa-se IPs (ASNs) ou domínios (FQDN).  
    \# Como os IPs mudam muito, o ideal é usar as regras WalledGarden por domínio (FQDN) do OpenNDS.  
      
    list walledgarden\_custom 'api.mercadopago.com'  
    list walledgarden\_custom 'nubank.com.br'  
    list walledgarden\_custom 'bradesco.com.br'  
    \# (Adicionar os domínios essenciais dos bancos locais)

**Nota Estratégica:** Como vimos na documentação de arquitetura, garantir que **todos** os bancos funcionem perfeitamente apenas com listas de domínios pode dar dores de cabeça, pois os bancos usam dezenas de serviços Cloud invisíveis.

*Alternativa Avançada (Traffic Shaping):* Quando o seu Python recebe o pedido de PIX, em vez de manter o cliente no preauthenticated\_users, o Python autoriza o cliente no OpenNDS, mas aplica um limite de velocidade extremo no próprio roteador (ex: ndsctl auth \<MAC\> \--download 128 \--upload 128). Aos 128kbps, as redes sociais não carregam, mas os bancos funcionam. Aos 10 minutos, se não pagou, o Python corta (ndsctl deauth \<MAC\>).

## **5\. Como o Python "Conversa" com o Roteador**

O seu código Python (no Notebook) vai precisar de executar comandos no roteador para libertar ou cortar a internet. Isso é feito através do comando ndsctl do OpenNDS.

Como o Python está noutra máquina (10.0.0.10), ele tem duas formas de dar ordens ao roteador (10.0.0.1):

1. **Via SSH:** O Python faz um login silencioso por SSH no roteador e roda o comando.  
   * *Comando para Libertar Internet Total:* ssh root@10.0.0.1 "ndsctl auth AA:BB:CC:DD:EE:FF"  
   * *Comando para Cortar Internet:* ssh root@10.0.0.1 "ndsctl deauth AA:BB:CC:DD:EE:FF"  
2. **Via FAS API:** O OpenNDS expõe a própria API se for configurado para isso, permitindo que o Python envie requisições HTTP para libertar clientes.

*(Para o nosso projeto, a execução via SSH com a biblioteca paramiko do Python é geralmente a mais fácil de implementar e debugar).*

## **6\. Resumo da Instalação (Checklist)**

* \[ \] Roteador TP-Link com OpenWrt instalado.  
* \[ \] IP Fixo configurado para o Notebook (10.0.0.10).  
* \[ \] Pacote opennds instalado.  
* \[ \] Ficheiro /etc/config/opennds editado para apontar para 10.0.0.10:8000.  
* \[ \] Regras de preauthenticated\_users configuradas para permitir o tráfego local para o Notebook.  
* \[ \] Serviço OpenNDS reiniciado: /etc/init.d/opennds restart.
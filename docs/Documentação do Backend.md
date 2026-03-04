# **Documentação do Backend: Servidor Python (Hotspot Híbrido: PIX \+ Vouchers)**

## **1\. Visão Geral e Responsabilidades**

O servidor Backend funcionará no computador portátil ligado à rede local (atrás do roteador OpenWrt). Ele é o responsável por orquestrar toda a lógica de negócio do seu provedor de Wi-Fi.

**Principais funções:**

1. Servir os ficheiros HTML/JS do Frontend (Captive Portal).  
2. Gerar os códigos PIX comunicando com a API bancária.  
3. **Validar códigos PIN (Vouchers) inseridos por clientes que pagaram em dinheiro físico.**  
4. Comunicar com o roteador (OpenNDS) para libertar ou bloquear o acesso à internet de um endereço MAC.  
5. Gerir o tempo limite de cada utilizador e cortar o acesso quando o tempo expirar.

## **2\. Stack Tecnológica**

* **Linguagem:** Python 3.10+  
* **Framework Web:** FastAPI (extremamente rápido e nativamente assíncrono).  
* **Servidor ASGI:** Uvicorn (para correr a aplicação FastAPI).  
* **Base de Dados:** SQLite (via biblioteca SQLAlchemy para ORM).  
* **Agendador de Tarefas:** APScheduler (para verificar de minuto a minuto quem tem o tempo expirado).

## **3\. Estrutura da Base de Dados (Resumo)**

*(Veja o ficheiro documentacao\_banco\_dados.md para a estrutura completa).*

O backend irá interagir com 4 tabelas principais:

* usuarios: Identificados pelo MAC Address.  
* planos: Os pacotes disponíveis para venda via PIX.  
* transacoes\_pix: Guarda o ID do PIX e verifica se já foi pago (Polling).  
* **vouchers (Novo):** Guarda os códigos alfanuméricos gerados pelo administrador para venda em dinheiro vivo.

## **4\. Endpoints (Rotas da API)**

O sistema agora possui rotas divididas em três categorias: Rotas Públicas (Frontend), Rotas de Negócio (PIX e Vouchers) e Rotas de Administração.

### **4.1. Rotas de Interface (Frontend)**

* **GET /login**  
  * **Ação:** Retorna o ficheiro index.html. É a primeira rota chamada pelo OpenNDS.

### **4.2. Rotas de Negócio (PIX)**

* **POST /api/gerar-pix**  
  * **Ação:** Recebe o mac\_address e o plano\_id. Cria a intenção de pagamento no MercadoPago, ativa o "Walled Garden" (10 minutos) no roteador e retorna o QR Code.  
* **GET /api/status-pix/{txid}**  
  * **Ação:** Rota de *Polling* para o telemóvel do cliente verificar se o PIX já "caiu". Liberta a internet total se o status for "pago".

### **4.3. Rotas de Negócio (Vouchers / Dinheiro Físico)**

* **POST /api/resgatar-voucher**  
  * **Ação:** Recebe o mac\_address e o codigo (PIN) digitado pelo cliente.  
  * **Lógica Interna:** 1\. Procura o código na tabela vouchers.  
    2\. Se o status for disponivel, muda para usado.  
    3\. Regista o mac\_address\_usado e a data\_uso.  
    4\. Atualiza o utilizador para ativo e calcula o fim\_acesso com base na duração do voucher.  
    5\. Dispara o comando de liberação para o OpenNDS (ndsctl auth \<MAC\>).  
  * **Retorno:** JSON indicando {"sucesso": true} ou {"sucesso": false, "erro": "Código inválido"}.

### **4.4. Rotas de Administração (Apenas para o Dono)**

Estas rotas devem ter uma proteção simples (como uma password no cabeçalho ou um ecrã de login separado) para que os clientes não acedam.

* **POST /api/admin/gerar-voucher**  
  * **Ação:** Gera um ou mais códigos aleatórios (ex: K9M-3XT) vinculados a uma duração específica.  
  * **Parâmetros Esperados:** {"duracao\_minutos": 60, "quantidade": 1}  
  * **Retorno:** A lista de códigos gerados para o administrador anotar ou imprimir.

## **5\. Como o Python "Conversa" com o Roteador (OpenNDS)**

Sempre que um PIX for confirmado ou um **Voucher for resgatado com sucesso**, o Python precisa executar um comando no roteador para abrir o acesso à internet.

Como o Python está noutra máquina (10.0.0.10) e o roteador em outra (10.0.0.1), o método mais seguro é o Python enviar um comando via SSH usando a biblioteca paramiko:

import paramiko

def libertar\_mac\_no\_roteador(mac\_address):  
    cliente \= paramiko.SSHClient()  
    cliente.set\_missing\_host\_key\_policy(paramiko.AutoAddPolicy())  
    \# Liga ao roteador OpenWrt  
    cliente.connect('10.0.0.1', username='root', password='sua\_senha')  
      
    \# Executa o comando do OpenNDS para autorizar a navegação  
    stdin, stdout, stderr \= cliente.exec\_command(f'ndsctl auth {mac\_address}')  
    cliente.close()

## **6\. Tarefas em Segundo Plano (Background Jobs)**

Usando o APScheduler, o Python executa tarefas de manutenção invisíveis:

1. **Job de Expiração de Acesso (Corre a cada 1 minuto):**  
   * Faz uma query: SELECT mac\_address FROM usuarios WHERE status \= 'ativo' AND fim\_acesso \< AGORA().  
   * Para cada utilizador expirado, altera o status para bloqueado no SQLite.  
   * Aciona a função SSH executando o comando de corte: ndsctl deauth \<MAC\_ADDRESS\>.  
     *(Nota: Isso afeta tanto quem pagou por PIX quanto quem usou Voucher. O corte é implacável baseado no tempo).*  
2. **Job de Limpeza de PIX (Corre a cada 15 minutos):**  
   * Cancela as transações PIX pendentes que já passaram do limite de 10 minutos.
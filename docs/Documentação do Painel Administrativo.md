# **Documentação do Painel Administrativo (Frontend & Backend)**

## **1\. Visão Geral**

O Painel Administrativo é a interface de gestão exclusiva para o dono do provedor (Administrador). Ele não fica hospedado na nuvem, mas sim localmente no Notebook (Servidor Python).

* **Como Acessar:** Pelo navegador do próprio Notebook (ou de um dispositivo autorizado na rede local) através do endereço http://localhost:8000/admin ou http://10.0.0.10:8000/admin.  
* **Objetivo:** Acompanhar o faturamento, gerar Vouchers para vendas em dinheiro e gerenciar os usuários conectados.

## **2\. Segurança e Autenticação (Backend)**

Como o painel roda na mesma rede em que os clientes estão conectados, a segurança é primordial.

* **Proteção de Rotas:** Todas as rotas que começam com /api/admin/... ou /admin/... exigirão autenticação.  
* **Método:** Autenticação baseada em Sessão (Cookies) ou Token JWT (JSON Web Token).  
* **Credenciais Únicas:** Haverá apenas um usuário (o Administrador) configurado diretamente no arquivo .env do Python, sem necessidade de tabela de "funcionários" no banco de dados para simplificar.  
  * Exemplo no .env: ADMIN\_USER="astrolink", ADMIN\_PASS="senha\_super\_segura"

## **3\. Estrutura do Frontend (Telas do Admin)**

O painel será construído com **HTML/CSS (TailwindCSS)** e **JavaScript**, com um design "Dashboard" limpo, preferencialmente em modo escuro (Dark Mode) para conforto visual.

### **Tela 1: Login**

* **Elementos:** Campo de Usuário, Campo de Senha, Botão "Entrar".  
* **Ação:** O JS envia as credenciais para o backend. Se correto, recebe um Token de acesso e redireciona para o Dashboard.

### **Tela 2: Dashboard (Visão Geral)**

* **Elementos:**  
  * **Cards de Métricas:** \* "Faturamento Hoje (PIX \+ Dinheiro)"  
    * "Usuários Online Agora"  
    * "Vouchers Disponíveis"  
  * **Gráfico Simples (Opcional):** Vendas dos últimos 7 dias.

### **Tela 3: Gestão de Vouchers (A Tela do Dinheiro)**

* **Elementos:**  
  * **Formulário de Geração:** Select para escolher o plano (ex: 24 Horas) e input para "Quantidade" (ex: gerar 5 códigos). Botão "Gerar Vouchers".  
  * **Área de Resultado:** Um quadro grande mostrando os códigos recém-gerados (ex: A7X-92P, M4K-1LW) com um botão de "Imprimir" ou "Copiar".  
  * **Tabela de Histórico:** Lista dos últimos vouchers gerados, mostrando o Status (disponivel ou usado) e quando foi ativado.

### **Tela 4: Gestão de Clientes (Controle da Rede)**

* **Elementos:**  
  * **Tabela de Usuários Ativos:** Mostra o MAC Address, o Plano que compraram, se foi PIX ou Voucher, e uma contagem regressiva de quanto tempo falta para a internet deles cair.  
  * **Botão "Derrubar" (Kick):** Um botão vermelho ao lado de cada usuário.  
* **Ação:** Se você suspeitar que alguém burlou a rede ou precisa derrubar um cliente, clica no botão. O JS chama a rota de "kick" no Backend, que entra via SSH no roteador e corta a internet do MAC Address na hora.

## **4\. Endpoints do Backend (API Admin)**

O servidor FastAPI (Python) terá um conjunto de rotas separadas apenas para o painel.

### **4.1. Autenticação**

* **POST /api/admin/login**  
  * **Ação:** Valida o usuário e senha com as variáveis do .env. Retorna um token JWT ou define um Cookie de sessão seguro.

### **4.2. Dados do Dashboard**

* **GET /api/admin/dashboard-stats**  
  * **Ação:** Faz consultas COUNT e SUM no banco SQLite para retornar o faturamento do dia e total de clientes ativos.  
  * **Retorno:** {"faturamento\_hoje": 150.00, "online\_agora": 12}

### **4.3. Módulo de Vouchers**

* **POST /api/admin/gerar-vouchers**  
  * **Ação:** Recebe {"plano\_id": 2, "quantidade": 5}. Gera códigos alfanuméricos aleatórios (excluindo letras que confundem, como 'O' e '0'), insere na tabela vouchers e retorna a lista.  
* **GET /api/admin/vouchers**  
  * **Ação:** Retorna a lista paginada dos vouchers para preencher a tabela de histórico.

### **4.4. Módulo de Controle de Rede**

* **GET /api/admin/usuarios-ativos**  
  * **Ação:** Retorna todos os usuários onde status \== 'ativo'.  
* **POST /api/admin/derrubar-usuario**  
  * **Ação:** Recebe um mac\_address. Altera o status dele no SQLite para bloqueado, zera o fim\_acesso e **executa imediatamente o comando ndsctl deauth \<MAC\> via SSH no roteador OpenWrt**.
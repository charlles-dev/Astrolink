# **Documentação da Base de Dados: Hotspot PIX e Vouchers**

## **1\. Visão Geral e Arquitetura**

* **Motor de Base de Dados:** SQLite 3  
* **Localização:** Ficheiro local no computador portátil (ex: hotspot\_producao.db).  
* **Justificativa:** O SQLite é ideal para sistemas embutidos e locais. Como o volume de transações não será de dezenas de milhares por segundo, ele oferece um desempenho excelente com **custo zero** de infraestrutura e manutenção (basta copiar o ficheiro .db para um pen drive para fazer um backup completo).

## **2\. Diagrama de Entidade-Relacionamento (DER)**

Abaixo está o mapeamento visual de como as tabelas se relacionam. Temos quatro entidades principais: **Utilizadores**, **Planos** (PIX), **Transações PIX** e agora os **Vouchers** (para pagamento em dinheiro físico).

erDiagram  
    USUARIOS ||--o{ TRANSACOES\_PIX : "realiza"  
    PLANOS ||--o{ TRANSACOES\_PIX : "gera"  
    USUARIOS ||--o{ VOUCHERS : "resgata"

    USUARIOS {  
        INTEGER id PK  
        VARCHAR mac\_address UK  
        VARCHAR status  
        DATETIME fim\_acesso  
        VARCHAR ip\_atual  
        DATETIME criado\_em  
    }

    PLANOS {  
        INTEGER id PK  
        VARCHAR nome  
        DECIMAL preco  
        INTEGER duracao\_minutos  
        BOOLEAN ativo  
    }

    TRANSACOES\_PIX {  
        INTEGER id PK  
        VARCHAR txid UK  
        VARCHAR mac\_address FK  
        INTEGER plano\_id FK  
        VARCHAR status  
        DATETIME data\_criacao  
    }

    VOUCHERS {  
        INTEGER id PK  
        VARCHAR codigo UK  
        INTEGER duracao\_minutos  
        VARCHAR status  
        VARCHAR mac\_address\_usado FK  
        DATETIME data\_criacao  
        DATETIME data\_uso  
    }

## **3\. Dicionário de Dados (Estrutura das Tabelas)**

### **3.1. Tabela: usuarios**

Regista os dispositivos (telemóveis, tablets) que se ligam à rede. A chave de identificação é o Endereço MAC do dispositivo.

| Coluna | Tipo (SQLite) | Restrições | Descrição |
| :---- | :---- | :---- | :---- |
| id | INTEGER | PRIMARY KEY, AUTOINCREMENT | Identificador único interno. |
| mac\_address | VARCHAR(17) | UNIQUE, NOT NULL | Endereço físico do telemóvel (ex: 00:1A:2B:3C:4D:5E). |
| status | VARCHAR(20) | NOT NULL, DEFAULT 'bloqueado' | Estado atual (bloqueado, walled\_garden, ativo). |
| fim\_acesso | DATETIME | NULLABLE | Data e hora exata em que a internet deve ser cortada. |
| ip\_atual | VARCHAR(15) | NULLABLE | O último IP local (ex: 10.0.0.55) atribuído pelo roteador. |
| criado\_em | DATETIME | DEFAULT CURRENT\_TIMESTAMP | Data em que o dispositivo conectou pela primeira vez. |

### **3.2. Tabela: planos**

Catálogo de pacotes de internet disponíveis para venda automática via PIX.

| Coluna | Tipo (SQLite) | Restrições | Descrição |
| :---- | :---- | :---- | :---- |
| id | INTEGER | PRIMARY KEY, AUTOINCREMENT | Identificador do plano. |
| nome | VARCHAR(50) | NOT NULL | Nome de exibição (ex: "Pacote 1 Hora", "Passe 24 Horas"). |
| preco | DECIMAL(10,2) | NOT NULL | Valor em Reais (R$). |
| duracao\_minutos | INTEGER | NOT NULL | Tempo de acesso que o plano concede (ex: 60, 1440). |
| ativo | BOOLEAN | DEFAULT 1 (True) | Se 1, aparece no portal. Se 0, fica oculto. |

### **3.3. Tabela: transacoes\_pix**

Regista as intenções de compra via PIX para verificação de pagamento.

| Coluna | Tipo (SQLite) | Restrições | Descrição |
| :---- | :---- | :---- | :---- |
| id | INTEGER | PRIMARY KEY, AUTOINCREMENT | Identificador interno. |
| txid | VARCHAR(100) | UNIQUE, NOT NULL | ID da transação gerado pela API do banco (MercadoPago). |
| mac\_address | VARCHAR(17) | FOREIGN KEY, NOT NULL | O dispositivo que tentou comprar. |
| plano\_id | INTEGER | FOREIGN KEY, NOT NULL | O plano selecionado. |
| status | VARCHAR(20) | NOT NULL, DEFAULT 'pendente' | Estado do pagamento (pendente, pago, expirado). |
| data\_criacao | DATETIME | DEFAULT CURRENT\_TIMESTAMP | Quando o QR Code foi gerado. |

### **3.4. Tabela: vouchers (Nova)**

Regista os códigos gerados manualmente pelo administrador para vendas em dinheiro físico.

| Coluna | Tipo (SQLite) | Restrições | Descrição |
| :---- | :---- | :---- | :---- |
| id | INTEGER | PRIMARY KEY, AUTOINCREMENT | Identificador interno. |
| codigo | VARCHAR(20) | UNIQUE, NOT NULL | O código que o cliente vai digitar (ex: A7X-92P). |
| duracao\_minutos | INTEGER | NOT NULL | Tempo de acesso que este código concede. |
| status | VARCHAR(20) | NOT NULL, DEFAULT 'disponivel' | Estado do código (disponivel, usado, cancelado). |
| mac\_address\_usado | VARCHAR(17) | FOREIGN KEY, NULLABLE | O MAC Address do telemóvel que utilizou este código. Fica nulo até ser resgatado. |
| data\_criacao | DATETIME | DEFAULT CURRENT\_TIMESTAMP | Quando o administrador gerou o código. |
| data\_uso | DATETIME | NULLABLE | Data e hora exata em que o cliente digitou o código no portal. |

## **4\. Scripts de Criação (SQL DDL)**

\-- Criar a tabela de Utilizadores  
CREATE TABLE usuarios (  
    id INTEGER PRIMARY KEY AUTOINCREMENT,  
    mac\_address VARCHAR(17) UNIQUE NOT NULL,  
    status VARCHAR(20) NOT NULL DEFAULT 'bloqueado',  
    fim\_acesso DATETIME,  
    ip\_atual VARCHAR(15),  
    criado\_em DATETIME DEFAULT CURRENT\_TIMESTAMP  
);

\-- Criar a tabela de Planos  
CREATE TABLE planos (  
    id INTEGER PRIMARY KEY AUTOINCREMENT,  
    nome VARCHAR(50) NOT NULL,  
    preco DECIMAL(10,2) NOT NULL,  
    duracao\_minutos INTEGER NOT NULL,  
    ativo BOOLEAN DEFAULT 1  
);

\-- Criar a tabela de Transações PIX  
CREATE TABLE transacoes\_pix (  
    id INTEGER PRIMARY KEY AUTOINCREMENT,  
    txid VARCHAR(100) UNIQUE NOT NULL,  
    mac\_address VARCHAR(17) NOT NULL,  
    plano\_id INTEGER NOT NULL,  
    status VARCHAR(20) NOT NULL DEFAULT 'pendente',  
    data\_criacao DATETIME DEFAULT CURRENT\_TIMESTAMP,  
    FOREIGN KEY (mac\_address) REFERENCES usuarios(mac\_address),  
    FOREIGN KEY (plano\_id) REFERENCES planos(id)  
);

\-- Criar a tabela de Vouchers  
CREATE TABLE vouchers (  
    id INTEGER PRIMARY KEY AUTOINCREMENT,  
    codigo VARCHAR(20) UNIQUE NOT NULL,  
    duracao\_minutos INTEGER NOT NULL,  
    status VARCHAR(20) NOT NULL DEFAULT 'disponivel',  
    mac\_address\_usado VARCHAR(17),  
    data\_criacao DATETIME DEFAULT CURRENT\_TIMESTAMP,  
    data\_uso DATETIME,  
    FOREIGN KEY (mac\_address\_usado) REFERENCES usuarios(mac\_address)  
);

\-- Inserir os planos padrão (Semente/Seed)  
INSERT INTO planos (nome, preco, duracao\_minutos) VALUES   
('Acesso Rápido \- 1 Hora', 5.00, 60),  
('Diária Completa \- 24 Horas', 15.00, 1440),  
('Passe Semanal \- 7 Dias', 50.00, 10080);

## **5\. Estratégia de Cópias de Segurança (Backups)**

1. **Backup Automático (Local):** O próprio script Python será configurado para, todos os dias à meia-noite, fazer uma cópia do ficheiro hotspot.db para uma pasta chamada backups/hotspot\_YYYY\_MM\_DD.db.  
2. **Backup Externo (Pen Drive):** Recomenda-se ligar uma Pen Drive ao portátil e configurar um pequeno script no Windows/Linux para copiar os ficheiros de backup para essa Pen Drive diariamente.
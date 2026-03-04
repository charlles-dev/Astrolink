# **🌌 Astrolink**

**Sistema de Gestão de Hotspot Wi-Fi Híbrido para Zonas Remotas (Starlink \+ PIX Automatizado \+ Vouchers em Dinheiro).**

O **Astrolink** é uma solução de infraestrutura e software de baixo custo (Zero-Cost Software) projetada para democratizar e rentabilizar o acesso à internet em áreas remotas (comunidades ribeirinhas, rotas turísticas isoladas, zonas rurais) onde as redes móveis tradicionais não chegam.

Utilizando uma antena **Starlink** como *backhaul*, um roteador comum com **OpenWrt** e um servidor local em **Python**, o sistema cria um Captive Portal avançado. O projeto opera num **modelo híbrido**, permitindo tanto a compra autónoma via PIX (através de um "Walled Garden") quanto a venda tradicional a dinheiro físico (através de Vouchers/PINs gerados pelo administrador).

## **🚀 Principais Funcionalidades**

* **Suporte a Pagamentos Híbridos:** Venda 100% automatizada via PIX ou venda manual com geração de códigos PIN (Vouchers) para clientes com dinheiro físico.
* **Captive Portal Automatizado:** Interceção de tráfego de utilizadores não autenticados na rede Wi-Fi.
* **Walled Garden para PIX:** Liberação temporária (10 minutos) restrita a IPs/Domínios de bancos e APIs de pagamento para quem não tem internet.
* **Bypass de CGNAT da Starlink:** Sistema de *Active Polling* local para confirmar recebimentos de PIX, contornando a falta de IP público fixo da Starlink.
* **Gestão Totalmente Local (Offline-First):** Base de dados SQLite e servidor FastAPI a correr localmente. A gestão do tempo e dos acessos não depende de servidores Cloud.
* **Desconexão Automática:** Corte de acesso imediato e implacável após a expiração do tempo contratado.

## **🛠️ Stack Tecnológica e Arquitetura**

O projeto foi desenhado para utilizar hardware acessível e 100% de software Open-Source.

### **Software**

* **Roteamento e Firewall:** OpenWrt \+ OpenNDS (Forwarding Authentication Service \- FAS)
* **Backend:** Python 3.10+ \+ FastAPI (Orquestração, Integração PIX e Gestão de Vouchers)
* **Base de Dados:** SQLite \+ SQLAlchemy (Leve, 4 tabelas relacionais)
* **Frontend (Portal):** HTML5, CSS3 (TailwindCSS) e Vanilla JS (Mobile-First)

### **Hardware Base Recomendado**

* **Internet:** Starlink (Gen 5 ou superior)
* **Access Point Outdoor:** TP-Link Omada EAP225-Outdoor (Cobertura 360º de longo alcance)
* **Roteador Gestor:** TP-Link Archer C6/C7 (Compatível com OpenWrt)
* **Servidor Local:** Computador Portátil padrão (baixo consumo, com bateria nativa para quedas de energia)

## **🗺️ Como Funciona (A Jornada do Utilizador)**

O sistema atende a dois perfis de clientes em simultâneo:

### **Fluxo A: Cliente Autónomo (Paga com PIX)**

1. **Ligação:** O cliente liga-se à rede "Astrolink" e o portal abre.
2. **Seleção e Walled Garden:** Escolhe um plano (ex: 24 Horas). O backend gera o PIX e liberta a rede apenas para apps bancárias por 10 minutos.
3. **Pagamento:** O cliente abre a sua app do banco, paga, e o servidor deteta automaticamente.
4. **Liberação:** O Python envia um comando via SSH ao roteador para libertar a internet total.

### **Fluxo B: Cliente Tradicional (Paga em Dinheiro)**

1. **Compra Física:** O cliente entrega o dinheiro em espécie ao administrador.
2. **Geração:** O administrador gera um PIN no seu painel oculto (ex: B4R-C0Z) e entrega ao cliente.
3. **Resgate:** O cliente liga-se à rede, insere o PIN na página inicial e o sistema liberta o acesso imediatamente, calculando a data de expiração.

## **⚙️ Configuração do Ambiente (Instalação)**

*(Instruções detalhadas para replicação do projeto. Consulte a documentação completa na pasta docs/ do repositório).*

### **1\. Configuração do Roteador (OpenWrt)**

O roteador atua como Gateway (10.0.0.1) e deve ter o OpenNDS configurado para redirecionar utilizadores para o servidor Python (10.0.0.10:8000).

### **2\. Configuração do Backend (Python)**

### **2\. Configuração do Backend (Python)**

```bash
# Clonar o repositório
git clone https://github.com/seu-usuario/astrolink.git
cd astrolink/backend

# Criar o ambiente virtual e instalar dependências
python -m venv venv
source venv/bin/activate # ou venv\Scripts\activate no Windows
pip install -r requirements.txt

# Configurar as variáveis de ambiente
cp .env.example .env
# Edite o .env com o seu Token do MercadoPago e passwords locais

# Iniciar o servidor
uvicorn main:app --host 0.0.0.0 --port 8000
```

## **🏗️ Estrutura Física (Mastro e Topologia)**

Para garantir segurança contra furtos, animais e intempéries em zonas remotas, o projeto recomenda uma topologia de **Torre Única**:

* Toda a inteligência (Starlink, Roteador, Portátil) fica trancada numa base segura.
* Apenas um cabo blindado sobe o mastro (via PoE) para alimentar a antena Omada EAP225-Outdoor.
* Desenhos de serralheria e estrutura modular do mastro em aço galvanizado estão disponíveis na documentação.

## **🤝 Contribuição**

Contribuições são muito bem-vindas\! Se mora em áreas remotas, atua com telecomunicações rurais ou é programador Python/OpenWrt, sinta-se à vontade para abrir uma *Issue* ou enviar um *Pull Request*.

## **📄 Licença**

Este projeto está sob a licença MIT \- veja o ficheiro [LICENSE](https://www.google.com/search?q=LICENSE) para mais detalhes.
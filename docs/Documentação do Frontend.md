# **Documentação do Frontend: Captive Portal (Hotspot Híbrido: PIX \+ Vouchers)**

## **1\. Visão Geral e Princípios de Design**

O Frontend é a interface visual que o cliente vê assim que liga o telemóvel à sua rede Wi-Fi Starlink. O OpenNDS (no roteador) interceta a navegação e força a abertura destas páginas.

**Princípios Fundamentais:**

* **Mobile-First:** 99% dos acessos serão via telemóvel. O design tem de ser responsivo, com botões grandes e de fácil leitura sob a luz do sol.  
* **Extrema Leveza:** Como a página é carregada *antes* de a internet estar totalmente liberada, não devemos usar imagens pesadas ou bibliotecas complexas.  
* **Modelo Híbrido:** O portal deve oferecer claramente duas opções: escolher um plano para pagar via PIX **OU** inserir um código (PIN) para quem já pagou em dinheiro.

## **2\. Stack Tecnológica Recomendada**

Para manter o custo zero e o desempenho máximo no seu notebook local:

* **HTML5 e CSS3:** Estrutura base.  
* **Tailwind CSS (via CDN):** Para uma estilização rápida e moderna.  
* **Vanilla JavaScript (JS puro):** Para gerir os cliques, extrair dados do URL e comunicar com o seu backend em Python.  
* **FontAwesome / Lucide (Opcional):** Apenas para ícones básicos (Wi-Fi, cadeado, dinheiro).

## **3\. Estrutura de Ecrãs (Views)**

O sistema precisará de apenas 3 ecrãs principais. Pode ser construído como uma *Single Page Application* (SPA) simples, escondendo e mostrando \<div\>s com JavaScript.

### **Ecrã 1: Seleção de Plano & Resgate de Voucher (Landing Page)**

É a página inicial. O OpenNDS envia o utilizador para cá anexando o endereço MAC dele no URL.

* **Elementos UI:**  
  * Logótipo do seu projeto (ex: "Astrolink Wi-Fi").  
  * **Área A (Automático):** Botões de Planos (ex: "1 Hora \- R$ 5", "24 Horas \- R$ 15"). *Ação: Gera o PIX e vai para o Ecrã 2\.*  
  * **Divisor:** "OU"  
  * **Área B (Dinheiro Físico):** Um campo de texto grande \[ \_ \_ \_ \- \_ \_ \_ \] com o placeholder "Tem um código PIN?".  
  * Botão: *"Ativar Internet"*. *Ação: Valida o PIN e vai direto para o Ecrã 3\.*

### **Ecrã 2: Pagamento PIX (Jardim Murado Ativo)**

Abre se o cliente escolheu um plano PIX.

* **Elementos UI:**  
  * Cronómetro em contagem decrescente (ex: "Tem 09:59 minutos para pagar").  
  * **Imagem do QR Code PIX** e **Código "Copia e Cola"**.  
  * Botão grande: *"Copiar Código PIX"*.  
  * Instrução: *"A sua internet para apps de banco está liberada por 10 minutos."*  
  * Indicador de carregamento: *"A aguardar pagamento..."*  
* **Ação do JS:** A página faz *Polling* (verifica a cada 5 segundos) no Python. Se pago, vai para o Ecrã 3\.

### **Ecrã 3: Sucesso (Acesso Liberado)**

Abre quando o PIX é pago OU quando um Voucher é validado com sucesso.

* **Elementos UI:**  
  * Ícone verde de confirmação.  
  * Mensagem: *"Acesso Liberado\! Boa navegação."*  
  * Resumo: *"Tempo disponível: 24 Horas."*  
  * Botão: *"Começar a Navegar"*.  
* **Ação do JS:** Ao carregar, o sistema avisa o roteador OpenNDS que o MAC tem permissão total.

## **4\. Comunicação Frontend \<-\> Backend (Lógica JavaScript)**

### **Passo A: Capturar Dados Iniciais (OpenNDS \-\> Frontend)**

// Exemplo de URL vinda do roteador: \[http://10.0.0.10:8000/?mac=00:1A:2B:3C:4D:5E\&tok=xyz\](http://10.0.0.10:8000/?mac=00:1A:2B:3C:4D:5E\&tok=xyz)  
const urlParams \= new URLSearchParams(window.location.search);  
const userMac \= urlParams.get('mac');  
const gatewayToken \= urlParams.get('tok'); 

### **Passo B1: Solicitar o PIX**

Se o cliente clica no plano "1 Hora":

fetch('/api/gerar-pix', {  
    method: 'POST',  
    headers: { 'Content-Type': 'application/json' },  
    body: JSON.stringify({ mac: userMac, plano\_id: 1 })  
})  
.then(res \=\> res.json())  
.then(data \=\> {  
    mostrarEcraPagamento(data.qr\_code, data.pix\_copia\_cola, data.txid);  
});

### **Passo B2: Resgatar Voucher (Novo Fluxo de Dinheiro Vivo)**

Se o cliente digita o código (ex: "A7X-92P") e clica em Ativar:

const codigoPin \= document.getElementById('inputVoucher').value;

fetch('/api/resgatar-voucher', {  
    method: 'POST',  
    headers: { 'Content-Type': 'application/json' },  
    body: JSON.stringify({ mac: userMac, codigo: codigoPin })  
})  
.then(res \=\> res.json())  
.then(data \=\> {  
    if (data.sucesso) {  
        // Pula o pagamento e vai direto para a tela de sucesso  
        window.location.href \= \`/sucesso.html?tok=${gatewayToken}\`;  
    } else {  
        alert("Código inválido ou já utilizado. Tente novamente.");  
    }  
});

### **Passo C: Verificar Pagamento PIX (Polling)**

No ecrã de pagamento, a cada 5 segundos:

setInterval(() \=\> {  
    fetch(\`/api/verificar-pagamento?txid=${currentTxid}\`)  
    .then(res \=\> res.json())  
    .then(data \=\> {  
        if (data.status \=== 'pago') {  
            window.location.href \= \`/sucesso.html?tok=${gatewayToken}\`;  
        }  
    });  
}, 5000);

## **5\. Requisitos de UI/UX e Boas Práticas**

1. **Clareza Visual (PIX vs Dinheiro):** É fundamental que a interface separe visualmente a área de "Comprar online (PIX)" da área "Já tenho um código". Use cores diferentes (ex: verde para o PIX, azul ou cinza para o Voucher).  
2. **Prevenção de Cache:** Use metatags no HTML (Cache-Control: no-cache) para garantir que o celular do cliente não carregue um Ecrã de Sucesso antigo da semana passada.  
3. **Feedback de Teclado (Vouchers):** O campo de digitação do Voucher deve ter autocapitalize="characters" no HTML para forçar o teclado do celular a digitar letras maiúsculas, evitando erros do cliente.
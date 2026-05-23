# Documentacao do Roteador

## Papel do Roteador

O roteador OpenWrt com OpenNDS controla quem pode navegar. O Astrolink nao
substitui o OpenNDS; ele atua como sistema de negocio e controle:

- recebe o cliente pelo portal cativo
- valida voucher ou inicia PIX
- registra sessao no banco
- autoriza ou desconecta o MAC no OpenNDS

## Desenvolvimento Local

Por padrao, a integracao real fica desligada:

```env
OPENNDS_ENABLED=false
```

Assim o backend responde normalmente sem depender de roteador fisico.

## Roteador Real

Para testar com OpenNDS:

```env
OPENNDS_ENABLED=true
OPENNDS_SSH_HOST=192.168.1.1
OPENNDS_SSH_PORT=22
OPENNDS_SSH_USER=root
OPENNDS_SSH_KEY_PATH=C:\Users\charl\.ssh\id_ed25519
OPENNDS_SSH_TIMEOUT=10s
OPENNDS_AUTH_RETRIES=3
```

## Comandos Usados

Liberar cliente:

```sh
ndsctl auth AA:BB:CC:DD:EE:FF 86400
```

Desconectar cliente:

```sh
ndsctl deauth AA:BB:CC:DD:EE:FF
```

Checar status manualmente:

```sh
ndsctl status
ndsctl clients
```

## Fluxo Esperado

1. Cliente conecta ao Wi-Fi.
2. OpenNDS redireciona para o portal Astrolink.
3. Portal recebe `mac`, `ip` e `token`.
4. Cliente resgata voucher ou inicia PIX.
5. Backend cria/atualiza sessao.
6. Backend chama `ndsctl auth`.
7. Quando expirar ou houver acao manual, backend chama `ndsctl deauth`.

## Pendencias

- Script ou endpoint para configurar OpenNDS automaticamente.
- Parser de `ndsctl clients`.
- Health check real do roteador.
- Traffic shaping por plano.
- Job de expiracao de sessoes.

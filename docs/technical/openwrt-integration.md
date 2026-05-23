# Integracao OpenWrt/OpenNDS

## Estado Atual

O backend Go controla o OpenNDS via SSH. A camada fica em:

```text
node/internal/gateway/
  gateway.go
  opennds.go
  ssh.go
```

Em desenvolvimento, `OPENNDS_ENABLED=false` usa um controller no-op. Em roteador
real, `OPENNDS_ENABLED=true` usa SSH para executar `ndsctl`.

## Variaveis de Ambiente

```env
OPENNDS_ENABLED=false
OPENNDS_SSH_HOST=192.168.1.1
OPENNDS_SSH_PORT=22
OPENNDS_SSH_USER=root
OPENNDS_SSH_KEY_PATH=C:\Users\charl\.ssh\id_ed25519
OPENNDS_SSH_TIMEOUT=10s
OPENNDS_AUTH_RETRIES=3
```

## Comandos Executados

Autorizar:

```sh
ndsctl auth AA:BB:CC:DD:EE:FF 86400
```

Desautorizar:

```sh
ndsctl deauth AA:BB:CC:DD:EE:FF
```

O valor de duracao vem do plano resgatado pelo voucher ou escolhido no PIX.

## Fluxo Voucher + OpenNDS

1. Portal chama `POST /api/voucher/resgatar`.
2. Backend valida o voucher.
3. Backend cria/atualiza `usuarios_mac`.
4. Backend chama `gateway.Authorize`.
5. `OpenNDSController` monta `ndsctl auth`.
6. `SSHRunner` executa o comando no roteador.
7. API retorna `roteador_autorizado=true` ou `false`.

Falha no roteador nao desfaz o resgate do voucher nesta fase; a resposta indica
`roteador_autorizado=false` para permitir tratamento posterior.

## Fluxo Admin Disconnect

1. Admin chama `POST /admin/usuarios/:mac/desconectar`.
2. Backend chama `gateway.Deauthorize`.
3. `OpenNDSController` monta `ndsctl deauth`.
4. Se SSH falhar, API responde `502 roteador_indisponivel`.

## Configuracao Manual do OpenNDS

Exemplo de redirect para o portal em dev:

```text
http://127.0.0.1:5173/?mac=%h&ip=%i&token=%t
```

Em uma rede real, troque `127.0.0.1` pelo IP do servidor Astrolink acessivel
pelos clientes.

O backend deve estar acessivel em:

```text
http://<ip-do-servidor>:5000
```

## Testes

A camada OpenNDS tem testes usando runner falso, sem exigir roteador real:

```powershell
cd node
go test ./internal/gateway
```

## Pendencias

- Endpoint/FAS especifico para validacao chamada pelo OpenNDS.
- Aplicacao automatica de configuracao UCI no roteador.
- Parser de `ndsctl clients`.
- Traffic shaping por plano.
- Health check real do roteador.
- Job de expiracao chamando `ndsctl deauth`.

# Schema do Banco Local

O banco local usa PostgreSQL. A migration atual esta em:

```text
node/migrations/000001_initial_schema.up.sql
```

Admin cloud e multi-tenancy estao pausados nesta fase; este documento cobre
apenas o no local.

## Diagrama Resumido

```text
planos
  |-- usuarios_mac
  |-- transacoes_pix
  |-- vouchers -- voucher_usos

roteadores -- usuarios_mac

system_settings
blacklist_mac
walled_garden
logs
sessoes_admin
```

## `planos`

Pacotes exibidos no portal e usados por vouchers/PIX.

Campos principais:

- `id`
- `nome`
- `descricao`
- `preco`
- `duracao_minutos`
- `dados_mb`
- `velocidade_down`
- `velocidade_up`
- `recomendado`
- `ativo`
- `visivel_portal`
- `ordem`

## `usuarios_mac`

Estado de acesso de cada dispositivo.

Campos principais:

- `mac`
- `ip_atual`
- `nome`
- `status`: `ativo`, `expirado`, `bloqueado`, `walled_garden`
- `plano_id`
- `inicio_acesso`
- `fim_acesso`
- `dados_consumidos_mb`
- `dados_limite_mb`
- `roteador_id`

## `transacoes_pix`

Cobrancas PIX. Nesta fase o PIX e demonstrativo.

Campos principais:

- `txid`
- `mac`
- `plano_id`
- `valor`
- `status`: `pendente`, `aprovado`, `cancelado`, `expirado`
- `pix_copia_cola`
- `qr_code_base64`
- `mp_payment_id`
- `webhook_at`

## `voucher_lotes`

Agrupa vouchers criados em massa.

Campos principais:

- `quantidade`
- `plano_id`
- `criado_por`

## `vouchers`

Codigos resgataveis no portal.

Campos principais:

- `codigo`
- `plano_id`
- `tipo`: `single_use` ou `universal`
- `usos_maximos`
- `usos_atuais`
- `validade_em`
- `ativo`
- `prefixo`
- `lote_id`

## `voucher_usos`

Historico de resgate.

Campos principais:

- `voucher_id`
- `mac`
- `ip`
- `tempo_adicionado_min`

## `roteadores`

Inventario local de roteadores OpenWrt/OpenNDS.

Campos principais:

- `nome`
- `ip`
- `porta_ssh`
- `usuario_ssh`
- `chave_ssh_path`
- `status`
- `ultimo_ping_ms`
- `versao_openwrt`
- `versao_opennds`
- `ativo`

## `blacklist_mac`

MACs bloqueados manualmente ou por regra futura.

## `walled_garden`

Hosts ou redes liberadas antes da autenticacao.

## `system_settings`

Configuracoes key/value usadas pelo portal e integracoes.

Seeds atuais:

- `hotspot_nome`
- `hotspot_logo_url`
- `cor_primaria`
- `cor_fundo`
- `url_pos_conexao`
- `coleta_nome`
- `mp_access_token`
- `mp_public_key`

## `logs`

Tabela para auditoria. Ainda precisa ser ligada aos fluxos principais.

## `sessoes_admin`

Base para sessoes de admin. A autenticacao atual ainda e simples e sera
substituida por JWT/refresh real.

## Pendencias de Schema

- Adicionar migrations incrementais.
- Completar CRUD admin para planos/vouchers.
- Ligar Mercado Pago real.
- Persistir eventos OpenNDS.
- Implementar backup/restore.
- Separar tabelas cloud em docs proprias quando o cloud voltar para o escopo.

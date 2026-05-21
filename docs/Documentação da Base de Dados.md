# Documentacao da Base de Dados

## Motor

O banco local do Astrolink e PostgreSQL. Em desenvolvimento ele roda pelo
`docker-compose.dev.yml`; em modo simples o backend tambem consegue operar com
store em memoria quando `DATABASE_URL` nao esta configurado.

O schema aplicado no banco esta em:

`node/migrations/000001_initial_schema.up.sql`

## Tabelas Principais

| Tabela | Funcao |
|---|---|
| `planos` | Planos vendidos no portal |
| `usuarios_mac` | Estado de acesso por MAC |
| `transacoes_pix` | Cobrancas PIX demonstrativas |
| `voucher_lotes` | Lotes de vouchers gerados |
| `vouchers` | Codigos de acesso |
| `voucher_usos` | Historico de resgate de vouchers |
| `roteadores` | Dados dos roteadores OpenWrt/OpenNDS |
| `blacklist_mac` | MACs bloqueados |
| `walled_garden` | Hosts liberados antes da autenticacao |
| `system_settings` | Configuracoes de white-label e integracoes |
| `logs` | Auditoria e eventos |
| `sessoes_admin` | Sessoes de admin local |

## Relacoes Essenciais

```text
planos -> vouchers
planos -> usuarios_mac
planos -> transacoes_pix
vouchers -> voucher_usos
roteadores -> usuarios_mac
```

## Seeds Atuais

A migration inicial cria dois planos:

- `Acesso 24 Horas`
- `Acesso 1 Hora`

Tambem cria settings padrao como `hotspot_nome`, `cor_primaria`,
`url_pos_conexao`, `mp_access_token` e `mp_public_key`.

Em modo memoria, o backend adiciona tambem:

- plano `Pacote Semanal`
- voucher `TEST-1234`
- voucher universal `UNIV-0000`

## Backup

Backups automatizados ainda nao foram implementados. Enquanto isso, use dump do
Postgres:

```powershell
docker compose -f docker-compose.dev.yml exec postgres pg_dump -U astrolink astrolink > backup.sql
```

## Pendencias

- Migrations incrementais para as proximas fases.
- Job de expiracao de sessoes.
- Backup e restore pelo admin local.
- Auditoria consistente em `logs`.

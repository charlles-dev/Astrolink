# Multi-Tenancy no Cloud Panel

## Modelo de Isolamento

O Cloud Panel usa **Row Level Security (RLS)** do PostgreSQL para isolamento de dados entre tenants. Cada operador (tenant) vê apenas seus próprios dados — isso é garantido em nível de banco de dados, não apenas na aplicação.

---

## JWT e Tenant ID

Ao fazer login via Supabase Auth, o JWT gerado contém o `tenant_id`:

```sql
-- Função chamada ao criar conta ou fazer login
CREATE OR REPLACE FUNCTION auth.jwt_tenant_id()
RETURNS UUID AS $$
  SELECT tenant_id FROM tenant_members
  WHERE user_id = auth.uid()
  LIMIT 1;
$$ LANGUAGE SQL SECURITY DEFINER;
```

```typescript
// Ao fazer login, buscar tenant do usuário
const { data: member } = await supabase
  .from('tenant_members')
  .select('tenant_id, role')
  .eq('user_id', user.id)
  .single()

// Setar no contexto da sessão para o RLS usar
```

---

## Row Level Security (RLS)

### Política base para nodes

```sql
-- Habilitar RLS em todas as tabelas
ALTER TABLE nodes ENABLE ROW LEVEL SECURITY;
ALTER TABLE node_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE node_metrics ENABLE ROW LEVEL SECURITY;
ALTER TABLE node_commands ENABLE ROW LEVEL SECURITY;
ALTER TABLE subscriptions ENABLE ROW LEVEL SECURITY;

-- Política: usuário vê apenas nós do seu tenant
CREATE POLICY "nodes_tenant_isolation" ON nodes
  FOR ALL
  USING (
    tenant_id IN (
      SELECT tenant_id FROM tenant_members
      WHERE user_id = auth.uid()
      AND accepted_at IS NOT NULL  -- apenas convites aceitos
    )
  );

-- Política de inserção: verificar limite de nós do plano
CREATE POLICY "nodes_insert_limit" ON nodes
  FOR INSERT
  WITH CHECK (
    -- Verificar se tenant não excedeu o limite de nós
    (
      SELECT COUNT(*) FROM nodes n
      WHERE n.tenant_id = tenant_id
    ) < (
      SELECT max_nos FROM tenants t
      WHERE t.id = tenant_id
    )
  );
```

### Política para node_events (otimizada para leitura)

```sql
-- Usando tenant_id direto na tabela (desnormalizado intencionalmente)
-- Evita JOIN no WHERE que seria necessário se só tivesse node_id
CREATE POLICY "events_tenant_isolation" ON node_events
  FOR SELECT
  USING (
    tenant_id IN (
      SELECT tenant_id FROM tenant_members
      WHERE user_id = auth.uid()
    )
  );

-- Inserção: apenas o serviço de sync (via service_role key)
CREATE POLICY "events_service_insert" ON node_events
  FOR INSERT
  WITH CHECK (auth.role() = 'service_role');
```

### Política para métricas públicas do mapa

```sql
-- Mapa público: qualquer um pode ler nós marcados como públicos
CREATE POLICY "nodes_public_map" ON nodes
  FOR SELECT
  USING (
    public_listing = TRUE
    AND status = 'online'
    -- Não expor campos sensíveis: RLS sozinha não resolve,
    -- usar view com campos específicos
  );

-- View pública (sem campos sensíveis)
CREATE VIEW public_nodes AS
  SELECT id, public_nome, public_descricao, public_foto_url,
         public_site, public_horario, lat, lng, cidade, estado,
         status, updated_at
  FROM nodes
  WHERE public_listing = TRUE AND status = 'online';
```

---

## Roles por Tenant

```typescript
// Types
type TenantRole = 'owner' | 'admin' | 'viewer'

// Permissões por role
const PERMISSIONS: Record<TenantRole, string[]> = {
  owner: ['*'],  // tudo
  admin: [
    'nodes:read', 'nodes:write', 'nodes:command',
    'events:read', 'metrics:read',
    'team:invite', 'team:remove',
    'billing:read',  // apenas leitura do billing
  ],
  viewer: [
    'nodes:read',
    'events:read', 'metrics:read',
  ],
}

// Middleware de verificação de permissão (Edge Function)
export function requirePermission(permission: string) {
  return async (req: Request) => {
    const { user } = await supabase.auth.getUser()
    const { data: member } = await supabase
      .from('tenant_members')
      .select('role')
      .eq('user_id', user.id)
      .single()

    const perms = PERMISSIONS[member.role]
    if (!perms.includes('*') && !perms.includes(permission)) {
      return new Response('Forbidden', { status: 403 })
    }
  }
}
```

---

## Onboarding Multi-Tenant

### Criação de Tenant

```typescript
// Edge Function: create-tenant
// Chamada após verificação de email

export default async function handler(req: Request) {
  const { nome, slug } = await req.json()
  const { user } = await supabase.auth.getUser()

  // Verificar se slug está disponível
  const { count } = await supabase
    .from('tenants')
    .select('*', { count: 'exact' })
    .eq('slug', slug)
  if (count > 0) return Response.json({ error: 'slug_unavailable' }, { status: 409 })

  // Criar tenant
  const { data: tenant } = await supabase
    .from('tenants')
    .insert({ nome, slug, plano_billing: 'free', max_nos: 1 })
    .select()
    .single()

  // Adicionar criador como owner
  await supabase.from('tenant_members').insert({
    tenant_id: tenant.id,
    user_id: user.id,
    role: 'owner',
    accepted_at: new Date().toISOString(),
  })

  // Criar subscription free
  await supabase.from('subscriptions').insert({
    tenant_id: tenant.id,
    plano: 'free',
    status: 'active',
    periodo_inicio: new Date().toISOString(),
    periodo_fim: new Date(Date.now() + 100 * 365 * 24 * 60 * 60 * 1000).toISOString(),  // free = sem fim
  })

  return Response.json({ tenant })
}
```

---

## Convites de Equipe

```typescript
// Edge Function: invite-member

export default async function handler(req: Request) {
  const { email, role, tenant_id } = await req.json()
  const { user } = await supabase.auth.getUser()

  // Verificar se quem convida tem permissão (owner ou admin)
  const { data: inviter } = await supabase
    .from('tenant_members')
    .select('role')
    .eq('user_id', user.id)
    .eq('tenant_id', tenant_id)
    .single()

  if (!['owner', 'admin'].includes(inviter.role)) {
    return Response.json({ error: 'insufficient_permissions' }, { status: 403 })
  }

  // Admin não pode convidar outro owner
  if (inviter.role === 'admin' && role === 'owner') {
    return Response.json({ error: 'cannot_assign_owner' }, { status: 403 })
  }

  // Gerar token de convite (expira em 7 dias)
  const token = crypto.randomUUID()
  const expiresAt = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000)

  await supabase.from('invitations').insert({
    tenant_id, email, role, token,
    invited_by: user.id,
    expires_at: expiresAt.toISOString(),
  })

  // Enviar email com link de convite
  await sendInviteEmail({
    to: email,
    inviterName: user.user_metadata.name,
    tenantName: tenant.nome,
    role,
    acceptUrl: `https://app.astrolink.app/convite/${token}`,
  })

  return Response.json({ success: true })
}
```

---

## Limites e Quotas

```typescript
// Verificações de limite executadas antes de operações críticas

// 1. Limite de nós por plano
async function canAddNode(tenantId: string): Promise<boolean> {
  const { data: tenant } = await supabase
    .from('tenants')
    .select('max_nos')
    .eq('id', tenantId)
    .single()

  const { count } = await supabase
    .from('nodes')
    .select('*', { count: 'exact' })
    .eq('tenant_id', tenantId)

  return count < tenant.max_nos
}

// 2. Limite de membros da equipe
const TEAM_LIMITS: Record<string, number> = {
  free: 1,      // apenas o owner
  pro: 3,       // owner + 2
  business: -1, // ilimitado
}

// 3. Retenção de dados por plano
const DATA_RETENTION_DAYS: Record<string, number> = {
  free: 30,       // 30 dias de histórico
  pro: 365,       // 12 meses
  business: -1,   // ilimitado
}
```

---

## Isolamento de Storage (Logos e Fotos)

```typescript
// Supabase Storage com RLS por pasta tenant
// Estrutura: storage/public/tenants/{tenant_id}/{arquivo}

// Policy Storage (configurada via Dashboard ou SQL)
// Apenas o tenant pode escrever na sua pasta
// Leitura pública (logos são públicas)

const uploadLogo = async (tenantId: string, file: File) => {
  const ext = file.name.split('.').pop()
  const path = `tenants/${tenantId}/logo.${ext}`

  const { data, error } = await supabase.storage
    .from('public')
    .upload(path, file, { upsert: true })

  const { data: { publicUrl } } = supabase.storage
    .from('public')
    .getPublicUrl(path)

  return publicUrl
}
```

---

## Auditoria Multi-Tenant

```sql
-- Tabela de auditoria (Cloud)
CREATE TABLE audit_log (
  id          BIGSERIAL PRIMARY KEY,
  tenant_id   UUID NOT NULL,
  user_id     UUID REFERENCES auth.users(id),
  acao        VARCHAR(100) NOT NULL,
  -- ex: node.added, member.invited, node.command.mac_ban
  recurso_tipo VARCHAR(50),
  recurso_id   TEXT,
  payload     JSONB,
  ip          INET,
  created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- RLS: tenant vê apenas seu audit log
ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;
CREATE POLICY "audit_tenant_isolation" ON audit_log
  FOR SELECT USING (
    tenant_id IN (
      SELECT tenant_id FROM tenant_members WHERE user_id = auth.uid()
    )
  );
```

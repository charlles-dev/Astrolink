export interface Settings {
  hotspot_nome: string
  hotspot_logo_url: string
  cor_primaria: string
  cor_secundaria?: string
  cor_fundo: string
  mensagem_boas_vindas: string
  url_pos_conexao: string
  coleta_nome: boolean
  mostrar_velocidade?: boolean
}

export interface Plano {
  id: number
  nome: string
  descricao?: string
  preco: string
  duracao_minutos: number | null
  duracao_formatada: string
  dados_mb: number | null
  velocidade_down: number
  velocidade_up: number
  recomendado: boolean
  ativo: boolean
  visivel_portal: boolean
  ordem: number
}

export interface PlanosResponse {
  planos: Plano[]
}

export interface AdminPlanBody {
  nome: string
  descricao: string
  preco: number
  duracao_minutos: number | null
  dados_mb: number | null
  velocidade_down: number
  velocidade_up: number
  recomendado: boolean
  ativo: boolean
  visivel_portal: boolean
  ordem: number
}

export interface AdminPlanResponse {
  plano: Plano
}

export interface SessaoStatus {
  ativa: boolean
  plano?: string
  fim_acesso?: string
  tempo_restante_segundos?: number
  dados_consumidos_mb?: number
}

export interface PixTransaction {
  txid: string
  valor: string
  descricao: string
  pix_copia_cola: string
  qr_code_base64: string
  expira_em: string
  expira_em_segundos: number
  status?: string
}

export interface PixStatusResponse {
  txid: string
  status: string
  expira_em: string
}

export interface GerarPixBody {
  plano_id: number
  mac: string
  ip: string
  nome?: string
}

export interface ResgatarVoucherBody {
  codigo: string
  mac: string
  ip: string
}

export interface ResgatarVoucherResponse {
  sucesso: boolean
  plano: string
  tempo_adicionado_minutos: number
  fim_acesso: string
  tempo_restante_segundos: number
  acesso_anterior: boolean
  roteador_autorizado?: boolean
}

export interface DeviceInfo {
  mac: string
  ip: string
  token: string
}

export interface AdminLoginBody {
  usuario: string
  senha: string
  totp_codigo?: string
}

export interface AdminLoginResponse {
  access_token: string
  refresh_token: string
  expires_in: number
  token_type: string
}

export interface AdminHealthCheck {
  status: string
  latencia_ms?: number
}

export interface AdminRouterHealth {
  total: number
  online: number
  offline: number
}

export interface AdminHealthResponse {
  status: string
  versao: string
  uptime_segundos: number
  checks: {
    banco_dados: AdminHealthCheck
    redis: AdminHealthCheck
    rabbitmq: AdminHealthCheck
    mercadopago: AdminHealthCheck
    roteadores: AdminRouterHealth
  }
}

export interface AdminUser {
  id: number
  mac: string
  ip_atual?: string
  nome?: string
  status: string
  plano?: {
    id: number
    nome: string
  }
  inicio_acesso?: string
  fim_acesso?: string
  tempo_restante_segundos: number
  dados_consumidos_mb: number
  roteador?: {
    id: number
    nome: string
  }
}

export interface AdminUserSession {
  inicio_acesso?: string
  fim_acesso?: string
  plano?: {
    id: number
    nome: string
  }
  status: string
  tempo_restante_segundos: number
  dados_consumidos_mb: number
  valor?: string
  origem: string
}

export interface AdminUserDetail {
  usuario: AdminUser
  sessao_atual?: AdminUserSession
  historico_sessoes: AdminUserSession[]
  total_sessoes: number
  total_gasto: string
  ultima_visita?: string
}

export interface AdminUsersResponse {
  total: number
  page: number
  limit: number
  usuarios: AdminUser[]
}

export interface AdminExtendUserBody {
  minutos: number
}

export interface AdminBanUserBody {
  motivo: string
}

export interface AdminUserResponse {
  usuario: AdminUser
}

export interface AdminRouter {
  id: number
  nome: string
  ip: string
  porta_ssh: number
  usuario_ssh: string
  chave_ssh_path?: string
  status: string
  ultimo_ping_ms?: number
  ultimo_check_at?: string
  versao_openwrt?: string
  versao_opennds?: string
  ativo: boolean
  usuarios_ativos: number
  created_at?: string
  updated_at?: string
}

export interface AdminRouterBody {
  nome: string
  ip: string
  porta_ssh: number
  usuario_ssh: string
  chave_ssh_path: string
  ativo: boolean
}

export interface AdminRoutersResponse {
  roteadores: AdminRouter[]
}

export interface AdminRouterResponse {
  roteador: AdminRouter
}

export interface AdminRouterDiagnosticResponse {
  status: string
  roteador: Partial<AdminRouter>
  diagnostico: unknown
  erro?: string
}

export interface AdminSpeedtestResponse {
  roteador_id: number
  download_mbps: number
  upload_mbps: number
  status: string
  mensagem: string
  medido_em: string
  roteador_status: string
}

export interface AdminBlacklistEntry {
  id: number
  mac: string
  motivo?: string
  criado_por: string
  created_at?: string
}

export interface AdminBlacklistResponse {
  blacklist: AdminBlacklistEntry[]
  total: number
}

export interface AdminBlacklistBody {
  mac: string
  motivo: string
}

export interface AdminBlacklistEntryResponse {
  entrada: AdminBlacklistEntry
}

export interface AdminWalledGardenEntry {
  id: number
  host: string
  descricao?: string
  tipo: string
  sistema: boolean
  created_at?: string
}

export interface AdminWalledGardenResponse {
  walled_garden: AdminWalledGardenEntry[]
  total: number
}

export interface AdminWalledGardenBody {
  host: string
  descricao: string
  tipo: string
}

export interface AdminWalledGardenEntryResponse {
  entrada: AdminWalledGardenEntry
}

export interface AdminVoucher {
  id: number
  codigo: string
  plano: {
    id: number
    nome: string
  }
  tipo: string
  usos_maximos?: number | null
  usos_atuais: number
  validade_em?: string | null
  ativo: boolean
  prefixo?: string
  lote_id?: number
  created_at?: string
}

export type AdminVoucherStatusFilter = 'ativo' | 'inativo' | 'todos'

export interface AdminVoucherFilters {
  status?: AdminVoucherStatusFilter
  plano_id?: number
  codigo?: string
  lote_id?: number
}

export interface AdminVouchersResponse {
  total: number
  vouchers: AdminVoucher[]
}

export interface AdminVoucherResponse {
  voucher: AdminVoucher
}

export interface GenerateAdminVouchersBody {
  plano_id: number
  quantidade: number
  tipo?: string
  usos_maximos?: number | null
  validade_dias?: number | null
  prefixo?: string
}

export interface GenerateAdminVouchersResponse {
  lote_id: number
  quantidade: number
  vouchers: AdminVoucher[]
}

export type AdminPaymentStatusFilter = 'pendente' | 'aprovado' | 'cancelado' | 'expirado' | ''

export interface AdminPaymentFilters {
  status?: AdminPaymentStatusFilter
  inicio?: string
  fim?: string
}

export interface AdminPaymentTotals {
  pendente: number
  aprovado: number
  cancelado: number
  expirado: number
  valor_total: string
}

export interface AdminPayment {
  txid: string
  status: string
  valor: string
  descricao: string
  mac: string
  plano_id: number
  plano?: {
    id: number
    nome: string
  }
  created_at: string
  expira_em: string
}

export interface AdminPaymentsResponse {
  total: number
  totais: AdminPaymentTotals
  pagamentos: AdminPayment[]
}

export interface AdminPaymentReportResponse {
  periodo: {
    de: string
    ate: string
  }
  total: number
  totais: AdminPaymentTotals
  pagamentos: AdminPayment[]
}

export interface AdminLogFilters {
  nivel?: string
  tipo?: string
  texto?: string
}

export interface AdminLog {
  timestamp: string
  nivel: string
  tipo: string
  mensagem: string
  detalhes?: unknown
}

export interface AdminLogsResponse {
  total: number
  logs: AdminLog[]
}

export interface AdminBackupResponse {
  arquivo?: string
  tamanho_bytes?: number
  mensagem?: string
  created_at?: string
}

export interface AdminRestoreBackupBody {
  arquivo: string
  confirmacao: string
}

export interface SetupField {
  key: string
  label: string
  description: string
  secret: boolean
  configured: boolean
  value?: string
}

export interface SetupGroup {
  label: string
  fields: SetupField[]
}

export interface SetupStatus {
  requires_restart: boolean
  writable: boolean
  groups: Record<string, SetupGroup>
}

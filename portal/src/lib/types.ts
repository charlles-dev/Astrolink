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

export interface AdminUsersResponse {
  total: number
  page: number
  limit: number
  usuarios: AdminUser[]
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

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

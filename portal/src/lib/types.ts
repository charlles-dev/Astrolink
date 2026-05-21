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
}

export interface DeviceInfo {
  mac: string
  ip: string
  token: string
}

import type {
  GerarPixBody,
  AdminHealthResponse,
  AdminPlanBody,
  AdminPlanResponse,
  AdminLoginBody,
  AdminLoginResponse,
  AdminVouchersResponse,
  AdminUsersResponse,
  GenerateAdminVouchersBody,
  GenerateAdminVouchersResponse,
  PixStatusResponse,
  PixTransaction,
  PlanosResponse,
  ResgatarVoucherBody,
  ResgatarVoucherResponse,
  SessaoStatus,
  Settings
} from './types'

interface ErrorPayload {
  erro?: string
  mensagem?: string
}

export class APIError extends Error {
  status: number
  code: string

  constructor(status: number, code: string, message: string) {
    super(message)
    this.name = 'APIError'
    this.status = status
    this.code = code
  }
}

export function createApiClient(baseURL = '') {
  async function request<T>(
    method: string,
    path: string,
    body?: unknown,
    token?: string
  ): Promise<T> {
    const headers: Record<string, string> = {}
    if (body) headers['Content-Type'] = 'application/json'
    if (token) headers.Authorization = `Bearer ${token}`

    const response = await fetch(`${baseURL}${path}`, {
      method,
      headers: Object.keys(headers).length ? headers : undefined,
      body: body ? JSON.stringify(body) : undefined
    })

    const contentType = response.headers.get('content-type') ?? ''
    const data = contentType.includes('application/json') ? await response.json() : null

    if (!response.ok) {
      const payload = (data ?? {}) as ErrorPayload
      throw new APIError(
        response.status,
        payload.erro ?? 'erro_interno',
        payload.mensagem ?? 'Nao foi possivel completar a requisicao'
      )
    }

    return data as T
  }

  return {
    getSettings: () => request<Settings>('GET', '/api/settings'),
    getPlanos: () => request<PlanosResponse>('GET', '/api/planos'),
    getSessaoStatus: (mac: string) =>
      request<SessaoStatus>('GET', `/api/sessao/status?mac=${encodeURIComponent(mac)}`),
    gerarPix: (body: GerarPixBody) => request<PixTransaction>('POST', '/api/pix/gerar', body),
    getPixStatus: (txid: string) =>
      request<PixStatusResponse>('GET', `/api/pix/status/${encodeURIComponent(txid)}`),
    resgatarVoucher: (body: ResgatarVoucherBody) =>
      request<ResgatarVoucherResponse>('POST', '/api/voucher/resgatar', body),
    loginAdmin: (body: AdminLoginBody) =>
      request<AdminLoginResponse>('POST', '/admin/auth/login', body),
    getAdminHealth: (token: string) =>
      request<AdminHealthResponse>('GET', '/admin/sistema/saude', undefined, token),
    getAdminPlanos: (token: string) =>
      request<PlanosResponse>('GET', '/admin/planos', undefined, token),
    createAdminPlano: (token: string, body: AdminPlanBody) =>
      request<AdminPlanResponse>('POST', '/admin/planos', body, token),
    updateAdminPlano: (token: string, id: number, body: AdminPlanBody) =>
      request<AdminPlanResponse>('PUT', `/admin/planos/${id}`, body, token),
    updateAdminPlanoStatus: (token: string, id: number, ativo: boolean) =>
      request<AdminPlanResponse>('PATCH', `/admin/planos/${id}/status`, { ativo }, token),
    getAdminUsuarios: (token: string) =>
      request<AdminUsersResponse>('GET', '/admin/usuarios', undefined, token),
    disconnectAdminUsuario: (token: string, mac: string) =>
      request<{ sucesso: boolean }>(
        'POST',
        `/admin/usuarios/${encodeURIComponent(mac)}/desconectar`,
        {},
        token
      ),
    getAdminVouchers: (token: string) =>
      request<AdminVouchersResponse>('GET', '/admin/vouchers', undefined, token),
    generateAdminVouchers: (token: string, body: GenerateAdminVouchersBody) =>
      request<GenerateAdminVouchersResponse>('POST', '/admin/vouchers/gerar', body, token)
  }
}

export const api = createApiClient()

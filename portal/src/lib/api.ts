import type {
  GerarPixBody,
  AdminHealthResponse,
  AdminBackupResponse,
  AdminLogFilters,
  AdminLogsResponse,
  AdminPaymentFilters,
  AdminPaymentsResponse,
  AdminPlanBody,
  AdminPlanResponse,
  AdminRestoreBackupBody,
  AdminLoginBody,
  AdminLoginResponse,
  AdminVoucherFilters,
  AdminVoucherResponse,
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
  SetupStatus,
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
  function buildQuery(filters?: object) {
    const params = new URLSearchParams()
    Object.entries(filters ?? {}).forEach(([key, value]: [string, unknown]) => {
      if (value === undefined || value === null || value === '') return
      const stringValue = String(value).trim()
      if (stringValue) params.set(key, stringValue)
    })
    const query = params.toString()
    return query ? `?${query}` : ''
  }

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

  async function requestBlob(path: string, token: string): Promise<Blob> {
    const response = await fetch(`${baseURL}${path}`, {
      method: 'GET',
      headers: { Authorization: `Bearer ${token}` }
    })

    if (!response.ok) {
      const contentType = response.headers.get('content-type') ?? ''
      const data = contentType.includes('application/json') ? await response.json() : null
      const payload = (data ?? {}) as ErrorPayload
      throw new APIError(
        response.status,
        payload.erro ?? 'erro_interno',
        payload.mensagem ?? 'Nao foi possivel completar a requisicao'
      )
    }

    return response.blob()
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
    getSetupStatus: (token: string) =>
      request<SetupStatus>('GET', '/admin/setup/status', undefined, token),
    updateSetupEnv: (values: Record<string, string>, token: string) =>
      request<SetupStatus>('PUT', '/admin/setup/env', { values }, token),
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
    getAdminVouchers: (token: string, filters?: AdminVoucherFilters) =>
      request<AdminVouchersResponse>('GET', `/admin/vouchers${buildQuery(filters)}`, undefined, token),
    deactivateAdminVoucher: (token: string, id: number) =>
      request<AdminVoucherResponse>('PATCH', `/admin/vouchers/${id}/desativar`, undefined, token),
    exportAdminVouchers: (token: string, filters?: AdminVoucherFilters) =>
      requestBlob(`/admin/vouchers/export.csv${buildQuery(filters)}`, token),
    generateAdminVouchers: (token: string, body: GenerateAdminVouchersBody) =>
      request<GenerateAdminVouchersResponse>('POST', '/admin/vouchers/gerar', body, token),
    getAdminPagamentos: (token: string, filters?: AdminPaymentFilters) =>
      request<AdminPaymentsResponse>(
        'GET',
        `/admin/pagamentos${buildQuery(filters)}`,
        undefined,
        token
      ),
    exportAdminPagamentos: (token: string, filters?: AdminPaymentFilters) =>
      requestBlob(`/admin/pagamentos/export.csv${buildQuery(filters)}`, token),
    getAdminLogs: (token: string, filters?: AdminLogFilters) =>
      request<AdminLogsResponse>('GET', `/admin/logs${buildQuery(filters)}`, undefined, token),
    exportAdminLogs: (token: string, filters?: AdminLogFilters) =>
      requestBlob(`/admin/logs/export.csv${buildQuery(filters)}`, token),
    createAdminBackup: (token: string) =>
      request<AdminBackupResponse>('POST', '/admin/backup', {}, token),
    restoreAdminBackup: (token: string, body: AdminRestoreBackupBody) =>
      request<AdminBackupResponse>('POST', '/admin/backup/restaurar', body, token)
  }
}

export const api = createApiClient()

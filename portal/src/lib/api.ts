import type {
  GerarPixBody,
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
  async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
    const response = await fetch(`${baseURL}${path}`, {
      method,
      headers: body ? { 'Content-Type': 'application/json' } : undefined,
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
      request<ResgatarVoucherResponse>('POST', '/api/voucher/resgatar', body)
  }
}

export const api = createApiClient()

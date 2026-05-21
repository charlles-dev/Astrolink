import { afterEach, describe, expect, it, vi } from 'vitest'
import { APIError, createApiClient } from './api'

afterEach(() => vi.restoreAllMocks())

describe('createApiClient', () => {
  it('loads plans from /api/planos', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        new Response(
          JSON.stringify({
            planos: [
              {
                id: 1,
                nome: 'Acesso 24 Horas',
                preco: '15.00',
                duracao_minutos: 1440,
                duracao_formatada: '24 horas',
                dados_mb: null,
                velocidade_down: 10,
                velocidade_up: 5,
                recomendado: true,
                ativo: true,
                visivel_portal: true,
                ordem: 1
              }
            ]
          }),
          { status: 200, headers: { 'content-type': 'application/json' } }
        )
      )
    )

    const api = createApiClient('')
    const result = await api.getPlanos()

    expect(result.planos[0].nome).toBe('Acesso 24 Horas')
    expect(fetch).toHaveBeenCalledWith('/api/planos', expect.objectContaining({ method: 'GET' }))
  })

  it('throws APIError with backend message', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        new Response(JSON.stringify({ erro: 'nao_encontrado', mensagem: 'voucher nao encontrado' }), {
          status: 404,
          headers: { 'content-type': 'application/json' }
        })
      )
    )

    const api = createApiClient('')

    await expect(
      api.resgatarVoucher({ codigo: 'XXXX-9999', mac: 'AA:BB:CC:DD:EE:FF', ip: '192.168.1.50' })
    ).rejects.toMatchObject({
      status: 404,
      code: 'nao_encontrado',
      message: 'voucher nao encontrado'
    } satisfies Partial<APIError>)
  })

  it('logs in to the local admin API', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        new Response(
          JSON.stringify({
            access_token: 'token-123',
            refresh_token: 'refresh-123',
            expires_in: 28800,
            token_type: 'Bearer'
          }),
          { status: 200, headers: { 'content-type': 'application/json' } }
        )
      )
    )

    const api = createApiClient('')
    const result = await api.loginAdmin({ usuario: 'admin', senha: 'admin123' })

    expect(result.access_token).toBe('token-123')
    expect(fetch).toHaveBeenCalledWith(
      '/admin/auth/login',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ usuario: 'admin', senha: 'admin123' })
      })
    )
  })

  it('sends bearer token to admin endpoints', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        new Response(JSON.stringify({ total: 0, page: 1, limit: 50, usuarios: [] }), {
          status: 200,
          headers: { 'content-type': 'application/json' }
        })
      )
    )

    const api = createApiClient('')
    await api.getAdminUsuarios('token-123')

    expect(fetch).toHaveBeenCalledWith(
      '/admin/usuarios',
      expect.objectContaining({
        headers: expect.objectContaining({ Authorization: 'Bearer token-123' })
      })
    )
  })
})

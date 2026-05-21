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

  it('loads admin vouchers with bearer token', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        new Response(JSON.stringify({ vouchers: [{ id: 1, codigo: 'VIPA-1234' }] }), {
          status: 200,
          headers: { 'content-type': 'application/json' }
        })
      )
    )

    const api = createApiClient('')
    const result = await api.getAdminVouchers('token-123')

    expect(result.vouchers[0].codigo).toBe('VIPA-1234')
    expect(fetch).toHaveBeenCalledWith(
      '/admin/vouchers',
      expect.objectContaining({
        headers: expect.objectContaining({ Authorization: 'Bearer token-123' })
      })
    )
  })

  it('generates admin vouchers', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        new Response(JSON.stringify({ lote_id: 1, quantidade: 2, vouchers: [] }), {
          status: 201,
          headers: { 'content-type': 'application/json' }
        })
      )
    )

    const api = createApiClient('')
    await api.generateAdminVouchers('token-123', { plano_id: 2, quantidade: 2, prefixo: 'VIPA' })

    expect(fetch).toHaveBeenCalledWith(
      '/admin/vouchers/gerar',
      expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({ Authorization: 'Bearer token-123' }),
        body: JSON.stringify({ plano_id: 2, quantidade: 2, prefixo: 'VIPA' })
      })
    )
  })

  it('creates an admin plan with bearer token', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        new Response(JSON.stringify({ plano: { id: 3, nome: 'Noite Livre' } }), {
          status: 201,
          headers: { 'content-type': 'application/json' }
        })
      )
    )

    const body = {
      nome: 'Noite Livre',
      descricao: 'Acesso noturno',
      preco: 9.9,
      duracao_minutos: 480,
      dados_mb: 2048,
      velocidade_down: 20,
      velocidade_up: 8,
      recomendado: true,
      ativo: true,
      visivel_portal: false,
      ordem: 3
    }

    const api = createApiClient('')
    const result = await api.createAdminPlano('token-123', body)

    expect(result.plano.nome).toBe('Noite Livre')
    expect(fetch).toHaveBeenCalledWith(
      '/admin/planos',
      expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({ Authorization: 'Bearer token-123' }),
        body: JSON.stringify(body)
      })
    )
  })

  it('updates an admin plan and toggles status with bearer token', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValueOnce(
          new Response(JSON.stringify({ plano: { id: 3, nome: 'Noite Livre' } }), {
            status: 200,
            headers: { 'content-type': 'application/json' }
          })
        )
        .mockResolvedValueOnce(
          new Response(JSON.stringify({ plano: { id: 3, ativo: false } }), {
            status: 200,
            headers: { 'content-type': 'application/json' }
          })
        )
    )

    const body = {
      nome: 'Noite Livre',
      descricao: '',
      preco: 10,
      duracao_minutos: 480,
      dados_mb: null,
      velocidade_down: 20,
      velocidade_up: 8,
      recomendado: false,
      ativo: true,
      visivel_portal: true,
      ordem: 3
    }

    const api = createApiClient('')
    await api.updateAdminPlano('token-123', 3, body)
    await api.updateAdminPlanoStatus('token-123', 3, false)

    expect(fetch).toHaveBeenNthCalledWith(
      1,
      '/admin/planos/3',
      expect.objectContaining({
        method: 'PUT',
        headers: expect.objectContaining({ Authorization: 'Bearer token-123' }),
        body: JSON.stringify(body)
      })
    )
    expect(fetch).toHaveBeenNthCalledWith(
      2,
      '/admin/planos/3/status',
      expect.objectContaining({
        method: 'PATCH',
        headers: expect.objectContaining({ Authorization: 'Bearer token-123' }),
        body: JSON.stringify({ ativo: false })
      })
    )
  })
})

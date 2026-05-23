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

  it('loads local setup status with bearer token', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        new Response(
          JSON.stringify({
            requires_restart: false,
            writable: true,
            groups: {
              mercadopago: {
                label: 'Mercado Pago',
                fields: [
                  {
                    key: 'MERCADOPAGO_ACCESS_TOKEN',
                    label: 'Access token',
                    description: 'Token privado',
                    secret: true,
                    configured: true
                  }
                ]
              }
            }
          }),
          { status: 200, headers: { 'content-type': 'application/json' } }
        )
      )
    )

    const api = createApiClient('')
    const result = await api.getSetupStatus('token-123')

    expect(result.groups.mercadopago.fields[0].configured).toBe(true)
    expect(fetch).toHaveBeenCalledWith(
      '/admin/setup/status',
      expect.objectContaining({
        method: 'GET',
        headers: expect.objectContaining({ Authorization: 'Bearer token-123' })
      })
    )
  })

  it('updates local setup env values with bearer token', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        new Response(JSON.stringify({ requires_restart: true, writable: true, groups: {} }), {
          status: 200,
          headers: { 'content-type': 'application/json' }
        })
      )
    )

    const api = createApiClient('')
    const result = await api.updateSetupEnv({ ADMIN_USUARIO: 'novo-admin' }, 'token-123')

    expect(result.requires_restart).toBe(true)
    expect(fetch).toHaveBeenCalledWith(
      '/admin/setup/env',
      expect.objectContaining({
        method: 'PUT',
        headers: expect.objectContaining({
          Authorization: 'Bearer token-123',
          'Content-Type': 'application/json'
        }),
        body: JSON.stringify({ values: { ADMIN_USUARIO: 'novo-admin' } })
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

  it('loads admin vouchers with filters', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        new Response(JSON.stringify({ total: 0, vouchers: [] }), {
          status: 200,
          headers: { 'content-type': 'application/json' }
        })
      )
    )

    const api = createApiClient('')
    await api.getAdminVouchers('token-123', {
      status: 'inativo',
      plano_id: 2,
      codigo: 'vip',
      lote_id: 9
    })

    expect(fetch).toHaveBeenCalledWith(
      '/admin/vouchers?status=inativo&plano_id=2&codigo=vip&lote_id=9',
      expect.objectContaining({
        method: 'GET',
        headers: expect.objectContaining({ Authorization: 'Bearer token-123' })
      })
    )
  })

  it('deactivates an admin voucher', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        new Response(JSON.stringify({ voucher: { id: 7, codigo: 'VIPA-7777', ativo: false } }), {
          status: 200,
          headers: { 'content-type': 'application/json' }
        })
      )
    )

    const api = createApiClient('')
    const result = await api.deactivateAdminVoucher('token-123', 7)

    expect(result.voucher.ativo).toBe(false)
    expect(fetch).toHaveBeenCalledWith(
      '/admin/vouchers/7/desativar',
      expect.objectContaining({
        method: 'PATCH',
        headers: expect.objectContaining({ Authorization: 'Bearer token-123' })
      })
    )
  })

  it('exports admin vouchers as CSV with filters', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        new Response('codigo,status\nVIPA-1234,ativo\n', {
          status: 200,
          headers: { 'content-type': 'text/csv' }
        })
      )
    )

    const api = createApiClient('')
    const result = await api.exportAdminVouchers('token-123', {
      status: 'ativo',
      plano_id: 2,
      codigo: 'VIPA'
    })

    expect(result).toBeInstanceOf(Blob)
    expect(fetch).toHaveBeenCalledWith(
      '/admin/vouchers/export.csv?status=ativo&plano_id=2&codigo=VIPA',
      expect.objectContaining({
        method: 'GET',
        headers: expect.objectContaining({ Authorization: 'Bearer token-123' })
      })
    )
  })

  it('loads admin payments with filters', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        new Response(JSON.stringify({ total: 0, totais: {}, pagamentos: [] }), {
          status: 200,
          headers: { 'content-type': 'application/json' }
        })
      )
    )

    const api = createApiClient('')
    await api.getAdminPagamentos('token-123', {
      status: 'aprovado',
      inicio: '2026-05-01',
      fim: '2026-05-21'
    })

    expect(fetch).toHaveBeenCalledWith(
      '/admin/pagamentos?status=aprovado&inicio=2026-05-01&fim=2026-05-21',
      expect.objectContaining({
        method: 'GET',
        headers: expect.objectContaining({ Authorization: 'Bearer token-123' })
      })
    )
  })

  it('exports admin payments as CSV with filters', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        new Response('txid,status\npix-1,aprovado\n', {
          status: 200,
          headers: { 'content-type': 'text/csv' }
        })
      )
    )

    const api = createApiClient('')
    const result = await api.exportAdminPagamentos('token-123', {
      status: 'pendente',
      inicio: '2026-05-01'
    })

    expect(result).toBeInstanceOf(Blob)
    expect(fetch).toHaveBeenCalledWith(
      '/admin/pagamentos/export.csv?status=pendente&inicio=2026-05-01',
      expect.objectContaining({
        method: 'GET',
        headers: expect.objectContaining({ Authorization: 'Bearer token-123' })
      })
    )
  })

  it('loads and exports admin payment reports', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValueOnce(
          new Response(
            JSON.stringify({
              periodo: { de: '2026-05-01T00:00:00Z', ate: '2026-06-01T00:00:00Z' },
              total: 0,
              totais: { pendente: 0, aprovado: 0, cancelado: 0, expirado: 0, valor_total: '0.00' },
              pagamentos: []
            }),
            { status: 200, headers: { 'content-type': 'application/json' } }
          )
        )
        .mockResolvedValueOnce(
          new Response('txid,status\n', {
            status: 200,
            headers: { 'content-type': 'text/csv' }
          })
        )
        .mockResolvedValueOnce(
          new Response('%PDF-1.4\n', {
            status: 200,
            headers: { 'content-type': 'application/pdf' }
          })
        )
    )

    const api = createApiClient('')
    const report = await api.getAdminPagamentosRelatorio('token-123', {
      inicio: '2026-05-01',
      fim: '2026-05-31'
    })
    const csv = await api.exportAdminPagamentosRelatorio('token-123', { status: 'aprovado' })
    const pdf = await api.exportAdminPagamentosRelatorioPDF('token-123', { status: 'aprovado' })

    expect(report.total).toBe(0)
    expect(csv).toBeInstanceOf(Blob)
    expect(pdf).toBeInstanceOf(Blob)
    expect(fetch).toHaveBeenNthCalledWith(
      1,
      '/admin/pagamentos/relatorio?inicio=2026-05-01&fim=2026-05-31',
      expect.objectContaining({
        method: 'GET',
        headers: expect.objectContaining({ Authorization: 'Bearer token-123' })
      })
    )
    expect(fetch).toHaveBeenNthCalledWith(
      2,
      '/admin/pagamentos/relatorio?status=aprovado&formato=csv',
      expect.objectContaining({
        method: 'GET',
        headers: expect.objectContaining({ Authorization: 'Bearer token-123' })
      })
    )
    expect(fetch).toHaveBeenNthCalledWith(
      3,
      '/admin/pagamentos/relatorio?status=aprovado&formato=pdf',
      expect.objectContaining({
        method: 'GET',
        headers: expect.objectContaining({ Authorization: 'Bearer token-123' })
      })
    )
  })

  it('loads admin logs and exports them as CSV', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValueOnce(
          new Response(JSON.stringify({ total: 0, logs: [] }), {
            status: 200,
            headers: { 'content-type': 'application/json' }
          })
        )
        .mockResolvedValueOnce(
          new Response('timestamp,nivel\n2026-05-21,info\n', {
            status: 200,
            headers: { 'content-type': 'text/csv' }
          })
        )
    )

    const api = createApiClient('')
    await api.getAdminLogs('token-123', { nivel: 'erro', tipo: 'backup', texto: 'falha' })
    const csv = await api.exportAdminLogs('token-123', { texto: 'pix' })

    expect(csv).toBeInstanceOf(Blob)
    expect(fetch).toHaveBeenNthCalledWith(
      1,
      '/admin/logs?nivel=erro&tipo=backup&texto=falha',
      expect.objectContaining({
        method: 'GET',
        headers: expect.objectContaining({ Authorization: 'Bearer token-123' })
      })
    )
    expect(fetch).toHaveBeenNthCalledWith(
      2,
      '/admin/logs/export.csv?texto=pix',
      expect.objectContaining({
        method: 'GET',
        headers: expect.objectContaining({ Authorization: 'Bearer token-123' })
      })
    )
  })

  it('requests an admin backup', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        new Response(JSON.stringify({ mensagem: 'Backup iniciado' }), {
          status: 202,
          headers: { 'content-type': 'application/json' }
        })
      )
    )

    const api = createApiClient('')
    const result = await api.createAdminBackup('token-123')

    expect(result.mensagem).toBe('Backup iniciado')
    expect(fetch).toHaveBeenCalledWith(
      '/admin/backup',
      expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({ Authorization: 'Bearer token-123' }),
        body: JSON.stringify({})
      })
    )
  })

  it('requests a protected admin restore', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        new Response(JSON.stringify({ erro: 'restore_indisponivel', mensagem: 'Restore protegido' }), {
          status: 501,
          headers: { 'content-type': 'application/json' }
        })
      )
    )

    const api = createApiClient('')

    await expect(
      api.restoreAdminBackup('token-123', {
        arquivo: 'backup.sql',
        confirmacao: 'RESTAURAR'
      })
    ).rejects.toMatchObject({
      status: 501,
      code: 'restore_indisponivel',
      message: 'Restore protegido'
    } satisfies Partial<APIError>)

    expect(fetch).toHaveBeenCalledWith(
      '/admin/backup/restaurar',
      expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({ Authorization: 'Bearer token-123' }),
        body: JSON.stringify({ arquivo: 'backup.sql', confirmacao: 'RESTAURAR' })
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

  it('manages local network endpoints with bearer token', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValueOnce(
          new Response(JSON.stringify({ roteadores: [] }), {
            status: 200,
            headers: { 'content-type': 'application/json' }
          })
        )
        .mockResolvedValueOnce(
          new Response(JSON.stringify({ entrada: { mac: 'AA:BB:CC:DD:EE:FF' } }), {
            status: 201,
            headers: { 'content-type': 'application/json' }
          })
        )
        .mockResolvedValueOnce(
          new Response(JSON.stringify({ walled_garden: [] }), {
            status: 200,
            headers: { 'content-type': 'application/json' }
          })
        )
    )

    const api = createApiClient('')
    await api.getAdminRouters('token-123')
    await api.addAdminBlacklist('token-123', {
      mac: 'AA:BB:CC:DD:EE:FF',
      motivo: 'teste'
    })
    await api.getAdminWalledGarden('token-123')

    expect(fetch).toHaveBeenNthCalledWith(
      1,
      '/admin/rede/roteadores',
      expect.objectContaining({
        method: 'GET',
        headers: expect.objectContaining({ Authorization: 'Bearer token-123' })
      })
    )
    expect(fetch).toHaveBeenNthCalledWith(
      2,
      '/admin/rede/blacklist',
      expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({ Authorization: 'Bearer token-123' }),
        body: JSON.stringify({ mac: 'AA:BB:CC:DD:EE:FF', motivo: 'teste' })
      })
    )
    expect(fetch).toHaveBeenNthCalledWith(
      3,
      '/admin/rede/walled-garden',
      expect.objectContaining({
        method: 'GET',
        headers: expect.objectContaining({ Authorization: 'Bearer token-123' })
      })
    )
  })
})

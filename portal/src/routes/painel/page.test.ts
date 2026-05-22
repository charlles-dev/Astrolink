import { fireEvent, render, screen, waitFor } from '@testing-library/svelte'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import Page from './+page.svelte'
import { APIError, api } from '$lib/api'

vi.mock('$lib/api', async (importOriginal) => {
  const actual = await importOriginal<typeof import('$lib/api')>()

  return {
    APIError: actual.APIError,
    api: {
      loginAdmin: vi.fn(),
      getAdminHealth: vi.fn(),
      getAdminPlanos: vi.fn(),
      getAdminUsuarios: vi.fn(),
      getAdminVouchers: vi.fn(),
      getAdminPagamentos: vi.fn(),
      getAdminLogs: vi.fn()
    }
  }
})

const mockApi = vi.mocked(api)

beforeEach(() => {
  sessionStorage.clear()
  mockApi.loginAdmin.mockResolvedValue({
    access_token: 'token-123',
    refresh_token: 'refresh-123',
    expires_in: 28800,
    token_type: 'Bearer'
  })
  mockApi.getAdminHealth.mockResolvedValue({
    status: 'ok',
    versao: '0.1.0',
    uptime_segundos: 10,
    checks: {
      banco_dados: { status: 'ok' },
      redis: { status: 'ok' },
      rabbitmq: { status: 'ok' },
      mercadopago: { status: 'ok' },
      roteadores: { total: 0, online: 0, offline: 0 }
    }
  })
  mockApi.getAdminPlanos.mockResolvedValue({ planos: [] })
  mockApi.getAdminUsuarios.mockResolvedValue({ total: 0, page: 1, limit: 50, usuarios: [] })
  mockApi.getAdminVouchers.mockResolvedValue({ total: 0, vouchers: [] })
  mockApi.getAdminPagamentos.mockResolvedValue({
    total: 0,
    totais: { pendente: 0, aprovado: 0, cancelado: 0, expirado: 0, valor_total: '0.00' },
    pagamentos: []
  })
  mockApi.getAdminLogs.mockResolvedValue({ total: 0, logs: [] })
  vi.stubGlobal(
    'fetch',
    vi.fn(async () => new Response(null, { status: 204 }))
  )
})

afterEach(() => {
  vi.clearAllMocks()
  vi.unstubAllGlobals()
  sessionStorage.clear()
})

describe('Painel admin login', () => {
  it('preserves simple login without 2FA', async () => {
    render(Page)

    await fireEvent.submit(screen.getByRole('button', { name: 'Entrar' }).closest('form')!)

    await waitFor(() => {
      expect(mockApi.loginAdmin).toHaveBeenCalledWith({
        usuario: 'admin',
        senha: 'admin123'
      })
    })
  })

  it('shows the 2FA code field when the backend requires TOTP', async () => {
    mockApi.loginAdmin.mockRejectedValueOnce(
      new APIError(428, 'totp_obrigatorio', 'Informe o codigo 2FA')
    )

    render(Page)

    await fireEvent.submit(screen.getByRole('button', { name: 'Entrar' }).closest('form')!)

    expect(await screen.findByLabelText('Codigo 2FA')).toBeInTheDocument()
  })

  it('resubmits login with the TOTP code after 2FA is requested', async () => {
    mockApi.loginAdmin
      .mockRejectedValueOnce(new APIError(428, 'totp_obrigatorio', 'Informe o codigo 2FA'))
      .mockRejectedValueOnce(new APIError(401, 'nao_autenticado', 'Credenciais invalidas'))

    render(Page)

    await fireEvent.submit(screen.getByRole('button', { name: 'Entrar' }).closest('form')!)
    await fireEvent.input(await screen.findByLabelText('Codigo 2FA'), {
      target: { value: '123456' }
    })
    await fireEvent.submit(screen.getByRole('button', { name: 'Entrar' }).closest('form')!)

    await waitFor(() => {
      expect(mockApi.loginAdmin).toHaveBeenLastCalledWith({
        usuario: 'admin',
        senha: 'admin123',
        totp_codigo: '123456'
      })
    })
  })
})

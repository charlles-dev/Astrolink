import { fireEvent, render, screen, waitFor } from '@testing-library/svelte'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import Page from './+page.svelte'
import { api } from '$lib/api'

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
      getAdminLogs: vi.fn(),
      getSetupStatus: vi.fn(),
      updateSetupEnv: vi.fn()
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
  mockApi.getSetupStatus.mockResolvedValue({ requires_restart: false, writable: true, groups: {} })
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

describe('/painel/pagamentos', () => {
  it('renders the payments page after login', async () => {
    render(Page)

    await fireEvent.submit(screen.getByRole('button', { name: 'Entrar' }).closest('form')!)

    expect(await screen.findByRole('heading', { name: 'Pagamentos' })).toBeInTheDocument()
  })

  it('formats totals, explains truncated payments and clears filters', async () => {
    mockApi.getAdminPagamentos.mockResolvedValue({
      total: 11,
      totais: { pendente: 1, aprovado: 10, cancelado: 0, expirado: 0, valor_total: '1234.56' },
      pagamentos: Array.from({ length: 11 }, (_, index) => ({
        txid: `pix-${index + 1}`,
        status: 'aprovado',
        valor: '15.5',
        descricao: 'Acesso 24 Horas',
        mac: `AA:BB:CC:DD:EE:${String(index).padStart(2, '0')}`,
        plano_id: 1,
        plano: { id: 1, nome: 'Acesso 24 Horas' },
        created_at: '2026-05-21T10:00:00Z',
        expira_em: '2026-05-21T10:30:00Z'
      }))
    })

    render(Page)

    await fireEvent.submit(screen.getByRole('button', { name: 'Entrar' }).closest('form')!)

    expect(await screen.findByText('R$ 1.234,56')).toBeInTheDocument()
    expect(screen.getAllByText('R$ 15,50')).toHaveLength(10)
    expect(screen.getByText('Mostrando 10 de 11 pagamentos (limite de 10).')).toBeInTheDocument()

    await fireEvent.change(screen.getByLabelText('Status do pagamento'), {
      target: { value: 'aprovado' }
    })
    await fireEvent.input(screen.getByLabelText('Início'), { target: { value: '2026-05-01' } })
    await fireEvent.input(screen.getByLabelText('Fim'), { target: { value: '2026-05-21' } })
    await fireEvent.click(screen.getByRole('button', { name: 'Aplicar filtros de pagamentos' }))

    await waitFor(() => {
      expect(mockApi.getAdminPagamentos).toHaveBeenLastCalledWith('token-123', {
        status: 'aprovado',
        inicio: '2026-05-01',
        fim: '2026-05-21'
      })
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Limpar filtros de pagamentos' }))

    await waitFor(() => {
      expect(mockApi.getAdminPagamentos).toHaveBeenLastCalledWith('token-123', {})
    })
    expect(screen.getByLabelText('Status do pagamento')).toHaveValue('')
    expect(screen.getByLabelText('Início')).toHaveValue('')
    expect(screen.getByLabelText('Fim')).toHaveValue('')
  })
})

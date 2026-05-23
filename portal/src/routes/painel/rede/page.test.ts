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
      updateSetupEnv: vi.fn(),
      getAdminRouters: vi.fn(),
      createAdminRouter: vi.fn(),
      updateAdminRouter: vi.fn(),
      deleteAdminRouter: vi.fn(),
      diagnoseAdminRouter: vi.fn(),
      speedtestAdminRouter: vi.fn(),
      getAdminBlacklist: vi.fn(),
      addAdminBlacklist: vi.fn(),
      deleteAdminBlacklist: vi.fn(),
      getAdminWalledGarden: vi.fn(),
      addAdminWalledGarden: vi.fn(),
      deleteAdminWalledGarden: vi.fn()
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
      mercadopago: { status: 'disabled' },
      roteadores: { total: 1, online: 1, offline: 0 }
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
  mockApi.getAdminRouters.mockResolvedValue({
    roteadores: [
      {
        id: 1,
        nome: 'Roteador Principal',
        ip: '192.168.1.1',
        porta_ssh: 22,
        usuario_ssh: 'root',
        status: 'online',
        ativo: true,
        usuarios_ativos: 2
      }
    ]
  })
  mockApi.createAdminRouter.mockResolvedValue({
    roteador: {
      id: 2,
      nome: 'Roteador Patio',
      ip: '192.168.1.2',
      porta_ssh: 22,
      usuario_ssh: 'root',
      status: 'online',
      ativo: true,
      usuarios_ativos: 0
    }
  })
  mockApi.getAdminBlacklist.mockResolvedValue({ total: 0, blacklist: [] })
  mockApi.getAdminWalledGarden.mockResolvedValue({
    total: 1,
    walled_garden: [
      {
        id: 1,
        host: 'api.mercadopago.com',
        descricao: 'Mercado Pago API',
        tipo: 'dominio',
        sistema: true
      }
    ]
  })
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

describe('/painel/rede', () => {
  it('loads network data after login', async () => {
    render(Page)

    await fireEvent.submit(screen.getByRole('button', { name: 'Entrar' }).closest('form')!)

    expect(await screen.findByRole('heading', { name: 'Rede local' })).toBeInTheDocument()
    expect(screen.getByText('Roteador Principal')).toBeInTheDocument()
    expect(screen.getByText('api.mercadopago.com')).toBeInTheDocument()
    expect(mockApi.getAdminRouters).toHaveBeenCalledWith('token-123')
  })

  it('creates a router from the local network form', async () => {
    render(Page)

    await fireEvent.submit(screen.getByRole('button', { name: 'Entrar' }).closest('form')!)
    await screen.findByText('Roteador Principal')

    await fireEvent.input(screen.getByLabelText('Nome'), {
      target: { value: 'Roteador Patio' }
    })
    await fireEvent.input(screen.getByLabelText('IP'), {
      target: { value: '192.168.1.2' }
    })
    await fireEvent.click(screen.getByRole('button', { name: 'Cadastrar roteador' }))

    await waitFor(() => {
      expect(mockApi.createAdminRouter).toHaveBeenCalledWith('token-123', {
        nome: 'Roteador Patio',
        ip: '192.168.1.2',
        porta_ssh: 22,
        usuario_ssh: 'root',
        chave_ssh_path: '',
        ativo: true
      })
    })
  })
})

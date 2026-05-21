import { fireEvent, render, screen } from '@testing-library/svelte'
import { describe, expect, it, vi } from 'vitest'

import AdminDashboard from './AdminDashboard.svelte'

describe('AdminDashboard', () => {
  it('renders health, plans and connected users', () => {
    render(AdminDashboard, {
      props: {
        health: {
          status: 'healthy',
          versao: '0.1.0',
          uptime_segundos: 0,
          checks: {
            banco_dados: { status: 'memory', latencia_ms: 0 },
            redis: { status: 'mock' },
            rabbitmq: { status: 'mock' },
            mercadopago: { status: 'mock' },
            roteadores: { total: 1, online: 1, offline: 0 }
          }
        },
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
        ],
        vouchers: [
          {
            id: 1,
            codigo: 'VIPA-1234',
            plano: { id: 1, nome: 'Acesso 24 Horas' },
            tipo: 'single_use',
            usos_atuais: 0,
            ativo: true
          }
        ],
        usuarios: [
          {
            id: 1,
            mac: 'AA:BB:CC:DD:EE:FF',
            ip_atual: '192.168.1.50',
            status: 'ativo',
            plano: { id: 1, nome: 'Acesso 24 Horas' },
            tempo_restante_segundos: 3600,
            dados_consumidos_mb: 120
          }
        ],
        loading: false,
        actionMessage: '',
        onRefresh: vi.fn(),
        onDisconnect: vi.fn(),
        onGenerateVouchers: vi.fn(),
        onLogout: vi.fn()
      }
    })

    expect(screen.getByRole('heading', { name: 'Painel local' })).toBeInTheDocument()
    expect(screen.getByText('Banco')).toBeInTheDocument()
    expect(screen.getByText('memory')).toBeInTheDocument()
    expect(screen.getAllByText('Acesso 24 Horas').length).toBeGreaterThan(0)
    expect(screen.getByText('AA:BB:CC:DD:EE:FF')).toBeInTheDocument()
    expect(screen.getByText('VIPA-1234')).toBeInTheDocument()
  })

  it('requests user disconnect from the row action', async () => {
    const onDisconnect = vi.fn()

    render(AdminDashboard, {
      props: {
        health: null,
        planos: [],
        vouchers: [],
        usuarios: [
          {
            id: 1,
            mac: 'AA:BB:CC:DD:EE:FF',
            status: 'ativo',
            tempo_restante_segundos: 0,
            dados_consumidos_mb: 0
          }
        ],
        loading: false,
        actionMessage: '',
        onRefresh: vi.fn(),
        onDisconnect,
        onGenerateVouchers: vi.fn(),
        onLogout: vi.fn()
      }
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Desconectar AA:BB:CC:DD:EE:FF' }))

    expect(onDisconnect).toHaveBeenCalledWith('AA:BB:CC:DD:EE:FF')
  })

  it('submits single-use voucher generation with validity days', async () => {
    const onGenerateVouchers = vi.fn()

    render(AdminDashboard, {
      props: {
        health: null,
        planos: [
          {
            id: 2,
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
        ],
        vouchers: [],
        usuarios: [],
        loading: false,
        actionMessage: '',
        onRefresh: vi.fn(),
        onDisconnect: vi.fn(),
        onGenerateVouchers,
        onLogout: vi.fn()
      }
    })

    await fireEvent.input(screen.getByLabelText('Prefixo'), { target: { value: 'BARCO' } })
    await fireEvent.input(screen.getByLabelText('Quantidade'), { target: { value: '3' } })
    await fireEvent.input(screen.getByLabelText('Validade (dias)'), { target: { value: '7' } })
    await fireEvent.click(screen.getByRole('button', { name: 'Gerar vouchers' }))

    expect(onGenerateVouchers).toHaveBeenCalledWith({
      plano_id: 2,
      quantidade: 3,
      tipo: 'single_use',
      validade_dias: 7,
      prefixo: 'BARCO'
    })
  })

  it('submits universal voucher generation with max uses', async () => {
    const onGenerateVouchers = vi.fn()

    render(AdminDashboard, {
      props: {
        health: null,
        planos: [
          {
            id: 2,
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
        ],
        vouchers: [],
        usuarios: [],
        loading: false,
        actionMessage: '',
        onRefresh: vi.fn(),
        onDisconnect: vi.fn(),
        onGenerateVouchers,
        onLogout: vi.fn()
      }
    })

    await fireEvent.click(screen.getByLabelText('Universal'))
    await fireEvent.input(screen.getByLabelText('Prefixo'), { target: { value: 'pub' } })
    await fireEvent.input(screen.getByLabelText('Quantidade'), { target: { value: '4' } })
    await fireEvent.input(screen.getByLabelText('Usos maximos'), { target: { value: '25' } })
    await fireEvent.click(screen.getByRole('button', { name: 'Gerar vouchers' }))

    expect(onGenerateVouchers).toHaveBeenCalledWith({
      plano_id: 2,
      quantidade: 4,
      tipo: 'universal',
      usos_maximos: 25,
      prefixo: 'PUB'
    })
  })
})

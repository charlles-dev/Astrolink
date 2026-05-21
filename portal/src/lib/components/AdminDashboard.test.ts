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
        onLogout: vi.fn()
      }
    })

    expect(screen.getByRole('heading', { name: 'Painel local' })).toBeInTheDocument()
    expect(screen.getByText('Banco memory')).toBeInTheDocument()
    expect(screen.getByText('Acesso 24 Horas')).toBeInTheDocument()
    expect(screen.getByText('AA:BB:CC:DD:EE:FF')).toBeInTheDocument()
  })

  it('requests user disconnect from the row action', async () => {
    const onDisconnect = vi.fn()

    render(AdminDashboard, {
      props: {
        health: null,
        planos: [],
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
        onLogout: vi.fn()
      }
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Desconectar AA:BB:CC:DD:EE:FF' }))

    expect(onDisconnect).toHaveBeenCalledWith('AA:BB:CC:DD:EE:FF')
  })
})

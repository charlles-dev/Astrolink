import { render, screen, within } from '@testing-library/svelte'
import { describe, expect, it } from 'vitest'

import AdminLiveEventsPanel from './AdminLiveEventsPanel.svelte'

describe('AdminLiveEventsPanel', () => {
  it('renders connected state with snapshot counters', () => {
    render(AdminLiveEventsPanel, {
      props: {
        connected: true,
        lastEventAt: '2026-05-21T10:35:00',
        snapshot: {
          usuarios: { ativos: 12, total: 38 },
          vouchers: { ativos: 7, total: 22 },
          pix: { pendente: 3, aprovado: 19 },
          logs: 144
        },
        events: []
      }
    })

    expect(screen.getByRole('heading', { name: 'Eventos ao vivo' })).toBeInTheDocument()
    expect(screen.getByText('Conectado')).toBeInTheDocument()
    expect(screen.getByText('Recebendo eventos em tempo real.')).toBeInTheDocument()
    expect(screen.getByText(/Última atualização: 21\/05.*10:35/)).toBeInTheDocument()

    expect(screen.getByText('Usuários')).toBeInTheDocument()
    expect(screen.getByText('12/38')).toBeInTheDocument()
    expect(screen.getByText('Vouchers')).toBeInTheDocument()
    expect(screen.getByText('7/22')).toBeInTheDocument()
    expect(screen.getByText('PIX')).toBeInTheDocument()
    expect(screen.getByText('3 pendente')).toBeInTheDocument()
    expect(screen.getByText('19 aprovado')).toBeInTheDocument()
    expect(screen.getByText('Logs')).toBeInTheDocument()
    expect(screen.getByText('144')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Ver logs' })).toHaveAttribute('href', '/painel/logs')
  })

  it('renders disconnected state without snapshot', () => {
    render(AdminLiveEventsPanel, {
      props: {
        connected: false,
        lastEventAt: '',
        snapshot: null,
        events: []
      }
    })

    expect(screen.getByText('Desconectado')).toBeInTheDocument()
    expect(screen.getByText('Aguardando reconexao do canal ao vivo.')).toBeInTheDocument()
    expect(screen.getByText('Sem snapshot recebido.')).toBeInTheDocument()
    expect(screen.getByText('Nenhum evento recebido')).toBeInTheDocument()
  })

  it('renders a compact list of recent events', () => {
    render(AdminLiveEventsPanel, {
      props: {
        connected: true,
        lastEventAt: '2026-05-21T10:35:00',
        snapshot: null,
        events: [
          {
            id: 'evt-1',
            tipo: 'pix.aprovado',
            mensagem: 'Pagamento PIX confirmado para AA:BB:CC:DD:EE:FF.',
            timestamp: '2026-05-21T10:35:00'
          },
          {
            id: 'evt-2',
            tipo: 'usuario.conectado',
            mensagem: 'Cliente liberado no roteador principal.',
            timestamp: '2026-05-21T10:33:00'
          }
        ]
      }
    })

    const items = screen.getAllByRole('article')

    expect(items).toHaveLength(2)
    expect(within(items[0]).getByText('pix.aprovado')).toBeInTheDocument()
    expect(within(items[0]).getByText('Pagamento PIX confirmado para AA:BB:CC:DD:EE:FF.')).toBeInTheDocument()
    expect(within(items[0]).getByText(/21\/05.*10:35/)).toBeInTheDocument()
    expect(within(items[1]).getByText('usuario.conectado')).toBeInTheDocument()
    expect(within(items[1]).getByText('Cliente liberado no roteador principal.')).toBeInTheDocument()
  })
})

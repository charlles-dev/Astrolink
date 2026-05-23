import { fireEvent, render, screen } from '@testing-library/svelte'
import { describe, expect, it, vi } from 'vitest'

import AdminUsersPanel from './AdminUsersPanel.svelte'

describe('AdminUsersPanel', () => {
  it('shows a loading state instead of an empty state while users are loading', () => {
    render(AdminUsersPanel, {
      props: {
        usuarios: [],
        loading: true,
        onDisconnect: vi.fn()
      }
    })

    expect(screen.getByText('Carregando usuários')).toBeInTheDocument()
    expect(screen.queryByText('Nenhum usuário registrado')).not.toBeInTheDocument()
  })

  it('exposes local user actions for extending and banning access', async () => {
    const onExtendUser = vi.fn()
    const onBanUser = vi.fn()
    render(AdminUsersPanel, {
      props: {
        usuarios: [
          {
            id: 1,
            mac: 'AA:BB:CC:DD:EE:FF',
            ip_atual: '192.168.1.50',
            status: 'ativo',
            tempo_restante_segundos: 120,
            dados_consumidos_mb: 8
          }
        ],
        loading: false,
        onDisconnect: vi.fn(),
        onExtendUser,
        onBanUser
      }
    })

    await fireEvent.input(screen.getByLabelText('Minutos para AA:BB:CC:DD:EE:FF'), {
      target: { value: '90' }
    })
    await fireEvent.click(screen.getByRole('button', { name: 'Estender' }))
    await fireEvent.input(screen.getByLabelText('Motivo do bloqueio para AA:BB:CC:DD:EE:FF'), {
      target: { value: 'abuso' }
    })
    await fireEvent.click(screen.getByRole('button', { name: 'Banir' }))

    expect(onExtendUser).toHaveBeenCalledWith('AA:BB:CC:DD:EE:FF', 90)
    expect(onBanUser).toHaveBeenCalledWith('AA:BB:CC:DD:EE:FF', 'abuso')
  })
})

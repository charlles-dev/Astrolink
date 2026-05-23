import { render, screen } from '@testing-library/svelte'
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
})

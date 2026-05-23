import { fireEvent, render, screen } from '@testing-library/svelte'
import { describe, expect, it, vi } from 'vitest'

import AdminSetupPanel from './AdminSetupPanel.svelte'
import type { SetupStatus } from '../../types'

const setupStatus: SetupStatus = {
  requires_restart: false,
  writable: true,
  groups: {
    mercadopago: {
      label: 'Mercado Pago',
      fields: [
        {
          key: 'MERCADOPAGO_ACCESS_TOKEN',
          label: 'Access token',
          description: 'Token privado do Mercado Pago',
          secret: true,
          configured: true
        },
        {
          key: 'MERCADOPAGO_PAYER_EMAIL',
          label: 'E-mail pagador',
          description: 'E-mail padrao',
          secret: false,
          configured: true,
          value: 'cliente@example.com'
        }
      ]
    },
    admin: {
      label: 'Admin',
      fields: [
        {
          key: 'ADMIN_USUARIO',
          label: 'Usuario admin',
          description: 'Login do painel',
          secret: false,
          configured: true,
          value: 'admin'
        }
      ]
    }
  }
}

describe('AdminSetupPanel', () => {
  it('keeps configured secret blank out of the patch', async () => {
    const onSaveSetup = vi.fn()

    render(AdminSetupPanel, {
      props: {
        setupStatus,
        loading: false,
        onSaveSetup
      }
    })

    expect(screen.getByLabelText('Access token')).toHaveAttribute('placeholder', 'Configurado')

    await fireEvent.input(screen.getByLabelText('E-mail pagador'), {
      target: { value: 'novo@example.com' }
    })
    await fireEvent.click(screen.getByRole('button', { name: 'Salvar setup local' }))

    expect(onSaveSetup).toHaveBeenCalledWith({
      MERCADOPAGO_PAYER_EMAIL: 'novo@example.com'
    })
  })

  it('sends edited fields as a setup patch', async () => {
    const onSaveSetup = vi.fn()

    render(AdminSetupPanel, {
      props: {
        setupStatus,
        loading: false,
        onSaveSetup
      }
    })

    await fireEvent.input(screen.getByLabelText('Usuario admin'), {
      target: { value: 'operador' }
    })
    await fireEvent.click(screen.getByRole('button', { name: 'Salvar setup local' }))

    expect(onSaveSetup).toHaveBeenCalledWith({
      ADMIN_USUARIO: 'operador'
    })
  })
})

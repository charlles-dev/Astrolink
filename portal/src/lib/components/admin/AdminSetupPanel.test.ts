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
          label: 'Usuário admin',
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
    expect(screen.getAllByText('Configurado').length).toBeGreaterThan(0)
    expect(screen.getByText('Deixe em branco para manter o valor atual.')).toBeInTheDocument()

    await fireEvent.input(screen.getByLabelText('E-mail pagador'), {
      target: { value: 'novo@example.com' }
    })
    await fireEvent.click(screen.getByRole('button', { name: 'Salvar setup local' }))

    expect(onSaveSetup).toHaveBeenCalledWith({
      MERCADOPAGO_PAYER_EMAIL: 'novo@example.com'
    })
  })

  it('shows per-field configured badges and a persistent restart alert', () => {
    render(AdminSetupPanel, {
      props: {
        setupStatus: {
          ...setupStatus,
          requires_restart: true,
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
                  key: 'MERCADOPAGO_WEBHOOK_SECRET',
                  label: 'Webhook secret',
                  description: 'Assinatura do webhook',
                  secret: true,
                  configured: false
                }
              ]
            }
          }
        },
        loading: false,
        onSaveSetup: vi.fn()
      }
    })

    expect(screen.queryByText('Reiniciar')).not.toBeInTheDocument()
    expect(screen.getByRole('status', { name: 'Reinício necessário' })).toHaveTextContent(
      'Reinicie o serviço para aplicar as alterações de setup local.'
    )
    expect(screen.getByText('Não configurado')).toBeInTheDocument()
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

    await fireEvent.input(screen.getByLabelText('Usuário admin'), {
      target: { value: 'operador' }
    })
    await fireEvent.click(screen.getByRole('button', { name: 'Salvar setup local' }))

    expect(onSaveSetup).toHaveBeenCalledWith({
      ADMIN_USUARIO: 'operador'
    })
  })
})

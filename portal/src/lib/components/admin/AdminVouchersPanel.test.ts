import { fireEvent, render, screen, within } from '@testing-library/svelte'
import { afterEach, describe, expect, it, vi } from 'vitest'

import AdminVouchersPanel from './AdminVouchersPanel.svelte'
import type { AdminVoucher, Plano } from '../../types'

const planoBase: Plano = {
  id: 1,
  nome: 'Acesso 24 Horas',
  descricao: 'Internet por um dia',
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

const vouchers: AdminVoucher[] = [
  {
    id: 7,
    codigo: 'VIPA-7777',
    plano: { id: 1, nome: 'Acesso 24 Horas' },
    tipo: 'single_use',
    usos_atuais: 0,
    usos_maximos: 1,
    lote_id: 12,
    validade_em: '2026-06-01T12:00:00Z',
    ativo: true
  },
  {
    id: 8,
    codigo: 'PUB-8888',
    plano: { id: 1, nome: 'Acesso 24 Horas' },
    tipo: 'universal',
    usos_atuais: 3,
    usos_maximos: 25,
    ativo: true
  }
]

describe('AdminVouchersPanel', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('prints a PDF-ready voucher sheet with batch summary and customer instructions', async () => {
    const print = vi.fn()
    vi.stubGlobal('print', print)

    render(AdminVouchersPanel, {
      props: {
        planos: [planoBase],
        vouchers,
        loading: false
      }
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Gerar folha PDF' }))

    expect(print).toHaveBeenCalledTimes(1)

    const sheet = screen.getByTestId('voucher-print-sheet')
    const tickets = within(sheet).getAllByRole('article')

    expect(within(sheet).getByRole('heading', { name: 'Astrolink' })).toBeInTheDocument()
    expect(within(sheet).getByText('Folha para PDF ou impressao')).toBeInTheDocument()
    expect(within(sheet).getByText('2 vouchers')).toBeInTheDocument()
    expect(within(sheet).getAllByText('Lote 12').length).toBeGreaterThan(0)
    expect(within(sheet).getAllByText('Acesso 24 Horas').length).toBeGreaterThan(0)
    expect(within(sheet).getByText('Recorte nas linhas pontilhadas')).toBeInTheDocument()

    expect(tickets).toHaveLength(2)
    expect(within(tickets[0]).getByText('VIPA-7777')).toBeInTheDocument()
    expect(within(tickets[0]).getByText('Acesso 24 Horas')).toBeInTheDocument()
    expect(within(tickets[0]).getByText('Uso unico')).toBeInTheDocument()
    expect(within(tickets[0]).getByText('0/1')).toBeInTheDocument()
    expect(within(tickets[0]).getByText('Lote 12')).toBeInTheDocument()
    expect(within(tickets[0]).getByText('01/06/2026')).toBeInTheDocument()
    expect(within(tickets[0]).getByText('Digite este codigo no portal do Wi-Fi.')).toBeInTheDocument()
    expect(within(tickets[0]).getByText('Valido somente ate a data indicada.')).toBeInTheDocument()
    expect(within(tickets[1]).getByText('Universal')).toBeInTheDocument()
    expect(within(tickets[1]).getByText('3/25')).toBeInTheDocument()
  })
})

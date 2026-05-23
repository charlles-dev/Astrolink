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

function makeVoucher(id: number): AdminVoucher {
  return {
    id,
    codigo: `VIP-${String(id).padStart(4, '0')}`,
    plano: { id: 1, nome: 'Acesso 24 Horas' },
    tipo: 'single_use',
    usos_atuais: 0,
    usos_maximos: 1,
    ativo: true
  }
}

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

  it('disables voucher controls while loading', () => {
    render(AdminVouchersPanel, {
      props: {
        planos: [planoBase],
        vouchers,
        loading: true
      }
    })

    expect(screen.getByLabelText('Prefixo')).toBeDisabled()
    expect(screen.getByLabelText('Quantidade')).toBeDisabled()
    expect(screen.getByLabelText('Código')).toBeDisabled()
    expect(screen.getByLabelText('Lote')).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Gerar vouchers' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Aplicar filtros' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Exportar CSV' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Gerar folha PDF' })).toBeDisabled()
  })

  it('shows the visible count when the voucher list is limited to eight', () => {
    render(AdminVouchersPanel, {
      props: {
        planos: [planoBase],
        vouchers: Array.from({ length: 10 }, (_, index) => makeVoucher(index + 1)),
        loading: false
      }
    })

    expect(screen.getByText('Mostrando 8 de 10 vouchers')).toBeInTheDocument()
    expect(screen.getByText('Há mais 2 vouchers nos filtros atuais.')).toBeInTheDocument()
    expect(screen.queryByText('VIP-0009')).not.toBeInTheDocument()
  })

  it('asks for inline context before deactivating a voucher and can cancel', async () => {
    const onDeactivateVoucher = vi.fn()

    render(AdminVouchersPanel, {
      props: {
        planos: [planoBase],
        vouchers,
        loading: false,
        onDeactivateVoucher
      }
    })

    const row = screen.getByRole('article', { name: 'Voucher VIPA-7777' })
    await fireEvent.click(within(row).getByRole('button', { name: 'Desativar VIPA-7777' }))

    expect(within(row).getByText('Desativar VIPA-7777? O código não poderá ser usado em novas ativações.')).toBeInTheDocument()

    await fireEvent.click(within(row).getByRole('button', { name: 'Cancelar desativação de VIPA-7777' }))
    expect(onDeactivateVoucher).not.toHaveBeenCalled()
    expect(within(row).queryByText('Desativar VIPA-7777? O código não poderá ser usado em novas ativações.')).not.toBeInTheDocument()

    await fireEvent.click(within(row).getByRole('button', { name: 'Desativar VIPA-7777' }))
    await fireEvent.click(within(row).getByRole('button', { name: 'Confirmar desativação de VIPA-7777' }))

    expect(onDeactivateVoucher).toHaveBeenCalledWith(7)
  })
})

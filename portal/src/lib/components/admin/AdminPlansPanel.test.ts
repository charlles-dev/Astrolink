import { fireEvent, render, screen, within } from '@testing-library/svelte'
import { describe, expect, it, vi } from 'vitest'

import AdminPlansPanel from './AdminPlansPanel.svelte'
import type { Plano } from '../../types'

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

describe('AdminPlansPanel', () => {
  it('submits a new plan with admin fields', async () => {
    const onSavePlan = vi.fn()

    render(AdminPlansPanel, {
      props: {
        planos: [],
        loading: false,
        onSavePlan
      }
    })

    await fireEvent.input(screen.getByLabelText('Nome'), { target: { value: 'Noite Livre' } })
    await fireEvent.input(screen.getByLabelText('Descrição'), { target: { value: 'Acesso noturno' } })
    await fireEvent.input(screen.getByLabelText('Preço'), { target: { value: '9.90' } })
    await fireEvent.input(screen.getByLabelText('Duração (min)'), { target: { value: '480' } })
    await fireEvent.input(screen.getByLabelText('Dados (MB)'), { target: { value: '2048' } })
    await fireEvent.input(screen.getByLabelText('Download (Mbps)'), { target: { value: '20' } })
    await fireEvent.input(screen.getByLabelText('Upload (Mbps)'), { target: { value: '8' } })
    await fireEvent.input(screen.getByLabelText('Ordem'), { target: { value: '3' } })
    await fireEvent.click(screen.getByLabelText('Recomendado'))
    await fireEvent.click(screen.getByLabelText('Visível no portal'))
    await fireEvent.click(screen.getByRole('button', { name: 'Salvar plano' }))

    expect(onSavePlan).toHaveBeenCalledWith({
      nome: 'Noite Livre',
      descricao: 'Acesso noturno',
      preco: 9.9,
      duracao_minutos: 480,
      dados_mb: 2048,
      velocidade_down: 20,
      velocidade_up: 8,
      recomendado: true,
      ativo: true,
      visivel_portal: false,
      ordem: 3
    })
  })

  it('edits an existing plan and can cancel editing', async () => {
    const onSavePlan = vi.fn()

    render(AdminPlansPanel, {
      props: {
        planos: [planoBase],
        loading: false,
        onSavePlan
      }
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Editar Acesso 24 Horas' }))
    expect(screen.getByText('Editando: Acesso 24 Horas')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Cancelar edição' })).toBeInTheDocument()

    await fireEvent.input(screen.getByLabelText('Preço'), { target: { value: '18.50' } })
    await fireEvent.click(screen.getByRole('button', { name: 'Atualizar plano' }))

    expect(onSavePlan).toHaveBeenCalledWith(
      expect.objectContaining({ preco: 18.5 }),
      planoBase.id
    )

    await fireEvent.click(screen.getByRole('button', { name: 'Editar Acesso 24 Horas' }))
    await fireEvent.click(screen.getByRole('button', { name: 'Cancelar edição' }))

    expect(screen.getByRole('button', { name: 'Salvar plano' })).toBeInTheDocument()
  })

  it('blocks invalid numeric values instead of submitting them as zero', async () => {
    const onSavePlan = vi.fn()

    render(AdminPlansPanel, {
      props: {
        planos: [],
        loading: false,
        onSavePlan
      }
    })

    await fireEvent.input(screen.getByLabelText('Nome'), { target: { value: 'Noite Livre' } })
    await fireEvent.input(screen.getByLabelText('Preço'), { target: { value: 'abc' } })
    await fireEvent.input(screen.getByLabelText('Download (Mbps)'), { target: { value: '' } })
    await fireEvent.input(screen.getByLabelText('Upload (Mbps)'), { target: { value: 'abc' } })
    await fireEvent.click(screen.getByRole('button', { name: 'Salvar plano' }))

    expect(screen.getByText('Informe um preço válido.')).toBeInTheDocument()
    expect(screen.getByText('Informe o download em Mbps.')).toBeInTheDocument()
    expect(screen.getByText('Informe o upload em Mbps.')).toBeInTheDocument()
    expect(onSavePlan).not.toHaveBeenCalled()
  })

  it('renders plan state and toggles active status', async () => {
    const onTogglePlanStatus = vi.fn()

    render(AdminPlansPanel, {
      props: {
        planos: [{ ...planoBase, ativo: false }],
        loading: false,
        onTogglePlanStatus
      }
    })

    const row = screen.getByRole('article', { name: 'Acesso 24 Horas' })

    expect(within(row).getByText('Inativo')).toBeInTheDocument()
    expect(within(row).getByText('Recomendado')).toBeInTheDocument()
    expect(within(row).getByText('R$ 15,00')).toBeInTheDocument()
    expect(within(row).getByText('24 horas')).toBeInTheDocument()

    await fireEvent.click(within(row).getByRole('button', { name: 'Ativar Acesso 24 Horas' }))

    expect(onTogglePlanStatus).toHaveBeenCalledWith(1, true)
  })
})

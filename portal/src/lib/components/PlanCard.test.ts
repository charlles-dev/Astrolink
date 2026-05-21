import { render, screen } from '@testing-library/svelte'
import { describe, expect, it, vi } from 'vitest'

import PlanCard from './PlanCard.svelte'

describe('PlanCard', () => {
  it('renders plan name, price, duration and recommended badge', () => {
    render(PlanCard, {
      props: {
        plano: {
          id: 1,
          nome: 'Acesso 24 Horas',
          descricao: 'Um dia completo',
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
        },
        onSelect: vi.fn()
      }
    })

    expect(screen.getByText('Acesso 24 Horas')).toBeInTheDocument()
    expect(screen.getByText('R$ 15,00')).toBeInTheDocument()
    expect(screen.getByText('RECOMENDADO')).toBeInTheDocument()
  })
})

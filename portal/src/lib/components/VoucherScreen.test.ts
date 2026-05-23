import { render, screen } from '@testing-library/svelte'
import { describe, expect, it, vi } from 'vitest'

import VoucherScreen from './VoucherScreen.svelte'

describe('VoucherScreen', () => {
  it('shows backend voucher errors inline', () => {
    render(VoucherScreen, {
      props: {
        error: 'voucher nao encontrado',
        onBack: vi.fn(),
        onSubmit: vi.fn()
      }
    })

    expect(screen.getByRole('alert')).toHaveTextContent('voucher nao encontrado')
  })
})

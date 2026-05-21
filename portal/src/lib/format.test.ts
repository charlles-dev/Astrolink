import { describe, expect, it } from 'vitest'
import { formatCurrency, formatCountdown, maskVoucherCode } from './format'

describe('formatCurrency', () => {
  it('formats backend decimal strings as BRL', () => {
    expect(formatCurrency('15.00')).toBe('R$ 15,00')
  })
})

describe('formatCountdown', () => {
  it('formats seconds as compact hours and minutes', () => {
    expect(formatCountdown(86399)).toBe('23h 59m')
  })
})

describe('maskVoucherCode', () => {
  it('uppercases and groups voucher codes', () => {
    expect(maskVoucherCode('test1234')).toBe('TEST-1234')
  })
})

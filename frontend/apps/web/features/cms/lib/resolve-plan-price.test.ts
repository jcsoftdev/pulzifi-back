import { describe, expect, it } from 'bun:test'
import { formatCents } from './resolve-plan-price'

describe('formatCents', () => {
  it('formats whole-dollar usd amounts without decimals', () => {
    expect(formatCents(2000, 'usd')).toBe('$20')
  })

  it('formats fractional usd amounts with two decimals', () => {
    expect(formatCents(2050, 'usd')).toBe('$20.50')
  })

  it('returns "Custom" for null amounts', () => {
    expect(formatCents(null, 'usd')).toBe('Custom')
  })

  it('prefixes the ISO code for non-usd currencies', () => {
    expect(formatCents(1000, 'eur')).toBe('EUR 10')
  })
})

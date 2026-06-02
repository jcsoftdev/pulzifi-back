import { describe, expect, test } from 'bun:test'
import { mapAlertType } from './notification-api'

describe('mapAlertType', () => {
  const cases: Array<{ input: string; expected: ReturnType<typeof mapAlertType> }> = [
    { input: 'content_change', expected: 'warning' },
    { input: 'error', expected: 'error' },
    { input: 'unknown_type', expected: 'info' },
    { input: '', expected: 'info' },
  ]
  for (const c of cases) {
    test(`"${c.input}" → ${c.expected}`, () => {
      expect(mapAlertType(c.input)).toBe(c.expected)
    })
  }
})

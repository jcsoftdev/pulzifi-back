import { describe, expect, test } from 'bun:test'
import { mapLoginResponse, transformUser } from './auth-api'

type UserBackend = Parameters<typeof transformUser>[0]
type LoginBackend = Parameters<typeof mapLoginResponse>[0]

const baseUser: UserBackend = {
  id: 'u1',
  name: 'Ada',
  email: 'ada@example.com',
  role: 'OWNER',
  created_at: '2026-01-01T00:00:00Z',
}

describe('transformUser', () => {
  test('maps snake_case timestamps to camelCase', () => {
    const out = transformUser({ ...baseUser, updated_at: '2026-02-02T00:00:00Z' })
    expect(out.createdAt).toBe('2026-01-01T00:00:00Z')
    expect(out.updatedAt).toBe('2026-02-02T00:00:00Z')
  })

  describe('tenant coercion', () => {
    const cases: Array<{ name: string; input: UserBackend['tenant']; expected: string | null | undefined }> = [
      { name: 'undefined stays undefined', input: undefined, expected: undefined },
      { name: 'null stays null', input: null, expected: null },
      { name: 'value is preserved', input: 'acme', expected: 'acme' },
    ]
    for (const c of cases) {
      test(c.name, () => {
        expect(transformUser({ ...baseUser, tenant: c.input }).tenant).toBe(c.expected)
      })
    }
  })

  describe('organization nesting', () => {
    test('absent organization → undefined', () => {
      expect(transformUser(baseUser).organization).toBeUndefined()
    })

    test('null organization → undefined', () => {
      expect(transformUser({ ...baseUser, organization: null }).organization).toBeUndefined()
    })

    test('present organization is mapped through', () => {
      const out = transformUser({
        ...baseUser,
        organization: {
          id: 'o1',
          name: 'Acme',
          subdomain: 'acme',
          planCode: 'pro',
          featureFlags: { beta: true },
        },
      })
      expect(out.organization).toEqual({
        id: 'o1',
        name: 'Acme',
        subdomain: 'acme',
        planCode: 'pro',
        featureFlags: { beta: true },
      })
    })

    test('missing featureFlags defaults to empty object', () => {
      const out = transformUser({
        ...baseUser,
        organization: {
          id: 'o1',
          name: 'Acme',
          subdomain: 'acme',
          planCode: 'pro',
        } as NonNullable<UserBackend['organization']>,
      })
      expect(out.organization?.featureFlags).toEqual({})
    })
  })
})

describe('mapLoginResponse', () => {
  const cases: Array<{ name: string; input: LoginBackend; expectedExpiry: number }> = [
    { name: 'prefers expires_in', input: { expires_in: 900, expiresIn: 111 }, expectedExpiry: 900 },
    { name: 'falls back to expiresIn', input: { expiresIn: 1200 }, expectedExpiry: 1200 },
    { name: 'defaults to 3600 when both absent', input: {}, expectedExpiry: 3600 },
  ]
  for (const c of cases) {
    test(`expiry — ${c.name}`, () => {
      expect(mapLoginResponse(c.input).expiresIn).toBe(c.expectedExpiry)
    })
  }

  test('null tenant becomes undefined', () => {
    expect(mapLoginResponse({ tenant: null }).tenant).toBeUndefined()
  })

  test('tenant and nonce are preserved', () => {
    const out = mapLoginResponse({ tenant: 'acme', nonce: 'n-123' })
    expect(out.tenant).toBe('acme')
    expect(out.nonce).toBe('n-123')
  })
})

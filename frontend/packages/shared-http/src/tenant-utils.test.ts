import { beforeAll, describe, expect, test } from 'bun:test'

// The `env` module captures process.env at import time, so we set the app domain
// BEFORE importing the module under test to make every case deterministic
// regardless of the ambient shell environment.
let extractTenantFromHostname: (hostname: string) => string | null

beforeAll(async () => {
  process.env.NEXT_PUBLIC_APP_DOMAIN = 'pulzifi.com'
  const mod = await import('./tenant-utils')
  extractTenantFromHostname = mod.extractTenantFromHostname
})

describe('extractTenantFromHostname', () => {
  const cases: Array<{ name: string; hostname: string; want: string | null }> = [
    // Configured app domain (NEXT_PUBLIC_APP_DOMAIN = pulzifi.com)
    { name: 'tenant subdomain on app domain', hostname: 'acme.pulzifi.com', want: 'acme' },
    { name: 'bare app domain is not a tenant', hostname: 'pulzifi.com', want: null },
    { name: 'app-domain match is case-insensitive', hostname: 'ACME.PULZIFI.COM', want: 'acme' },
    { name: 'port is stripped before extraction', hostname: 'acme.pulzifi.com:3001', want: 'acme' },

    // Local development domains (fall through to label-based parsing)
    { name: 'localhost subdomain', hostname: 'tenant1.localhost', want: 'tenant1' },
    { name: 'localhost subdomain with port', hostname: 'tenant1.localhost:3001', want: 'tenant1' },
    { name: 'lvh.me wildcard subdomain', hostname: 'acme.lvh.me', want: 'acme' },

    // Non-tenant hostnames
    { name: 'bare localhost', hostname: 'localhost', want: null },
    { name: 'empty hostname', hostname: '', want: null },
    { name: 'two-label non-localhost base domain', hostname: 'example.com', want: null },

    // Reserved subdomains are never tenants — filtered on BOTH the label-based
    // path (lvh.me) and the configured-app-domain path (pulzifi.com), so the
    // result is deterministic regardless of env import timing.
    { name: 'www prefix on lvh.me', hostname: 'www.lvh.me', want: null },
    { name: 'api prefix on lvh.me', hostname: 'api.lvh.me', want: null },
    { name: 'admin prefix on lvh.me', hostname: 'admin.lvh.me', want: null },
    { name: 'www prefix on app domain', hostname: 'www.pulzifi.com', want: null },
    { name: 'api prefix on app domain', hostname: 'api.pulzifi.com', want: null },
  ]

  for (const c of cases) {
    test(c.name, () => {
      expect(extractTenantFromHostname(c.hostname)).toBe(c.want)
    })
  }
})

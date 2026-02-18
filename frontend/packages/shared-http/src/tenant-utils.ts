/**
 * Tenant extraction utilities (shared across frontend)
 * Handles subdomain-based multi-tenancy
 */

/**
 * Extracts tenant from hostname
 * @param hostname - The hostname (e.g., "tenant1.localhost", "tenant1.app.com", "www.app.com")
 * @returns tenant name or null if not found
 *
 * Examples:
 * - "tenant1.localhost" → "tenant1"
 * - "tenant1.app.com" → "tenant1"
 * - "www.app.com" → null
 * - "localhost" → null
 */
export function extractTenantFromHostname(hostname: string): string | null {
  const normalizedHostname = hostname.split(':')[0]?.toLowerCase() ?? ''
  if (!normalizedHostname) {
    return null
  }

  const appDomain = process.env.NEXT_PUBLIC_APP_DOMAIN?.toLowerCase()
  if (appDomain) {
    if (normalizedHostname === appDomain) {
      return null
    }

    if (normalizedHostname.endsWith(`.${appDomain}`)) {
      const tenantPart = normalizedHostname.slice(0, -(appDomain.length + 1))
      return tenantPart || null
    }
  }

  const parts = normalizedHostname.split('.')

  // Need at least 2 parts for a subdomain (e.g., tenant.localhost)
  if (parts.length < 2) {
    return null
  }

  // If not localhost and only two labels, treat it as base domain (non-tenant)
  // Example: pulzifi.local -> null
  if (parts.length === 2 && parts[1] !== 'localhost') {
    return null
  }

  const subdomain = parts[0] ?? ''

  // Ignore common prefixes that aren't tenants
  const ignoredSubdomains = [
    'www',
    'api',
    'admin',
    'app',
  ]
  if (ignoredSubdomains.includes(subdomain.toLowerCase())) {
    return null
  }

  return subdomain
}

/**
 * Gets tenant from current window location (client-side only)
 * Returns default tenant from env if no subdomain found
 */
export function getTenantFromWindow(): string | null {
  if (globalThis.window === undefined) {
    return null
  }

  const tenant = extractTenantFromHostname(globalThis.window.location.hostname)

  // Fallback to default tenant for development
  if (!tenant && typeof process !== 'undefined' && process.env?.NEXT_PUBLIC_DEFAULT_TENANT) {
    return process.env.NEXT_PUBLIC_DEFAULT_TENANT
  }

  return tenant
}

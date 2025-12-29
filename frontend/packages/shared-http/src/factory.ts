import type { IHttpClient } from './types'
import { AxiosHttpClient } from './axios-client'
import { FetchHttpClient } from './fetch-client'
import { getTokenProvider } from './token-provider'
// import { headers } from 'next/headers'

// Client-side: Use relative URL or same host to avoid CORS issues with subdomains
const getClientApiUrl = (): string => {
  if (typeof window === 'undefined') return ''
  // Use same host as current page to avoid cross-site issues
  // If on volkswagen.localhost:9090, API will be http://volkswagen.localhost:9090
  return `${window.location.protocol}//${window.location.host}`
}

// Server-side: Use configured backend URL
const SERVER_API_URL =
  process.env.NEXT_SERVER_API_URL || process.env.NEXT_PUBLIC_API_URL || 'http://localhost:9090'

function extractTenantFromHostname(): string | null {
  if (globalThis.window === undefined) return null

  const hostname = globalThis.location.hostname
  const subdomain = hostname.split('.')[0]

  if (!subdomain || subdomain === 'localhost') return null

  const isNumericSubdomain = /^\d+$/.exec(subdomain)
  if (!isNumericSubdomain) {
    return subdomain
  }

  return null
}

/**
 * Extract tenant from server-side request headers or hostname
 */
async function extractTenantFromServer(): Promise<string | null> {
  try {
    const headersList = await (await import('next/headers')).headers()
    const host = headersList.get('host') || ''

    // Extract subdomain from host header
    const subdomain = host.split('.')[0]

    // Skip localhost and ports - no tenant available
    if (!subdomain || subdomain === 'localhost' || subdomain.includes(':')) {
      return null
    }

    // Skip numeric subdomains (like IP addresses)
    const isNumericSubdomain = /^\d+$/.exec(subdomain)
    if (isNumericSubdomain) {
      return null
    }

    return subdomain
  } catch {
    return null
  }
}

/**
 * Create HTTP client for server-side usage (SSR, Server Actions, API Routes)
 * Uses configured token provider for dynamic token fetching
 */
export async function createServerHttpClient(): Promise<IHttpClient> {
  const provider = getTokenProvider()
  const headers: Record<string, string> = {}

  const tenant = await extractTenantFromServer()
  if (tenant) {
    headers['X-Tenant'] = tenant
  }

  return new FetchHttpClient(SERVER_API_URL, headers, provider || undefined)
}

/**
 * Create HTTP client for client-side usage (Browser)
 * Uses configured token provider for dynamic token fetching
 * Extracts tenant from subdomain automatically
 * Uses same host as current page to avoid CORS issues
 */
export async function createClientHttpClient(): Promise<IHttpClient> {
  const provider = getTokenProvider()
  const headers: Record<string, string> = {}

  const tenant = extractTenantFromHostname()
  if (tenant) {
    headers['X-Tenant'] = tenant
  }

  // Use same host as current page to prevent cross-site issues
  const apiUrl = getClientApiUrl()
  return new AxiosHttpClient(apiUrl, headers, provider || undefined)
}

/**
 * Get appropriate HTTP client based on environment (SSR vs Browser)
 */
export async function getHttpClient(): Promise<IHttpClient> {
  if (globalThis.window === undefined) {
    return await createServerHttpClient()
  }
  return await createClientHttpClient()
}

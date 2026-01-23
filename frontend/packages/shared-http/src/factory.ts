import type { IHttpClient } from './types'
import { AxiosHttpClient } from './axios-client'
import { FetchHttpClient } from './fetch-client'
import { getTokenProvider } from './token-provider'
import { extractTenantFromHostname, getTenantFromWindow } from './tenant-utils'

// Client-side: Use base domain to go through Nginx reverse proxy
const getClientApiUrl = (): string => {
  if (typeof window === 'undefined') return ''

  const hostname = window.location.hostname
  const protocol = window.location.protocol

  if (hostname.includes('.app.local')) {
    return `${protocol}//app.local`
  }

  return `${protocol}//${hostname}`
}

/**
 * Build server API URL - simplified without dynamic imports
 */
function getServerApiUrl(): string {
  // Always use localhost in development to reach Docker services
  return 'http://localhost'
}

/**
 * Get tenant from auth session - static import to avoid memory leaks
 */
async function getTenantFromAuth(): Promise<string | null> {
  try {
    // Static import - tree-shaken if not used
    const { auth } = await import('@workspace/auth')
    const { isExtendedSession } = await import('@workspace/auth')
    const session = await auth()
    return isExtendedSession(session) ? (session.tenant ?? null) : null
  } catch {
    return null
  }
}

/**
 * Create HTTP client for server-side usage (SSR, Server Actions, API Routes)
 * Gets tenant from auth session
 */
export async function createServerHttpClient(): Promise<IHttpClient> {
  const provider = getTokenProvider()
  const headers: Record<string, string> = {}

  // Get tenant from auth session
  const tenant = await getTenantFromAuth()
  if (tenant) {
    headers['X-Tenant'] = tenant
  }

  const apiUrl = getServerApiUrl()
  return new FetchHttpClient(apiUrl, headers, provider || undefined)
}

/**
 * Create HTTP client for browser usage (Client Components, useEffect, event handlers)
 * Uses AxiosHttpClient with automatic tenant extraction from subdomain
 * Communicates with backend through Nginx reverse proxy
 */
export async function createBrowserHttpClient(): Promise<IHttpClient> {
  const provider = getTokenProvider()
  const headers: Record<string, string> = {}

  const tenant = getTenantFromWindow()
  if (tenant) {
    headers['X-Tenant'] = tenant
  }

  // Use same host as current page to prevent cross-site issues
  const apiUrl = getClientApiUrl()
  return new AxiosHttpClient(apiUrl, headers, provider || undefined)
}

/**
 * Get HTTP client based on environment
 * - Server-side (SSR, Server Actions, API Routes): Returns FetchHttpClient
 * - Client-side (Browser, useEffect): Returns AxiosHttpClient
 */
export async function getHttpClient(): Promise<IHttpClient> {
  if (globalThis.window === undefined) {
    return await createServerHttpClient()
  }
  return await createBrowserHttpClient()
}

import type { IHttpClient } from './types'
import { AxiosHttpClient } from './axios-client'
import { FetchHttpClient } from './fetch-client'
import { extractTenantFromHostname, getTenantFromWindow } from './tenant-utils'

// Client-side: Use base domain to go through Nginx reverse proxy
const getClientApiUrl = (): string => {
  if (globalThis.window === undefined) return ''

  const hostname = globalThis.window.location.hostname
  const protocol = globalThis.window.location.protocol

  if (hostname.includes('.app.local')) {
    return `${protocol}//app.local`
  }

  return `${protocol}//${hostname}`
}

/**
 * Build server API URL - simplified without dynamic imports
 */
function getServerApiUrl(): string {
  const configuredApiUrl =
    process.env.SERVER_API_URL ?? process.env.API_URL ?? process.env.NEXT_PUBLIC_API_URL

  if (configuredApiUrl) {
    try {
      return new URL(configuredApiUrl).origin
    } catch {
      try {
        return new URL(`http://${configuredApiUrl}`).origin
      } catch {
        // fall through to default
      }
    }
  }

  // Default backend gateway for local development
  return 'http://localhost:9090'
}

async function getServerForwardHeaders(): Promise<Record<string, string>> {
  const forwarded: Record<string, string> = {}

  try {
    const { headers } = await import('next/headers')
    const incoming = await headers()

    const cookie = incoming.get('cookie')
    if (cookie) {
      forwarded.Cookie = cookie
    }

    const host = incoming.get('host')
    if (host) {
      const tenant = extractTenantFromHostname(host)
      if (tenant) {
        forwarded['X-Tenant'] = tenant
      }
    }
  } catch {
    // No-op outside Next.js server runtime
  }

  return forwarded
}

/**
 * Create HTTP client for server-side usage (SSR, Server Actions, API Routes)
 * Gets tenant from auth session
 */
export async function createServerHttpClient(): Promise<IHttpClient> {
  const headers = await getServerForwardHeaders()

  const apiUrl = getServerApiUrl()
  return new FetchHttpClient(apiUrl, headers)
}

/**
 * Create HTTP client for browser usage (Client Components, useEffect, event handlers)
 * Uses AxiosHttpClient with automatic tenant extraction from subdomain
 * Communicates with backend through Nginx reverse proxy
 */
export async function createBrowserHttpClient(): Promise<IHttpClient> {
  const headers: Record<string, string> = {}

  const tenant = getTenantFromWindow()
  if (tenant) {
    headers['X-Tenant'] = tenant
  }

  // Use same host as current page to prevent cross-site issues
  const apiUrl = getClientApiUrl()
  return new AxiosHttpClient(apiUrl, headers)
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

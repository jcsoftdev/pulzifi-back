import { AxiosHttpClient } from './axios-client'
import { env } from './env'
import { FetchHttpClient } from './fetch-client'
import { extractTenantFromHostname, getTenantFromWindow } from './tenant-utils'
import type { IHttpClient } from './types'

// Client-side: call the backend directly when NEXT_PUBLIC_API_URL is set,
// bypassing the Next.js rewrite proxy. Falls back to same-origin (proxy) otherwise.
const getClientApiUrl = (): string => {
  if (env.NEXT_PUBLIC_API_URL) {
    return env.NEXT_PUBLIC_API_URL
  }
  if (globalThis.window !== undefined) {
    return globalThis.window.location.origin
  }
  return 'http://localhost:3000'
}

// BFF routes (/api/auth/...) must always go through the Next.js server.
const getBffBaseUrl = (): string => {
  if (globalThis.window !== undefined) {
    return globalThis.window.location.origin
  }
  return 'http://localhost:3000'
}

/**
 * Build server API URL - simplified without dynamic imports
 */
function getServerApiUrl(): string {
  const configuredApiUrl = env.SERVER_API_URL
  console.log('[HTTP Factory] Configured SERVER_API_URL:', configuredApiUrl)
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
  console.warn('[HTTP Factory] Invalid or missing SERVER_API_URL. Defaulting to http://localhost:9090')
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
 * Create HTTP client for Next.js BFF routes (/api/auth/...).
 * Always routes through the Next.js server regardless of NEXT_PUBLIC_API_URL.
 */
export async function createBffHttpClient(): Promise<IHttpClient> {
  const headers: Record<string, string> = {}
  const tenant = getTenantFromWindow()
  if (tenant) {
    headers['X-Tenant'] = tenant
  }
  return new AxiosHttpClient(getBffBaseUrl(), headers)
}

/**
 * Create HTTP client for browser usage (Client Components, useEffect, event handlers)
 * Uses AxiosHttpClient with automatic tenant extraction from subdomain
 * Communicates with backend directly via NEXT_PUBLIC_API_URL
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

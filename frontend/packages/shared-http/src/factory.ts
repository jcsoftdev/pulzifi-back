import { AxiosHttpClient } from './axios-client'
import { env } from './env'
import { FetchHttpClient } from './fetch-client'
import { extractTenantFromHostname, getTenantFromWindow } from './tenant-utils'
import type { IHttpClient } from './types'

// Browser API calls always use the current page origin so cookies are sent as
// same-origin requests. Locally, Caddy on :3000 routes /api/v1/* straight to
// the Go backend. In production the Next.js rewrite proxy handles it.
const getClientApiUrl = (): string => {
  return globalThis.window.location.origin
}

function getServerApiUrl(): string {
  const configuredApiUrl = env.SERVER_API_URL
  if (!configuredApiUrl) {
    throw new Error('SERVER_API_URL is not configured')
  }
  try {
    return new URL(configuredApiUrl).origin
  } catch {
    throw new Error(`SERVER_API_URL is invalid: "${configuredApiUrl}"`)
  }
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
 */
export async function createBffHttpClient(): Promise<IHttpClient> {
  const headers: Record<string, string> = {}
  const tenant = getTenantFromWindow()
  if (tenant) {
    headers['X-Tenant'] = tenant
  }
  return new AxiosHttpClient(getClientApiUrl(), headers)
}

/**
 * Create HTTP client for browser usage (Client Components, useEffect, event handlers).
 * Uses same-origin so cookies are always included without cross-origin restrictions.
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

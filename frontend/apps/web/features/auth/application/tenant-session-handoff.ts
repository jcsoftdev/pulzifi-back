'use client'

import { env } from '@/lib/env'
import type { LoginResponse } from '@workspace/services'

export type HostInfo = {
  hostname: string
  isLocalhost: boolean
  protocol: string
  port: string
  baseDomain: string
}

export function getHostInfo(): HostInfo {
  const hostname = globalThis.location.hostname
  const isLocalhost =
    hostname === 'localhost' || hostname === '127.0.0.1' || hostname.endsWith('.localhost')
  const appBaseUrl = env.NEXT_PUBLIC_APP_BASE_URL
  const base = isLocalhost && appBaseUrl ? new URL(appBaseUrl) : null
  const protocol = base ? base.protocol : globalThis.location.protocol
  const port = base
    ? base.port
    : isLocalhost
      ? globalThis.location.port || '3000'
      : globalThis.location.port
  const appDomain = env.NEXT_PUBLIC_APP_DOMAIN
  let baseDomain = appDomain === 'localhost' && !isLocalhost ? undefined : appDomain
  if (!baseDomain)
    baseDomain = base
      ? base.hostname.split('.').slice(-2).join('.')
      : isLocalhost
        ? 'localhost'
        : hostname.split('.').slice(-2).join('.')
  return {
    hostname,
    isLocalhost,
    protocol,
    port,
    baseDomain,
  }
}

const PENDING_PLAN_COOKIE = 'pz_pending_plan'

/**
 * Persists the pricing-page plan selection in a cross-subdomain cookie so it
 * survives redirects that drop query params (session-expired bounces, manual
 * login, OAuth). Read + cleared by the onboarding wizard.
 */
export function setPendingPlanCookie(plan: string, cycle: string): void {
  const { baseDomain } = getHostInfo()
  const value = encodeURIComponent(`${plan}:${cycle}`)
  document.cookie = `${PENDING_PLAN_COOKIE}=${value}; domain=.${baseDomain}; path=/; max-age=900; SameSite=Lax`
}

export function readPendingPlanCookie(): { plan: string; cycle: string } | null {
  const match = document.cookie.match(new RegExp(`(?:^|; )${PENDING_PLAN_COOKIE}=([^;]+)`))
  if (!match?.[1]) return null
  const [plan, cycle] = decodeURIComponent(match[1]).split(':')
  if (!plan) return null
  return { plan, cycle: cycle || 'monthly' }
}

export function clearPendingPlanCookie(): void {
  const { baseDomain } = getHostInfo()
  document.cookie = `${PENDING_PLAN_COOKIE}=; domain=.${baseDomain}; path=/; max-age=0; SameSite=Lax`
}

/**
 * Performs the cross-subdomain session handoff after a successful login.
 *
 * - If `loginResponse.tenant` is null/undefined → calls `onNoTenant()`.
 * - Otherwise builds the tenant callback URL and navigates, going through
 *   `set-base-session` when currently on a subdomain and a nonce is present.
 */
export function performTenantHandoff(
  loginResponse: Pick<LoginResponse, 'tenant' | 'nonce'>,
  redirectTo: string,
  onNoTenant: () => void
): void {
  const tenant = loginResponse.tenant
  if (!tenant) {
    onNoTenant()
    return
  }

  const { hostname, protocol, port, baseDomain } = getHostInfo()
  const targetHost = `${tenant}.${baseDomain}`
  const portSuffix = port ? `:${port}` : ''
  const tenantCallbackUrl = new URL(`${protocol}//${targetHost}${portSuffix}/api/auth/callback`)
  if (loginResponse.nonce) tenantCallbackUrl.searchParams.set('nonce', loginResponse.nonce)
  tenantCallbackUrl.searchParams.set('redirectTo', redirectTo)

  const isOnSubdomain = hostname !== baseDomain
  if (isOnSubdomain && loginResponse.nonce) {
    const baseSessionUrl = new URL(
      `${protocol}//${baseDomain}${portSuffix}/api/auth/set-base-session`
    )
    baseSessionUrl.searchParams.set('nonce', loginResponse.nonce)
    baseSessionUrl.searchParams.set('tenant', tenant)
    baseSessionUrl.searchParams.set('returnTo', tenantCallbackUrl.toString())
    globalThis.location.href = baseSessionUrl.toString()
  } else {
    globalThis.location.href = tenantCallbackUrl.toString()
  }
}

'use client'

import { AuthApi } from '@workspace/services'
import { useRouter, useSearchParams } from 'next/navigation'
import { Suspense, useEffect, useState } from 'react'
import { LoginForm } from '@/features/auth/ui/login-form'
import { env } from '@/lib/env'

type LoginFormBlockData = {
  blockType: 'login-form'
  headline?: string
  subheadline?: string
}

type Props = {
  block: LoginFormBlockData
}

export function LoginFormBlock({ block }: Props) {
  return (
    <Suspense
      fallback={<main className="flex flex-1 items-center justify-center px-4 pb-12 pt-24" />}
    >
      <LoginFormInner block={block} />
    </Suspense>
  )
}

function LoginFormInner({ block }: Props) {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string>()
  const [infoBanner, setInfoBanner] = useState<string>()

  const headline = block.headline ?? 'Welcome back'
  const subheadline = block.subheadline ?? 'Enter your credentials to continue'

  useEffect(() => {
    const errorParam = searchParams.get('error')
    if (errorParam === 'SessionExpired') setError('Your session has expired. Please sign in again.')
    else if (errorParam === 'PendingApproval')
      setInfoBanner(
        'Your account is pending approval by an administrator. Please check back later.'
      )
    if (searchParams.get('registered') === 'true')
      setInfoBanner('Registration successful! Please wait for admin approval before logging in.')
  }, [
    searchParams,
  ])

  const getHostInfo = () => {
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

  const handleLogin = async (credentials: { email: string; password: string }) => {
    setIsLoading(true)
    setError(undefined)
    try {
      const loginResponse = await AuthApi.login(credentials)
      const tenant = loginResponse.tenant
      if (!tenant) {
        router.push('/')
        router.refresh()
        return
      }
      const { hostname, protocol, port, baseDomain } = getHostInfo()
      const targetHost = `${tenant}.${baseDomain}`
      const redirectTo = searchParams.get('callbackUrl') || '/'
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
    } catch (err: unknown) {
      const axiosError = err as {
        response?: {
          status?: number
          data?: {
            code?: string
          }
        }
      }
      if (axiosError?.response?.status === 403) {
        const code = axiosError.response?.data?.code
        setError(
          code === 'USER_REJECTED'
            ? 'Your account has been rejected. Please contact support.'
            : 'Your account is pending approval by an administrator.'
        )
      } else {
        setError('Invalid email or password')
      }
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <main className="flex flex-1 items-center justify-center px-4 pb-12 pt-24">
      <div className="w-full max-w-[440px] rounded-2xl border border-[var(--pz-card-border)] bg-white p-8 shadow-[var(--pz-card-shadow-rest)]">
        <div className="mb-7">
          <div className="mb-3 inline-flex items-center gap-2 rounded-full bg-[var(--pz-accent)]/10 px-3 py-1">
            <span className="size-1.5 rounded-full bg-[var(--pz-accent)]" />
            <span className="text-xs font-semibold uppercase tracking-widest text-[var(--pz-accent)]">
              Secure sign-in
            </span>
          </div>
          <h1 className="font-heading text-3xl font-bold leading-tight text-[var(--pz-ink)]">
            {headline}
          </h1>
          <p className="mt-1.5 text-sm text-[var(--pz-ink-2)]">{subheadline}</p>
        </div>

        {infoBanner && (
          <div className="mb-5 rounded-xl border border-blue-200 bg-blue-50 p-3 text-sm text-blue-700">
            {infoBanner}
          </div>
        )}

        <LoginForm onSubmit={handleLogin} isLoading={isLoading} error={error} />
      </div>
    </main>
  )
}

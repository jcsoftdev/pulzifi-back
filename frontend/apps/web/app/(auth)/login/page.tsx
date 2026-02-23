'use client'

import { AuthApi } from '@workspace/services'
import { env } from '@/lib/env'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@workspace/ui/components/atoms/card'
import Link from 'next/link'
import { useRouter, useSearchParams } from 'next/navigation'
import { useEffect, useState } from 'react'
import { LoginForm } from '@/features/auth/ui/login-form'

export default function LoginPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string>()
  const [infoBanner, setInfoBanner] = useState<string>()

  useEffect(() => {
    const errorParam = searchParams.get('error')
    if (errorParam === 'SessionExpired') {
      setError('Your session has expired. Please sign in again.')
    } else if (errorParam === 'PendingApproval') {
      setInfoBanner('Your account is pending approval by an administrator. Please check back later.')
    }

    if (searchParams.get('registered') === 'true') {
      setInfoBanner('Registration successful! Please wait for admin approval before logging in.')
    }
  }, [
    searchParams,
  ])

  const getHostInfo = () => {
    const hostname = globalThis.location.hostname
    const isLocalhost = hostname === 'localhost' || hostname === '127.0.0.1' || hostname.endsWith('.localhost')

    // NEXT_PUBLIC_APP_BASE_URL is only for local dev (e.g. behind an HTTPS proxy).
    // In production we always derive protocol/port from the actual browser URL to
    // avoid a stale build-time value sending users to localhost:PORT.
    const appBaseUrl = env.NEXT_PUBLIC_APP_BASE_URL
    const base = (isLocalhost && appBaseUrl) ? new URL(appBaseUrl) : null

    let protocol: string
    if (base) {
      protocol = base.protocol
    } else {
      protocol = globalThis.location.protocol
    }

    let port: string | undefined
    if (base) {
      port = base.port
    } else if (isLocalhost) {
      port = globalThis.location.port || '3000'
    } else {
      port = globalThis.location.port
    }

    const appDomain = env.NEXT_PUBLIC_APP_DOMAIN
    // Ignore NEXT_PUBLIC_APP_DOMAIN=localhost when not actually on localhost
    // (prevents stale build-time value from breaking production redirects)
    let baseDomain = (appDomain === 'localhost' && !isLocalhost) ? undefined : appDomain
    if (!baseDomain) {
      if (base) {
        baseDomain = base.hostname.split('.').slice(-2).join('.')
      } else if (isLocalhost) {
        baseDomain = 'localhost'
      } else {
        baseDomain = hostname.split('.').slice(-2).join('.')
      }
    }

    return { hostname, isLocalhost, protocol, port, baseDomain, base }
  }

  const buildTenantCallbackUrl = (protocol: string, targetHost: string, port?: string, nonce?: string | null, redirectTo = '/') => {
    const portSuffix = port ? `:${port}` : ''
    const tenantCallbackUrl = new URL(`${protocol}//${targetHost}${portSuffix}/api/auth/callback`)
    if (nonce) {
      tenantCallbackUrl.searchParams.set('nonce', nonce)
    }
    tenantCallbackUrl.searchParams.set('redirectTo', redirectTo)
    return tenantCallbackUrl
  }

  const buildBaseSessionUrl = (protocol: string, baseDomain: string, port?: string, nonce?: string | null, tenant?: string, returnTo?: string) => {
    const portSuffix = port ? `:${port}` : ''
    const baseSessionUrl = new URL(`${protocol}//${baseDomain}${portSuffix}/api/auth/set-base-session`)
    if (nonce) baseSessionUrl.searchParams.set('nonce', nonce)
    if (tenant) baseSessionUrl.searchParams.set('tenant', tenant)
    if (returnTo) baseSessionUrl.searchParams.set('returnTo', returnTo)
    return baseSessionUrl
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

      const { hostname, protocol, port, baseDomain, isLocalhost } = getHostInfo()
      const targetHost = `${tenant}.${baseDomain}`
      const redirectTo = searchParams.get('callbackUrl') || '/'

      const tenantCallbackUrl = buildTenantCallbackUrl(protocol, targetHost, port, loginResponse.nonce, redirectTo)

      // When logging in from a tenant subdomain we also need to set cookies at
      // the base domain so the main domain recognises the session.
      // Redirect chain: base/set-base-session → tenant/callback → app
      const isOnSubdomain = hostname !== baseDomain
      if (isOnSubdomain && loginResponse.nonce) {
        const baseSessionUrl = buildBaseSessionUrl(protocol, baseDomain, port, loginResponse.nonce, tenant, tenantCallbackUrl.toString())
        globalThis.location.href = baseSessionUrl.toString()
      } else {
        globalThis.location.href = tenantCallbackUrl.toString()
      }
    } catch (err: unknown) {
      const axiosError = err as {
        response?: {
          status?: number
          data?: {
            error?: string
            code?: string
          }
        }
      }

      if (axiosError?.response?.status === 403) {
        const code = axiosError.response.data?.code
        if (code === 'USER_REJECTED') {
          setError('Your account has been rejected. Please contact support.')
        } else {
          setError('Your account is pending approval by an administrator.')
        }
      } else {
        setError('Invalid email or password')
      }
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-background px-4">
      <div className="w-full max-w-md">
        <Card>
          <CardHeader className="text-center">
            <CardTitle className="text-3xl">Welcome back</CardTitle>
            <CardDescription>Sign in to your account to continue</CardDescription>
          </CardHeader>
          <CardContent>
            {infoBanner && (
              <div className="text-sm bg-blue-50 dark:bg-blue-950 text-blue-700 dark:text-blue-300 p-3 rounded-md border border-blue-200 dark:border-blue-800 mb-4">
                {infoBanner}
              </div>
            )}

            <LoginForm onSubmit={handleLogin} isLoading={isLoading} error={error} />

            <div className="mt-6 text-center text-sm text-muted-foreground">
              Don't have an account?{' '}
              <Link href="/register" className="text-primary hover:underline font-medium">
                Sign up
              </Link>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

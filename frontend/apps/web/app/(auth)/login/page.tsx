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

  // Check for session expired error
  useEffect(() => {
    const errorParam = searchParams.get('error')
    if (errorParam === 'SessionExpired') {
      setError('Your session has expired. Please sign in again.')
    }
  }, [
    searchParams,
  ])

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

      // Cookies are set on the base domain by the BFF, so they're
      // automatically available on all tenant subdomains. Just redirect.
      const protocol = globalThis.location.protocol
      const port = globalThis.location.port
      const hostname = globalThis.location.hostname

      const appDomain = env.NEXT_PUBLIC_APP_DOMAIN
      let baseDomain = appDomain
      if (!baseDomain) {
        if (hostname === 'localhost' || hostname === '127.0.0.1') {
          baseDomain = 'localhost'
        } else if (hostname.endsWith('.localhost')) {
          // e.g. acme.localhost → baseDomain = "localhost"
          baseDomain = 'localhost'
        } else {
          // e.g. acme.pulzifi.com → baseDomain = "pulzifi.com"
          baseDomain = hostname.split('.').slice(-2).join('.')
        }
      }

      const targetHost = `${tenant}.${baseDomain}`
      const portSuffix = port ? `:${port}` : ''
      const redirectTo = searchParams.get('callbackUrl') || '/'

      // Always redirect via the callback route so cookies are set at the
      // correct tenant subdomain origin.
      const callbackUrl = new URL(`${protocol}//${targetHost}${portSuffix}/api/auth/callback`)
      if (loginResponse.nonce) {
        callbackUrl.searchParams.set('nonce', loginResponse.nonce)
      }
      callbackUrl.searchParams.set('redirectTo', redirectTo)
      globalThis.location.href = callbackUrl.toString()
    } catch (error) {
      console.error(error)
      setError('Invalid email or password')
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

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
    }

    if (searchParams.get('registered') === 'true') {
      setInfoBanner('Registration successful! Please wait for admin approval before logging in.')
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

      const protocol = globalThis.location.protocol
      const port = globalThis.location.port
      const hostname = globalThis.location.hostname

      const appDomain = env.NEXT_PUBLIC_APP_DOMAIN
      let baseDomain = appDomain
      if (!baseDomain) {
        if (hostname === 'localhost' || hostname === '127.0.0.1') {
          baseDomain = 'localhost'
        } else if (hostname.endsWith('.localhost')) {
          baseDomain = 'localhost'
        } else {
          baseDomain = hostname.split('.').slice(-2).join('.')
        }
      }

      const targetHost = `${tenant}.${baseDomain}`
      const portSuffix = port ? `:${port}` : ''
      const redirectTo = searchParams.get('callbackUrl') || '/'

      const callbackUrl = new URL(`${protocol}//${targetHost}${portSuffix}/api/auth/callback`)
      if (loginResponse.nonce) {
        callbackUrl.searchParams.set('nonce', loginResponse.nonce)
      }
      callbackUrl.searchParams.set('redirectTo', redirectTo)
      globalThis.location.href = callbackUrl.toString()
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

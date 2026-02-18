'use client'

import { useState, useEffect } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { LoginForm } from '@/features/auth/ui/login-form'
import { AuthApi } from '@workspace/services'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@workspace/ui/components/atoms/card'
import Link from 'next/link'

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

      // Redirigir al tenant correcto
      const protocol = globalThis.location.protocol
      const port = globalThis.location.port
      const hostname = globalThis.location.hostname

      const appDomain = process.env.NEXT_PUBLIC_APP_DOMAIN
      let baseDomain = appDomain

      // Si no hay variable de entorno, intentar inferir (fallback)
      if (!baseDomain) {
        if (hostname === 'localhost' || hostname === '127.0.0.1') {
          baseDomain = 'localhost'
        } else {
          const parts = hostname.split('.')
          baseDomain = parts.slice(-2).join('.')
        }
      }

      // Construir el host destino: tenant.dominioBase
      const targetHost = `${tenant}.${baseDomain}`
      const currentHost = hostname

      // Verificar si ya estamos en el subdominio correcto
      if (currentHost === targetHost) {
        router.push('/')
        router.refresh()
      } else {
        // Construir URL completa
        const portSuffix = port ? `:${port}` : ''
        const tenantUrl = `${protocol}//${targetHost}${portSuffix}`

        globalThis.location.href = tenantUrl
      }
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

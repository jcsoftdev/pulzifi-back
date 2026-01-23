'use client'

import { useState, useEffect } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { signIn, getSession } from 'next-auth/react'
import { LoginForm } from '@/features/auth/ui/login-form'
import type { ExtendedSession } from '@workspace/auth'
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
      console.log('Client: Starting login for', credentials.email)

      // Hacer signIn de NextAuth (que internamente llama a AuthApi.login)
      const result = await signIn('credentials', {
        email: credentials.email,
        password: credentials.password,
        redirect: false,
      })

      console.log('Client: SignIn result:', result)

      if (result?.error) {
        console.error('Client: SignIn error:', result.error)
        setError('Invalid email or password')
        return
      }

      if (!result?.ok) {
        console.error('Client: SignIn not ok')
        setError('Login failed. Please try again.')
        return
      }

      console.log('Client: Login successful, waiting for session...')

      // Esperar un momento para que NextAuth actualice la sesión
      await new Promise((resolve) => setTimeout(resolve, 500))

      // Usar getSession de NextAuth que garantiza obtener la sesión actualizada
      const session = (await getSession()) as ExtendedSession | null

      console.log('Client: Session data:', session)

      const tenant = session?.tenant

      if (!tenant) {
        console.warn('No tenant in session, redirecting to home')
        router.push('/')
        router.refresh()
        return
      }

      console.log('Client: Login successful, redirecting to tenant:', tenant)

      // Redirigir al tenant correcto
      const protocol = window.location.protocol
      const port = window.location.port
      const hostname = window.location.hostname

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
        console.log('Already on correct tenant, redirecting to /')
        router.push('/')
        router.refresh()
      } else {
        // Construir URL completa
        const portSuffix = port ? `:${port}` : ''
        const tenantUrl = `${protocol}//${targetHost}${portSuffix}`

        console.log('Redirecting to:', tenantUrl)
        window.location.href = tenantUrl
      }
    } catch (err) {
      console.error('Login error:', err)
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

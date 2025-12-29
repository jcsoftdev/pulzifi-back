'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { signIn } from 'next-auth/react'
import { LoginForm } from '@/features/auth/ui/login-form'
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
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string>()

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

      // Obtener la sesión para extraer el tenant
      const sessionResponse = await fetch('/api/auth/session')
      const session = await sessionResponse.json()

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

      // Extraer el hostname sin puerto
      const hostname = window.location.hostname

      // Determinar si es localhost o dominio real
      const isLocalhost = hostname === 'localhost' || hostname === '127.0.0.1'

      if (isLocalhost) {
        // En localhost, usar subdominio.localhost:puerto
        // Ejemplo: volkswagen.localhost:3000
        const tenantUrl = `${protocol}//${tenant}.localhost${port ? `:${port}` : ''}`

        // Si ya estamos en el tenant correcto, solo redirigir a home
        if (hostname === `${tenant}.localhost`) {
          console.log('Already on correct tenant, redirecting to /')
          router.push('/')
          router.refresh()
        } else {
          console.log('Redirecting to:', tenantUrl)
          window.location.href = tenantUrl
        }
      } else {
        // En producción, extraer el dominio base
        const parts = hostname.split('.')
        const baseDomain = parts.slice(-2).join('.') // Obtiene 'app.com' de 'volkswagen.app.com'

        // Si ya estamos en el tenant correcto, solo redirigir a home
        if (hostname === `${tenant}.${baseDomain}` || hostname.startsWith(`${tenant}.`)) {
          console.log('Already on correct tenant, redirecting to /')
          router.push('/')
          router.refresh()
        } else {
          // Redirigir al subdominio del tenant
          const tenantUrl = `${protocol}//${tenant}.${baseDomain}`
          console.log('Redirecting to:', tenantUrl)
          window.location.href = tenantUrl
        }
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

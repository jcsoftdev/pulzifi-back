'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { RegisterForm } from '@/features/auth/ui/register-form'
import { AuthApi } from '@workspace/services'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@workspace/ui/components/atoms/card'
import Link from 'next/link'

export default function RegisterPage() {
  const router = useRouter()
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string>()

  const handleRegister = async (data: {
    email: string
    password: string
    firstName: string
    lastName: string
  }) => {
    setIsLoading(true)
    setError(undefined)

    try {
      await AuthApi.register(data)

      // Redirect to login after successful registration
      router.push('/login?registered=true')
    } catch (err: unknown) {
      const error = err as {
        response?: {
          data?: {
            error?: string
          }
        }
        message?: string
      }
      const errorMessage =
        error?.response?.data?.error || error?.message || 'Registration failed. Please try again.'
      setError(errorMessage)
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-background px-4">
      <div className="w-full max-w-md">
        <Card>
          <CardHeader className="text-center">
            <CardTitle className="text-3xl">Create an account</CardTitle>
            <CardDescription>Sign up to get started with Pulzifi</CardDescription>
          </CardHeader>
          <CardContent>
            <RegisterForm onSubmit={handleRegister} isLoading={isLoading} error={error} />

            <div className="mt-6 text-center text-sm text-muted-foreground">
              Already have an account?{' '}
              <Link href="/login" className="text-primary hover:underline font-medium">
                Sign in
              </Link>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

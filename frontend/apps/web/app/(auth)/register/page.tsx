'use client'

import { AuthApi } from '@workspace/services'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@workspace/ui/components/atoms/card'
import Link from 'next/link'
import { useState } from 'react'
import type { RegisterData } from '@/features/auth/domain/types'
import { RegisterForm } from '@/features/auth/ui/register-form'

export default function RegisterPage() {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string>()
  const [submitted, setSubmitted] = useState(false)

  const handleRegister = async (data: RegisterData) => {
    setIsLoading(true)
    setError(undefined)

    try {
      await AuthApi.register(data)
      setSubmitted(true)
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

  if (submitted) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background px-4">
        <div className="w-full max-w-md">
          <Card>
            <CardHeader className="text-center">
              <CardTitle className="text-3xl">Registration submitted!</CardTitle>
              <CardDescription>
                Your account is pending approval by an administrator. You will be able to log in once
                your account has been approved.
              </CardDescription>
            </CardHeader>
            <CardContent className="text-center">
              <Link href="/login" className="text-primary hover:underline font-medium">
                Back to login
              </Link>
            </CardContent>
          </Card>
        </div>
      </div>
    )
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

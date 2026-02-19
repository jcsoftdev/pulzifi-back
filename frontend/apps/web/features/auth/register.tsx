'use client'

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@workspace/ui/components/atoms/card'
import Link from 'next/link'
import { useRegister } from './application/use-register'
import { RegisterForm } from './ui/register-form'

export function RegisterFeature() {
  const { register, isLoading, error, submitted, checkSubdomain, subdomainStatus, subdomainMessage } = useRegister()

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
            <RegisterForm
              onSubmit={register}
              isLoading={isLoading}
              error={error}
              onSubdomainChange={checkSubdomain}
              subdomainStatus={subdomainStatus}
              subdomainMessage={subdomainMessage}
            />

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

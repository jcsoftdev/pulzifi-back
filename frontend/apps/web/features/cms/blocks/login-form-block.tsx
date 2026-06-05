'use client'

import { AuthApi } from '@workspace/services'
import { useRouter, useSearchParams } from 'next/navigation'
import { Suspense, useEffect, useState } from 'react'
import { performTenantHandoff } from '@/features/auth/application/tenant-session-handoff'
import { LoginForm } from '@/features/auth/ui/login-form'

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

  const headline = block.headline ?? 'Welcome back'
  const subheadline = block.subheadline ?? 'Enter your credentials to continue'

  useEffect(() => {
    const errorParam = searchParams.get('error')
    if (errorParam === 'SessionExpired') setError('Your session has expired. Please sign in again.')
  }, [
    searchParams,
  ])

  const handleLogin = async (credentials: { email: string; password: string }) => {
    setIsLoading(true)
    setError(undefined)
    try {
      const loginResponse = await AuthApi.login(credentials)
      const redirectTo = searchParams.get('callbackUrl') || '/'
      performTenantHandoff(loginResponse, redirectTo, () => {
        router.push('/')
        router.refresh()
      })
    } catch {
      setError('Invalid email or password')
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

        <LoginForm onSubmit={handleLogin} isLoading={isLoading} error={error} />
      </div>
    </main>
  )
}

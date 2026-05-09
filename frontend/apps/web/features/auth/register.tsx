'use client'

import { CheckCircle2 } from 'lucide-react'
import Link from 'next/link'
import { Navbar } from '@/features/landing'
import { useRegister } from './application/use-register'
import { RegisterForm } from './ui/register-form'

export function RegisterFeature() {
  const {
    register,
    isLoading,
    error,
    submitted,
    checkSubdomain,
    subdomainStatus,
    subdomainMessage,
  } = useRegister()

  if (submitted) {
    return (
      <div className="flex min-h-screen flex-col bg-[var(--pz-page-bg)]">
        <div className="mx-auto w-full max-w-[1280px] px-3 pt-3">
          <Navbar />
        </div>
        <div className="h-28 shrink-0" />
        <main className="flex flex-1 items-center justify-center px-4 py-8">
          <div className="w-full max-w-[480px] rounded-2xl border border-[var(--pz-card-border)] bg-white p-10 text-center shadow-[var(--pz-card-shadow-rest)]">
            <div className="mx-auto mb-6 flex size-14 items-center justify-center rounded-full bg-[var(--pz-accent)]/10">
              <CheckCircle2 className="size-7 text-[var(--pz-accent)]" />
            </div>
            <h1 className="font-heading text-3xl font-bold leading-tight text-[var(--pz-ink)]">
              Registration submitted!
            </h1>
            <p className="mt-3 text-base leading-6 text-[var(--pz-ink-2)]">
              Your account is pending approval by an administrator. You will be able to log in once
              your account has been approved.
            </p>
            <Link
              href="/login"
              className="mt-8 inline-flex h-12 items-center rounded-full bg-[var(--pz-accent)] px-8 text-sm font-medium text-white shadow-[var(--pz-shadow-accent)] transition-[opacity,transform] hover:opacity-90 hover:scale-[1.02] active:scale-[0.98]"
            >
              Back to login
            </Link>
          </div>
        </main>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen flex-col bg-[var(--pz-page-bg)]">
      <div className="mx-auto w-full max-w-[1280px] px-3 pt-3">
        <Navbar />
      </div>
      <div className="h-20 shrink-0" />
      <main className="flex flex-1 items-center justify-center px-4 py-6">
        <div className="w-full max-w-[580px] rounded-2xl border border-[var(--pz-card-border)] bg-white p-6 shadow-[var(--pz-card-shadow-rest)] md:p-8">
          <div className="mb-4">
            <div className="mb-2 inline-flex items-center gap-2 rounded-full bg-[var(--pz-accent)]/10 px-3 py-1">
              <span className="size-1.5 rounded-full bg-[var(--pz-accent)]" />
              <span className="text-xs font-semibold uppercase tracking-widest text-[var(--pz-accent)]">Free 14-day trial</span>
            </div>
            <h1 className="font-heading text-2xl font-bold leading-tight text-[var(--pz-ink)]">
              Create your account
            </h1>
            <p className="mt-1 text-sm text-[var(--pz-ink-2)]">No credit card required</p>
          </div>

          <RegisterForm
            onSubmit={register}
            isLoading={isLoading}
            error={error}
            onSubdomainChange={checkSubdomain}
            subdomainStatus={subdomainStatus}
            subdomainMessage={subdomainMessage}
          />
        </div>
      </main>

    </div>
  )
}

'use client'

import { useRegister } from '@/features/auth/application/use-register'
import { RegisterForm } from '@/features/auth/ui/register-form'

type RegisterFormBlockData = {
  blockType: 'register-form'
  headline?: string
  subheadline?: string
  trialBadge?: string
}

type Props = {
  block: RegisterFormBlockData
}

export function RegisterFormBlock({ block }: Props) {
  const {
    register,
    isLoading,
    error,
    checkSubdomain,
    subdomainStatus,
    subdomainMessage,
  } = useRegister()

  const headline = block.headline ?? 'Create your account'
  const subheadline = block.subheadline ?? 'No credit card required'
  const trialBadge = block.trialBadge ?? 'Free 14-day trial'

  return (
    <main className="flex flex-1 items-center justify-center px-4 pb-12 pt-24">
      <div className="w-full max-w-[520px] rounded-2xl border border-[var(--pz-card-border)] bg-white p-8 shadow-[var(--pz-card-shadow-rest)] md:p-10">
        <div className="mb-8">
          <div className="mb-3 inline-flex items-center gap-2 rounded-full bg-[var(--pz-accent)]/10 px-3 py-1">
            <span className="size-1.5 rounded-full bg-[var(--pz-accent)]" />
            <span className="text-xs font-semibold uppercase tracking-widest text-[var(--pz-accent)]">
              {trialBadge}
            </span>
          </div>
          <h1 className="font-heading text-3xl font-bold leading-tight text-[var(--pz-ink)]">
            {headline}
          </h1>
          <p className="mt-1.5 text-sm text-[var(--pz-ink-2)]">{subheadline}</p>
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
  )
}

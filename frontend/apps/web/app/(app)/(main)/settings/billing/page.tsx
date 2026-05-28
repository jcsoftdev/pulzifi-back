import { Suspense } from 'react'
import { BillingTab } from '@/features/billing'

export const metadata = {
  title: 'Billing',
}

/**
 * /settings/billing — Stripe subscription management.
 * Auth is guaranteed by the (main) layout's AuthGuard.
 * BillingTab is a client component (hooks for subscription + checkout).
 * Wrapped in Suspense because BillingTab calls useSearchParams() (reads the
 * ?promo= apply-link param) — Next.js requires a boundary for that hook.
 */
export default function BillingPage() {
  return (
    <Suspense>
      <BillingTab />
    </Suspense>
  )
}

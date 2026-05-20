import { BillingTab } from '@/features/billing'

export const metadata = {
  title: 'Billing',
}

/**
 * /settings/billing — Stripe subscription management.
 * Auth is guaranteed by the (main) layout's AuthGuard.
 * BillingTab is a client component (hooks for subscription + checkout).
 */
export default function BillingPage() {
  return <BillingTab />
}

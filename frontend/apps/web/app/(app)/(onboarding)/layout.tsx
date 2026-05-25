export const dynamic = 'force-dynamic'

import { AuthApi } from '@workspace/services'
import { UnauthorizedError } from '@workspace/shared-http'
import { redirect } from 'next/navigation'

/**
 * Onboarding layout — requires auth but does NOT require a tenant.
 *
 * Users with status='approved' and no org land here after OAuth.
 * Kept outside (main) to avoid the AuthGuard tenant-check loop.
 */
export default async function OnboardingLayout({ children }: { children: React.ReactNode }) {
  try {
    await AuthApi.getCurrentUser()
  } catch (error: unknown) {
    // Re-throw Next.js internal errors (redirect, notFound) — they must propagate
    if (error && typeof error === 'object' && 'digest' in error) {
      throw error
    }
    if (error instanceof UnauthorizedError) {
      redirect('/login')
    }
    // Any other error — treat as unauthenticated
    redirect('/login')
  }

  return <>{children}</>
}

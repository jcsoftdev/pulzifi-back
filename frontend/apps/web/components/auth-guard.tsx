import { AuthApi } from '@workspace/services'
import { UnauthorizedError } from '@workspace/shared-http'
import { redirect } from 'next/navigation'
import type { ReactNode } from 'react'
import { SessionRefresher } from './session-refresher'

interface AuthGuardProps {
  children: ReactNode
}

/**
 * Auth Guard - Protects authenticated routes
 *
 * - Checks if user session is valid
 * - On expired access token, renders SessionRefresher (client-side refresh)
 * - Works at layout level for all protected routes
 * - Approved users with no org (tenant === null) are redirected to /onboarding
 */
export async function AuthGuard({ children }: AuthGuardProps) {
  try {
    const user = await AuthApi.getCurrentUser()
    // Defense in depth: users without an org must complete onboarding.
    // AuthGuard only runs inside (main) route group; (onboarding) has its own layout,
    // so this redirect can never loop.
    if (user.tenant == null) {
      redirect('/onboarding')
    }
    // Users who haven't answered the onboarding questions go to the wizard (step 2).
    if (!user.onboardingCompleted) {
      redirect('/onboarding')
    }
  } catch (error) {
    // Re-throw Next.js internal errors (redirect, notFound) — they must propagate
    if (error && typeof error === 'object' && 'digest' in error) {
      throw error
    }
    if (error instanceof UnauthorizedError) {
      return <SessionRefresher />
    }
    console.error('[AuthGuard] Unexpected error, redirecting to /login:', error)
    redirect('/login')
  }

  return children
}

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
 */
export async function AuthGuard({ children }: AuthGuardProps) {
  try {
    const user = await AuthApi.getCurrentUser()
    if (user.status && user.status !== 'approved') {
      redirect('/login?error=PendingApproval')
    }
  } catch (error) {
    // Re-throw Next.js internal errors (redirect, notFound) — they must propagate
    if (error && typeof error === 'object' && 'digest' in error) {
      throw error
    }
    if (error instanceof UnauthorizedError) {
      // Don't redirect to login — the browser still has a valid refresh_token
      // cookie. Render a client component that refreshes the session and then
      // triggers a full page reload with fresh cookies.
      return <SessionRefresher />
    }
    redirect('/login')
  }

  return children
}

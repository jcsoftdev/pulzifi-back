import { AuthApi } from '@workspace/services'
import { UnauthorizedError } from '@workspace/shared-http'
import { redirect } from 'next/navigation'
import type { ReactNode } from 'react'

interface AuthGuardProps {
  children: ReactNode
}

/**
 * Auth Guard - Protects authenticated routes
 *
 * - Checks if user session is valid
 * - Redirects to base domain login if session expired
 * - Works at layout level for all protected routes
 */
export async function AuthGuard({ children }: AuthGuardProps) {
  try {
    const user = await AuthApi.getCurrentUser()
    if (user.status && user.status !== 'approved') {
      redirect('/login?error=PendingApproval')
    }
  } catch (error) {
    if (error instanceof UnauthorizedError) {
      redirect('/login')
    }
    redirect('/login')
  }

  return children
}

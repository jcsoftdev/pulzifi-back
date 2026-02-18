import type { ReactNode } from 'react'
import { redirect } from 'next/navigation'
import { AuthApi } from '@workspace/services'
import { UnauthorizedError } from '@workspace/shared-http'

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
    await AuthApi.getCurrentUser()
  } catch (error) {
    if (error instanceof UnauthorizedError) {
      redirect('/login')
    }
    redirect('/login')
  }

  return children
}

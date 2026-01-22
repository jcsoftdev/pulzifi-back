import { ReactNode } from 'react'
import { redirect } from 'next/navigation'
import { auth, type ExtendedSession } from '@workspace/auth'

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
  const session = (await auth()) as ExtendedSession | null

  // If no session or refresh token error, redirect to login
  if (!session || session.error === 'RefreshAccessTokenError') {
    // Redirect to base domain login (NextAuth will clear the cookie automatically)
    const protocol = process.env.NODE_ENV === 'production' ? 'https' : 'http'
    const hostname = process.env.NEXT_PUBLIC_APP_DOMAIN || 'app.local'
    const port = process.env.PORT ? `:${process.env.PORT}` : ''

    const loginUrl = `${protocol}://${hostname}${port}/login`
    redirect(loginUrl)
  }

  return children
}

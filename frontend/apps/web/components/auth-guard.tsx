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
    // Redirect to login (AuthGuard runs on server, so we can't easily get current host without headers)
    // But for subdomains, we want to redirect to the login page of the CURRENT domain if possible,
    // or fallback to the base domain.
    // Since this is a server component, we rely on the hardcoded base domain for safety.
    // If you want to support subdomain login redirects, you'd need to pass headers() here.

    // For now, let's just redirect to /login relative to the current root,
    // but redirect() requires an absolute URL or path.
    // Using a relative path '/login' will keep the current domain!
    redirect('/login')
  }

  return children
}

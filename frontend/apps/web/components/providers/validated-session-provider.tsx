'use client'

import { SessionProvider as NextAuthSessionProvider } from 'next-auth/react'
import { useSessionValidator } from '@/hooks/use-session-validator'
import type { ReactNode } from 'react'

interface SessionProviderProps {
  children: ReactNode
}

/**
 * Wrapper component that validates the session on client side
 * This is used internally by ValidatedSessionProvider
 */
function SessionValidator({ children }: Readonly<{ children: ReactNode }>) {
  useSessionValidator()
  return <>{children}</>
}

/**
 * Enhanced SessionProvider that includes automatic session validation
 * 
 * This provider wraps NextAuth's SessionProvider and adds automatic
 * validation and redirect logic when the refresh token fails.
 * 
 * Usage:
 * ```tsx
 * <ValidatedSessionProvider>
 *   <YourApp />
 * </ValidatedSessionProvider>
 * ```
 */
export function ValidatedSessionProvider({ children }: Readonly<SessionProviderProps>) {
  return (
    <NextAuthSessionProvider>
      <SessionValidator>{children}</SessionValidator>
    </NextAuthSessionProvider>
  )
}

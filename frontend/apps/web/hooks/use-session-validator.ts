'use client'

import { useSession } from 'next-auth/react'
import { useRouter } from 'next/navigation'
import { useEffect, useRef } from 'react'

/**
 * Hook to validate session on client side and redirect to login if there's a refresh error
 * 
 * This hook should be used in client components that require authentication.
 * It will automatically redirect to the login page if the session has a refresh error.
 */
export function useSessionValidator() {
  const { data: session, status, update } = useSession()
  const router = useRouter()
  const hasRedirectedRef = useRef(false)

  useEffect(() => {
    // Only run on client side
    if (typeof window === 'undefined') return

    // Check if session has an error (from failed refresh token)
    if (session?.error === 'RefreshAccessTokenError' && !hasRedirectedRef.current) {
      console.error('[useSessionValidator] Refresh token error detected, redirecting to login', {
        error: session.error,
        sessionStatus: status,
      })
      
      hasRedirectedRef.current = true
      
      // Clear the session and redirect to login
      router.push('/login?error=SessionExpired')
    }
  }, [session, status, router])

  useEffect(() => {
    // Poll session every 30 seconds to check for errors
    const interval = setInterval(async () => {
      if (status === 'authenticated') {
        console.log('[useSessionValidator] Polling session for updates')
        await update()
      }
    }, 30000) // 30 seconds

    return () => clearInterval(interval)
  }, [status, update])

  return {
    session,
    status,
    isLoading: status === 'loading',
    isAuthenticated: status === 'authenticated' && !session?.error,
    hasError: !!session?.error,
  }
}

'use client'

import type { Session } from 'next-auth'
import { useSession } from 'next-auth/react'
import { useEffect } from 'react'

interface SessionWithToken extends Session {
  accessToken?: string
}

/**
 * Hook to sync token from NextAuth session to global variable
 * This token is read by NextAuthTokenProvider in @workspace/auth
 */
export function useAuthTokenSync() {
  const { data: session } = useSession()

  useEffect(() => {
    const token = (session as SessionWithToken)?.accessToken ?? null
    // Sync to global variable used by NextAuthTokenProvider
    ;(globalThis as any).__authToken__ = token
  }, [session])
}

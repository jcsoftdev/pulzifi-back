'use client'

import type { Session } from 'next-auth'
import { useSession } from 'next-auth/react'
import { useEffect } from 'react'

interface SessionWithToken extends Session {
  accessToken?: string
}

// Global variable to store the current token (updated by the hook)
let currentToken: string | null = null

export function getStoredToken(): string | null {
  return currentToken
}

export function setStoredToken(token: string | null): void {
  currentToken = token
}

/**
 * Hook to sync token from NextAuth session to global variable
 * This avoids calling getSession() which makes HTTP requests
 */
export function useAuthTokenSync() {
  const { data: session } = useSession()

  useEffect(() => {
    const token = (session as SessionWithToken)?.accessToken ?? null
    setStoredToken(token)
  }, [
    session,
  ])
}

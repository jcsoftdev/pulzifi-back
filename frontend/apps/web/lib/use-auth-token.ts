/**
 * Client-side token storage
 * This is a simple in-memory storage that syncs with the session
 */

import { useEffect } from 'react'
import { useSession } from 'next-auth/react'
import type { Session } from 'next-auth'

interface SessionWithToken extends Session {
  accessToken?: string
}

let currentToken: string | null = null

export function updateClientToken(token: string | null): void {
  currentToken = token
}

export function getStoredToken(): string | null {
  return currentToken
}

/**
 * Hook that syncs the access token from next-auth session to in-memory storage
 */
export function useAuthTokenSync(): void {
  const { data: session } = useSession()

  useEffect(() => {
    const token = (session as SessionWithToken)?.accessToken ?? null
    updateClientToken(token)
  }, [session])
}

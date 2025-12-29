'use client'

import { createContext, useContext, useEffect, useRef } from 'react'
import type { Session } from 'next-auth'
import { useSession } from 'next-auth/react'
import { updateClientToken } from './token-manager'

interface SessionWithToken extends Session {
  accessToken?: string
}

interface AuthContextValue {
  accessToken: string | null
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const { data: session, status } = useSession()
  const initializedRef = useRef(false)

  // Initialize token immediately (synchronously) on first render
  if (!initializedRef.current && session) {
    const token = (session as SessionWithToken)?.accessToken ?? null
    updateClientToken(token)
    initializedRef.current = true
  }

  useEffect(() => {
    const newToken = (session as SessionWithToken)?.accessToken ?? null
    if (process.env.NODE_ENV === 'development') {
      console.log('[AuthProvider] Session status:', status)
      console.log('[AuthProvider] Token:', newToken ? `${newToken.substring(0, 30)}...` : 'null')
    }
    updateClientToken(newToken)
  }, [
    session,
    status,
  ])

  const accessToken = (session as SessionWithToken)?.accessToken ?? null

  return (
    <AuthContext.Provider
      value={{
        accessToken,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAccessToken(): string | null {
  const context = useContext(AuthContext)
  if (!context) {
    return null
  }
  return context.accessToken
}

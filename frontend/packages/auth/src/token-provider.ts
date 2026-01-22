import type { ITokenProvider } from '@workspace/shared-http'
import { auth } from './auth'
import { ExtendedSession } from './extended-types'

/**
 * NextAuth Token Provider Implementation
 *
 * Server-side: Uses auth() to get session (auth() already has global lock)
 * Client-side: Uses token from global storage (synced by useAuthTokenSync hook in app)
 */
class NextAuthTokenProvider implements ITokenProvider {
  async getServerToken(): Promise<string | null> {
    try {
      const session = await auth() as ExtendedSession | null
      
      // If session has error, it means refresh failed
      if (session?.error === 'RefreshAccessTokenError') {
        return null
      }
      
      return session?.accessToken ?? null
    } catch (error) {
      console.error('[NextAuthTokenProvider] Session error:', error)
      return null
    }
  }

  async getClientToken(): Promise<string | null> {
    // Get token from global variable (must be set by app using useAuthTokenSync)
    return (globalThis as any).__authToken__ ?? null
  }
}

export const tokenProvider = new NextAuthTokenProvider()
export type { ITokenProvider }

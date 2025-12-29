import type { ITokenProvider } from '@workspace/shared-http'
import { auth } from './auth'
import { getStoredToken } from './use-auth-token'

/**
 * NextAuth Token Provider Implementation
 *
 * Server-side: Uses auth() to get session
 * Client-side: Uses token from useSession hook (no HTTP calls)
 */
class NextAuthTokenProvider implements ITokenProvider {
  async getServerToken(): Promise<string | null> {
    try {
      const session = await auth()
      return session?.accessToken ?? null
    } catch (error) {
      console.error('[NextAuthTokenProvider] Server error:', error)
      return null
    }
  }

  async getClientToken(): Promise<string | null> {
    // Get token from global variable (synced by useAuthTokenSync hook)
    return getStoredToken()
  }
}

export const tokenProvider = new NextAuthTokenProvider()

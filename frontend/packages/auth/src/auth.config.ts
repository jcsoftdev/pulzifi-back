import Credentials from 'next-auth/providers/credentials'
import type { User, Session } from 'next-auth'
import type { JWT } from 'next-auth/jwt'
import { AuthApi } from '@workspace/services'
import type { ExtendedJWT, ExtendedUser, ExtendedSession } from './extended-types'

/**
 * Refreshes the access token using the refresh token
 */
let refreshCount = 0
async function refreshAccessToken(token: ExtendedJWT): Promise<JWT> {
  try {
    refreshCount++
    const currentRefreshId = refreshCount
    
    if (!token.refreshToken) {
      throw new Error('No refresh token available')
    }

    const refreshToken = token.refreshToken
    const refreshTokenStr = String(refreshToken)
    const tokenPreview = `${refreshTokenStr.substring(0, 10)}...${refreshTokenStr.substring(refreshTokenStr.length - 10)}`
    console.log(`[Refresh #${currentRefreshId}] Starting token refresh with token: ${tokenPreview}`)

    const response = await AuthApi.refreshToken(
      refreshToken,
      token.tenant
    )
    
    const newTokenPreview = `${response.refreshToken.substring(0, 10)}...${response.refreshToken.substring(response.refreshToken.length - 10)}`
    console.log(`[Refresh #${currentRefreshId}] Token refresh successful, new token: ${newTokenPreview}`)
    return {
      ...token,
      accessToken: response.accessToken,
      refreshToken: response.refreshToken,
      accessTokenExpires: Date.now() + response.expiresIn * 1000,
      tenant: response.tenant || token.tenant, // Preserve tenant from response or original token
    } as JWT
  } catch (error) {
    console.error(`[Refresh #${refreshCount}] Token refresh FAILED:`, error)

    // When refresh fails (401 = revoked token), clear everything
    // This will trigger a logout on next auth() call
    return {
      error: 'RefreshAccessTokenError',
    } as JWT
  }
}

const cookieDomain = process.env.NODE_ENV === 'production' ? '.pulzifi.com' : undefined

const authConfig = {
  secret: process.env.AUTH_SECRET || process.env.NEXTAUTH_SECRET,
  trustHost: true,
  
  providers: [
    Credentials({
      credentials: {
        email: {
          label: 'Email',
          type: 'email',
        },
        password: {
          label: 'Password',
          type: 'password',
        },
      },
      authorize: async (credentials) => {
        const email = credentials?.email as string | undefined
        const password = credentials?.password as string | undefined

        if (!email || !password) {
          return null
        }

        try {
          const response = await AuthApi.login({
            email,
            password,
          })

          if (!response.accessToken) {
            return null
          }

          const user = {
            id: email,
            email,
            name: email.split('@')[0],
            accessToken: response.accessToken,
            refreshToken: response.refreshToken,
            tenant: response.tenant,
          } as User

          return user
        } catch (error) {
          console.error('[Auth] Login error:', error)
          return null
        }
      },
    }),
  ],
  callbacks: {
    authorized({ auth }: { auth: Session | null }) {
      return !!auth?.user
    },
    async jwt({ token, user }: { token: JWT; user: User }): Promise<JWT> {
      // Initial sign in
      if (user && 'accessToken' in user) {
        const extendedUser = user as ExtendedUser
        return {
          ...token,
          accessToken: extendedUser.accessToken,
          refreshToken: extendedUser.refreshToken,
          accessTokenExpires: Date.now() + 5 * 1000, // 5 seconds for testing
          id: extendedUser.id,
          email: extendedUser.email,
          name: extendedUser.name,
          tenant: extendedUser.tenant,
        } as JWT
      }

      // If there's already an error, don't try to refresh again
      if (token.error) {
        return token
      }

      // Return previous token if the access token has not expired yet
      if (Date.now() < (token.accessTokenExpires as number)) {
        return token
      }

      // Access token has expired, try to refresh it
      return refreshAccessToken(token as ExtendedJWT)
    },
    async session({ session, token }: { session: Session; token: JWT }): Promise<Session> {
      const extSession = session as ExtendedSession
      const extToken = token as ExtendedJWT
      
      // Ensure user object exists
      extSession.user ??= {
        id: extToken.id as string,
        email: extToken.email as string,
        name: extToken.name as string,
      }
      
      // Always preserve tenant, even if token is expired/errored
      if (extToken.tenant) {
        extSession.tenant = extToken.tenant
      }
      
      if (extToken.accessToken) {
        extSession.user.id = extToken.id as string
        extSession.accessToken = extToken.accessToken
      }

      // If there was an error during token refresh, propagate it
      if (extToken.error) {
        extSession.error = extToken.error
      }

      return extSession as Session
    },
  },
  pages: {
    signIn: '/login',
  },
  session: {
    strategy: 'jwt' as const,
    maxAge: 60 * 60, // 1 hour
  },
  cookies: {
    sessionToken: {
      name: 'authjs.session-token',
      options: {
        httpOnly: true,
        secure: process.env.NODE_ENV === 'production',
        sameSite: 'lax' as const,
        path: '/',
        domain: cookieDomain,
      },
    },
    callbackUrl: {
      name: 'authjs.callback-url',
      options: {
        sameSite: 'lax' as const,
        path: '/',
        domain: cookieDomain,
      },
    },
    csrfToken: {
      name: 'authjs.csrf-token',
      options: {
        httpOnly: true,
        sameSite: 'lax' as const,
        path: '/',
        domain: cookieDomain,
      },
    },
    pkceCodeVerifier: {
      name: 'authjs.pkce.code_verifier',
      options: {
        httpOnly: true,
        sameSite: 'lax' as const,
        path: '/',
        domain: cookieDomain,
      },
    },
  },
  debug: process.env.NODE_ENV === 'development',
}

export default authConfig

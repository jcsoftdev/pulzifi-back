import Credentials from 'next-auth/providers/credentials'
import { AuthApi } from '@workspace/services'
import type { User, Session } from 'next-auth'
import type { JWT } from 'next-auth/jwt'

/**
 * Refreshes the access token using the refresh token
 */
async function refreshAccessToken(token: JWT): Promise<JWT> {
  try {
    if (!token.refreshToken) {
      throw new Error('No refresh token available')
    }

    const response = await AuthApi.refreshToken(token.refreshToken as string)

    return {
      ...token,
      accessToken: response.accessToken,
      refreshToken: response.refreshToken ?? token.refreshToken,
      accessTokenExpires: Date.now() + response.expiresIn * 1000,
    }
  } catch (error) {
    console.error('[Auth] Error refreshing access token:', error)

    return {
      ...token,
      error: 'RefreshAccessTokenError',
    }
  }
}

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
          const loginResponse = await AuthApi.login({
            email,
            password,
          })

          if (!loginResponse.accessToken) {
            return null
          }

          return {
            id: email,
            email,
            name: email.split('@')[0],
            accessToken: loginResponse.accessToken,
            refreshToken: loginResponse.refreshToken,
            tenant: loginResponse.tenant,
          } as User
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
    async jwt({ token, user }: { token: JWT; user: User }) {
      // Initial sign in
      if (user?.accessToken) {
        return {
          accessToken: user.accessToken,
          refreshToken: user.refreshToken,
          accessTokenExpires: Date.now() + 15 * 60 * 1000, // 15 minutes
          id: user.id,
          tenant: user.tenant,
        }
      }

      // Return previous token if the access token has not expired yet
      if (Date.now() < (token.accessTokenExpires as number)) {
        return token
      }

      // Access token has expired, try to refresh it
      return refreshAccessToken(token)
    },
    async session({ session, token }: { session: Session; token: JWT }) {
      if (token?.accessToken) {
        session.user.id = token.id as string
        session.accessToken = token.accessToken
        session.tenant = token.tenant as string
      }

      // If there was an error during token refresh, propagate it
      if (token.error) {
        session.error = token.error as string
      }

      return session
    },
  },
  pages: {
    signIn: '/login',
  },
  session: {
    strategy: 'jwt' as const,
    maxAge: 60 * 60, // 1 hour
  },
  debug: process.env.NODE_ENV === 'development',
}

export default authConfig

import Credentials from 'next-auth/providers/credentials'
import { AuthApi } from '@workspace/services'
import type { User, Session } from 'next-auth'
import type { JWT } from 'next-auth/jwt'

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
      if (user?.accessToken) {
        token.accessToken = user.accessToken
        token.id = user.id
        token.tenant = user.tenant
      }
      return token
    },
    async session({ session, token }: { session: Session; token: JWT }) {
      if (token?.accessToken) {
        session.user.id = token.id as string
        // Expose accessToken to client (needed for API calls)
        // This is safe because we're sending it as Bearer token anyway
        session.accessToken = token.accessToken as string
        session.tenant = token.tenant as string
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

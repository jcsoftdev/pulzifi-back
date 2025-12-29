import { DefaultSession } from 'next-auth'

declare module 'next-auth' {
  interface Session {
    accessToken?: string
    refreshToken?: string
    tenant?: string
    user: {
      id: string
    } & DefaultSession['user']
  }

  interface User {
    accessToken?: string
    tenant?: string
  }
}

declare module 'next-auth/jwt' {
  interface JWT {
    accessToken?: string
    id?: string
    tenant?: string
  }
}

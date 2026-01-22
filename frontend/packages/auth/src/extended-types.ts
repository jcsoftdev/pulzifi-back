import type { Session, User } from 'next-auth'
import type { JWT } from 'next-auth/jwt'

/**
 * Extended Session with authentication properties
 */
export interface ExtendedSession extends Session {
  tenant?: string
  accessToken?: string
  error?: string
}

/**
 * Extended User with backend authentication tokens
 */
export interface ExtendedUser extends User {
  accessToken: string
  refreshToken: string
  tenant: string
}

/**
 * Extended JWT with all token properties
 */
export interface ExtendedJWT extends JWT {
  accessToken?: string
  refreshToken?: string
  accessTokenExpires?: number
  tenant?: string
  id?: string
  error?: string
}

/**
 * Type guards
 */
export function isExtendedSession(session: Session | null): session is ExtendedSession {
  return session !== null
}

export function isExtendedUser(user: User): user is ExtendedUser {
  return 'accessToken' in user && 'refreshToken' in user
}

export function isExtendedJWT(token: JWT): token is ExtendedJWT {
  return token !== null
}

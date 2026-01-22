import NextAuth from 'next-auth'
import type { NextAuthResult } from 'next-auth'
import authConfig from './auth.config'

const nextAuth = NextAuth(authConfig)

export const handlers: NextAuthResult['handlers'] = nextAuth.handlers
export const signIn: NextAuthResult['signIn'] = nextAuth.signIn
export const signOut: NextAuthResult['signOut'] = nextAuth.signOut

export const auth: NextAuthResult['auth'] = nextAuth.auth

export { default as authConfig } from './auth.config'

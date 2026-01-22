import NextAuth from 'next-auth'
import type { NextAuthResult } from 'next-auth'
import authConfig from './auth.config'

const nextAuth = NextAuth(authConfig)

export const handlers: NextAuthResult['handlers'] = nextAuth.handlers
export const signIn: NextAuthResult['signIn'] = nextAuth.signIn
export const signOut: NextAuthResult['signOut'] = nextAuth.signOut

/**
 * Global lock to prevent concurrent auth() calls that trigger simultaneous refreshes
 * This is critical because in SSR, multiple requests can arrive simultaneously and each
 * would try to refresh the token, causing race conditions and token revocation issues
 */
let authPromise: Promise<any> | null = null
let authCallCount = 0

/**
 * Wrapped auth() with global concurrency lock
 * Ensures only ONE refresh happens at a time across all concurrent requests
 */
export const auth = (async (...args: any[]) => {
  authCallCount++
  const callId = authCallCount
  
  // If an auth call is already in progress, wait for it
  if (authPromise) {
    console.log(`[Auth Lock] Call #${callId} WAITING for existing auth call`)
    const result = await authPromise
    console.log(`[Auth Lock] Call #${callId} RECEIVED result from waiting`)
    return result
  }

  console.log(`[Auth Lock] Call #${callId} STARTING new auth call`)
  // Start new auth call and store the promise globally
  authPromise = (nextAuth.auth as any)(...args)
  
  try {
    const result = await authPromise
    console.log(`[Auth Lock] Call #${callId} COMPLETED successfully`)
    return result
  } catch (error) {
    console.log(`[Auth Lock] Call #${callId} FAILED:`, error)
    throw error
  } finally {
    // Clear the promise after completion (success or error)
    authPromise = null
  }
}) as NextAuthResult['auth']

export { default as authConfig } from './auth.config'

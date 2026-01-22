'use client'

import { useEffect } from 'react'
import { setTokenProvider } from '@workspace/shared-http'
import { tokenProvider } from '@workspace/auth'
import { useAuthTokenSync } from '@/lib/use-auth-token'

/**
 * Client-side initialization component
 * Registers the token provider and syncs token from useSession hook
 */
export function ClientInit() {
  // Sync token from useSession hook (no HTTP calls)
  useAuthTokenSync()

  useEffect(() => {
    // Register token provider on client mount
    setTokenProvider(tokenProvider)
  }, [])

  return null
}

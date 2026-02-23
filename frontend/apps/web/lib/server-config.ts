import { env } from '@workspace/shared-http'

/**
 * Returns the Go backend origin for server-side BFF calls.
 * Throws at request time if SERVER_API_URL is not configured — fails fast
 * rather than silently calling a wrong host.
 */
export function getBackendOrigin(): string {
  const apiBase = env.SERVER_API_URL
  if (!apiBase) {
    throw new Error('SERVER_API_URL is not configured')
  }
  try {
    return new URL(apiBase).origin
  } catch {
    throw new Error(`SERVER_API_URL is invalid: "${apiBase}"`)
  }
}

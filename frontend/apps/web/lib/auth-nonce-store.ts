/**
 * Short-lived in-memory store for cross-subdomain cookie exchange.
 *
 * After login at localhost:3000 the BFF stores the JWT tokens here under a
 * one-time nonce, then the browser redirects to
 *   tenant.localhost:3000/api/auth/callback?nonce=<nonce>
 *
 * The callback route (running in the same Node.js process at the tenant
 * origin) retrieves the tokens, sets HttpOnly cookies scoped to that origin,
 * and redirects to the app root.
 *
 * Entries expire after 30 seconds — they only need to survive the redirect.
 */

interface NonceEntry {
  accessToken: string
  refreshToken: string
  expiresIn: number
  expiresAt: number // Date.now() + TTL
}

const NONCE_TTL_MS = 30_000

// Keyed on Symbol so the Map survives Hot Module Replacement in dev.
const STORE_KEY = Symbol.for('@pulzifi/auth-nonce-store')
const g = globalThis as typeof globalThis & { [STORE_KEY]?: Map<string, NonceEntry> }
if (!g[STORE_KEY]) g[STORE_KEY] = new Map()
const store = g[STORE_KEY]

export function saveNonce(
  nonce: string,
  tokens: { accessToken: string; refreshToken: string; expiresIn: number }
): void {
  // Prune expired entries on every write to avoid unbounded growth.
  const now = Date.now()
  for (const [k, v] of store) {
    if (v.expiresAt < now) store.delete(k)
  }
  store.set(nonce, { ...tokens, expiresAt: now + NONCE_TTL_MS })
}

export function consumeNonce(nonce: string): Omit<NonceEntry, 'expiresAt'> | null {
  const entry = store.get(nonce)
  if (!entry || entry.expiresAt < Date.now()) {
    store.delete(nonce)
    return null
  }
  store.delete(nonce)
  const { expiresAt: _, ...tokens } = entry
  return tokens
}

/**
 * Read nonce tokens WITHOUT consuming the entry.
 * Used by set-base-session so the same nonce can later be consumed by the
 * tenant callback.
 */
export function peekNonce(nonce: string): Omit<NonceEntry, 'expiresAt'> | null {
  const entry = store.get(nonce)
  if (!entry || entry.expiresAt < Date.now()) {
    store.delete(nonce)
    return null
  }
  const { expiresAt: _, ...tokens } = entry
  return tokens
}

/**
 * Token Provider Interface - Strategy Pattern
 *
 * Allows apps to inject their own token retrieval strategy
 * without breaking package independence.
 *
 * Example:
 * - Base implementation: uses cookies
 * - NextAuth implementation: uses getServerSession/getSession
 * - Test implementation: returns mock tokens
 */

export interface ITokenProvider {
  getServerToken(): Promise<string | null>
  getClientToken(): Promise<string | null>
}

/**
 * Global token provider registry (Singleton Pattern)
 * Apps should configure this at startup using setTokenProvider()
 */
class TokenProviderRegistry {
  private provider: ITokenProvider | null = null

  setProvider(provider: ITokenProvider): void {
    this.provider = provider
  }

  getProvider(): ITokenProvider | null {
    return this.provider
  }

  hasProvider(): boolean {
    return this.provider !== null
  }
}

// Singleton instance
const REGISTRY_KEY = Symbol.for('@workspace/shared-http/token-provider-registry')

const globalRegistry = globalThis as unknown as {
  [REGISTRY_KEY]: TokenProviderRegistry
}

if (!globalRegistry[REGISTRY_KEY]) {
  globalRegistry[REGISTRY_KEY] = new TokenProviderRegistry()
}

const registry = globalRegistry[REGISTRY_KEY]

/**
 * Configure the global token provider
 * Call this in your app's initialization (e.g., root layout, _app.tsx)
 */
export function setTokenProvider(provider: ITokenProvider): void {
  registry.setProvider(provider)
}

/**
 * Get the configured token provider
 * Returns null if not configured
 */
export function getTokenProvider(): ITokenProvider | null {
  return registry.getProvider()
}

/**
 * Check if a token provider is configured
 */
export function hasTokenProvider(): boolean {
  return registry.hasProvider()
}

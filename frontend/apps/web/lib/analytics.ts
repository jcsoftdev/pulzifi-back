/**
 * Meta Pixel event tracking.
 *
 * The base pixel is mounted in the root layout and only loads on the apex
 * (public marketing/signup surface), mirroring the GA gate. On tenant
 * subdomains `window.fbq` is undefined, so these calls safely no-op.
 */
export function trackFbEvent(
  event: FbqStandardEvent,
  params?: Record<string, unknown>
): void {
  if (typeof window === 'undefined' || typeof window.fbq !== 'function') return
  window.fbq('track', event, params)
}

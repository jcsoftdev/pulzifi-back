// Shared token refresh state — ensures only one refresh call is in-flight at a
// time across all HTTP clients (Axios interceptor, raw fetch, etc.).

let isRefreshing = false
let refreshSubscribers: ((success: boolean) => void)[] = []
let isRedirectingToLogin = false

function subscribeTokenRefresh(cb: (success: boolean) => void) {
  refreshSubscribers.push(cb)
}

function notifyRefreshSubscribers(success: boolean) {
  refreshSubscribers.forEach((cb) => cb(success))
  refreshSubscribers = []
}

function redirectToLogin(): void {
  if (globalThis.window === undefined) return
  if (isRedirectingToLogin) return
  isRedirectingToLogin = true

  const { protocol, host } = globalThis.window.location
  const hostWithoutPort = host.split(':')[0] ?? host
  const port = host.includes(':') ? `:${host.split(':')[1]}` : ''
  let baseDomainHost = host
  if (hostWithoutPort.endsWith('.localhost')) {
    baseDomainHost = `localhost${port}`
  } else {
    const parts = hostWithoutPort.split('.')
    if (parts.length > 2) baseDomainHost = `${parts.slice(1).join('.')}${port}`
  }
  globalThis.window.location.href = `${protocol}//${baseDomainHost}/login`
}

async function doRefresh(): Promise<boolean> {
  const res = await fetch('/api/auth/refresh', {
    method: 'POST',
    credentials: 'include',
  })
  return res.ok
}

/**
 * Attempt a token refresh and retry the original request.
 * Queues concurrent callers behind a single in-flight refresh.
 * Returns `true` if the refresh succeeded (caller should retry),
 * or redirects to login on failure.
 */
export async function refreshAndRetry(): Promise<boolean> {
  if (isRedirectingToLogin) return false

  if (isRefreshing) {
    return new Promise<boolean>((resolve) => {
      subscribeTokenRefresh((success) => resolve(success))
    })
  }

  isRefreshing = true
  try {
    const ok = await doRefresh()
    if (!ok) throw new Error('Refresh failed')
    notifyRefreshSubscribers(true)
    return true
  } catch {
    notifyRefreshSubscribers(false)
    redirectToLogin()
    return false
  } finally {
    isRefreshing = false
  }
}

export function getIsRedirectingToLogin(): boolean {
  return isRedirectingToLogin
}

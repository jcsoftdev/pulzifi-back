/**
 * Maps an API error to a user-facing message for the super-admin panels.
 * Handles both HttpError (SSR fetch client, `status`) and AxiosError
 * (browser client, `response.status`).
 */
function errorStatus(err: unknown): number | undefined {
  if (typeof err !== 'object' || err === null) return undefined
  const e = err as {
    status?: unknown
    response?: {
      status?: unknown
    }
  }
  if (typeof e.status === 'number') return e.status
  if (typeof e.response?.status === 'number') return e.response.status
  return undefined
}

export function describeLoadError(err: unknown, resource: string): string {
  switch (errorStatus(err)) {
    case 401:
      return 'Your session has expired. Please log in again.'
    case 403:
      return `You need SUPER_ADMIN role to manage ${resource}.`
    default:
      return `Failed to load ${resource}. Please try again.`
  }
}

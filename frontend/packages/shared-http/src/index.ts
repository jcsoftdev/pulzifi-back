export type { IHttpClient, RequestConfig } from './types'
export { UnauthorizedError, HttpError } from './types'
export { AxiosHttpClient } from './axios-client'
export { FetchHttpClient } from './fetch-client'
export {
  createServerHttpClient,
  createBrowserHttpClient,
  getHttpClient,
} from './factory'

// Tenant utilities
export { extractTenantFromHostname, getTenantFromWindow } from './tenant-utils'

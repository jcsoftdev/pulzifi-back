export { AxiosHttpClient } from './axios-client'
export {
  createBrowserHttpClient,
  createServerHttpClient,
  getHttpClient,
} from './factory'
export { FetchHttpClient } from './fetch-client'
// Tenant utilities
export { extractTenantFromHostname, getTenantFromWindow } from './tenant-utils'
export type { IHttpClient, RequestConfig } from './types'
export { HttpError, UnauthorizedError } from './types'

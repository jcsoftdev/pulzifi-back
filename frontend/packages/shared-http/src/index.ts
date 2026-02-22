export { AxiosHttpClient } from './axios-client'
export { env } from './env'
export {
  createBffHttpClient,
  createBrowserHttpClient,
  createServerHttpClient,
  getHttpClient,
} from './factory'
export { FetchHttpClient } from './fetch-client'
// Tenant utilities
export { extractTenantFromHostname, getTenantFromWindow } from './tenant-utils'
export type { HttpResponse, IHttpClient, RequestConfig } from './types'
export { HttpError, UnauthorizedError } from './types'

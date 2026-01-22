export type { IHttpClient, RequestConfig } from './types'
export { UnauthorizedError, HttpError } from './types'
export { AxiosHttpClient } from './axios-client'
export { FetchHttpClient } from './fetch-client'
export {
  setTokenProvider,
  getTokenProvider,
  hasTokenProvider,
  type ITokenProvider,
} from './token-provider'
export {
  createServerHttpClient,
  createBrowserHttpClient,
  getHttpClient,
} from './factory'

// Tenant utilities
export { extractTenantFromHostname, getTenantFromWindow } from './tenant-utils'

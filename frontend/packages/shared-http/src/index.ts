export type { IHttpClient, RequestConfig } from './types'
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
  createClientHttpClient,
  getHttpClient,
} from './factory'

// Re-export for convenience
export { UnauthorizedError } from './fetch-client'

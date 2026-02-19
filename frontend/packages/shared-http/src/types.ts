/**
 * HTTP Client Interface
 * SOLID: Dependency Inversion Principle - Depend on abstractions, not concretions
 */

export interface RequestConfig {
  headers?: Record<string, string>
  params?: Record<string, string>
  cache?: RequestCache
  credentials?: RequestCredentials
  next?: {
    revalidate?: number
    tags?: string[]
  }
  /** When true, returns HttpResponse<T> with data + headers instead of just T */
  withHeaders?: boolean
}

/** Full HTTP response including data, headers and status code */
export interface HttpResponse<T> {
  data: T
  headers: Headers
  status: number
}

export interface IHttpClient {
  get<T>(url: string, config: RequestConfig & { withHeaders: true }): Promise<HttpResponse<T>>
  get<T>(url: string, config?: RequestConfig): Promise<T>

  post<T>(url: string, data: unknown, config: RequestConfig & { withHeaders: true }): Promise<HttpResponse<T>>
  post<T>(url: string, data?: unknown, config?: RequestConfig): Promise<T>

  put<T>(url: string, data: unknown, config: RequestConfig & { withHeaders: true }): Promise<HttpResponse<T>>
  put<T>(url: string, data?: unknown, config?: RequestConfig): Promise<T>

  patch<T>(url: string, data: unknown, config: RequestConfig & { withHeaders: true }): Promise<HttpResponse<T>>
  patch<T>(url: string, data?: unknown, config?: RequestConfig): Promise<T>

  delete<T>(url: string, config: RequestConfig & { withHeaders: true }): Promise<HttpResponse<T>>
  delete<T>(url: string, config?: RequestConfig): Promise<T>
}

/**
 * HTTP Errors
 */
export class UnauthorizedError extends Error {
  constructor() {
    super('Unauthorized')
    this.name = 'UnauthorizedError'
  }
}

export class HttpError extends Error {
  constructor(
    public readonly status: number,
    public readonly statusText: string,
    public readonly path: string,
    message: string
  ) {
    super(message)
    this.name = 'HttpError'
  }
}

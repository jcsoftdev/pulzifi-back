/**
 * HTTP Client Interface
 * SOLID: Dependency Inversion Principle - Depend on abstractions, not concretions
 */

export interface RequestConfig {
  headers?: Record<string, string>
  params?: Record<string, string>
  cache?: RequestCache
  next?: {
    revalidate?: number
    tags?: string[]
  }
}

export interface IHttpClient {
  get<T>(url: string, config?: RequestConfig): Promise<T>
  post<T>(url: string, data?: unknown, config?: RequestConfig): Promise<T>
  put<T>(url: string, data?: unknown, config?: RequestConfig): Promise<T>
  patch<T>(url: string, data?: unknown, config?: RequestConfig): Promise<T>
  delete<T>(url: string, config?: RequestConfig): Promise<T>
}

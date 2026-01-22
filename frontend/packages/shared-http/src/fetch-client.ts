import type { IHttpClient, RequestConfig } from './types'
import type { ITokenProvider } from './token-provider'
import { UnauthorizedError, HttpError } from './types'

export { UnauthorizedError, HttpError }

export class FetchHttpClient implements IHttpClient {
  constructor(
    private readonly baseURL: string,
    private readonly defaultHeaders?: Record<string, string>,
    private readonly tokenProvider?: ITokenProvider
  ) {}

  private debug(message: string, data?: unknown): void {
    // Disabled in favor of console performance - use browser DevTools Network tab instead
    // if (process.env.NODE_ENV === 'development') {
    //   console.debug(`[FetchHttpClient] ${message}`, data)
    // }
  }

  private debugError(message: string, error: unknown): void {
    // Only log actual errors, not debug info
    if (process.env.NODE_ENV === 'development') {
      console.error(`[FetchHttpClient] ${message}`, error)
    }
  }

  private buildUrl(url: string, params?: Record<string, string>): string {
    const fullUrl = new URL(url, this.baseURL)
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        fullUrl.searchParams.append(key, value)
      })
    }
    return fullUrl.toString()
  }

  private async request<T>(url: string, config: RequestInit & RequestConfig = {}): Promise<T> {
    const { params, headers, ...fetchConfig } = config
    const fullUrl = this.buildUrl(url, params)

    const dynamicHeaders: Record<string, string> = {}
    if (this.tokenProvider) {
      const isServer = typeof window === 'undefined'
      const token = isServer 
        ? await this.tokenProvider.getServerToken()
        : await this.tokenProvider.getClientToken()

      if (token) {
        dynamicHeaders.Authorization = `Bearer ${token}`
      }
    }

    const finalHeaders = {
      'Content-Type': 'application/json',
      ...this.defaultHeaders,
      ...dynamicHeaders,
      ...headers,
    }

    const response = await fetch(fullUrl, {
      ...fetchConfig,
      headers: finalHeaders,
    })

    if (response.status === 401) {
      throw new UnauthorizedError()
    }

    if (!response.ok) {
      const contentType = response.headers.get('content-type')
      let errorMessage = `HTTP Error: ${response.status} ${response.statusText}`
      let errorDetails: unknown = null

      if (contentType?.includes('application/json')) {
        try {
          errorDetails = await response.json()
          errorMessage = (errorDetails as any).error || (errorDetails as any).message || errorMessage
        } catch {
          // Ignore JSON parse errors
        }
      } else if (contentType?.includes('text/plain')) {
        try {
          errorMessage = await response.text()
        } catch {
          // Ignore text parsing errors
        }
      }

      // Trim error message to remove trailing newlines and whitespace
      errorMessage = errorMessage.trim()

      this.debugError(`Request failed: ${url}`, {
        status: response.status,
        statusText: response.statusText,
        message: errorMessage,
        details: errorDetails,
      })

      throw new HttpError(response.status, response.statusText, url, errorMessage)
    }

    const contentType = response.headers.get('content-type')
    if (!contentType?.includes('application/json')) {
      const message = `Expected JSON response but got ${contentType}`
      this.debugError(`Content-Type mismatch: ${url}`, { contentType, expected: 'application/json' })
      throw new Error(message)
    }

    const data = await response.json()
    return data
  }

  async get<T>(url: string, config?: RequestConfig): Promise<T> {
    return this.request<T>(url, {
      ...config,
      method: 'GET',
    })
  }

  async post<T>(url: string, data?: unknown, config?: RequestConfig): Promise<T> {
    return this.request<T>(url, {
      ...config,
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async put<T>(url: string, data?: unknown, config?: RequestConfig): Promise<T> {
    return this.request<T>(url, {
      ...config,
      method: 'PUT',
      body: JSON.stringify(data),
    })
  }

  async patch<T>(url: string, data?: unknown, config?: RequestConfig): Promise<T> {
    return this.request<T>(url, {
      ...config,
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  }

  async delete<T>(url: string, config?: RequestConfig): Promise<T> {
    return this.request<T>(url, {
      ...config,
      method: 'DELETE',
    })
  }
}

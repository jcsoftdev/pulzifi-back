import type { IHttpClient, RequestConfig } from './types'
import type { ITokenProvider } from './token-provider'

export class UnauthorizedError extends Error {
  constructor() {
    super('Unauthorized')
    this.name = 'UnauthorizedError'
  }
}

export class FetchHttpClient implements IHttpClient {
  constructor(
    private readonly baseURL: string,
    private readonly defaultHeaders?: Record<string, string>,
    private readonly tokenProvider?: ITokenProvider
  ) {}

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

    const dynamicHeaders: Record<string, string> = {}
    if (this.tokenProvider) {
      const token = await this.tokenProvider.getServerToken()
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

    const response = await fetch(this.buildUrl(url, params), {
      ...fetchConfig,
      headers: finalHeaders,
    })

    if (response.status === 401) {
      throw new UnauthorizedError()
    }

    if (!response.ok) {
      const contentType = response.headers.get('content-type')
      let errorMessage = `HTTP Error: ${response.status} ${response.statusText}`

      if (contentType?.includes('application/json')) {
        try {
          const errorData = await response.json()
          errorMessage = errorData.error || errorData.message || errorMessage
        } catch {
          // Ignore JSON parse errors
        }
      }

      throw new Error(errorMessage)
    }

    const contentType = response.headers.get('content-type')
    if (!contentType?.includes('application/json')) {
      throw new Error(`Expected JSON response but got ${contentType}`)
    }

    return response.json()
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

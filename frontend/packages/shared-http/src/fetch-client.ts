import type { IHttpClient, RequestConfig } from './types'
import { UnauthorizedError, HttpError } from './types'

export { UnauthorizedError, HttpError } from './types'

export class FetchHttpClient implements IHttpClient {
  constructor(
    private readonly baseURL: string,
    private readonly defaultHeaders?: Record<string, string>
  ) {}

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

  private buildHeaders(headers?: HeadersInit): Record<string, string> {
    return {
      'Content-Type': 'application/json',
      ...this.defaultHeaders,
      ...(headers as Record<string, string> | undefined),
    }
  }

  private async parseHttpError(response: Response, url: string): Promise<never> {
    const contentType = response.headers.get('content-type')
    let errorMessage = `HTTP Error: ${response.status} ${response.statusText}`
    let errorDetails: unknown = null

    if (contentType?.includes('application/json')) {
      try {
        errorDetails = await response.json()
        if (typeof errorDetails === 'object' && errorDetails !== null) {
          const maybeError = (errorDetails as { error?: unknown }).error
          const maybeMessage = (errorDetails as { message?: unknown }).message
          const parsedError = typeof maybeError === 'string' ? maybeError : undefined
          const parsedMessage = typeof maybeMessage === 'string' ? maybeMessage : undefined
          errorMessage = parsedError ?? parsedMessage ?? errorMessage
        }
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

    errorMessage = errorMessage.trim()

    this.debugError(`Request failed: ${url}`, {
      status: response.status,
      statusText: response.statusText,
      message: errorMessage,
      details: errorDetails,
    })

    throw new HttpError(response.status, response.statusText, url, errorMessage)
  }

  private async parseJsonResponse<T>(response: Response, url: string): Promise<T> {
    const contentType = response.headers.get('content-type')
    if (!contentType?.includes('application/json')) {
      const message = `Expected JSON response but got ${contentType}`
      this.debugError(`Content-Type mismatch: ${url}`, {
        contentType,
        expected: 'application/json',
      })
      throw new Error(message)
    }

    const data = await response.json()
    return data as T
  }

  private async request<T>(url: string, config: RequestInit & RequestConfig = {}): Promise<T> {
    const { params, headers, ...fetchConfig } = config
    const fullUrl = this.buildUrl(url, params)

    const finalHeaders = this.buildHeaders(headers)

    const response = await fetch(fullUrl, {
      ...fetchConfig,
      credentials: fetchConfig.credentials ?? 'include',
      headers: finalHeaders,
    })

    if (response.status === 401) {
      throw new UnauthorizedError()
    }

    if (!response.ok) {
      return this.parseHttpError(response, url)
    }

    return this.parseJsonResponse<T>(response, url)
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

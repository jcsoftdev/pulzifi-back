import axios, { type AxiosError, type AxiosInstance, type AxiosRequestConfig } from 'axios'
import { env } from './env'
import { getTenantFromWindow } from './tenant-utils'
import type { HttpResponse, IHttpClient, RequestConfig } from './types'

// Shared refresh state — ensures only one refresh call is in-flight at a time
let isRefreshing = false
let refreshSubscribers: ((success: boolean) => void)[] = []

function subscribeTokenRefresh(cb: (success: boolean) => void) {
  refreshSubscribers.push(cb)
}

function notifyRefreshSubscribers(success: boolean) {
  refreshSubscribers.forEach((cb) => cb(success))
  refreshSubscribers = []
}

export class AxiosHttpClient implements IHttpClient {
  private readonly client: AxiosInstance

  constructor(baseURL: string, defaultHeaders?: Record<string, string>) {
    this.client = axios.create({
      baseURL,
      headers: {
        'Content-Type': 'application/json',
        ...defaultHeaders,
      },
      timeout: 30000,
      // Enable credentials to send cookies automatically
      withCredentials: true,
    })

    // Request interceptor: Add tenant from subdomain
    this.client.interceptors.request.use(
      async (config) => {
        // Extract tenant from subdomain (client-side only)
        if (globalThis.window !== undefined) {
          const tenant = getTenantFromWindow()
          if (tenant && config.headers) {
            config.headers['X-Tenant'] = tenant
          }
        }

        return config
      },
      (error) => {
        this.debugError('Request interceptor error', error)
        return Promise.reject(error)
      }
    )

    this.client.interceptors.response.use(
      (response) => response,
      async (error: AxiosError) => {
        const originalRequest = error.config as AxiosRequestConfig & { _retry?: boolean }
        if (error.response?.status === 401 && !originalRequest._retry) {
          return this.handleUnauthorized(error, originalRequest)
        }
        throw error
      }
    )
  }

  private redirectToLogin(): void {
    if (globalThis.window === undefined) return
    const { protocol, host } = globalThis.window.location
    const hostWithoutPort = host.split(':')[0] ?? host
    const port = host.includes(':') ? `:${host.split(':')[1]}` : ''
    let baseDomainHost = host
    if (hostWithoutPort.endsWith('.localhost')) {
      baseDomainHost = `localhost${port}`
    } else {
      const parts = hostWithoutPort.split('.')
      if (parts.length > 2) baseDomainHost = `${parts.slice(1).join('.')}${port}`
    }
    globalThis.window.location.href = `${protocol}//${baseDomainHost}/login`
  }

  private async handleUnauthorized(
    error: AxiosError,
    originalRequest: AxiosRequestConfig & { _retry?: boolean }
  ): Promise<unknown> {
    if (isRefreshing) {
      // Queue this request until the in-flight refresh completes
      return new Promise((resolve, reject) => {
        subscribeTokenRefresh((success) => {
          if (success) resolve(this.client(originalRequest))
          else reject(error)
        })
      })
    }

    originalRequest._retry = true
    isRefreshing = true

    try {
      const refreshResponse = await fetch('/api/auth/refresh', {
        method: 'POST',
        credentials: 'include',
      })
      if (!refreshResponse.ok) throw new Error('Refresh failed')
      notifyRefreshSubscribers(true)
      return this.client(originalRequest)
    } catch {
      notifyRefreshSubscribers(false)
      this.redirectToLogin()
      throw error
    } finally {
      isRefreshing = false
    }
  }

  private debugError(message: string, error: unknown): void {
    // Only log actual errors
    if (env.NODE_ENV === 'development') {
      console.error(`[AxiosHttpClient] ${message}`, error)
    }
  }

  private convertConfig(config?: RequestConfig): AxiosRequestConfig {
    return {
      headers: config?.headers,
      params: config?.params,
    }
  }

  get<T>(url: string, config: RequestConfig & { withHeaders: true }): Promise<HttpResponse<T>>
  get<T>(url: string, config?: RequestConfig): Promise<T>
  async get<T>(url: string, config?: RequestConfig): Promise<T | HttpResponse<T>> {
    const response = await this.client.get<T>(url, this.convertConfig(config))
    if (config?.withHeaders) {
      return { data: response.data, headers: new Headers(response.headers as Record<string, string>), status: response.status }
    }
    return response.data
  }

  post<T>(url: string, data: unknown, config: RequestConfig & { withHeaders: true }): Promise<HttpResponse<T>>
  post<T>(url: string, data?: unknown, config?: RequestConfig): Promise<T>
  async post<T>(url: string, data?: unknown, config?: RequestConfig): Promise<T | HttpResponse<T>> {
    const response = await this.client.post<T>(url, data, this.convertConfig(config))
    if (config?.withHeaders) {
      return { data: response.data, headers: new Headers(response.headers as Record<string, string>), status: response.status }
    }
    return response.data
  }

  put<T>(url: string, data: unknown, config: RequestConfig & { withHeaders: true }): Promise<HttpResponse<T>>
  put<T>(url: string, data?: unknown, config?: RequestConfig): Promise<T>
  async put<T>(url: string, data?: unknown, config?: RequestConfig): Promise<T | HttpResponse<T>> {
    const response = await this.client.put<T>(url, data, this.convertConfig(config))
    if (config?.withHeaders) {
      return { data: response.data, headers: new Headers(response.headers as Record<string, string>), status: response.status }
    }
    return response.data
  }

  patch<T>(url: string, data: unknown, config: RequestConfig & { withHeaders: true }): Promise<HttpResponse<T>>
  patch<T>(url: string, data?: unknown, config?: RequestConfig): Promise<T>
  async patch<T>(url: string, data?: unknown, config?: RequestConfig): Promise<T | HttpResponse<T>> {
    const response = await this.client.patch<T>(url, data, this.convertConfig(config))
    if (config?.withHeaders) {
      return { data: response.data, headers: new Headers(response.headers as Record<string, string>), status: response.status }
    }
    return response.data
  }

  delete<T>(url: string, config: RequestConfig & { withHeaders: true }): Promise<HttpResponse<T>>
  delete<T>(url: string, config?: RequestConfig): Promise<T>
  async delete<T>(url: string, config?: RequestConfig): Promise<T | HttpResponse<T>> {
    const response = await this.client.delete<T>(url, this.convertConfig(config))
    if (config?.withHeaders) {
      return { data: response.data, headers: new Headers(response.headers as Record<string, string>), status: response.status }
    }
    return response.data
  }
}

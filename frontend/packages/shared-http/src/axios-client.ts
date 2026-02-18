import axios, { type AxiosInstance, type AxiosRequestConfig, type AxiosError } from 'axios'
import type { IHttpClient, RequestConfig } from './types'
import { getTenantFromWindow } from './tenant-utils'

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
      (error: AxiosError) => {
        // 401 is handled by AuthGuard wrapper at layout level
        return Promise.reject(error)
      }
    )
  }

  private debugError(message: string, error: unknown): void {
    // Only log actual errors
    if (process.env.NODE_ENV === 'development') {
      console.error(`[AxiosHttpClient] ${message}`, error)
    }
  }

  private convertConfig(config?: RequestConfig): AxiosRequestConfig {
    return {
      headers: config?.headers,
      params: config?.params,
    }
  }

  async get<T>(url: string, config?: RequestConfig): Promise<T> {
    const response = await this.client.get<T>(url, this.convertConfig(config))
    return response.data
  }

  async post<T>(url: string, data?: unknown, config?: RequestConfig): Promise<T> {
    const response = await this.client.post<T>(url, data, this.convertConfig(config))
    return response.data
  }

  async put<T>(url: string, data?: unknown, config?: RequestConfig): Promise<T> {
    const response = await this.client.put<T>(url, data, this.convertConfig(config))
    return response.data
  }

  async patch<T>(url: string, data?: unknown, config?: RequestConfig): Promise<T> {
    const response = await this.client.patch<T>(url, data, this.convertConfig(config))
    return response.data
  }

  async delete<T>(url: string, config?: RequestConfig): Promise<T> {
    const response = await this.client.delete<T>(url, this.convertConfig(config))
    return response.data
  }
}

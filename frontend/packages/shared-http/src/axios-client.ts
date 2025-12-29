import axios, { type AxiosInstance, type AxiosRequestConfig } from 'axios'
import type { IHttpClient, RequestConfig } from './types'
import type { ITokenProvider } from './token-provider'

export class AxiosHttpClient implements IHttpClient {
  private readonly client: AxiosInstance

  constructor(
    baseURL: string,
    defaultHeaders?: Record<string, string>,
    private readonly tokenProvider?: ITokenProvider
  ) {
    this.client = axios.create({
      baseURL,
      headers: {
        'Content-Type': 'application/json',
        ...defaultHeaders,
      },
      timeout: 30000,
      // Enable credentials to send cookies automatically (NextAuth)
      withCredentials: true,
    })

    // Interceptor para agregar tenant desde subdominio y token si está disponible
    this.client.interceptors.request.use(
      async (config) => {
        // Extract tenant from subdomain (client-side only)
        if (typeof window !== 'undefined') {
          const hostname = window.location.hostname
          const parts = hostname.split('.')
          // If hostname is like: tenant.localhost or tenant.app.com
          if (parts.length >= 2 && parts[0] !== 'www') {
            const tenant = parts[0]
            if (config.headers && tenant) {
              config.headers['X-Tenant'] = tenant
            }
          }
        }

        // Try to add token from provider (fallback if not using cookies)
        if (this.tokenProvider) {
          const token = await this.tokenProvider.getClientToken()
          if (token && config.headers) {
            config.headers.Authorization = `Bearer ${token}`
          }
        }
        return config
      },
      (error) => Promise.reject(error)
    )

    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401 && typeof window !== 'undefined') {
          window.location.href = '/login'
        }
        return Promise.reject(error)
      }
    )
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

import type { IHttpClient } from './types'
import { AxiosHttpClient } from './axios-client'
import { FetchHttpClient } from './fetch-client'
import { TokenManager } from './token-manager'

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

export async function createServerHttpClient(): Promise<IHttpClient> {
  const token = await TokenManager.getServerToken()
  
  const headers: Record<string, string> = {}
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  return new FetchHttpClient(API_URL, headers)
}

export function createClientHttpClient(): IHttpClient {
  const token = TokenManager.getClientToken()
  
  const headers: Record<string, string> = {}
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  return new AxiosHttpClient(API_URL, headers)
}

export async function getHttpClient(): Promise<IHttpClient> {
  if (typeof window === 'undefined') {
    return createServerHttpClient()
  } else {
    return createClientHttpClient()
  }
}

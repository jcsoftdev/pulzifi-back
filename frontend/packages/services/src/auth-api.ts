import { getHttpClient } from '@workspace/shared-http'

// Internal: Backend response types (snake_case from Go)
interface UserBackendDto {
  id: string
  name: string
  email: string
  role: string
  avatar?: string
  tenant?: string | null
  created_at: string
  updated_at?: string
}

interface LoginBackendResponse {
  expires_in?: number
  tenant?: string | null
  expiresIn?: number
  nonce?: string
}

interface RegisterBackendResponse {
  user_id: string
  email: string
  first_name: string
  last_name: string
  status: string
  message: string
}

// Exported: Frontend types (camelCase)
export interface User {
  id: string
  name: string
  email: string
  role: string
  avatar?: string
  tenant?: string
  createdAt: string
  updatedAt?: string
}

export interface LoginDto {
  email: string
  password: string
}

export interface LoginResponse {
  expiresIn: number
  tenant?: string
  nonce?: string
}

// Helper: Transform backend to frontend format
function transformUser(backend: UserBackendDto): User {
  return {
    id: backend.id,
    name: backend.name,
    email: backend.email,
    role: backend.role,
    avatar: backend.avatar,
    tenant: backend.tenant ?? undefined,
    createdAt: backend.created_at,
    updatedAt: backend.updated_at,
  }
}

function mapLoginResponse(response: LoginBackendResponse): LoginResponse {
  return {
    expiresIn: response.expires_in ?? response.expiresIn ?? 3600,
    tenant: response.tenant ?? undefined,
    nonce: response.nonce,
  }
}

export const AuthApi = {
  async getCurrentUser(): Promise<User> {
    const http = await getHttpClient()
    const response = await http.get<UserBackendDto>('/api/v1/auth/me')
    return transformUser(response)
  },

  async login(credentials: LoginDto): Promise<LoginResponse> {
    const http = await getHttpClient()
    // Call the Next.js BFF route (/api/auth/login) — it proxies to the backend
    // server-side and forwards Set-Cookie back as same-origin so the browser stores it.
    // HttpOnly cookies cannot be read or set from JS directly.
    const response = await http.post<LoginBackendResponse>('/api/auth/login', credentials)
    return mapLoginResponse(response)
  },

  async logout(): Promise<void> {
    const http = await getHttpClient()
    // Use the Next.js BFF route so Set-Cookie (clear) is forwarded as same-origin
    await http.post('/api/auth/logout', {})
  },

  async checkSubdomain(subdomain: string): Promise<{ available: boolean; message?: string }> {
    const http = await getHttpClient()
    return http.get<{ available: boolean; message?: string }>(
      `/api/v1/auth/check-subdomain?subdomain=${encodeURIComponent(subdomain)}`
    )
  },

  async register(data: {
    email: string
    password: string
    firstName: string
    lastName: string
    organizationName: string
    organizationSubdomain: string
  }): Promise<{ status: string; message: string }> {
    const http = await getHttpClient()
    const response = await http.post<RegisterBackendResponse>('/api/v1/auth/register', {
      email: data.email,
      password: data.password,
      firstName: data.firstName,
      lastName: data.lastName,
      organization_name: data.organizationName,
      organization_subdomain: data.organizationSubdomain,
    })
    return {
      status: response.status,
      message: response.message,
    }
  },
}

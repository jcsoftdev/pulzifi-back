import { getHttpClient } from '@workspace/shared-http'

// Internal: Backend response types (snake_case from Go)
interface UserBackendDto {
  id: string
  name: string
  email: string
  role: string
  avatar?: string
  created_at: string
  updated_at?: string
}

interface LoginBackendResponse {
  session_id?: string
  expires_in?: number
  tenant?: string | null
  sessionId?: string
  expiresIn?: number
}

// Exported: Frontend types (camelCase)
export interface User {
  id: string
  name: string
  email: string
  role: string
  avatar?: string
  createdAt: string
  updatedAt?: string
}

export interface LoginDto {
  email: string
  password: string
}

export interface LoginResponse {
  sessionId?: string
  expiresIn: number
  tenant?: string
}

// Helper: Transform backend to frontend format
function transformUser(backend: UserBackendDto): User {
  return {
    id: backend.id,
    name: backend.name,
    email: backend.email,
    role: backend.role,
    avatar: backend.avatar,
    createdAt: backend.created_at,
    updatedAt: backend.updated_at,
  }
}

function mapLoginResponse(response: LoginBackendResponse): LoginResponse {
  const sessionId = response.session_id ?? response.sessionId
  const expiresIn = response.expires_in ?? response.expiresIn ?? 3600
  const tenant = response.tenant ?? undefined

  return {
    sessionId,
    expiresIn,
    tenant,
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
    const response = await http.post<LoginBackendResponse>('/api/v1/auth/login', credentials)
    return mapLoginResponse(response)
  },

  async logout(): Promise<void> {
    const http = await getHttpClient()
    await http.post('/api/v1/auth/logout', {})
  },

  async register(data: {
    email: string
    password: string
    firstName: string
    lastName: string
  }): Promise<User> {
    const http = await getHttpClient()
    const response = await http.post<UserBackendDto>('/api/v1/auth/register', data)
    return transformUser(response)
  },
}

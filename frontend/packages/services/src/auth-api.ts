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
  access_token: string
  refresh_token: string
  token_type: string
  expires_in: number
  tenant?: string
}

interface RegisterBackendResponse {
  user_id: string
  email: string
  first_name: string
  last_name: string
  message: string
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
  accessToken: string
  refreshToken: string
  tokenType: string
  expiresIn: number
  tenant?: string
}

export interface RegisterDto {
  email: string
  password: string
  firstName: string
  lastName: string
}

export interface RegisterResponse {
  userId: string
  email: string
  firstName: string
  lastName: string
  message: string
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

export const AuthApi = {
  async getCurrentUser(): Promise<User> {
    const http = await getHttpClient()
    const response = await http.get<UserBackendDto>('/api/v1/auth/me')
    return transformUser(response)
  },

  async login(credentials: LoginDto): Promise<LoginResponse> {
    const http = await getHttpClient()
    const response = await http.post<LoginBackendResponse>('/api/v1/auth/login', credentials)
    return {
      accessToken: response.access_token,
      refreshToken: response.refresh_token,
      tokenType: response.token_type,
      expiresIn: response.expires_in,
      tenant: response.tenant,
    }
  },

  async register(data: RegisterDto): Promise<RegisterResponse> {
    const http = await getHttpClient()
    const backendData = {
      email: data.email,
      password: data.password,
      first_name: data.firstName,
      last_name: data.lastName,
    }
    const response = await http.post<RegisterBackendResponse>('/api/v1/auth/register', backendData)
    return {
      userId: response.user_id,
      email: response.email,
      firstName: response.first_name,
      lastName: response.last_name,
      message: response.message,
    }
  },

  async logout(): Promise<void> {
    const http = await getHttpClient()
    await http.post('/api/v1/auth/logout', {})
  },
}

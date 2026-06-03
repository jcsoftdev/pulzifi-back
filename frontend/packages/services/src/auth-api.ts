import { createBffHttpClient, getHttpClient } from '@workspace/shared-http'

// Internal: Backend response types (snake_case from Go)
interface OrganizationBackendDto {
  id: string
  name: string
  subdomain: string
  planCode: string
  featureFlags: Record<string, unknown>
}

interface UserBackendDto {
  id: string
  name: string
  email: string
  role: string
  status?: string
  avatar?: string
  tenant?: string | null
  organization?: OrganizationBackendDto | null
  onboarding_completed?: boolean
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
  organization_subdomain: string
  status: string
  message: string
}

// Exported: Frontend types (camelCase)
export interface UserOrganization {
  id: string
  name: string
  subdomain: string
  planCode: string
  featureFlags: Record<string, unknown>
}

export interface User {
  id: string
  name: string
  email: string
  role: string
  status?: string
  avatar?: string
  tenant?: string | null
  organization?: UserOrganization
  onboardingCompleted: boolean
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
export function transformUser(backend: UserBackendDto): User {
  return {
    id: backend.id,
    name: backend.name,
    email: backend.email,
    role: backend.role,
    status: backend.status,
    avatar: backend.avatar,
    tenant: backend.tenant === undefined ? undefined : (backend.tenant ?? null),
    organization: backend.organization
      ? {
          id: backend.organization.id,
          name: backend.organization.name,
          subdomain: backend.organization.subdomain,
          planCode: backend.organization.planCode,
          featureFlags: backend.organization.featureFlags ?? {},
        }
      : undefined,
    onboardingCompleted: backend.onboarding_completed ?? false,
    createdAt: backend.created_at,
    updatedAt: backend.updated_at,
  }
}

export function mapLoginResponse(response: LoginBackendResponse): LoginResponse {
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
    const http = await createBffHttpClient()
    const response = await http.post<LoginBackendResponse>('/api/auth/login', credentials)
    return mapLoginResponse(response)
  },

  async logout(): Promise<void> {
    const http = await createBffHttpClient()
    await http.post('/api/auth/logout', {})
  },

  async updateProfile(data: { firstName: string; lastName: string }): Promise<User> {
    const http = await getHttpClient()
    const response = await http.put<UserBackendDto>('/api/v1/auth/me', {
      first_name: data.firstName,
      last_name: data.lastName,
    })
    return transformUser(response)
  },

  async changePassword(data: { currentPassword: string; newPassword: string }): Promise<void> {
    const http = await getHttpClient()
    await http.put('/api/v1/auth/me/password', {
      current_password: data.currentPassword,
      new_password: data.newPassword,
    })
  },

  async checkSubdomain(subdomain: string): Promise<{
    available: boolean
    message?: string
  }> {
    const http = await createBffHttpClient()
    return http.post<{
      available: boolean
      message?: string
    }>('/api/v1/auth/check-subdomain', {
      subdomain,
    })
  },

  async submitOnboarding(body: {
    org_name: string
    subdomain: string
    company_size?: string
    business_type?: string
    competitor_challenges?: string[]
    website_url?: string
  }): Promise<{ subdomain: string; trial_ends_at: string }> {
    const http = await getHttpClient()
    return http.post<{ subdomain: string; trial_ends_at: string }>(
      '/api/v1/auth/onboarding',
      body
    )
  },

  async submitOnboardingProfile(body: {
    company_size?: string
    business_type?: string
    competitor_challenges?: string[]
    website_url?: string
  }): Promise<{ onboarding_completed: boolean }> {
    const http = await getHttpClient()
    return http.post<{ onboarding_completed: boolean }>(
      '/api/v1/auth/onboarding/profile',
      body
    )
  },

  async register(data: {
    email: string
    password: string
    firstName: string
    lastName: string
    organizationName: string
    organizationSubdomain: string
  }): Promise<{
    organizationSubdomain: string
    status: string
    message: string
  }> {
    const http = await createBffHttpClient()
    const response = await http.post<RegisterBackendResponse>('/api/v1/auth/register', {
      email: data.email,
      password: data.password,
      firstName: data.firstName,
      lastName: data.lastName,
      organization_name: data.organizationName,
      organization_subdomain: data.organizationSubdomain,
    })
    return {
      organizationSubdomain: response.organization_subdomain,
      status: response.status,
      message: response.message,
    }
  },
}

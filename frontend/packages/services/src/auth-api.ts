import { createBffHttpClient, getHttpClient } from '@workspace/shared-http'

// Internal: Backend response types (snake_case from Go)
interface UserBackendDto {
  id: string
  name: string
  email: string
  role: string
  status?: string
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

interface PlatformInvitationDetailsBackendDto {
  email: string
  invited_by_name: string
  expires_at: string
}

interface AcceptInvitationSuccessBackendDto {
  nonce: string
  tenant: string
  expires_in: number
}

interface AcceptInvitationSessionFailureBackendDto {
  error: 'session_issue_failed'
  org_provisioned: true
  login_url: string
}

// Exported: Frontend types (camelCase)
export interface User {
  id: string
  name: string
  email: string
  role: string
  status?: string
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

export interface PlatformInvitationDetails {
  email: string
  invitedByName: string
  expiresAt: string
}

export interface AcceptInvitationPayload {
  firstName: string
  lastName: string
  password: string
  organizationName: string
  organizationSubdomain: string
}

export interface AcceptInvitationSuccess {
  nonce: string
  tenant: string
  expiresIn: number
}

export interface AcceptInvitationSessionFailure {
  error: 'session_issue_failed'
  orgProvisioned: true
  loginUrl: string
}

export type AcceptInvitationResponse = AcceptInvitationSuccess | AcceptInvitationSessionFailure

// Helper: Transform backend to frontend format
function transformUser(backend: UserBackendDto): User {
  return {
    id: backend.id,
    name: backend.name,
    email: backend.email,
    role: backend.role,
    status: backend.status,
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

  async getInvitation(token: string): Promise<PlatformInvitationDetails> {
    const http = await getHttpClient()
    const response = await http.get<PlatformInvitationDetailsBackendDto>(
      `/api/v1/auth/invitations/${encodeURIComponent(token)}`
    )
    return {
      email: response.email,
      invitedByName: response.invited_by_name,
      expiresAt: response.expires_at,
    }
  },

  async acceptInvitation(
    token: string,
    payload: AcceptInvitationPayload
  ): Promise<AcceptInvitationResponse> {
    const http = await getHttpClient()
    const body = {
      first_name: payload.firstName,
      last_name: payload.lastName,
      password: payload.password,
      organization_name: payload.organizationName,
      organization_subdomain: payload.organizationSubdomain,
    }

    try {
      const response = await http.post<AcceptInvitationSuccessBackendDto>(
        `/api/v1/auth/invitations/${encodeURIComponent(token)}/accept`,
        body
      )
      return {
        nonce: response.nonce,
        tenant: response.tenant,
        expiresIn: response.expires_in,
      }
    } catch (err: unknown) {
      // Backend returns 500 with { error: "session_issue_failed", org_provisioned, login_url }
      // when org/user creation succeeded but session issuance failed.
      const axiosErr = err as {
        response?: {
          status?: number
          data?:
            | AcceptInvitationSessionFailureBackendDto
            | {
                error?: string
              }
        }
      }
      const data = axiosErr?.response?.data as AcceptInvitationSessionFailureBackendDto | undefined
      if (data?.error === 'session_issue_failed' && data.org_provisioned) {
        return {
          error: 'session_issue_failed',
          orgProvisioned: true,
          loginUrl: data.login_url,
        }
      }
      throw err
    }
  },

  async register(data: {
    email: string
    password: string
    firstName: string
    lastName: string
    organizationName: string
    organizationSubdomain: string
  }): Promise<{
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
      status: response.status,
      message: response.message,
    }
  },
}

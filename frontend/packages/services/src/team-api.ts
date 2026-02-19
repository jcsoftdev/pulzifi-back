import { getHttpClient } from '@workspace/shared-http'

// Internal: Backend response types (snake_case from Go)
interface TeamMemberBackendDto {
  id: string
  user_id: string
  role: string
  first_name: string
  last_name: string
  email: string
  avatar_url?: string
  invited_by?: string
  joined_at: string
}

interface ListMembersBackendResponse {
  members: TeamMemberBackendDto[]
}

// Exported: Frontend types (camelCase)
export interface TeamMember {
  id: string
  userId: string
  role: string // "OWNER" | "ADMIN" | "MEMBER"
  firstName: string
  lastName: string
  email: string
  avatarUrl?: string
  invitedBy?: string
  joinedAt: string
}

export interface InviteMemberDto {
  email: string
  role: string
}

export interface UpdateMemberDto {
  role: string
}

// Helper: Transform backend to frontend format
function transformMember(backend: TeamMemberBackendDto): TeamMember {
  return {
    id: backend.id,
    userId: backend.user_id,
    role: backend.role,
    firstName: backend.first_name,
    lastName: backend.last_name,
    email: backend.email,
    avatarUrl: backend.avatar_url,
    invitedBy: backend.invited_by,
    joinedAt: backend.joined_at,
  }
}

export const TeamApi = {
  async listMembers(): Promise<TeamMember[]> {
    const http = await getHttpClient()
    const response = await http.get<ListMembersBackendResponse>('/api/v1/team/members')
    return response.members.map(transformMember)
  },

  async inviteMember(data: InviteMemberDto): Promise<TeamMember> {
    const http = await getHttpClient()
    const response = await http.post<TeamMemberBackendDto>('/api/v1/team/members', data)
    return transformMember(response)
  },

  async updateMember(memberId: string, data: UpdateMemberDto): Promise<void> {
    const http = await getHttpClient()
    await http.put(`/api/v1/team/members/${memberId}`, data)
  },

  async removeMember(memberId: string): Promise<void> {
    const http = await getHttpClient()
    await http.delete(`/api/v1/team/members/${memberId}`)
  },
}

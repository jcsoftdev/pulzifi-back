export type MemberRole = 'OWNER' | 'ADMIN' | 'MEMBER'

export interface TeamMember {
  id: string
  userId: string
  role: MemberRole
  firstName: string
  lastName: string
  email: string
  avatarUrl?: string
  invitedBy?: string
  joinedAt: string
}

export function memberFullName(member: TeamMember): string {
  return `${member.firstName} ${member.lastName}`.trim() || member.email
}

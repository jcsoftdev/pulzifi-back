export type InvitationStatus = 'pending' | 'accepted' | 'revoked' | 'expired'

export interface PlatformInvitation {
  id: string
  email: string
  status: InvitationStatus
  invitedBy: string
  expiresAt: string
  emailSentAt: string | null
  emailError: string | null
  resentCount: number
  createdAt: string
}

export interface InviteToPlatformResponse {
  id: string
  email: string
  expiresAt: string
  emailDelivery: 'sent' | 'failed'
}

export interface PlatformInvitationDetails {
  email: string
  invitedByName: string
  expiresAt: string
}

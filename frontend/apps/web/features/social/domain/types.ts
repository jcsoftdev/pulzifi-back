// Re-export all service-layer types so feature code imports from one place.
export type {
  ChangeType,
  ChangeSummary,
  CaptionDiff,
  CreateSocialProfileDto,
  Platform,
  ProfileData,
  SocialChange,
  SocialChangeDetail,
  SocialPost,
  SocialProfile,
  SocialProfileDetail,
  SocialQuotaStatus,
  SocialSnapshot,
  TextDiff,
  UpdateSocialProfileDto,
} from '@workspace/services/social-api'

// ─── UI-level models ─────────────────────────────────────────────────────────

export const INTERVAL_PRESETS: { value: number; label: string }[] = [
  { value: 1440, label: 'Daily (every 24h)' },
  { value: 720, label: 'Every 12 hours' },
  { value: 360, label: 'Every 6 hours' },
  { value: 120, label: 'Every 2 hours' },
]

export const PLATFORM_LABELS: Record<string, string> = {
  instagram: 'Instagram',
  tiktok: 'TikTok',
  facebook: 'Facebook',
}

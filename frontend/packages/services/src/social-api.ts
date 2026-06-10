import { getHttpClient } from '@workspace/shared-http'

// ─── Platform & change-type enums ──────────────────────────────────────────

export type Platform = 'instagram' | 'tiktok' | 'facebook'

export type ChangeType =
  | 'new_post'
  | 'removed_post'
  | 'caption_edited'
  | 'bio_changed'
  | 'followers_changed'
  | 'avatar_changed'
  | 'display_name_changed'

// ─── Backend DTOs (snake_case — as returned by Go) ─────────────────────────

interface SocialProfileBackendDto {
  id: string
  workspace_id: string
  platform: Platform
  handle: string
  display_name: string
  avatar_url: string
  is_active: boolean
  check_interval_minutes: number
  next_check_at: string | null
  last_checked_at: string | null
  consecutive_failures: number
  created_at: string
  updated_at: string
}

interface PostBackendDto {
  external_id: string
  caption: string
  media_url: string
  stored_media_url: string
  posted_at: string
  likes_count: number
  comments_count: number
}

interface ProfileDataBackendDto {
  bio: string
  display_name: string
  avatar_url: string
  followers_count: number
  following_count: number
  posts_count: number
  posts: PostBackendDto[]
}

interface TextDiffBackendDto {
  from: string
  to: string
}

interface CaptionDiffBackendDto {
  external_id: string
  from: string
  to: string
}

interface ChangeSummaryBackendDto {
  followers_delta?: number
  bio?: TextDiffBackendDto
  display_name?: TextDiffBackendDto
  new_posts?: string[]
  removed_posts?: string[]
  edited_captions?: CaptionDiffBackendDto[]
  avatar_changed?: boolean
}

interface SocialChangeBackendDto {
  id: string
  profile_id: string
  from_snapshot_id: string
  to_snapshot_id: string
  change_types: ChangeType[]
  summary?: ChangeSummaryBackendDto
  created_at: string
}

interface SocialQuotaBackendDto {
  checks_used: number
  checks_limit: number | null
  resets_at: string
}

interface SocialSnapshotBackendDto {
  id: string
  profile_id: string
  captured_at: string
  status: 'success' | 'failed'
  error: string | null
  followers_count: number | null
  posts_count: number | null
  data: ProfileDataBackendDto | null
}

interface GetProfileBackendDto {
  profile: SocialProfileBackendDto
  latest_snapshot: SocialSnapshotBackendDto | null
}

interface SocialSnapshotSummaryBackendDto {
  id: string
  profile_id: string
  captured_at: string
  status: 'success' | 'failed'
  error?: string
  followers_count: number
  posts_count: number
}

interface GetChangeBackendDto {
  change: SocialChangeBackendDto
  before_data: ProfileDataBackendDto | null
  after_data: ProfileDataBackendDto | null
}

// ─── Exported frontend types (camelCase) ───────────────────────────────────

export interface SocialPost {
  externalId: string
  caption: string
  mediaUrl: string
  storedMediaUrl: string
  postedAt: string
  likesCount: number
  commentsCount: number
}

export interface ProfileData {
  bio: string
  displayName: string
  avatarUrl: string
  followersCount: number
  followingCount: number
  postsCount: number
  posts: SocialPost[]
}

export interface TextDiff {
  from: string
  to: string
}

export interface CaptionDiff {
  externalId: string
  from: string
  to: string
}

export interface ChangeSummary {
  followersDelta: number
  bio?: TextDiff
  displayName?: TextDiff
  newPosts: string[]
  removedPosts: string[]
  editedCaptions: CaptionDiff[]
  avatarChanged: boolean
}

export interface SocialProfile {
  id: string
  workspaceId: string
  platform: Platform
  handle: string
  displayName: string
  avatarUrl: string
  isActive: boolean
  checkIntervalMinutes: number
  nextCheckAt: string | null
  lastCheckedAt: string | null
  consecutiveFailures: number
  createdAt: string
  updatedAt: string
}

export interface SocialChange {
  id: string
  profileId: string
  fromSnapshotId: string
  toSnapshotId: string
  changeTypes: ChangeType[]
  summary?: ChangeSummary
  createdAt: string
}

export interface SocialQuotaStatus {
  checksUsed: number
  checksLimit: number | null
  resetsAt: string
}

export interface SocialSnapshot {
  id: string
  profileId: string
  capturedAt: string
  status: 'success' | 'failed'
  error: string | null
  followersCount: number | null
  postsCount: number | null
  data: ProfileData | null
}

export interface SocialSnapshotSummary {
  id: string
  profileId: string
  capturedAt: string
  status: 'success' | 'failed'
  error?: string
  followersCount: number
  postsCount: number
}

export interface SocialProfileDetail {
  profile: SocialProfile
  latestSnapshot: SocialSnapshot | null
}

export interface SocialChangeDetail {
  change: SocialChange
  beforeData: ProfileData | null
  afterData: ProfileData | null
}

// ─── Input DTOs ─────────────────────────────────────────────────────────────

export interface CreateSocialProfileDto {
  workspaceId: string
  platform: Platform
  handle: string
  checkIntervalMinutes: number
}

export interface UpdateSocialProfileDto {
  checkIntervalMinutes?: number
  isActive?: boolean
}

// ─── Transformers ───────────────────────────────────────────────────────────

function transformPost(p: PostBackendDto): SocialPost {
  return {
    externalId: p.external_id,
    caption: p.caption,
    mediaUrl: p.media_url,
    storedMediaUrl: p.stored_media_url,
    postedAt: p.posted_at,
    likesCount: p.likes_count,
    commentsCount: p.comments_count,
  }
}

function transformProfileData(d: ProfileDataBackendDto): ProfileData {
  return {
    bio: d.bio ?? '',
    displayName: d.display_name ?? '',
    avatarUrl: d.avatar_url ?? '',
    followersCount: d.followers_count ?? 0,
    followingCount: d.following_count ?? 0,
    postsCount: d.posts_count ?? 0,
    posts: (d.posts ?? []).map(transformPost),
  }
}

function transformSummary(s: ChangeSummaryBackendDto): ChangeSummary {
  return {
    followersDelta: s.followers_delta ?? 0,
    bio: s.bio ? { from: s.bio.from, to: s.bio.to } : undefined,
    displayName: s.display_name ? { from: s.display_name.from, to: s.display_name.to } : undefined,
    newPosts: s.new_posts ?? [],
    removedPosts: s.removed_posts ?? [],
    editedCaptions: (s.edited_captions ?? []).map((c) => ({
      externalId: c.external_id,
      from: c.from,
      to: c.to,
    })),
    avatarChanged: s.avatar_changed ?? false,
  }
}

function transformProfile(p: SocialProfileBackendDto): SocialProfile {
  return {
    id: p.id,
    workspaceId: p.workspace_id,
    platform: p.platform,
    handle: p.handle,
    displayName: p.display_name,
    avatarUrl: p.avatar_url,
    isActive: p.is_active,
    checkIntervalMinutes: p.check_interval_minutes,
    nextCheckAt: p.next_check_at,
    lastCheckedAt: p.last_checked_at,
    consecutiveFailures: p.consecutive_failures,
    createdAt: p.created_at,
    updatedAt: p.updated_at,
  }
}

function transformSnapshot(s: SocialSnapshotBackendDto): SocialSnapshot {
  return {
    id: s.id,
    profileId: s.profile_id,
    capturedAt: s.captured_at,
    status: s.status,
    error: s.error,
    followersCount: s.followers_count,
    postsCount: s.posts_count,
    data: s.data ? transformProfileData(s.data) : null,
  }
}

function transformSnapshotSummary(s: SocialSnapshotSummaryBackendDto): SocialSnapshotSummary {
  return {
    id: s.id,
    profileId: s.profile_id,
    capturedAt: s.captured_at,
    status: s.status,
    error: s.error,
    followersCount: s.followers_count ?? 0,
    postsCount: s.posts_count ?? 0,
  }
}

function transformChange(c: SocialChangeBackendDto): SocialChange {
  return {
    id: c.id,
    profileId: c.profile_id,
    fromSnapshotId: c.from_snapshot_id,
    toSnapshotId: c.to_snapshot_id,
    changeTypes: c.change_types,
    summary: c.summary ? transformSummary(c.summary) : undefined,
    createdAt: c.created_at,
  }
}

// ─── API client ─────────────────────────────────────────────────────────────

export const SocialApi = {
  async createProfile(data: CreateSocialProfileDto): Promise<SocialProfile> {
    const http = await getHttpClient()
    const response = await http.post<SocialProfileBackendDto>(
      `/api/v1/workspaces/${data.workspaceId}/social-profiles`,
      {
        platform: data.platform,
        handle: data.handle,
        check_interval_minutes: data.checkIntervalMinutes,
      }
    )
    return transformProfile(response)
  },

  async listProfiles(workspaceId: string): Promise<SocialProfile[]> {
    const http = await getHttpClient()
    const response = await http.get<SocialProfileBackendDto[]>(
      `/api/v1/workspaces/${workspaceId}/social-profiles`
    )
    return (response ?? []).map(transformProfile)
  },

  async getProfile(profileId: string): Promise<SocialProfileDetail> {
    const http = await getHttpClient()
    const response = await http.get<GetProfileBackendDto>(`/api/v1/social-profiles/${profileId}`)
    return {
      profile: transformProfile(response.profile),
      latestSnapshot: response.latest_snapshot ? transformSnapshot(response.latest_snapshot) : null,
    }
  },

  async updateProfile(profileId: string, data: UpdateSocialProfileDto): Promise<SocialProfile> {
    const http = await getHttpClient()
    const payload: Record<string, unknown> = {}
    if (data.checkIntervalMinutes !== undefined)
      payload.check_interval_minutes = data.checkIntervalMinutes
    if (data.isActive !== undefined) payload.is_active = data.isActive
    const response = await http.patch<SocialProfileBackendDto>(
      `/api/v1/social-profiles/${profileId}`,
      payload
    )
    return transformProfile(response)
  },

  async deleteProfile(profileId: string): Promise<void> {
    const http = await getHttpClient()
    await http.delete(`/api/v1/social-profiles/${profileId}`)
  },

  async triggerCheck(profileId: string): Promise<SocialSnapshot> {
    const http = await getHttpClient()
    const response = await http.post<SocialSnapshotBackendDto>(
      `/api/v1/social-profiles/${profileId}/check`,
      {}
    )
    return transformSnapshot(response)
  },

  async listChanges(profileId: string): Promise<SocialChange[]> {
    const http = await getHttpClient()
    const response = await http.get<{ changes: SocialChangeBackendDto[] }>(
      `/api/v1/social-profiles/${profileId}/changes`
    )
    return (response.changes ?? []).map(transformChange)
  },

  async listWorkspaceChanges(
    workspaceId: string,
    limit = 50,
    offset = 0
  ): Promise<SocialChange[]> {
    const http = await getHttpClient()
    const response = await http.get<{ changes: SocialChangeBackendDto[] }>(
      `/api/v1/workspaces/${workspaceId}/social-changes?limit=${limit}&offset=${offset}`
    )
    return (response.changes ?? []).map(transformChange)
  },

  async getChange(changeId: string): Promise<SocialChangeDetail> {
    const http = await getHttpClient()
    const response = await http.get<GetChangeBackendDto>(`/api/v1/social-changes/${changeId}`)
    return {
      change: transformChange(response.change),
      beforeData: response.before_data ? transformProfileData(response.before_data) : null,
      afterData: response.after_data ? transformProfileData(response.after_data) : null,
    }
  },

  async listSnapshots(profileId: string, limit = 20, offset = 0): Promise<SocialSnapshotSummary[]> {
    const http = await getHttpClient()
    const response = await http.get<{ snapshots: SocialSnapshotSummaryBackendDto[] }>(
      `/api/v1/social-profiles/${profileId}/snapshots?limit=${limit}&offset=${offset}`
    )
    return (response.snapshots ?? []).map(transformSnapshotSummary)
  },

  async getSnapshot(snapshotId: string): Promise<SocialSnapshot | null> {
    const http = await getHttpClient()
    const response = await http.get<SocialSnapshotBackendDto | null>(
      `/api/v1/social-snapshots/${snapshotId}`
    )
    return response ? transformSnapshot(response) : null
  },

  async getQuotaStatus(workspaceId: string): Promise<SocialQuotaStatus> {
    const http = await getHttpClient()
    const response = await http.get<SocialQuotaBackendDto>(
      `/api/v1/workspaces/${workspaceId}/social-quota`
    )
    return {
      checksUsed: response.checks_used,
      checksLimit: response.checks_limit,
      resetsAt: response.resets_at,
    }
  },
}

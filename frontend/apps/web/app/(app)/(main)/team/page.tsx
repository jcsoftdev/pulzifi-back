import { AuthApi } from '@workspace/services'
import { TeamFeature } from '@/features/team'

export default async function TeamPage() {
  let currentUserId: string | undefined

  try {
    const user = await AuthApi.getCurrentUser()
    currentUserId = user.id
  } catch {
    // Not critical if user fetch fails — the feature will still render
  }

  return <TeamFeature currentUserId={currentUserId} />
}

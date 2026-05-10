import { TeamFeature } from '@/features/team'
import { AuthApi } from '@workspace/services'

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

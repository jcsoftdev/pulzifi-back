import { SettingsFeature } from '@/features/settings'
import { AuthApi } from '@workspace/services'

export default async function SettingsPage() {
  let userEmail: string | undefined

  try {
    const user = await AuthApi.getCurrentUser()
    userEmail = user.email
  } catch {
    // Not critical — feature still renders without email
  }

  return <SettingsFeature userEmail={userEmail} />
}

import { AuthApi, UsageApi } from '@workspace/services'
import { DashboardFeature } from '@/features/dashboard'

export default async function Page() {
  let userName = ''
  let monthlyChecks = 0
  let maxMonthlyChecks = 2000
  let usagePercent = 0

  try {
    const [user, checksData] = await Promise.all([
      AuthApi.getCurrentUser(),
      UsageApi.getChecksData(),
    ])
    userName = user.name.split(' ')[0] ?? user.name
    monthlyChecks = checksData.current
    maxMonthlyChecks = checksData.max || 2000
    usagePercent =
      checksData.max > 0 ? Math.round((checksData.current / checksData.max) * 100) : 0
  } catch {
    // Non-critical — feature still renders without this data
  }

  return (
    <DashboardFeature
      userName={userName}
      monthlyChecks={monthlyChecks}
      maxMonthlyChecks={maxMonthlyChecks}
      usagePercent={usagePercent}
    />
  )
}

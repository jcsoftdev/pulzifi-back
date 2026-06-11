import { ChecksTag } from '@workspace/ui/components/molecules'
import { UsageService } from '@/features/usage/domain/services/usage-service'

export async function ChecksDataLoader() {
  const checksData = await UsageService.getChecksData()
  if (!checksData) return null

  return (
    <div className="hidden md:block">
      <ChecksTag
        current={checksData.current}
        max={checksData.max}
        refillDate={checksData.refillDate}
      />
    </div>
  )
}

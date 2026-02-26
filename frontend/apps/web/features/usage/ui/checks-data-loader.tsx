import { UsageService } from '@/features/usage/domain/services/usage-service'
import { ChecksTag } from '@workspace/ui/components/molecules'

export async function ChecksDataLoader() {
  const checksData = await UsageService.getChecksData()

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

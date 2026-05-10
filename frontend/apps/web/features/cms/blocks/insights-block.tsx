import { InsightsSection } from '@/features/landing'

// v1: InsightsSection renders with its own internal data.
// CMS control of this section is deferred to a future iteration.
type Props = { block: { blockType: 'insights' } }

export function InsightsBlock(_props: Props) {
  return <InsightsSection />
}

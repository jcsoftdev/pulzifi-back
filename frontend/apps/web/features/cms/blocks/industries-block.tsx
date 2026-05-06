import { IndustriesSection } from '@/features/landing'

// v1: IndustriesSection renders with its own internal data.
// CMS control of this section is deferred to a future iteration.
type Props = { block: { blockType: 'industries' } }

export function IndustriesBlock(_props: Props) {
  return <IndustriesSection />
}

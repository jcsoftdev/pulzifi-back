import { IndustriesSection } from '@/features/landing'

type Props = {
  block: {
    blockType: 'industries'
    items?: { title: string; description: string; icon?: string | null }[] | null
  }
}

export function IndustriesBlock(_props: Props) {
  return <IndustriesSection />
}

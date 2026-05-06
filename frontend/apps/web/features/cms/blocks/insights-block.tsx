import { InsightsSection } from '@/features/landing'

type Props = {
  block: {
    blockType: 'insights'
    heading?: string | null
    subheading?: string | null
    items?: { title: string; description: string }[] | null
  }
}

export function InsightsBlock(_props: Props) {
  return <InsightsSection />
}

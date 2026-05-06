import { HowItWorksSection } from '@/features/landing'

type Props = {
  block: {
    blockType: 'how-it-works'
    steps?: { step: number; title: string; description: string }[] | null
  }
}

export function HowItWorksBlock({ block }: Props) {
  return <HowItWorksSection steps={block.steps ?? []} />
}

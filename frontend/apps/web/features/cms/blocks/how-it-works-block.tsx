import { HowItWorksSection } from '@/features/landing'

type Props = {
  block: {
    blockType: 'how-it-works'
    eyebrow?: string | null
    headline?: string | null
    headlineHighlight?: string | null
    subheadline?: string | null
    steps?:
      | {
          step: number
          icon?: string | null
          title: string
          description: string
          mockType?: 'url' | 'insight' | 'alerts' | null
          mockText?: string | null
        }[]
      | null
  }
}

export function HowItWorksBlock({ block }: Props) {
  return (
    <HowItWorksSection
      eyebrow={block.eyebrow ?? undefined}
      headline={block.headline ?? undefined}
      headlineHighlight={block.headlineHighlight ?? undefined}
      subheadline={block.subheadline ?? undefined}
      steps={
        block.steps?.map((s) => ({
          step: s.step,
          icon: s.icon ?? undefined,
          title: s.title,
          description: s.description,
          mockType: (s.mockType ?? 'url') as 'url' | 'insight' | 'alerts',
          mockText: s.mockText ?? undefined,
        })) ?? undefined
      }
    />
  )
}

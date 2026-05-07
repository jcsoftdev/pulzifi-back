import { ProblemSection } from '@/features/landing'

type Props = {
  block: {
    blockType: 'problem'
    eyebrow?: string | null
    headline?: string | null
    headlineHighlight?: string | null
    cards?: { metric?: string | null; label?: string | null; description?: string | null }[] | null
  }
}

export function ProblemBlock({ block }: Props) {
  return (
    <ProblemSection
      eyebrow={block.eyebrow ?? undefined}
      headline={block.headline ?? undefined}
      headlineHighlight={block.headlineHighlight ?? undefined}
      cards={
        block.cards?.map((c) => ({
          metric: c.metric ?? '',
          label: c.label ?? '',
          description: c.description ?? '',
        })) ?? undefined
      }
    />
  )
}

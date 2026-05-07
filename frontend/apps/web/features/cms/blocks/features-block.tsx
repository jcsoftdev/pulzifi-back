import { FeaturesSection } from '@/features/landing'

type Props = {
  block: {
    blockType: 'features'
    eyebrow?: string | null
    headline?: string | null
    headlineHighlight?: string | null
    intro?: string | null
    bullets?: { title?: string | null; description?: string | null }[] | null
    demoTitle?: string | null
    demoBadge?: string | null
    demoSite?: string | null
    demoChange?: string | null
    demoAnalysis?: string | null
    demoActions?: { label?: string | null }[] | null
  }
}

export function FeaturesBlock({ block }: Props) {
  return (
    <FeaturesSection
      eyebrow={block.eyebrow ?? undefined}
      headline={block.headline ?? undefined}
      headlineHighlight={block.headlineHighlight ?? undefined}
      intro={block.intro ?? undefined}
      bullets={
        block.bullets
          ?.filter((b) => b.title && b.description)
          .map((b) => ({ title: b.title!, description: b.description! })) ?? undefined
      }
      demoTitle={block.demoTitle ?? undefined}
      demoBadge={block.demoBadge ?? undefined}
      demoSite={block.demoSite ?? undefined}
      demoChange={block.demoChange ?? undefined}
      demoAnalysis={block.demoAnalysis ?? undefined}
      demoActions={
        block.demoActions
          ?.filter((a) => a.label)
          .map((a) => ({ label: a.label! })) ?? undefined
      }
    />
  )
}

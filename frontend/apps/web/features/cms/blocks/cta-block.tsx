import { CtaSection } from '@/features/landing'

type Props = {
  block: {
    blockType: 'cta'
    eyebrow?: string | null
    headline?: string | null
    headlineHighlight?: string | null
    subtext?: string | null
    primaryLabel?: string | null
    primaryHref?: string | null
    secondaryLabel?: string | null
    secondaryHref?: string | null
    riskNote?: string | null
    variant?: 'primary' | 'secondary' | null
  }
}

export function CtaBlock({ block }: Props) {
  return (
    <CtaSection
      eyebrow={block.eyebrow ?? undefined}
      headline={block.headline ?? undefined}
      headlineHighlight={block.headlineHighlight ?? undefined}
      subtext={block.subtext ?? undefined}
      primaryLabel={block.primaryLabel ?? undefined}
      primaryHref={block.primaryHref ?? undefined}
      riskNote={block.riskNote ?? undefined}
    />
  )
}

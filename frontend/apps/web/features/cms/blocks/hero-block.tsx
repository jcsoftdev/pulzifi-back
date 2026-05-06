import { HeroSection } from '@/features/landing'

type Props = {
  block: {
    blockType: 'hero'
    headline?: string | null
    subheadline?: string | null
    ctaLabel?: string | null
    ctaHref?: string | null
  }
}

export function HeroBlock({ block }: Props) {
  return (
    <HeroSection
      headline={block.headline ?? undefined}
      subheadline={block.subheadline ?? undefined}
      ctaLabel={block.ctaLabel ?? undefined}
      ctaHref={block.ctaHref ?? undefined}
    />
  )
}

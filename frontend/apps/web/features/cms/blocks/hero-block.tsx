import { HeroSection } from '@/features/landing'

type MediaRef = {
  url?: string | null
  alt?: string | null
  width?: number | null
  height?: number | null
}

type Props = {
  block: {
    blockType: 'hero'
    eyebrowBadge?: string | null
    eyebrowText?: string | null
    headline?: string | null
    headlineHighlight?: string | null
    subheadline?: string | null
    primaryCtaLabel?: string | null
    primaryCtaHref?: string | null
    secondaryCtaLabel?: string | null
    secondaryCtaHref?: string | null
    trustLine?: string | null
    image?: MediaRef | number | string | null
  }
}

export function HeroBlock({ block }: Props) {
  const image =
    typeof block.image === 'object' && block.image !== null
      ? {
          url: block.image.url ?? undefined,
          alt: block.image.alt ?? undefined,
          width: block.image.width ?? undefined,
          height: block.image.height ?? undefined,
        }
      : undefined

  return (
    <HeroSection
      eyebrowBadge={block.eyebrowBadge ?? undefined}
      eyebrowText={block.eyebrowText ?? undefined}
      headline={block.headline ?? undefined}
      headlineHighlight={block.headlineHighlight ?? undefined}
      subheadline={block.subheadline ?? undefined}
      primaryCtaLabel={block.primaryCtaLabel ?? undefined}
      primaryCtaHref={block.primaryCtaHref ?? undefined}
      secondaryCtaLabel={block.secondaryCtaLabel ?? undefined}
      secondaryCtaHref={block.secondaryCtaHref ?? undefined}
      trustLine={block.trustLine ?? undefined}
      image={image}
    />
  )
}

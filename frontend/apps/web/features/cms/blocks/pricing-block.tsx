import { PricingSection } from '@/features/landing'

type PlanDoc = {
  id: string | number
  name: string
  price: string
  priceAnnual?: string | null
  period?: string | null
  tagline?: string | null
  features?: { text?: string | null; included?: boolean | null }[] | null
  ctaLabel?: string | null
  ctaHref?: string | null
  highlighted?: boolean | null
  popularBadge?: string | null
}

type Props = {
  block: {
    blockType: 'pricing'
    eyebrow?: string | null
    headline?: string | null
    headlineHighlight?: string | null
    subheadline?: string | null
    guaranteeNote?: string | null
    plans?: (PlanDoc | string | number)[] | null
    billing?: {
      monthlyLabel?: string | null
      annualLabel?: string | null
      annualBadge?: string | null
      annualNote?: string | null
    } | null
    comparePlansHeadline?: string | null
    featuresLabel?: string | null
  }
}

function isPlanDoc(value: PlanDoc | string | number): value is PlanDoc {
  return typeof value === 'object' && value !== null && 'name' in value
}

export function PricingBlock({ block }: Props) {
  const populated = (block.plans ?? []).filter(isPlanDoc)

  return (
    <PricingSection
      eyebrow={block.eyebrow ?? undefined}
      headline={block.headline ?? undefined}
      headlineHighlight={block.headlineHighlight ?? undefined}
      subheadline={block.subheadline ?? undefined}
      guaranteeNote={block.guaranteeNote ?? undefined}
      billing={block.billing ? {
        monthlyLabel: block.billing.monthlyLabel ?? undefined,
        annualLabel: block.billing.annualLabel ?? undefined,
        annualBadge: block.billing.annualBadge ?? undefined,
        annualNote: block.billing.annualNote ?? undefined,
      } : undefined}
      comparePlansHeadline={block.comparePlansHeadline ?? undefined}
      featuresLabel={block.featuresLabel ?? undefined}
      plans={populated.map((p) => ({
        name: p.name,
        price: p.price,
        priceAnnual: p.priceAnnual ?? undefined,
        period: p.period ?? undefined,
        tagline: p.tagline ?? undefined,
        features:
          p.features?.map((f) => ({
            text: f.text ?? '',
            included: f.included ?? true,
          })) ?? undefined,
        ctaLabel: p.ctaLabel ?? undefined,
        ctaHref: p.ctaHref ?? undefined,
        highlighted: p.highlighted ?? false,
        popularBadge: p.popularBadge ?? undefined,
      }))}
    />
  )
}

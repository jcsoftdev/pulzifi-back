import { PricingSection } from '@/features/landing'

type Props = {
  block: {
    blockType: 'pricing'
    plans?: {
      name: string
      price: string
      period?: string | null
      features?: { text: string }[] | null
      ctaLabel?: string | null
      ctaHref?: string | null
      highlighted?: boolean | null
    }[] | null
  }
}

export function PricingBlock({ block }: Props) {
  return (
    <PricingSection
      plans={(block.plans ?? []).map((p) => ({
        name: p.name,
        price: p.price,
        period: p.period ?? undefined,
        features: (p.features ?? []).map((f) => f.text),
        ctaLabel: p.ctaLabel ?? undefined,
        ctaHref: p.ctaHref ?? undefined,
        highlighted: p.highlighted ?? false,
      }))}
    />
  )
}

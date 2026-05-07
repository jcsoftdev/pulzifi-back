'use client'

import { useCardStagger } from '../lib/gsap'
import { Eyebrow } from './components/eyebrow'
import { Highlight } from './components/highlight'
import { PricingCard } from './components/pricing-card'
import { SectionFrame } from './components/section-frame'

type PricingFeature = { text?: string; included?: boolean }
type PricingPlan = {
  name: string
  price: string
  period?: string
  tagline?: string
  features?: PricingFeature[]
  ctaLabel?: string
  ctaHref?: string
  highlighted?: boolean
  popularBadge?: string
}

type PricingSectionProps = {
  eyebrow?: string
  headline?: string
  headlineHighlight?: string
  subheadline?: string
  guaranteeNote?: string
  plans?: PricingPlan[]
}

export function PricingSection({
  eyebrow,
  headline,
  headlineHighlight,
  subheadline,
  guaranteeNote,
  plans,
}: Readonly<PricingSectionProps> = {}) {
  const items = plans ?? []
  const cardsRef = useCardStagger<HTMLDivElement>({ scale: true, stagger: 0.1, y: 32 })

  return (
    <SectionFrame id="pricing" bg="alt" className="relative overflow-hidden">
      {/* Gold blob — top center */}
      <div className="pointer-events-none absolute inset-0 -z-10 flex justify-center" aria-hidden>
        <div
          className="h-[400px] w-[600px] -translate-y-1/3"
          style={{
            background: 'radial-gradient(circle, color-mix(in srgb, var(--pz-accent-gold) 18%, transparent) 0%, transparent 65%)',
            filter: 'blur(100px)',
            opacity: 0.35,
          }}
        />
      </div>

      <div className="mx-auto max-w-2xl text-center">
        {eyebrow && <Eyebrow className="mb-4">{eyebrow}</Eyebrow>}
        <h2 className="font-heading text-4xl font-bold tracking-tight text-[var(--pz-ink)] md:text-5xl">
          {headline ?? 'Simple,'}{' '}
          {headlineHighlight ? (
            <Highlight tone="accent-gold">{headlineHighlight}</Highlight>
          ) : (
            <Highlight tone="accent-gold">Transparent Pricing</Highlight>
          )}
        </h2>
        {subheadline && (
          <p className="mt-4 text-lg leading-relaxed text-[var(--pz-ink-2)]">{subheadline}</p>
        )}
      </div>

      <div ref={cardsRef} className="mt-14 grid grid-cols-1 gap-6 md:grid-cols-3 md:gap-8">
        {items.map((plan) => (
          <div key={plan.name} data-pz-card>
            <PricingCard
              name={plan.name}
              price={plan.price}
              period={plan.period}
              description={plan.tagline}
              cta={plan.ctaLabel ?? 'Get Started'}
              ctaHref={plan.ctaHref ?? '/register'}
              features={(plan.features ?? []).map((f) => ({
                text: f.text ?? '',
                included: f.included ?? true,
              }))}
              popular={plan.highlighted ?? false}
              popularBadge={plan.popularBadge}
            />
          </div>
        ))}
      </div>

      {guaranteeNote && (
        <p className="mt-10 text-center text-sm leading-5 text-[var(--pz-ink-2)]">{guaranteeNote}</p>
      )}
    </SectionFrame>
  )
}

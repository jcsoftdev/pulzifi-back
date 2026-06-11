'use client'

import { type CSSProperties, useEffect, useState } from 'react'
import 'swiper/css'
import 'swiper/css/pagination'
import { A11y, Pagination } from 'swiper/modules'
import { Swiper, SwiperSlide } from 'swiper/react'
import { Eyebrow } from './components/eyebrow'
import { Highlight } from './components/highlight'
import { PricingCard } from './components/pricing-card'
import { PricingComparisonTable } from './components/pricing-comparison-table'
import { SectionFrame } from './components/section-frame'

/**
 * Returns true only after mount AND when the viewport is ≥ 768px (md breakpoint).
 * Initialises to false so the server render and the first client paint agree —
 * prevents hydration mismatches while keeping mobile stacked cards SSR-rendered.
 */
function useIsMdOrAbove(): boolean {
  const [isMd, setIsMd] = useState(false)
  useEffect(() => {
    const mq = window.matchMedia('(min-width: 768px)')
    setIsMd(mq.matches)
    const handler = (e: MediaQueryListEvent) => setIsMd(e.matches)
    mq.addEventListener('change', handler)
    return () => mq.removeEventListener('change', handler)
  }, [])
  return isMd
}

type PricingFeature = {
  text?: string
  included?: boolean
}
type PricingPlan = {
  name: string
  price: string
  priceAnnual?: string
  period?: string
  tagline?: string
  features?: PricingFeature[]
  ctaLabel?: string
  ctaHref?: string
  highlighted?: boolean
  popularBadge?: string
  planCode?: string
}

type BillingCopy = {
  monthlyLabel?: string
  annualLabel?: string
  annualBadge?: string
  annualNote?: string
}

type PricingSectionProps = {
  eyebrow?: string
  headline?: string
  headlineHighlight?: string
  subheadline?: string
  guaranteeNote?: string
  plans?: PricingPlan[]
  billing?: BillingCopy
  comparePlansHeadline?: string
  featuresLabel?: string
}

export function PricingSection({
  eyebrow,
  headline,
  headlineHighlight,
  subheadline,
  guaranteeNote,
  plans,
  billing,
  comparePlansHeadline,
  featuresLabel,
}: Readonly<PricingSectionProps> = {}) {
  const items = plans ?? []
  const [cycle, setCycle] = useState<'monthly' | 'annual'>('monthly')
  const isMd = useIsMdOrAbove()

  const hasAnnual = items.some((p) => p.priceAnnual)

  const swiperVars = {
    '--swiper-pagination-color': 'var(--pz-accent)',
    '--swiper-pagination-bullet-inactive-color': 'var(--pz-ink-2)',
  } as CSSProperties

  const renderCard = (plan: PricingPlan) => {
    const baseHref = plan.ctaHref ?? '/register'
    const ctaHref =
      plan.planCode && baseHref.includes('/register')
        ? `${baseHref}?plan=${encodeURIComponent(plan.planCode)}&cycle=${cycle === 'annual' ? 'yearly' : 'monthly'}`
        : baseHref
    return (
    <PricingCard
      name={plan.name}
      price={plan.price}
      priceAnnual={plan.priceAnnual}
      period={plan.period}
      description={plan.tagline}
      cta={plan.ctaLabel ?? 'Get Started'}
      ctaHref={ctaHref}
      features={(plan.features ?? []).map((f) => ({
        text: f.text ?? '',
        included: f.included ?? true,
      }))}
      popular={plan.highlighted ?? false}
      popularBadge={plan.popularBadge}
      billingCycle={cycle}
      annualNote={billing?.annualNote}
      featuresLabel={featuresLabel}
    />
  )
  }

  const tableColumns = items.map((plan) => ({
    name: plan.name,
    cta: plan.ctaLabel ?? 'Get Started',
    ctaHref: (() => {
      const base = plan.ctaHref ?? '/register'
      if (!plan.planCode || !base.includes('/register')) return base
      return `${base}?plan=${encodeURIComponent(plan.planCode)}&cycle=${cycle === 'annual' ? 'yearly' : 'monthly'}`
    })(),
    popular: plan.highlighted ?? false,
    features: (plan.features ?? []).map((f) => ({
      text: f.text ?? '',
      included: f.included ?? true,
    })),
  }))

  return (
    // biome-ignore lint/correctness/useUniqueElementIds: static nav anchor ID
    <SectionFrame id="pricing" bg="alt" className="relative overflow-hidden">
      {/* Gold blob — top center */}
      <div className="pointer-events-none absolute inset-0 -z-10 flex justify-center" aria-hidden>
        <div
          className="pz-blob-soft h-[400px] w-[600px] -translate-y-1/3"
          style={{
            background:
              'radial-gradient(circle, color-mix(in srgb, var(--pz-accent-gold) 18%, transparent) 0%, transparent 65%)',
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

        {/* Billing toggle */}
        {hasAnnual && (
          <div className="mt-8 inline-flex items-center gap-1 rounded-full border border-[var(--pz-card-border)] bg-white p-1 shadow-sm">
            <button
              type="button"
              onClick={() => setCycle('monthly')}
              className={`rounded-full px-5 py-2 text-sm font-medium transition-all duration-200 ${
                cycle === 'monthly'
                  ? 'bg-[var(--pz-dark-surface)] text-white shadow-sm'
                  : 'text-[var(--pz-ink-2)] hover:text-[var(--pz-ink)]'
              }`}
            >
              {billing?.monthlyLabel ?? 'Monthly'}
            </button>
            <button
              type="button"
              onClick={() => setCycle('annual')}
              className={`flex items-center gap-2 rounded-full px-5 py-2 text-sm font-medium transition-all duration-200 ${
                cycle === 'annual'
                  ? 'bg-[var(--pz-dark-surface)] text-white shadow-sm'
                  : 'text-[var(--pz-ink-2)] hover:text-[var(--pz-ink)]'
              }`}
            >
              {billing?.annualLabel ?? 'Annual'}
              <span className="rounded-full bg-[var(--pz-accent)] px-2 py-0.5 text-[10px] font-semibold text-white">
                {billing?.annualBadge ?? '2 months free'}
              </span>
            </button>
          </div>
        )}
      </div>

      {/* Mobile: stacked cards on the Y axis (no slider) */}
      <div className="mt-14 flex flex-col gap-6 md:hidden">
        {items.map((plan) => (
          <div key={plan.name} className="pt-2">
            {renderCard(plan)}
          </div>
        ))}
      </div>

      {/* Desktop: slider. Swiper is only rendered after mount + md breakpoint match
          to prevent Swiper JS from initialising on mobile (GPU layer / layout cost).
          The mobile stacked cards above are SSR-rendered and keep showing on mobile. */}
      {isMd && (
        <div className="mt-14">
          <Swiper
            modules={[Pagination, A11y]}
            className="w-full !overflow-visible !pb-12"
            style={swiperVars}
            slidesPerView="auto"
            spaceBetween={24}
            centerInsufficientSlides
            grabCursor
            pagination={{ clickable: true }}
          >
            {items.map((plan) => (
              <SwiperSlide key={plan.name} className="!h-auto !w-[300px] sm:!w-[340px]">
                <div className="h-full pt-2">{renderCard(plan)}</div>
              </SwiperSlide>
            ))}
          </Swiper>
        </div>
      )}

      {guaranteeNote && (
        <p className="mt-10 text-center text-sm leading-5 text-[var(--pz-ink-2)]">
          {guaranteeNote}
        </p>
      )}

      {/* Feature comparison table — desktop only; mobile uses the stacked cards above.
          Rendered only after mount + md match so the table DOM is never in the mobile
          document (avoids the hidden-but-present layout cost). */}
      {isMd && tableColumns.length > 0 && (
        <div className="mt-16">
          <h3 className="mb-8 text-center text-2xl font-bold tracking-tight text-[var(--pz-ink)]">
            {comparePlansHeadline ?? 'Compare plans'}
          </h3>
          <PricingComparisonTable plans={tableColumns} />
        </div>
      )}
    </SectionFrame>
  )
}

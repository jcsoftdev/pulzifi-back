'use client'

import { cn } from '@workspace/ui/lib/utils'
import { Check, X } from 'lucide-react'
import { useMemo } from 'react'
import { usePriceCountUp, useTilt } from '../../lib/gsap'
import { LandingButton } from './landing-button'

interface PricingFeature {
  text: string
  included?: boolean
}

interface PricingCardProps {
  name: string
  price: string
  priceAnnual?: string
  period?: string
  description?: string
  cta: string
  ctaHref?: string
  features: ReadonlyArray<PricingFeature>
  popular?: boolean
  popularBadge?: string
  billingCycle?: 'monthly' | 'annual'
  annualNote?: string
  featuresLabel?: string
}

function extractNumeric(price: string): number {
  const match = price.replace(/,/g, '').match(/\d+/)
  return match ? parseInt(match[0], 10) : 0
}

function PriceDisplay({
  price,
  priceAnnual,
  period: _period,
  billingCycle,
}: {
  price: string
  priceAnnual?: string
  period?: string
  billingCycle: 'monthly' | 'annual'
}) {
  const activePrice = billingCycle === 'annual' && priceAnnual ? priceAnnual : price
  const isNumeric = /\d/.test(activePrice)
  const numericVal = useMemo(
    () => extractNumeric(activePrice),
    [
      activePrice,
    ]
  )
  const prefix = isNumeric ? activePrice.replace(/[\d,]+.*/, '') : ''

  const countRef = usePriceCountUp<HTMLSpanElement>(0, numericVal, {
    duration: 1,
    prefix,
    delay: 0.1,
  })

  if (!isNumeric) {
    return (
      <span className="text-[52px] font-semibold leading-[56px] tracking-[-1.6px] text-current">
        {activePrice}
      </span>
    )
  }

  return (
    <span
      ref={countRef}
      className="text-[52px] font-semibold leading-[56px] tracking-[-1.6px] text-current"
    >
      {`${prefix}${numericVal}`}
    </span>
  )
}

export function PricingCard({
  name,
  price,
  priceAnnual,
  period,
  description,
  cta,
  ctaHref = '/register',
  features,
  popular,
  popularBadge,
  billingCycle = 'monthly',
  annualNote,
  featuresLabel,
}: Readonly<PricingCardProps>) {
  const tiltRef = useTilt<HTMLDivElement>({
    max: 5,
  })

  // A plan only has a real annual offer when its annual price is a positive
  // amount. Free ($0) has priceAnnual set but no annual concept, so it must not
  // get the annual treatment.
  const hasAnnualOffer = !!priceAnnual && extractNumeric(priceAnnual) > 0
  const hasPaidAnnual = billingCycle === 'annual' && hasAnnualOffer
  // priceAnnual is the full yearly total (Stripe unit_amount for the yearly
  // price). When annual is selected we show the per-month equivalent as the big
  // number (total / 12) and surface the yearly total in the small note below.
  const annualPrefix = priceAnnual ? priceAnnual.replace(/[\d,]+.*/, '') : ''
  const annualPerMonth = hasPaidAnnual
    ? `${annualPrefix}${Math.round(extractNumeric(priceAnnual as string) / 12)}`
    : undefined
  // Per-month framing always reads "/month" (annual just bills 12 of them up front).
  const periodLabel = period

  return (
    <div className="relative flex h-full flex-col pt-5">
      {popular && (
        <div className="absolute inset-x-0 top-0 z-10 flex justify-center">
          <span className="rounded-full bg-gradient-to-r from-[var(--pz-accent)] to-[var(--pz-accent-muted,#8b5cf6)] px-4 py-1.5 text-xs font-semibold uppercase tracking-[0.12em] text-white shadow-[0_8px_24px_-6px_rgba(106,53,224,0.6)]">
            {popularBadge ?? 'Most Popular'}
          </span>
        </div>
      )}

      <div
        ref={tiltRef}
        className={cn(
          'group/card relative flex flex-1 flex-col gap-6 overflow-hidden rounded-3xl border p-6 transition-all duration-300 hover:-translate-y-1 sm:p-8',
          popular
            ? 'border-transparent bg-[var(--pz-dark-surface)] text-white shadow-[0_24px_70px_-20px_rgba(106,53,224,0.55)] lg:scale-[1.03]'
            : 'border-[var(--pz-card-border)] bg-white text-[#111] shadow-[var(--pz-card-shadow-rest)] hover:border-[var(--pz-accent)]/30 hover:shadow-[0_24px_50px_-24px_rgba(17,17,17,0.28)]'
        )}
      >
        {/* Accent glow — featured plan only */}
        {popular && (
          <div
            aria-hidden
            className="pointer-events-none absolute -right-20 -top-24 h-56 w-56 rounded-full bg-[var(--pz-accent)] opacity-30 blur-[80px]"
          />
        )}

        <div className="relative flex flex-col gap-3.5">
          <h3
            className={cn(
              'text-xs font-semibold uppercase tracking-[0.18em]',
              popular ? 'text-white/70' : 'text-[var(--pz-accent)]'
            )}
          >
            {name}
          </h3>
          <div className="flex items-end gap-2">
            <PriceDisplay
              price={price}
              priceAnnual={annualPerMonth}
              period={period}
              billingCycle={billingCycle}
            />
            {periodLabel && (
              <span
                className={cn(
                  'pb-1.5 text-lg font-medium tracking-[-0.4px]',
                  popular ? 'text-white/50' : 'text-[#777]'
                )}
              >
                {periodLabel}
              </span>
            )}
          </div>
          {hasAnnualOffer && (
            <p
              className={cn(
                'text-xs font-medium',
                popular ? 'text-[var(--pz-accent-gold,#f59e0b)]' : 'text-[var(--pz-accent)]'
              )}
            >
              {hasPaidAnnual
                ? (annualNote ?? `Billed annually (${priceAnnual}/yr)`)
                : `Annual: ${priceAnnual}/yr`}
            </p>
          )}
          {description && (
            <p
              className={cn(
                'text-sm leading-5',
                popular ? 'text-white/65' : 'text-[var(--pz-ink-2)]'
              )}
            >
              {description}
            </p>
          )}
        </div>

        <LandingButton
          href={ctaHref}
          variant={popular ? 'primary' : 'dark'}
          className="relative w-full rounded-xl"
        >
          {cta}
        </LandingButton>

        <div
          className={cn(
            'relative flex flex-col gap-4 border-t pt-6',
            popular ? 'border-white/12' : 'border-[var(--pz-card-border)]'
          )}
        >
          <h4
            className={cn(
              'text-[11px] font-semibold uppercase tracking-[0.16em]',
              popular ? 'text-white/55' : 'text-[#999]'
            )}
          >
            {featuresLabel ?? 'Features'}
          </h4>
          <ul className="flex flex-col gap-3">
            {features.map((f) => {
              const included = f.included !== false
              return (
                <li key={f.text} className="flex items-start gap-3">
                  <span
                    className={cn(
                      'mt-0.5 flex size-5 shrink-0 items-center justify-center rounded-full',
                      included
                        ? 'bg-[var(--pz-accent)] shadow-[0_2px_8px_-2px_rgba(106,53,224,0.6)]'
                        : popular
                          ? 'bg-white/10'
                          : 'bg-black/[0.06]'
                    )}
                  >
                    {included ? (
                      <Check className="size-3 text-white" strokeWidth={3} />
                    ) : (
                      <X
                        className={cn('size-3', popular ? 'text-white/40' : 'text-[#999]')}
                        strokeWidth={3}
                      />
                    )}
                  </span>
                  <span
                    className={cn(
                      'text-[15px] leading-6',
                      included
                        ? popular
                          ? 'text-white/85'
                          : 'text-[var(--pz-ink-2)]'
                        : popular
                          ? 'text-white/35 line-through'
                          : 'text-[#aaa] line-through'
                    )}
                  >
                    {f.text}
                  </span>
                </li>
              )
            })}
          </ul>
        </div>
      </div>
    </div>
  )
}

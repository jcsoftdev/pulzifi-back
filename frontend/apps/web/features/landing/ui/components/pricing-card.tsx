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
  period,
  billingCycle,
}: {
  price: string
  priceAnnual?: string
  period?: string
  billingCycle: 'monthly' | 'annual'
}) {
  const activePrice = billingCycle === 'annual' && priceAnnual ? priceAnnual : price
  const isNumeric = /\d/.test(activePrice)
  const numericVal = useMemo(() => extractNumeric(activePrice), [activePrice])
  const prefix = isNumeric ? activePrice.replace(/[\d,]+.*/, '') : ''

  const countRef = usePriceCountUp<HTMLSpanElement>(0, numericVal, {
    duration: 1,
    prefix,
    delay: 0.1,
  })

  if (!isNumeric) {
    return (
      <span className="text-5xl font-medium leading-[56px] tracking-[-1.44px] text-[#111]">
        {activePrice}
      </span>
    )
  }

  return (
    <span
      ref={countRef}
      className="text-5xl font-medium leading-[56px] tracking-[-1.44px] text-[#111]"
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
  const tiltRef = useTilt<HTMLDivElement>({ max: 5 })

  return (
    <div className="relative pt-5">
      {popular && (
        <div className="absolute inset-x-0 top-0 flex justify-center z-1">
          <span className="rounded-full bg-[var(--pz-accent)] px-4 py-1.5 text-sm font-semibold leading-5 tracking-tight text-white shadow-sm">
            {popularBadge ?? 'Most Popular'}
          </span>
        </div>
      )}

      <div
        ref={tiltRef}
        className={cn(
          'flex flex-1 flex-col gap-6 rounded-2xl border border-[var(--pz-card-border)] bg-white p-5 shadow-[var(--pz-card-shadow-rest)] sm:p-[30px] transition-all duration-300',
          popular && 'ring-2 ring-[var(--pz-accent)] shadow-[var(--pz-shadow-accent-lg)]',
        )}
      >
        <div className="flex flex-col gap-3.5">
          <h3 className="text-xl font-medium leading-8 tracking-[-0.6px] text-[#111] capitalize">
            {name}
          </h3>
          <div className="flex items-end gap-2.5">
            <PriceDisplay
              price={price}
              priceAnnual={priceAnnual}
              period={period}
              billingCycle={billingCycle}
            />
            {period && (
              <span className="text-xl font-medium leading-8 tracking-[-0.6px] text-[#777]">
                {period}
              </span>
            )}
          </div>
          {billingCycle === 'annual' && priceAnnual && (
            <p className="text-xs text-[var(--pz-accent)] font-medium">
              {annualNote ?? 'Billed annually · 2 months free'}
            </p>
          )}
          {description && <p className="text-sm leading-5 text-[var(--pz-ink-2)]">{description}</p>}
        </div>

        <LandingButton
          href={ctaHref}
          variant={popular ? 'primary' : 'dark'}
          className="w-full rounded-[10px]"
        >
          {cta}
        </LandingButton>

        <div className="flex flex-col gap-4 p-3.5">
          <h4 className="text-xl font-medium leading-8 tracking-[-0.6px] text-[#111]">{featuresLabel ?? 'Features:'}</h4>
          <ul className="flex flex-col gap-2.5">
            {features.map((f) => {
              const included = f.included !== false
              return (
                <li key={f.text} className="flex items-center gap-3.5">
                  <span
                    className={cn(
                      'flex size-5 shrink-0 items-center justify-center rounded-full',
                      included ? 'bg-[var(--pz-dark-surface)]' : 'bg-black/10',
                    )}
                  >
                    {included ? (
                      <Check className="size-3 text-white" strokeWidth={3} />
                    ) : (
                      <X className="size-3 text-[#888]" strokeWidth={3} />
                    )}
                  </span>
                  <span
                    className={cn(
                      'text-base leading-6',
                      included ? 'text-[var(--pz-ink-2)]' : 'text-[#999] line-through',
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

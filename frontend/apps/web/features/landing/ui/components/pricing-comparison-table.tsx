'use client'

import { cn } from '@workspace/ui/lib/utils'
import { Check, Minus } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { LandingButton } from './landing-button'

interface PricingFeature {
  text: string
  included?: boolean
}

interface PlanColumn {
  name: string
  cta: string
  ctaHref?: string
  popular?: boolean
  features: ReadonlyArray<PricingFeature>
}

interface PricingComparisonTableProps {
  plans: ReadonlyArray<PlanColumn>
}

export function PricingComparisonTable({ plans }: Readonly<PricingComparisonTableProps>) {
  const headerRef = useRef<HTMLTableSectionElement>(null)
  const [sticky, setSticky] = useState(false)

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => { if (entry) setSticky(!entry.isIntersecting) },
      { threshold: 0, rootMargin: '-1px 0px 0px 0px' },
    )
    if (headerRef.current) observer.observe(headerRef.current)
    return () => observer.disconnect()
  }, [])

  // Build union of all feature labels in display order
  const allFeatures = Array.from(
    new Map(
      plans.flatMap((p) => p.features.map((f) => [f.text, f.text])),
    ).keys(),
  )

  return (
    <div className="w-full overflow-x-auto rounded-2xl border border-[var(--pz-card-border)]">
      <table className="w-full min-w-[640px] border-collapse text-sm">
        {/* sentinel row for sticky detection */}
        <thead ref={headerRef} aria-hidden className="h-0">
          <tr><th /></tr>
        </thead>

        {/* sticky header */}
        <thead
          className={cn(
            'transition-shadow duration-200',
            sticky && 'sticky top-0 z-20 shadow-md',
          )}
        >
          <tr className="bg-[var(--pz-page-bg)]">
            <th className="w-[40%] px-5 py-4 text-left text-base font-semibold text-[var(--pz-ink)]">
              Features
            </th>
            {plans.map((plan) => (
              <th
                key={plan.name}
                className={cn(
                  'px-4 py-4 text-center',
                  plan.popular && 'bg-[var(--pz-accent-pale,#f3f0ff)]',
                )}
              >
                <div className="flex flex-col items-center gap-2">
                  <span
                    className={cn(
                      'text-base font-semibold',
                      plan.popular ? 'text-[var(--pz-accent)]' : 'text-[var(--pz-ink)]',
                    )}
                  >
                    {plan.name}
                  </span>
                  <LandingButton
                    href={plan.ctaHref ?? '/register'}
                    variant={plan.popular ? 'primary' : 'dark'}
                    className="rounded-lg px-3 py-1.5 text-xs"
                  >
                    {plan.cta}
                  </LandingButton>
                </div>
              </th>
            ))}
          </tr>
        </thead>

        <tbody>
          {allFeatures.map((featureText, rowIdx) => (
            <tr
              key={featureText}
              className={cn(
                'border-t border-[var(--pz-card-border)] transition-colors duration-150 hover:bg-black/[0.02]',
                rowIdx % 2 === 0 ? 'bg-white' : 'bg-[var(--pz-page-bg)]',
              )}
            >
              <td className="px-5 py-3.5 text-[var(--pz-ink-2)]">{featureText}</td>
              {plans.map((plan) => {
                const feature = plan.features.find((f) => f.text === featureText)
                const included = feature ? (feature.included !== false) : false
                return (
                  <td
                    key={plan.name}
                    className={cn(
                      'px-4 py-3.5 text-center',
                      plan.popular && 'bg-[var(--pz-accent-pale,#f3f0ff)]/30',
                    )}
                  >
                    {included ? (
                      <Check className="mx-auto size-4 text-[var(--pz-accent)]" strokeWidth={2.5} />
                    ) : (
                      <Minus className="mx-auto size-4 text-[#ccc]" strokeWidth={2} />
                    )}
                  </td>
                )
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

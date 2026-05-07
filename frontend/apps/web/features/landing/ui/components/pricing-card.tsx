import { Check, X } from 'lucide-react'
import { cn } from '@workspace/ui/lib/utils'
import { LandingButton } from './landing-button'

interface PricingFeature {
  text: string
  included?: boolean
}

interface PricingCardProps {
  name: string
  price: string
  period?: string
  description?: string
  cta: string
  ctaHref?: string
  features: ReadonlyArray<PricingFeature>
  popular?: boolean
  popularBadge?: string
}

export function PricingCard({
  name,
  price,
  period,
  description,
  cta,
  ctaHref = '/register',
  features,
  popular,
  popularBadge,
}: Readonly<PricingCardProps>) {
  return (
    <div
      className={cn(
        'flex flex-1 flex-col gap-6 rounded-2xl border border-[var(--pz-card-border)] bg-white p-5 shadow-[var(--pz-card-shadow-rest)] sm:p-[30px] transition-all duration-300',
        popular && 'ring-2 ring-[var(--pz-accent)] md:-translate-y-2 shadow-[var(--pz-shadow-accent-lg)]',
      )}
    >
      <div className="flex flex-col gap-3.5">
        <div className="flex items-center justify-between">
          <h3 className="text-xl font-medium leading-8 tracking-[-0.6px] text-[#111] capitalize">
            {name}
          </h3>
          {popular && (
            <span className="rounded-full bg-[var(--pz-accent-pale)] px-2.5 py-1.5 text-sm font-medium leading-5 tracking-tight text-[var(--pz-dark-surface)]">
              {popularBadge ?? 'Most Popular'}
            </span>
          )}
        </div>
        <div className="flex items-end gap-2.5">
          <span className="text-5xl font-medium leading-[56px] tracking-[-1.44px] text-[#111] capitalize">
            {price}
          </span>
          {period && (
            <span className="text-xl font-medium leading-8 tracking-[-0.6px] text-[#777]">
              {period}
            </span>
          )}
        </div>
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
        <h4 className="text-xl font-medium leading-8 tracking-[-0.6px] text-[#111]">Features:</h4>
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
  )
}

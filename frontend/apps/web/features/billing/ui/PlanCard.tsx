'use client'

import { Badge } from '@workspace/ui/components/atoms/badge'
import { Button } from '@workspace/ui/components/atoms/button'
import type { Plan } from '../domain/plan'
import { formatPrice } from '../domain/plan'
import type { BillingCycle } from '../domain/subscription'

export type PlanRelation = 'current' | 'upgrade' | 'downgrade' | 'new'

interface PlanCardProps {
  plan: Plan
  billingCycle: BillingCycle
  /** How this plan relates to the user's current subscription. */
  relation: PlanRelation
  /** Enterprise plans bypass checkout/portal and route to a sales contact instead. */
  isEnterprise?: boolean
  isLoading?: boolean
  /** Visually highlight this card as the suggested default for users not on a
   *  paid plan yet (e.g. trial). Renders the same primary border as `current`
   *  but with a "Recommended" badge so the user is nudged toward an entry tier. */
  isRecommended?: boolean
  onChoose: (planId: string, billingCycle: BillingCycle, relation: PlanRelation) => void
}

function ctaLabel(relation: PlanRelation, isEnterprise: boolean, isLoading: boolean): string {
  if (isLoading) return 'Loading...'
  if (isEnterprise) return 'Contact sales'
  switch (relation) {
    case 'current':
      return 'Current plan'
    case 'upgrade':
      return 'Upgrade'
    case 'downgrade':
      return 'Downgrade'
    case 'new':
      return 'Choose plan'
  }
}

function ctaVariant(relation: PlanRelation): 'default' | 'secondary' | 'outline' {
  if (relation === 'current') return 'secondary'
  if (relation === 'downgrade') return 'outline'
  return 'default'
}

/**
 * Presentational card for a single plan in the plan selector.
 * Displays pricing for the selected billing cycle and a context-aware CTA
 * (Current / Upgrade / Downgrade / Choose / Contact sales).
 */
export function PlanCard({
  plan,
  billingCycle,
  relation,
  isEnterprise = false,
  isLoading = false,
  isRecommended = false,
  onChoose,
}: PlanCardProps) {
  const isCurrent = relation === 'current'
  const isHighlighted = isCurrent || isRecommended
  const price = billingCycle === 'monthly' ? plan.priceMonthly : plan.priceYearly
  const period = billingCycle === 'monthly' ? '/mo' : '/yr'
  const label = ctaLabel(relation, isEnterprise, isLoading)

  return (
    <div
      className={[
        'relative rounded-2xl border bg-card p-6 flex flex-col gap-4 transition-shadow',
        isHighlighted ? 'border-primary shadow-md' : 'border-border',
      ].join(' ')}
    >
      {/* Top badges. Priority: Current > Recommended > Most popular. */}
      <div className="absolute -top-3 left-6 flex gap-2">
        {isCurrent && (
          <Badge variant="default" className="text-xs">
            Current plan
          </Badge>
        )}
        {!isCurrent && isRecommended && (
          <Badge variant="default" className="text-xs">
            Recommended
          </Badge>
        )}
        {plan.isPopular && !isCurrent && !isRecommended && (
          <Badge variant="default" className="text-xs">
            Most popular
          </Badge>
        )}
      </div>

      {/* Plan name */}
      <div>
        <h4 className="text-base font-semibold text-foreground">{plan.name}</h4>
        {plan.description && (
          <p className="text-sm text-muted-foreground mt-0.5">{plan.description}</p>
        )}
      </div>

      {/* Price — Enterprise shows "Custom"; Trial shows "Free / 15 days". */}
      <div className="flex items-baseline gap-1">
        {isEnterprise ? (
          <span className="text-3xl font-bold text-foreground">Custom</span>
        ) : plan.code === 'trial' ? (
          <>
            <span className="text-3xl font-bold text-foreground">Free</span>
            <span className="text-sm text-muted-foreground">/ 15 days</span>
          </>
        ) : (
          <>
            <span className="text-3xl font-bold text-foreground">{formatPrice(price)}</span>
            <span className="text-sm text-muted-foreground">{period}</span>
          </>
        )}
      </div>

      {/* Features — one row per capability. Quantitative features render as
          "Label: value" with a primary ✓ bullet (every plan has SOMETHING
          for that row, even if just "1"). Boolean features render with ✓
          when included or ✗ when not — same row across all cards so the
          customer scans down to compare. */}
      {plan.features.length > 0 && (
        <ul className="space-y-2 text-sm flex-1">
          {plan.features.map((feature) => (
            <li key={feature.label} className="flex items-start gap-2.5">
              {feature.included ? (
                <span
                  className="mt-0.5 inline-flex h-4 w-4 shrink-0 items-center justify-center rounded-full bg-primary/15 text-primary"
                  aria-hidden
                >
                  <svg width="10" height="10" viewBox="0 0 12 12" fill="none">
                    <path
                      d="M2 6.5L5 9.5L10 3.5"
                      stroke="currentColor"
                      strokeWidth="2"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                  </svg>
                </span>
              ) : (
                <span
                  className="mt-0.5 inline-flex h-4 w-4 shrink-0 items-center justify-center rounded-full bg-muted text-muted-foreground/70"
                  aria-hidden
                >
                  <svg width="8" height="8" viewBox="0 0 12 12" fill="none">
                    <path
                      d="M3 3L9 9M9 3L3 9"
                      stroke="currentColor"
                      strokeWidth="2"
                      strokeLinecap="round"
                    />
                  </svg>
                </span>
              )}
              {feature.value !== undefined ? (
                <span className="flex-1 text-foreground">
                  <span className="text-muted-foreground">{feature.label}: </span>
                  <span className="font-medium">{feature.value}</span>
                </span>
              ) : (
                <span
                  className={
                    feature.included ? 'flex-1 text-foreground' : 'flex-1 text-muted-foreground/60'
                  }
                >
                  {feature.label}
                </span>
              )}
            </li>
          ))}
        </ul>
      )}

      {/* CTA — hidden when this card represents the user's active plan.
          A subtle "Active plan" label takes its place so the layout stays
          balanced across cards but the user is not invited to re-buy what
          they already have. */}
      {isCurrent ? (
        <div className="w-full mt-auto text-center text-sm text-muted-foreground py-2 rounded-md border border-dashed border-primary/40">
          Active plan
        </div>
      ) : (
        <Button
          type="button"
          variant={ctaVariant(relation)}
          className="w-full mt-auto"
          disabled={isLoading}
          onClick={() => onChoose(plan.id, billingCycle, relation)}
        >
          {label}
        </Button>
      )}
    </div>
  )
}

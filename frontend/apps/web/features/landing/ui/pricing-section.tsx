import { AnimatedSection } from './components/animated-section'
import { PricingCard } from './components/pricing-card'
import { SectionHeader } from './components/section-header'

type PricingSectionProps = {
  plans?: {
    name: string
    price: string
    period?: string
    features?: string[]
    ctaLabel?: string
    ctaHref?: string
    highlighted?: boolean
  }[]
}

export function PricingSection({ plans }: Readonly<PricingSectionProps> = {}) {
  const items = plans ?? []

  return (
    <section
      id="pricing"
      className="mx-auto max-w-[1256px] overflow-hidden rounded-3xl bg-[#29144c] px-6 py-12 md:px-[58px] md:py-16"
    >
      <div className="flex flex-col items-center gap-[60px]">
        <AnimatedSection>
          <SectionHeader
            title={
              <>
                Simple, <em className="font-heading">Transparent Pricing</em>
              </>
            }
            subtitle="Choose a plan that fits your business needs and budget. No hidden fees, no surprises, just straightforward pricing for powerful financial management."
            variant="dark"
          />
        </AnimatedSection>

        <div className="grid w-full gap-5 md:grid-cols-3">
          {items.map((plan, i) => (
            <AnimatedSection key={plan.name} animation="fade-up" delay={i * 150}>
              <PricingCard
                name={plan.name}
                price={plan.price}
                period={plan.period}
                description=""
                cta={plan.ctaLabel ?? 'Get Started'}
                ctaHref={plan.ctaHref ?? '/register'}
                features={plan.features ?? []}
                popular={plan.highlighted ?? false}
              />
            </AnimatedSection>
          ))}
        </div>
      </div>
    </section>
  )
}

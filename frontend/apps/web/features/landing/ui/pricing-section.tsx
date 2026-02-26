import { PRICING_PLANS } from '../lib/data'
import { AnimatedSection } from './components/animated-section'
import { PricingCard } from './components/pricing-card'
import { SectionHeader } from './components/section-header'

export function PricingSection() {
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
          {PRICING_PLANS.map((plan, i) => (
            <AnimatedSection key={plan.name} animation="fade-up" delay={i * 150}>
              <PricingCard {...plan} />
            </AnimatedSection>
          ))}
        </div>
      </div>
    </section>
  )
}

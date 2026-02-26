import { FEATURE_CARDS } from '../lib/data'
import { AnimatedSection } from './components/animated-section'
import { FeatureCard } from './components/feature-card'

export function FeaturesSection() {
  return (
    <div className="mx-auto flex max-w-[1256px] flex-col gap-2.5">
      {/* Row 1: 2 cards */}
      <div className="flex flex-col gap-2.5 md:flex-row">
        {FEATURE_CARDS.slice(0, 2).map((card, i) => (
          <AnimatedSection
            key={card.title}
            animation="fade-up"
            delay={i * 100}
            className="flex-1"
          >
            <FeatureCard {...card} />
          </AnimatedSection>
        ))}
      </div>
      {/* Row 2: 2 cards */}
      <div className="flex flex-col gap-2.5 md:flex-row">
        {FEATURE_CARDS.slice(2, 4).map((card, i) => (
          <AnimatedSection
            key={card.title}
            animation="fade-up"
            delay={i * 100 + 200}
            className="flex-1"
          >
            <FeatureCard {...card} />
          </AnimatedSection>
        ))}
      </div>
    </div>
  )
}

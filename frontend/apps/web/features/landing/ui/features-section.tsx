import { AnimatedSection } from './components/animated-section'
import { FeatureCard } from './components/feature-card'

type FeaturesSectionProps = {
  cards?: { title: string; description: string; image?: string | null }[]
}

export function FeaturesSection({ cards }: Readonly<FeaturesSectionProps> = {}) {
  const items = cards ?? []

  return (
    <div className="mx-auto flex max-w-[1256px] flex-col gap-2.5">
      {/* Row 1: 2 cards */}
      <div className="flex flex-col gap-2.5 md:flex-row">
        {items.slice(0, 2).map((card, i) => (
          <AnimatedSection
            key={card.title}
            animation="fade-up"
            delay={i * 100}
            className="flex-1"
          >
            <FeatureCard
              title={card.title}
              description={card.description}
              image={card.image ?? ''}
            />
          </AnimatedSection>
        ))}
      </div>
      {/* Row 2: 2 cards */}
      <div className="flex flex-col gap-2.5 md:flex-row">
        {items.slice(2, 4).map((card, i) => (
          <AnimatedSection
            key={card.title}
            animation="fade-up"
            delay={i * 100 + 200}
            className="flex-1"
          >
            <FeatureCard
              title={card.title}
              description={card.description}
              image={card.image ?? ''}
            />
          </AnimatedSection>
        ))}
      </div>
    </div>
  )
}

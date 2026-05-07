import { AnimatedSection } from './components/animated-section'
import { Eyebrow } from './components/eyebrow'
import { Highlight } from './components/highlight'
import { SectionFrame } from './components/section-frame'

type ProblemCard = { metric: string; label: string; description: string }

type ProblemSectionProps = {
  eyebrow?: string
  headline?: string
  headlineHighlight?: string
  cards?: ProblemCard[]
}

export function ProblemSection({
  eyebrow,
  headline,
  headlineHighlight,
  cards,
}: Readonly<ProblemSectionProps> = {}) {
  if (!cards?.length) return null
  return (
    <SectionFrame bg="alt" id="problem" className="relative overflow-hidden">
      {/* Warning-tone blob — top right */}
      <div className="pointer-events-none absolute inset-0 -z-10" aria-hidden>
        <div
          className="absolute right-0 top-0 h-[350px] w-[350px] translate-x-1/4 -translate-y-1/4"
          style={{
            background: 'radial-gradient(circle, color-mix(in srgb, var(--pz-accent-gold) 15%, transparent) 0%, transparent 65%)',
            filter: 'blur(90px)',
            opacity: 0.28,
          }}
        />
      </div>

      {/* Section header */}
      <div className="mx-auto max-w-2xl text-center">
        {eyebrow && <Eyebrow className="mb-4">{eyebrow}</Eyebrow>}
        <h2 className="font-heading text-4xl font-bold tracking-tight leading-[1.1] text-[var(--pz-ink)] md:text-5xl">
          {headline}{' '}
          {headlineHighlight && (
            <Highlight tone="accent">{headlineHighlight}</Highlight>
          )}
        </h2>
      </div>

      {/* 3-col card grid */}
      <div className="mt-14 grid grid-cols-1 gap-6 md:grid-cols-3 md:gap-8">
        {cards.map((card, i) => (
          <AnimatedSection
            as="article"
            key={card.label}
            animation="fade-up"
            delay={i * 100}
            className={[
              'rounded-xl border border-[var(--pz-card-border)] bg-white p-7',
              'shadow-[var(--pz-card-shadow-rest)]',
              'hover:shadow-[var(--pz-card-shadow-hover)] hover:-translate-y-1 transition-all duration-200',
            ].join(' ')}
          >
            <p className="font-heading text-5xl font-bold text-[var(--pz-accent)]">
              {card.metric}
            </p>
            <p className="mt-2 text-sm font-semibold uppercase tracking-wide text-[var(--pz-ink-2)]/70">
              {card.label}
            </p>
            <p className="mt-3 text-base leading-relaxed text-[var(--pz-ink-2)]">
              {card.description}
            </p>
          </AnimatedSection>
        ))}
      </div>
    </SectionFrame>
  )
}

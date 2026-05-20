import { AnimatedSection } from './components/animated-section'
import { Eyebrow } from './components/eyebrow'
import { Highlight } from './components/highlight'
import { LandingButton } from './components/landing-button'

type CtaSectionProps = {
  eyebrow?: string
  headline?: string
  headlineHighlight?: string
  subtext?: string
  primaryLabel?: string
  primaryHref?: string
  riskNote?: string
}

export function CtaSection({
  eyebrow,
  headline,
  headlineHighlight,
  subtext,
  primaryLabel,
  primaryHref,
  riskNote,
}: Readonly<CtaSectionProps> = {}) {
  return (
    // biome-ignore lint/correctness/useUniqueElementIds: static nav anchor ID
    <section id="cta" className="bg-[var(--pz-dark-surface)] py-24 text-white lg:py-32">
      <AnimatedSection animation="fade-up" className="mx-auto max-w-[900px] px-6 text-center">
        {eyebrow && (
          <Eyebrow tone="muted" className="mb-6 text-white/60">
            {eyebrow}
          </Eyebrow>
        )}

        <h2 className="font-heading text-4xl font-bold tracking-tight text-white md:text-5xl">
          {headline}{' '}
          {headlineHighlight && (
            <Highlight tone="accent-gold" italic>
              {headlineHighlight}
            </Highlight>
          )}
        </h2>

        {subtext && (
          <p className="mx-auto mt-6 max-w-xl text-lg leading-relaxed text-white/70">{subtext}</p>
        )}

        {primaryLabel && (
          <div className="mt-10">
            <LandingButton href={primaryHref ?? '/register'} variant="primary" size="lg" withArrow>
              {primaryLabel}
            </LandingButton>
          </div>
        )}

        {riskNote && <p className="mt-4 text-sm text-white/60">{riskNote}</p>}
      </AnimatedSection>
    </section>
  )
}

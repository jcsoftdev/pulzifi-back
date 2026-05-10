import { AnimatedSection } from './components/animated-section'
import { Eyebrow } from './components/eyebrow'
import { Highlight } from './components/highlight'
import { SectionFrame } from './components/section-frame'
import { StepMock } from './components/step-mock'

type Step = {
  step: number
  icon?: string
  title: string
  description: string
  mockType?: 'url' | 'insight' | 'alerts'
  mockText?: string
}

type HowItWorksSectionProps = {
  eyebrow?: string
  headline?: string
  headlineHighlight?: string
  subheadline?: string
  steps?: Step[]
}

export function HowItWorksSection({
  eyebrow,
  headline,
  headlineHighlight,
  subheadline,
  steps,
}: Readonly<HowItWorksSectionProps> = {}) {
  const items = steps ?? []

  return (
    <SectionFrame bg="white" id="how-it-works">
      {/* Section header */}
      <div className="mx-auto max-w-2xl text-center">
        {eyebrow && <Eyebrow className="mb-4">{eyebrow}</Eyebrow>}
        <h2 className="font-heading text-4xl font-bold tracking-tight leading-[1.1] text-[var(--pz-ink)] md:text-5xl">
          {headline ?? 'How it works'}
          {headlineHighlight && (
            <>
              {' '}
              <Highlight tone="accent">{headlineHighlight}</Highlight>
            </>
          )}
        </h2>
        {subheadline && (
          <p className="mt-4 text-lg leading-relaxed text-[var(--pz-ink-2)]">{subheadline}</p>
        )}
      </div>

      {/* 3-col step grid */}
      <div className="mt-14 grid grid-cols-1 gap-8 md:grid-cols-3 md:gap-10">
        {items.map((step, i) => (
          <AnimatedSection
            key={step.step}
            animation="fade-up"
            delay={i * 120}
            className="flex flex-col"
          >
            {/* Wireframe mock at top */}
            <StepMock
              mockType={step.mockType ?? 'url'}
              mockText={step.mockText}
            />

            {/* Numbered circle */}
            <div className="mt-6 flex size-8 items-center justify-center rounded-full bg-[var(--pz-accent)] text-white text-sm font-bold">
              {step.step}
            </div>

            {/* Title */}
            <h3 className="mt-4 text-xl font-semibold text-[var(--pz-ink)]">{step.title}</h3>

            {/* Body */}
            <p className="mt-2 text-base leading-relaxed text-[var(--pz-ink-2)]">
              {step.description}
            </p>
          </AnimatedSection>
        ))}
      </div>
    </SectionFrame>
  )
}

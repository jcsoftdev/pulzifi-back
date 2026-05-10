'use client'

import { Check, Sparkles } from 'lucide-react'
import { AnimatedSection } from './components/animated-section'
import { SectionHeader } from './components/section-header'

type FeaturesBullet = { title: string; description: string }
type FeaturesAction = { label: string }

type FeaturesSectionProps = {
  eyebrow?: string
  headline?: string
  headlineHighlight?: string
  intro?: string
  bullets?: FeaturesBullet[]
  demoTitle?: string
  demoBadge?: string
  demoSite?: string
  demoChange?: string
  demoAnalysis?: string
  demoActions?: FeaturesAction[]
}

export function FeaturesSection({
  eyebrow,
  headline,
  headlineHighlight,
  intro,
  bullets,
  demoTitle,
  demoBadge,
  demoSite,
  demoChange,
  demoAnalysis,
  demoActions,
}: Readonly<FeaturesSectionProps> = {}) {
  if (!bullets?.length && !demoAnalysis) return null
  return (
    <section className="mx-auto max-w-[1256px] rounded-3xl bg-white px-6 py-12 md:px-[58px] md:py-16">
      <div className="flex flex-col gap-12">
        <AnimatedSection>
          <SectionHeader
            badge={eyebrow}
            title={
              <>
                {headline}{' '}
                {headlineHighlight && (
                  <em
                    className="font-heading not-italic text-[var(--pz-accent)]"
                    style={{ fontStyle: 'italic' }}
                  >
                    {headlineHighlight}
                  </em>
                )}
              </>
            }
            subtitle={intro}
          />
        </AnimatedSection>

        <div className="grid items-start gap-10 md:grid-cols-2 md:gap-14">
          <AnimatedSection animation="slide-left" className="flex flex-col gap-6">
            {bullets?.map((b) => (
              <div key={b.title} className="flex gap-4">
                <span className="mt-1 flex size-7 shrink-0 items-center justify-center rounded-full bg-[var(--pz-accent)]/10">
                  <Check className="size-4 text-[var(--pz-accent)]" strokeWidth={3} />
                </span>
                <div>
                  <h3 className="font-heading text-xl font-medium leading-7 tracking-[-0.6px] text-[var(--pz-ink)]">
                    {b.title}
                  </h3>
                  <p className="mt-1.5 text-base leading-6 text-[var(--pz-ink-2)]">{b.description}</p>
                </div>
              </div>
            ))}
          </AnimatedSection>

          <AnimatedSection
            animation="slide-right"
            delay={120}
            className="relative rounded-3xl border border-black/5 bg-[var(--pz-page-bg)] p-6"
          >
            <div className="absolute -inset-4 -z-10 rounded-[32px] bg-[var(--pz-accent)]/10 blur-3xl" />
            {demoTitle && (
              <div className="flex items-center justify-between border-b border-black/5 pb-3">
                <span className="font-heading text-base font-medium text-[var(--pz-ink)]">
                  {demoTitle}
                </span>
                {demoBadge && (
                  <span className="inline-flex items-center gap-1.5 rounded-full bg-[var(--pz-accent-teal)]/10 px-2.5 py-1 text-[11px] font-medium text-[#0f766e]">
                    <span className="size-1.5 rounded-full bg-[var(--pz-accent-teal)]" />
                    {demoBadge}
                  </span>
                )}
              </div>
            )}
            {demoSite && (
              <p className="mt-4 text-[11px] uppercase tracking-wider text-[#888]">{demoSite}</p>
            )}
            {demoChange && (
              <p className="mt-1.5 text-base font-medium text-[var(--pz-ink)]">{demoChange}</p>
            )}
            {demoAnalysis && (
              <div className="mt-4 rounded-2xl border border-[var(--pz-accent)]/15 bg-white p-4">
                <div className="flex items-center gap-2">
                  <Sparkles className="size-4 text-[var(--pz-accent)]" />
                  <span className="text-xs font-semibold uppercase tracking-wider text-[var(--pz-dark-surface)]">
                    AI Analysis
                  </span>
                </div>
                <p className="mt-2 text-sm leading-relaxed text-[var(--pz-ink-2)]">{demoAnalysis}</p>
              </div>
            )}
            {demoActions && demoActions.length > 0 && (
              <div className="mt-4 flex flex-wrap gap-2">
                {demoActions.map((action) => (
                  <button
                    key={action.label}
                    type="button"
                    className="rounded-full border border-black/10 bg-white px-3.5 py-1.5 text-xs font-medium text-[var(--pz-ink)] transition hover:border-[var(--pz-accent)] hover:text-[var(--pz-accent)]"
                  >
                    {action.label}
                  </button>
                ))}
              </div>
            )}
          </AnimatedSection>
        </div>
      </div>
    </section>
  )
}

'use client'

import { useGSAP } from '@gsap/react'
import { cn } from '@workspace/ui/lib/utils'
import gsap from 'gsap'
import { ScrollTrigger } from 'gsap/ScrollTrigger'
import { useRef } from 'react'
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

function parseMetric(raw: string): { prefix: string; value: number; suffix: string } | null {
  const m = raw.match(/^([^0-9]*)([0-9]+(?:\.[0-9]+)?)(.*)$/)
  if (!m) return null
  return { prefix: m[1] ?? '', value: Number.parseFloat(m[2] ?? '0'), suffix: m[3] ?? '' }
}

function MetricCard({ card, index }: { card: ProblemCard; index: number }) {
  const metricRef = useRef<HTMLParagraphElement>(null)
  const parsed = parseMetric(card.metric)

  useGSAP(() => {
    if (typeof window === 'undefined') return
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return
    gsap.registerPlugin(ScrollTrigger)
    if (!metricRef.current || !parsed) return

    const obj = { val: 0 }
    gsap.to(obj, {
      val: parsed.value,
      duration: 1.4,
      ease: 'power2.out',
      delay: (index * 120) / 1000,
      scrollTrigger: {
        trigger: metricRef.current,
        start: 'top 92%',
        toggleActions: 'play none none none',
      },
      onUpdate() {
        if (!metricRef.current) return
        const display = parsed.value % 1 === 0
          ? Math.round(obj.val).toString()
          : obj.val.toFixed(1)
        metricRef.current.textContent = `${parsed.prefix}${display}${parsed.suffix}`
      },
    })
  }, { scope: metricRef })

  return (
    <AnimatedSection
      as="article"
      animation="fade-up"
      delay={index * 100}
      className={cn(
        'rounded-xl border border-[var(--pz-card-border)] bg-white p-7',
        'shadow-[var(--pz-card-shadow-rest)]',
        'hover:shadow-[var(--pz-card-shadow-hover)] hover:-translate-y-1 transition-all duration-200',
      )}
    >
      <p
        ref={metricRef}
        className="font-heading text-5xl font-bold text-[var(--pz-accent)]"
      >
        {parsed ? `${parsed.prefix}0${parsed.suffix}` : card.metric}
      </p>
      <p className="mt-2 text-sm font-semibold uppercase tracking-wide text-[var(--pz-ink-2)]/70">
        {card.label}
      </p>
      <p className="mt-3 text-base leading-relaxed text-[var(--pz-ink-2)]">
        {card.description}
      </p>
    </AnimatedSection>
  )
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

      <AnimatedSection animation="fade-up" className="mx-auto max-w-2xl text-center">
        {eyebrow && <Eyebrow className="mb-4">{eyebrow}</Eyebrow>}
        <h2 className="font-heading text-4xl font-bold tracking-tight leading-[1.1] text-[var(--pz-ink)] md:text-5xl">
          {headline}{' '}
          {headlineHighlight && (
            <Highlight tone="accent">{headlineHighlight}</Highlight>
          )}
        </h2>
      </AnimatedSection>

      <div className="mt-14 grid grid-cols-1 gap-6 md:grid-cols-3 md:gap-8">
        {cards.map((card, i) => (
          <MetricCard key={card.label} card={card} index={i} />
        ))}
      </div>
    </SectionFrame>
  )
}

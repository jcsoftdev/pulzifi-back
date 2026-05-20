'use client'

import { cn } from '@workspace/ui/lib/utils'
import Link from 'next/link'
import { useClipReveal, useHeadlineReveal, useHeroTimeline, useParallaxBlob } from '../lib/gsap'
import { type DashboardAlert, DashboardMock, type KpiItem } from './components/dashboard-mock'
import { Eyebrow } from './components/eyebrow'
import { LandingButton } from './components/landing-button'
import { UnderlineSwoosh } from './components/underline-swoosh'

type HeroSectionProps = {
  eyebrowBadge?: string
  eyebrowText?: string
  headline?: string
  headlineHighlight?: string
  subheadline?: string
  primaryCtaLabel?: string
  primaryCtaHref?: string
  secondaryCtaLabel?: string
  secondaryCtaHref?: string
  trustLine?: string
  dashboardAlerts?: DashboardAlert[]
  aiInsightTitle?: string
  aiInsightBody?: string
  kpis?: KpiItem[]
}

const fallbackAlerts: DashboardAlert[] = [
  {
    tone: 'signal',
    icon: '🔴',
    site: 'competitor.com — Pricing Change',
    title: 'Pro plan dropped from $89 → $69/mo',
    time: '2m ago',
  },
  {
    tone: 'amber',
    icon: '⚡',
    site: 'rival-startup.io — New Feature Page',
    title: '"AI Automation Suite" launched',
    time: '18m ago',
  },
  {
    tone: 'teal',
    icon: '📣',
    site: 'industry-leader.com — Messaging Shift',
    title: 'Homepage CTA changed to "Free Trial"',
    time: '1h ago',
  },
  {
    tone: 'ink',
    icon: '📄',
    site: 'regulator.gov — Policy Update',
    title: 'New compliance requirements — Q1 2026',
    time: '3h ago',
  },
]

const AVATAR_COLORS = [
  'var(--pz-accent)',
  'var(--pz-dark-surface)',
  'var(--pz-accent-teal)',
  'var(--pz-accent-gold)',
  '#0ea5e9',
]

export function HeroSection({
  eyebrowBadge,
  eyebrowText,
  headline,
  headlineHighlight,
  subheadline,
  primaryCtaLabel,
  primaryCtaHref,
  secondaryCtaLabel,
  secondaryCtaHref,
  trustLine,
  dashboardAlerts,
  aiInsightTitle,
  aiInsightBody,
  kpis,
}: Readonly<HeroSectionProps> = {}) {
  const sectionRef = useHeroTimeline<HTMLElement>()
  const blobRef = useParallaxBlob<HTMLDivElement>()
  const headlineRef = useHeadlineReveal<HTMLHeadingElement>({
    delay: 0.2,
    stagger: 0.06,
    duration: 0.95,
  })
  const dashboardRef = useClipReveal<HTMLDivElement>({
    direction: 'bottom',
    delay: 0.6,
    duration: 1.2,
    triggerOnMount: true,
  })
  const resolvedPrimaryHref = primaryCtaHref ?? '/register'
  const resolvedPrimaryLabel = primaryCtaLabel ?? 'Start Monitoring Free'
  const alerts = dashboardAlerts?.length ? dashboardAlerts : fallbackAlerts

  return (
    <section
      ref={sectionRef}
      className="relative overflow-hidden bg-white pt-32 pb-24 lg:pt-40 lg:pb-32"
    >
      {/* Gradient blob backdrop — parallax on scroll */}
      <div ref={blobRef} className="pointer-events-none absolute inset-0 -z-10" aria-hidden>
        <div
          data-pz-blob
          data-pz-blob-speed="0.12"
          className="absolute right-0 top-0 h-[600px] w-[600px] -translate-y-1/4 translate-x-1/4"
          style={{
            background: 'radial-gradient(circle, var(--pz-accent-tint) 0%, transparent 60%)',
            filter: 'blur(120px)',
            opacity: 0.3,
          }}
        />
        <div
          data-pz-blob
          data-pz-blob-speed="0.08"
          className="absolute bottom-0 left-0 h-[400px] w-[400px] translate-y-1/4 -translate-x-1/4"
          style={{
            background:
              'radial-gradient(circle, color-mix(in srgb, var(--pz-accent-teal) 15%, transparent) 0%, transparent 60%)',
            filter: 'blur(120px)',
            opacity: 0.25,
          }}
        />
      </div>

      {/* Inner frame */}
      <div className="mx-auto max-w-[1200px] px-6 lg:px-8">
        <div className="grid grid-cols-1 items-center gap-12 md:grid-cols-2 md:gap-16">
          {/* Left rail */}
          <div className="flex flex-col gap-8">
            {/* Eyebrow pill */}
            {(eyebrowBadge || eyebrowText) && (
              <div
                data-pz-hero-eyebrow
                className="inline-flex w-fit items-center gap-2 rounded-full border border-[var(--pz-card-border)] bg-white/70 px-3 py-1.5 backdrop-blur"
              >
                {eyebrowBadge && (
                  <span className="rounded bg-[var(--pz-accent)] px-1.5 py-0.5 text-[10px] font-bold uppercase tracking-wider text-white">
                    {eyebrowBadge}
                  </span>
                )}
                <Eyebrow tone="muted" className="normal-case tracking-normal text-xs">
                  {eyebrowText ?? 'Real-Time Web Monitoring'}
                </Eyebrow>
              </div>
            )}

            {/* H1 — words rise from yPercent: 110 mask via useHeadlineReveal */}
            <h1
              ref={headlineRef}
              className="font-heading text-5xl font-bold tracking-tight leading-[1.05] text-[var(--pz-ink)] md:text-6xl"
            >
              {headline ?? 'Everything you need to'}{' '}
              {headlineHighlight && (
                <span className="relative inline-block text-[var(--pz-accent)]">
                  <UnderlineSwoosh>{headlineHighlight}</UnderlineSwoosh>
                </span>
              )}
            </h1>

            {/* Subheadline */}
            <p data-pz-hero-sub className="max-w-xl text-lg leading-relaxed text-[var(--pz-ink-2)]">
              {subheadline ??
                'Automate your competitive intelligence. Get instant alerts when sites change, prices drop, or content updates.'}
            </p>

            {/* CTA pair */}
            <div data-pz-hero-cta className="flex flex-wrap items-center gap-4">
              <LandingButton
                href={resolvedPrimaryHref}
                variant="primary"
                size="lg"
                withArrow
                magnetic
                className="focus-visible:ring-2 focus-visible:ring-[var(--pz-accent)] focus-visible:ring-offset-2 focus-visible:outline-none"
              >
                {resolvedPrimaryLabel}
              </LandingButton>
              {secondaryCtaLabel && (
                <Link
                  href={secondaryCtaHref ?? '#how-it-works'}
                  className={cn(
                    'inline-flex items-center gap-1 text-sm font-medium text-[var(--pz-ink-2)]',
                    'transition-colors hover:text-[var(--pz-accent)]',
                    'focus-visible:ring-2 focus-visible:ring-[var(--pz-accent)] focus-visible:outline-none focus-visible:rounded'
                  )}
                >
                  {secondaryCtaLabel} →
                </Link>
              )}
            </div>

            {/* Trust strip */}
            {trustLine && (
              <div data-pz-hero-trust className="flex flex-wrap items-center gap-4">
                <div className="flex -space-x-2" aria-hidden>
                  {AVATAR_COLORS.map((c) => (
                    <span
                      key={c}
                      className="size-7 rounded-full border-2 border-white"
                      style={{
                        background: c,
                      }}
                    />
                  ))}
                </div>
                <div className="flex items-center gap-1" aria-hidden>
                  {[
                    1,
                    2,
                    3,
                    4,
                    5,
                  ].map((i) => (
                    // biome-ignore lint/a11y/noSvgWithoutTitle: decorative star rating icon
                    <svg
                      key={i}
                      viewBox="0 0 12 12"
                      className="size-3 fill-[var(--pz-accent-gold)]"
                    >
                      <path d="M6 0l1.5 3.9H12l-3.3 2.4 1.3 4L6 8.1 2 10.3l1.3-4L0 3.9h4.5z" />
                    </svg>
                  ))}
                </div>
                <p className="text-sm leading-5 text-[var(--pz-ink-2)]">{trustLine}</p>
              </div>
            )}
          </div>

          {/* Right rail — DashboardMock with clip-path uncover */}
          <div className="hidden md:block">
            <div ref={dashboardRef} className="relative">
              <DashboardMock
                kpis={kpis}
                alerts={alerts}
                aiInsightTitle={aiInsightTitle}
                aiInsightBody={aiInsightBody}
              />
              <div
                className="absolute -inset-6 -z-10 rounded-3xl bg-[var(--pz-accent)]/10 blur-3xl"
                aria-hidden
              />
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}

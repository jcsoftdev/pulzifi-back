'use client'

import { cn } from '@workspace/ui/lib/utils'
import Image from 'next/image'
import Link from 'next/link'
import { useClipReveal, useHeadlineReveal, useHeroTimeline, useParallaxBlob } from '../lib/gsap'
import { Eyebrow } from './components/eyebrow'
import { LandingButton } from './components/landing-button'
import { UnderlineSwoosh } from './components/underline-swoosh'

export type HeroImage = {
  url?: string | null
  alt?: string | null
  width?: number | null
  height?: number | null
}

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
  image?: HeroImage
}

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
  image,
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
          className="pz-blob-soft absolute right-0 top-0 h-[600px] w-[600px] -translate-y-1/4 translate-x-1/4"
          style={{
            background: 'radial-gradient(circle, var(--pz-accent-tint) 0%, transparent 60%)',
            filter: 'blur(120px)',
            opacity: 0.3,
          }}
        />
        <div
          data-pz-blob
          data-pz-blob-speed="0.08"
          className="pz-blob-soft absolute bottom-0 left-0 h-[400px] w-[400px] translate-y-1/4 -translate-x-1/4"
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
                <Image
                  src="/images/landing/clients-smiling.png"
                  alt="Pulzifi customers"
                  width={140}
                  height={40}
                  className="h-10 w-auto"
                />
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

          {/* Right rail — CMS hero image with clip-path uncover */}
          {image?.url && (
            <div className="hidden md:block">
              <div ref={dashboardRef} className="relative">
                <Image
                  src={image.url}
                  alt={image.alt ?? headline ?? 'Pulzifi'}
                  width={image.width ?? 1200}
                  height={image.height ?? 900}
                  className="h-auto w-full rounded-2xl"
                  priority
                  sizes="(max-width: 768px) 100vw, 50vw"
                />
              </div>
            </div>
          )}
        </div>
      </div>
    </section>
  )
}

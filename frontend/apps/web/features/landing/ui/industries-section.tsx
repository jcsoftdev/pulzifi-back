'use client'

import { useState } from 'react'
import { Sparkles } from 'lucide-react'
import { Dialog, DialogContent } from '@workspace/ui/components/atoms/dialog'
import { AnimatedSection } from './components/animated-section'
import { Eyebrow } from './components/eyebrow'
import { Highlight } from './components/highlight'
import { SectionFrame } from './components/section-frame'
import { TileTag } from './components/tile-tag'

type IndustryItem = {
  icon?: string
  title: string
  description: string
  realWin?: string
}

type IndustriesSectionProps = {
  compactMode?: boolean
  eyebrow?: string
  headline?: string
  headlineHighlight?: string
  subheadline?: string
  items?: IndustryItem[]
}

export function IndustriesSection({
  compactMode = true,
  eyebrow,
  headline,
  headlineHighlight,
  subheadline,
  items,
}: Readonly<IndustriesSectionProps> = {}) {
  const [activeItem, setActiveItem] = useState<IndustryItem | null>(null)

  if (!items?.length) return null

  if (!compactMode) {
    // Legacy 6-card grid — backwards compat
    return (
      <SectionFrame id="industries" bg="white">
        <AnimatedSection className="flex flex-col gap-12">
          <div className="mx-auto max-w-2xl text-center">
            {eyebrow && <Eyebrow className="mb-4">{eyebrow}</Eyebrow>}
            <h2 className="font-heading text-4xl font-bold tracking-tight text-[var(--pz-ink)] md:text-5xl">
              {headline}{' '}
              {headlineHighlight && <Highlight tone="accent">{headlineHighlight}</Highlight>}
            </h2>
            {subheadline && (
              <p className="mt-4 text-lg leading-relaxed text-[var(--pz-ink-2)]">{subheadline}</p>
            )}
          </div>

          <ul className="grid grid-cols-1 gap-4 sm:grid-cols-2 sm:gap-5 lg:grid-cols-3">
            {items.map((item, i) => (
              <AnimatedSection
                as="article"
                key={item.title}
                animation="fade-up"
                delay={(i % 3) * 100}
                className="group flex h-full flex-col gap-4 rounded-xl border border-[var(--pz-card-border)] bg-[var(--pz-page-bg)] p-6 transition-all duration-300 hover:-translate-y-1 hover:border-[var(--pz-accent)]/15 hover:bg-white hover:shadow-lg"
              >
                {item.icon && (
                  <span className="text-3xl" aria-hidden>
                    {item.icon}
                  </span>
                )}
                <h3 className="font-heading text-xl font-medium leading-7 tracking-[-0.6px] text-[var(--pz-ink)]">
                  {item.title}
                </h3>
                <p className="text-sm leading-relaxed text-[var(--pz-ink-2)]">{item.description}</p>
                {item.realWin && (
                  <div className="mt-auto rounded-xl border border-[var(--pz-accent-teal)]/20 bg-[#ecfdf5] p-3.5">
                    <p className="text-[11px] font-semibold uppercase tracking-wider text-[#0f766e]">
                      Real Win
                    </p>
                    <p className="mt-1.5 text-[13px] leading-relaxed text-[var(--pz-ink)]">
                      {item.realWin}
                    </p>
                  </div>
                )}
              </AnimatedSection>
            ))}
          </ul>
        </AnimatedSection>
      </SectionFrame>
    )
  }

  // Compact tile grid (default)
  return (
    <SectionFrame id="industries" bg="white">
      <div className="mx-auto max-w-2xl text-center">
        {eyebrow && <Eyebrow className="mb-4">{eyebrow}</Eyebrow>}
        <h2 className="font-heading text-4xl font-bold tracking-tight text-[var(--pz-ink)] md:text-5xl">
          {headline}{' '}
          {headlineHighlight && <Highlight tone="accent">{headlineHighlight}</Highlight>}
        </h2>
        {subheadline && (
          <p className="mt-4 text-lg leading-relaxed text-[var(--pz-ink-2)]">{subheadline}</p>
        )}
      </div>

      <div className="mt-12 grid grid-cols-2 gap-3 md:grid-cols-4 lg:grid-cols-6">
        {items.map((item, i) => (
          <AnimatedSection key={item.title} animation="fade-up" delay={i * 60}>
            <TileTag
              icon={
                item.icon ? (
                  item.icon
                ) : (
                  <Sparkles className="size-5 text-[var(--pz-accent)]" aria-hidden />
                )
              }
              label={item.title}
              onClick={() => setActiveItem(item)}
              className="w-full"
            />
          </AnimatedSection>
        ))}
      </div>

      <Dialog open={activeItem !== null} onOpenChange={(open) => !open && setActiveItem(null)}>
        <DialogContent className="max-w-md rounded-2xl p-8">
          {activeItem && (
            <div className="flex flex-col gap-5">
              {activeItem.icon ? (
                <span className="text-5xl leading-none" aria-hidden>
                  {activeItem.icon}
                </span>
              ) : (
                <Sparkles className="size-10 text-[var(--pz-accent)]" aria-hidden />
              )}
              <h3 className="font-heading text-2xl font-bold text-[var(--pz-ink)]">
                {activeItem.title}
              </h3>
              <p className="text-base leading-relaxed text-[var(--pz-ink-2)]">
                {activeItem.description}
              </p>
              {activeItem.realWin && (
                <div className="rounded-xl border border-[var(--pz-accent-teal)]/20 bg-[#ecfdf5] p-4">
                  <p className="text-[11px] font-semibold uppercase tracking-wider text-[#0f766e]">
                    Real Win
                  </p>
                  <p className="mt-2 text-sm leading-relaxed text-[var(--pz-ink)]">
                    {activeItem.realWin}
                  </p>
                </div>
              )}
            </div>
          )}
        </DialogContent>
      </Dialog>
    </SectionFrame>
  )
}

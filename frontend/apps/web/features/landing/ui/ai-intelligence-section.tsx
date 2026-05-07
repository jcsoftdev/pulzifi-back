'use client'

import { cn } from '@workspace/ui/lib/utils'
import Image from 'next/image'
import { useState } from 'react'
import { useTabCrossFade } from '../lib/gsap'
import { Eyebrow } from './components/eyebrow'
import { Highlight } from './components/highlight'
import { SectionFrame } from './components/section-frame'

export type AiIntelligenceTab = {
  label: string
  items: {
    title: string
    body?: string
    image?: string
  }[]
}

export type AiIntelligenceSectionProps = {
  eyebrow?: string
  headline?: string
  headlineHighlight?: string
  subheadline?: string
  tabs?: AiIntelligenceTab[]
}

const pillBase =
  'rounded-full px-4 py-2 text-sm font-medium transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--pz-accent)] focus-visible:ring-offset-2'
const activeStyle =
  'bg-[var(--pz-accent)] text-white shadow-[var(--pz-shadow-accent)]'
const inactiveStyle =
  'text-[var(--pz-ink-2)] hover:text-[var(--pz-ink)]'

function EmptyTab() {
  return (
    <div className="col-span-full rounded-xl border border-dashed border-[var(--pz-card-border)] p-10 text-center text-[var(--pz-ink-2)]/60 text-sm">
      No items configured for this tab yet.
    </div>
  )
}

function AiCard({
  title,
  body,
  image,
}: {
  title: string
  body?: string
  image?: string
}) {
  return (
    <article
      className={cn(
        'rounded-xl border border-[var(--pz-card-border)] bg-white p-6',
        'transition-shadow duration-200 hover:shadow-[var(--pz-card-shadow-hover)]',
      )}
    >
      {image ? (
        <div className="relative mb-5 overflow-hidden rounded-lg border border-[var(--pz-card-border)]">
          <Image
            src={image}
            alt={title}
            width={600}
            height={360}
            className="h-auto w-full object-cover"
          />
        </div>
      ) : (
        <div className="mb-5 flex aspect-[5/3] items-center justify-center rounded-lg border border-dashed border-[var(--pz-card-border)] bg-[var(--pz-page-bg-alt)] text-sm text-[var(--pz-ink-2)]/40">
          Screenshot mock
        </div>
      )}
      <h3 className="font-heading text-lg font-semibold text-[var(--pz-ink)]">{title}</h3>
      {body && (
        <p className="mt-2 text-sm leading-relaxed text-[var(--pz-ink-2)]">{body}</p>
      )}
    </article>
  )
}

export function AiIntelligenceSection({
  eyebrow,
  headline,
  headlineHighlight,
  subheadline,
  tabs = [],
}: AiIntelligenceSectionProps) {
  const [activeIndex, setActiveIndex] = useState(0)
  const panelsRef = useTabCrossFade<HTMLDivElement>(activeIndex)

  if (tabs.length === 0) {
    return (
      <SectionFrame bg="alt" id="ai-intelligence">
        <div className="text-center text-sm text-[var(--pz-ink-2)]/60">
          No AI Intelligence tabs configured yet.
        </div>
      </SectionFrame>
    )
  }

  function handleTabKeyDown(e: React.KeyboardEvent<HTMLButtonElement>) {
    const last = tabs.length - 1
    if (e.key === 'ArrowRight') {
      e.preventDefault()
      setActiveIndex((i) => (i === last ? 0 : i + 1))
    } else if (e.key === 'ArrowLeft') {
      e.preventDefault()
      setActiveIndex((i) => (i === 0 ? last : i - 1))
    } else if (e.key === 'Home') {
      e.preventDefault()
      setActiveIndex(0)
    } else if (e.key === 'End') {
      e.preventDefault()
      setActiveIndex(last)
    }
  }

  return (
    <SectionFrame bg="alt" id="ai-intelligence" className="relative overflow-hidden">
      {/* Teal blob — bottom left */}
      <div className="pointer-events-none absolute inset-0 -z-10" aria-hidden>
        <div
          className="absolute bottom-0 left-0 h-[400px] w-[500px] translate-y-1/4 -translate-x-1/4"
          style={{
            background: 'radial-gradient(circle, color-mix(in srgb, var(--pz-accent-teal) 18%, transparent) 0%, transparent 65%)',
            filter: 'blur(100px)',
            opacity: 0.3,
          }}
        />
      </div>
      {/* Section header */}
      <div className="mx-auto max-w-2xl text-center">
        {eyebrow && <Eyebrow className="mb-4">{eyebrow}</Eyebrow>}
        {headline && (
          <h2 className="font-heading text-4xl font-bold tracking-tight leading-[1.1] text-[var(--pz-ink)] md:text-5xl">
            {headline}{' '}
            {headlineHighlight && (
              <Highlight tone="accent-teal">{headlineHighlight}</Highlight>
            )}
          </h2>
        )}
        {subheadline && (
          <p className="mt-4 text-lg leading-relaxed text-[var(--pz-ink-2)]">{subheadline}</p>
        )}
      </div>

      {/* Tab list */}
      <div className="mt-10 flex justify-center">
        <div
          role="tablist"
          aria-label="AI capabilities"
          className="inline-flex gap-2 rounded-full border border-[var(--pz-card-border)] bg-white p-1.5"
        >
          {tabs.map((tab, i) => (
            <button
              key={i}
              type="button"
              role="tab"
              aria-selected={i === activeIndex}
              aria-controls={`ai-tab-panel-${i}`}
              id={`ai-tab-${i}`}
              tabIndex={i === activeIndex ? 0 : -1}
              onClick={() => setActiveIndex(i)}
              onKeyDown={handleTabKeyDown}
              className={cn(pillBase, i === activeIndex ? activeStyle : inactiveStyle)}
            >
              {tab.label}
            </button>
          ))}
        </div>
      </div>

      {/* Panels container — useTabCrossFade scoped ref */}
      <div ref={panelsRef} className="relative mt-12">
        {tabs.map((tab, i) => (
          <div
            key={i}
            data-pz-tab-panel
            role="tabpanel"
            id={`ai-tab-panel-${i}`}
            aria-labelledby={`ai-tab-${i}`}
            hidden={i !== activeIndex}
          >
            <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
              {tab.items.length === 0 ? (
                <EmptyTab />
              ) : (
                tab.items.map((item, j) => (
                  <AiCard key={j} title={item.title} body={item.body} image={item.image} />
                ))
              )}
            </div>
          </div>
        ))}
      </div>
    </SectionFrame>
  )
}

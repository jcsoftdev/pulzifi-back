'use client'

import { cn } from '@workspace/ui/lib/utils'
import {
  ArrowRight,
  BarChart3,
  Home,
  type LucideIcon,
  Megaphone,
  Rocket,
  Scale,
  ShoppingCart,
} from 'lucide-react'
import { useState } from 'react'
import { AnimatedSection } from './components/animated-section'
import { Eyebrow } from './components/eyebrow'
import { Highlight } from './components/highlight'
import { SectionFrame } from './components/section-frame'

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

const ACCENT_COLORS = [
  { bg: 'bg-violet-50', border: 'border-violet-200/60', badge: 'bg-violet-100 text-violet-700', dot: 'bg-violet-500', icon: 'text-violet-600', iconBg: 'bg-violet-100' },
  { bg: 'bg-sky-50', border: 'border-sky-200/60', badge: 'bg-sky-100 text-sky-700', dot: 'bg-sky-500', icon: 'text-sky-600', iconBg: 'bg-sky-100' },
  { bg: 'bg-emerald-50', border: 'border-emerald-200/60', badge: 'bg-emerald-100 text-emerald-700', dot: 'bg-emerald-500', icon: 'text-emerald-600', iconBg: 'bg-emerald-100' },
  { bg: 'bg-amber-50', border: 'border-amber-200/60', badge: 'bg-amber-100 text-amber-700', dot: 'bg-amber-500', icon: 'text-amber-600', iconBg: 'bg-amber-100' },
  { bg: 'bg-rose-50', border: 'border-rose-200/60', badge: 'bg-rose-100 text-rose-700', dot: 'bg-rose-500', icon: 'text-rose-600', iconBg: 'bg-rose-100' },
  { bg: 'bg-teal-50', border: 'border-teal-200/60', badge: 'bg-teal-100 text-teal-700', dot: 'bg-teal-500', icon: 'text-teal-600', iconBg: 'bg-teal-100' },
]

const ICONS: LucideIcon[] = [ShoppingCart, Megaphone, Home, Scale, Rocket, BarChart3]

function IndustryCard({ item, index, isActive, onClick }: {
  item: IndustryItem
  index: number
  isActive: boolean
  onClick: () => void
}) {
  const color = ACCENT_COLORS[index % ACCENT_COLORS.length] ?? ACCENT_COLORS[0]!
  const Icon = (ICONS[index % ICONS.length] ?? ShoppingCart) as LucideIcon

  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        'group relative flex flex-col gap-4 rounded-2xl border p-6 text-left',
        'transition-all duration-300 cursor-pointer',
        isActive
          ? `${color.bg} ${color.border} shadow-lg -translate-y-1`
          : 'bg-white border-[var(--pz-card-border)] hover:-translate-y-0.5 hover:shadow-md hover:border-[var(--pz-card-border-strong)]',
      )}
    >
      <div className="flex items-start justify-between">
        <div className={cn(
          'flex size-11 items-center justify-center rounded-xl',
          'transition-colors duration-300',
          isActive ? color.iconBg : 'bg-[var(--pz-page-bg-alt)]',
        )}>
          <Icon className={cn(
            'size-5 transition-colors duration-300',
            isActive ? color.icon : 'text-[var(--pz-ink-2)]',
          )} />
        </div>
        <div className={cn(
          'size-2 rounded-full mt-1 transition-all duration-300',
          isActive ? color.dot : 'bg-[var(--pz-card-border)]',
        )} />
      </div>

      <h3 className="font-semibold text-[17px] leading-snug tracking-tight text-[var(--pz-ink)]">
        {item.title}
      </h3>

      <p className={cn(
        'text-sm leading-relaxed transition-colors duration-300',
        isActive ? 'text-[var(--pz-ink)]' : 'text-[var(--pz-ink-2)]',
        'line-clamp-3',
      )}>
        {item.description}
      </p>

      {item.realWin && (
        <div className={cn(
          'mt-auto flex items-center gap-1.5 text-xs font-medium transition-colors duration-200',
          isActive ? color.icon : 'text-[var(--pz-ink-2)]/60 group-hover:text-[var(--pz-ink-2)]',
        )}>
          <span>Real win</span>
          <ArrowRight className="size-3" />
        </div>
      )}
    </button>
  )
}

export function IndustriesSection({
  eyebrow,
  headline,
  headlineHighlight,
  subheadline,
  items,
}: Readonly<IndustriesSectionProps> = {}) {
  const [activeIndex, setActiveIndex] = useState(0)

  if (!items?.length) return null

  const active = items[activeIndex]
  const color = ACCENT_COLORS[activeIndex % ACCENT_COLORS.length] ?? ACCENT_COLORS[0]!
  const ActiveIcon = (ICONS[activeIndex % ICONS.length] ?? ShoppingCart) as LucideIcon

  return (
    <SectionFrame id="usecases" bg="white">
      <AnimatedSection className="mx-auto max-w-2xl text-center">
        {eyebrow && <Eyebrow className="mb-4">{eyebrow}</Eyebrow>}
        <h2 className="font-heading text-4xl font-bold tracking-tight leading-[1.1] text-[var(--pz-ink)] md:text-5xl">
          {headline}{' '}
          {headlineHighlight && <Highlight tone="accent">{headlineHighlight}</Highlight>}
        </h2>
        {subheadline && (
          <p className="mt-4 text-lg leading-relaxed text-[var(--pz-ink-2)]">{subheadline}</p>
        )}
      </AnimatedSection>

      <div className="mt-16 grid grid-cols-1 gap-6 lg:grid-cols-5 lg:gap-8">
        <div className="grid grid-cols-2 gap-3 lg:col-span-3 lg:gap-4 content-start">
          {items.map((item, i) => (
            <AnimatedSection key={item.title} animation="fade-up" delay={i * 60}>
              <IndustryCard
                item={item}
                index={i}
                isActive={i === activeIndex}
                onClick={() => setActiveIndex(i)}
              />
            </AnimatedSection>
          ))}
        </div>

        <AnimatedSection
          animation="fade-up"
          delay={120}
          className="lg:col-span-2 lg:sticky lg:top-24 lg:self-start"
        >
          <div className={cn(
            'rounded-2xl border p-8 transition-all duration-500',
            color.bg, color.border,
          )}>
            <div className={cn(
              'flex size-16 items-center justify-center rounded-2xl mb-6 shadow-sm',
              color.iconBg,
            )}>
              <ActiveIcon className={cn('size-8', color.icon)} />
            </div>

            <span className={cn('inline-block rounded-full px-3 py-1 text-xs font-semibold mb-4', color.badge)}>
              Use Case {activeIndex + 1} of {items.length}
            </span>

            <h3 className="font-heading text-2xl font-bold leading-snug text-[var(--pz-ink)] mb-3">
              {active?.title}
            </h3>

            <p className="text-base leading-relaxed text-[var(--pz-ink-2)]">
              {active?.description}
            </p>

            {active?.realWin && (
              <div className="mt-6 rounded-xl bg-white/70 border border-white/80 p-5">
                <p className="text-[11px] font-bold uppercase tracking-widest text-[var(--pz-ink-2)]/60 mb-2">
                  Real Win
                </p>
                <p className="text-sm leading-relaxed text-[var(--pz-ink)] font-medium">
                  {active.realWin}
                </p>
              </div>
            )}

            <div className="mt-6 flex gap-1.5">
              {items.map((_, i) => (
                <button
                  key={i}
                  type="button"
                  onClick={() => setActiveIndex(i)}
                  className={cn(
                    'h-1.5 rounded-full transition-all duration-300',
                    i === activeIndex
                      ? `w-6 ${color.dot}`
                      : 'w-1.5 bg-[var(--pz-card-border)] hover:bg-[var(--pz-ink-2)]/30',
                  )}
                  aria-label={`Go to ${items[i]?.title ?? ''}`}
                />
              ))}
            </div>
          </div>
        </AnimatedSection>
      </div>
    </SectionFrame>
  )
}

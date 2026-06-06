import { cn } from '@workspace/ui/lib/utils'
import type { ReactNode } from 'react'
import { SectionAura } from './section-aura'
import { SectionPattern } from './section-pattern'

type BgVariant = 'white' | 'alt' | 'dark' | 'gray' | 'transparent'
type PatternVariant = 'grid' | 'dots' | 'none'

type SectionFrameProps = {
  bg?: BgVariant
  /** Decorative tiled texture behind the content. Defaults to the variant's natural pairing. */
  pattern?: PatternVariant
  id?: string
  className?: string
  children: ReactNode
}

const bgClasses: Record<BgVariant, string> = {
  white: 'bg-white',
  alt: 'bg-[var(--pz-page-bg-alt)]',
  gray: 'bg-gray-50',
  dark: 'bg-[var(--pz-dark-surface)]',
  transparent: 'bg-transparent',
}

// Grid reads on white surfaces; the faint dots read on the light-gray ones.
// Dark and transparent sections stay clean.
const defaultPattern: Record<BgVariant, PatternVariant> = {
  white: 'grid',
  alt: 'dots',
  gray: 'dots',
  dark: 'none',
  transparent: 'none',
}

const patternUrl: Record<Exclude<PatternVariant, 'none'>, string> = {
  grid: '/images/landing/pattern-grid.svg',
  dots: '/images/landing/pattern-dots.svg',
}

// The SVG only supplies the SHAPE (via mask alpha); the actual color comes from a
// design-system token so the texture re-tints with the theme. Both textures reuse
// the hairline token — one structural-line color, no magic opacity values.
const patternColor: Record<Exclude<PatternVariant, 'none'>, string> = {
  grid: 'var(--pz-pattern-line)',
  dots: 'var(--pz-pattern-line)',
}

export function SectionFrame({ bg = 'white', pattern, id, className, children }: SectionFrameProps) {
  const resolvedPattern = pattern ?? defaultPattern[bg]

  return (
    <section
      id={id}
      className={cn('relative overflow-hidden py-28 lg:py-36', bgClasses[bg], className)}
    >
      {resolvedPattern !== 'none' && (
        <>
          <SectionAura />
          <SectionPattern url={patternUrl[resolvedPattern]} color={patternColor[resolvedPattern]} />
        </>
      )}
      <div className="relative z-10 mx-auto max-w-[1200px] px-6 lg:px-8">{children}</div>
    </section>
  )
}

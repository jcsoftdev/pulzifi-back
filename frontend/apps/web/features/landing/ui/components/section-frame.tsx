import { cn } from '@workspace/ui/lib/utils'
import type { ReactNode } from 'react'

type BgVariant = 'white' | 'alt' | 'dark' | 'gray' | 'transparent'

type SectionFrameProps = {
  bg?: BgVariant
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

export function SectionFrame({ bg = 'white', id, className, children }: SectionFrameProps) {
  return (
    <section id={id} className={cn('py-28 lg:py-36', bgClasses[bg], className)}>
      <div className="mx-auto max-w-[1200px] px-6 lg:px-8">{children}</div>
    </section>
  )
}

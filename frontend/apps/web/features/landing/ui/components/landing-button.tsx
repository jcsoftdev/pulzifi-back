'use client'

import Link from 'next/link'
import { ArrowRight } from 'lucide-react'
import { cn } from '@workspace/ui/lib/utils'
import { useMagnetic } from '../../lib/gsap'

interface LandingButtonProps {
  href: string
  children: React.ReactNode
  variant?: 'primary' | 'outline' | 'dark'
  size?: 'default' | 'lg'
  withArrow?: boolean
  magnetic?: boolean
  className?: string
}

export function LandingButton({
  href,
  children,
  variant = 'primary',
  size = 'default',
  withArrow = false,
  magnetic = false,
  className,
}: Readonly<LandingButtonProps>) {
  const magneticRef = useMagnetic<HTMLSpanElement>({ enabled: magnetic, strength: 0.35 })

  const link = (
    <Link
      href={href}
      className={cn(
        'group inline-flex items-center justify-center gap-2 rounded-full font-medium',
        'transition-[background,color,box-shadow,opacity,transform] duration-300',
        'hover:scale-[1.02] active:scale-[0.98]',
        size === 'lg' ? 'h-14 px-8 text-base' : 'h-11 px-6 text-sm',
        variant === 'primary' &&
          'bg-[var(--pz-accent)] text-white hover:opacity-90 shadow-[var(--pz-shadow-accent)] hover:shadow-[var(--pz-shadow-accent-lg)]',
        variant === 'outline' &&
          'border border-[#ebebef] bg-white text-[var(--pz-ink)] hover:bg-gray-50',
        variant === 'dark' &&
          'bg-[var(--pz-dark-surface)] text-white hover:opacity-90',
        className,
      )}
    >
      {children}
      {withArrow && (
        <ArrowRight className="size-5 transition-transform duration-300 group-hover:translate-x-1" />
      )}
    </Link>
  )

  if (!magnetic) return link
  return (
    <span ref={magneticRef} className="inline-block">
      {link}
    </span>
  )
}

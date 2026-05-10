import { cn } from '@workspace/ui/lib/utils'
import type { ReactNode } from 'react'

type EyebrowTone = 'accent' | 'muted'

type EyebrowProps = {
  children: ReactNode
  tone?: EyebrowTone
  className?: string
}

const toneClasses: Record<EyebrowTone, string> = {
  accent: 'text-[var(--pz-accent)]',
  muted: 'text-gray-500',
}

export function Eyebrow({ children, tone = 'accent', className }: EyebrowProps) {
  return (
    <p
      className={cn(
        'text-xs font-semibold uppercase tracking-[0.18em]',
        toneClasses[tone],
        className,
      )}
    >
      {children}
    </p>
  )
}

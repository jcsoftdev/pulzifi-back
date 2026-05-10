import { cn } from '@workspace/ui/lib/utils'

type Tone = 'accent' | 'accent-gold' | 'accent-teal' | 'dark-surface'

interface HighlightProps {
  children: React.ReactNode
  tone?: Tone
  italic?: boolean
}

const toneClasses: Record<Tone, string> = {
  'accent': 'text-[var(--pz-accent)]',
  'accent-gold': 'text-[var(--pz-accent-gold)]',
  'accent-teal': 'text-[var(--pz-accent-teal)]',
  'dark-surface': 'text-[var(--pz-dark-surface)]',
}

export function Highlight({ children, tone = 'accent', italic = true }: Readonly<HighlightProps>) {
  return (
    <em className={cn('not-italic', toneClasses[tone], italic && 'italic')}>{children}</em>
  )
}

import { cn } from '@workspace/ui/lib/utils'
import type { ReactNode } from 'react'

type TileTagProps = {
  icon: ReactNode | string
  label: string
  onClick?: () => void
  className?: string
}

export function TileTag({ icon, label, onClick, className }: TileTagProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        'flex items-center gap-3 rounded-xl border border-[var(--pz-card-border,rgba(0,0,0,0.08))]',
        'bg-white p-4 transition',
        'hover:-translate-y-1 hover:shadow-[var(--pz-card-shadow-hover,0_8px_30px_-8px_rgba(0,0,0,0.10))]',
        'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--pz-accent)] focus-visible:ring-offset-2',
        className,
      )}
    >
      {typeof icon === 'string' ? (
        <span className="text-xl leading-none" aria-hidden>
          {icon}
        </span>
      ) : (
        icon
      )}
      <span className="text-sm font-medium text-[var(--pz-ink)]">{label}</span>
    </button>
  )
}

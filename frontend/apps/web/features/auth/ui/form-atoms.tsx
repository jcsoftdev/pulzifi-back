import { cn } from '@workspace/ui/lib/utils'

export function AuthLabel({ htmlFor, children }: { htmlFor: string; children: React.ReactNode }) {
  return (
    <label htmlFor={htmlFor} className="text-xs font-semibold uppercase tracking-wide text-[var(--pz-ink)]/50">
      {children}
    </label>
  )
}

export function ErrorBanner({ id, message }: { id: string; message: string }) {
  return (
    <div
      id={id}
      role="alert"
      className="rounded-xl border border-destructive/20 bg-destructive/5 px-4 py-3 text-sm text-destructive"
    >
      {message}
    </div>
  )
}

export function FieldMessage({
  id,
  children,
  variant = 'hint',
}: {
  id: string
  children: React.ReactNode
  variant?: 'hint' | 'error' | 'success'
}) {
  return (
    <p
      id={id}
      aria-live={variant !== 'hint' ? 'polite' : undefined}
      className={cn('text-xs', {
        'text-[var(--pz-ink)]/40': variant === 'hint',
        'text-destructive': variant === 'error',
        'text-green-600': variant === 'success',
      })}
    >
      {children}
    </p>
  )
}

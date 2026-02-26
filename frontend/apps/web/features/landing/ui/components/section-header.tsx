import { cn } from '@workspace/ui/lib/utils'

interface SectionHeaderProps {
  badge?: string
  badgeVariant?: 'light' | 'dark'
  title: React.ReactNode
  subtitle?: string
  align?: 'center' | 'left'
  variant?: 'light' | 'dark'
  className?: string
}

export function SectionHeader({
  badge,
  badgeVariant = 'light',
  title,
  subtitle,
  align = 'center',
  variant = 'light',
  className,
}: Readonly<SectionHeaderProps>) {
  const isDark = variant === 'dark'

  return (
    <div
      className={cn(
        'flex flex-col gap-4',
        align === 'center' ? 'items-center text-center' : 'items-start text-left',
        className
      )}
    >
      {badge && (
        <span
          className={cn(
            'inline-flex items-center justify-center rounded-full px-5 py-2.5 text-sm font-medium leading-5 tracking-tight',
            badgeVariant === 'light' || (!isDark && badgeVariant !== 'dark')
              ? 'bg-[#f2ebfd] text-[#29144c]'
              : 'bg-white/10 text-[#e1dbea]'
          )}
        >
          {badge}
        </span>
      )}
      <h2
        className={cn(
          'font-heading text-5xl font-medium leading-[56px] tracking-[-2.88px] max-w-3xl',
          'max-md:text-3xl max-md:leading-10 max-md:tracking-[-1.5px]',
          isDark ? 'text-white' : 'text-[#131313]'
        )}
      >
        {title}
      </h2>
      {subtitle && (
        <p
          className={cn(
            'max-w-xl text-base leading-6',
            isDark ? 'text-white/80' : 'text-[#444141]'
          )}
        >
          {subtitle}
        </p>
      )}
    </div>
  )
}

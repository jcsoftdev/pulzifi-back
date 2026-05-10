interface SectionBadgeProps {
  children: React.ReactNode
  variant?: 'light' | 'dark'
}

export function SectionBadge({ children, variant = 'light' }: Readonly<SectionBadgeProps>) {
  return (
    <span
      className={`inline-flex items-center justify-center rounded-full px-5 py-2.5 text-sm font-medium leading-5 tracking-tight ${
        variant === 'light'
          ? 'bg-[var(--pz-accent-tint)] text-[var(--pz-dark-surface)]'
          : 'bg-white/10 text-[#e1dbea]'
      }`}
    >
      {children}
    </span>
  )
}

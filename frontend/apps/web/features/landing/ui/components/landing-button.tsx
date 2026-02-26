import Link from 'next/link'
import { ArrowRight } from 'lucide-react'
import { cn } from '@workspace/ui/lib/utils'

interface LandingButtonProps {
  href: string
  children: React.ReactNode
  variant?: 'primary' | 'outline' | 'dark'
  size?: 'default' | 'lg'
  withArrow?: boolean
  className?: string
}

export function LandingButton({
  href,
  children,
  variant = 'primary',
  size = 'default',
  withArrow = false,
  className,
}: Readonly<LandingButtonProps>) {
  return (
    <Link
      href={href}
      className={cn(
        'inline-flex items-center justify-center gap-2 rounded-full font-medium transition-all duration-300',
        'hover:scale-[1.02] active:scale-[0.98]',
        size === 'lg' ? 'h-14 px-8 text-base' : 'h-11 px-6 text-sm',
        variant === 'primary' &&
          'bg-[#7c3aed] text-white hover:bg-[#6d28d9] shadow-[0_4px_20px_rgba(124,58,237,0.3)]',
        variant === 'outline' &&
          'border border-[#ebebef] bg-white text-[#121217] hover:bg-gray-50',
        variant === 'dark' &&
          'bg-[#29144c] text-white hover:bg-[#3d1d6e]',
        className
      )}
    >
      {children}
      {withArrow && <ArrowRight className="size-5 transition-transform group-hover:translate-x-0.5" />}
    </Link>
  )
}

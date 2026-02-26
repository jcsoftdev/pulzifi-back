'use client'

import { cn } from '@workspace/ui/lib/utils'
import { useInView } from '../../lib/animations'

interface AnimatedSectionProps {
  children: React.ReactNode
  className?: string
  animation?: 'fade-up' | 'fade-in' | 'slide-left' | 'slide-right' | 'scale'
  delay?: number
  as?: 'section' | 'div' | 'article'
  id?: string
}

export function AnimatedSection({
  children,
  className,
  animation = 'fade-up',
  delay = 0,
  as: Tag = 'div',
  id,
}: Readonly<AnimatedSectionProps>) {
  const [ref, isInView] = useInView<HTMLElement>()

  const animationClasses = {
    'fade-up': 'translate-y-8 opacity-0',
    'fade-in': 'opacity-0',
    'slide-left': '-translate-x-8 opacity-0',
    'slide-right': 'translate-x-8 opacity-0',
    scale: 'scale-95 opacity-0',
  }

  return (
    <Tag
      ref={ref as React.Ref<never>}
      id={id}
      className={cn(
        'transition-all duration-700 ease-out',
        isInView ? 'translate-x-0 translate-y-0 scale-100 opacity-100' : animationClasses[animation],
        className
      )}
      style={{ transitionDelay: `${delay}ms` }}
    >
      {children}
    </Tag>
  )
}

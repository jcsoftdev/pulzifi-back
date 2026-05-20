'use client'

import { useGSAP } from '@gsap/react'
import { cn } from '@workspace/ui/lib/utils'
import gsap from 'gsap'
import { ScrollTrigger } from 'gsap/ScrollTrigger'
import { useRef } from 'react'
import { usePreviewMode } from '../../lib/preview-mode'

interface AnimatedSectionProps {
  children: React.ReactNode
  className?: string
  animation?: 'fade-up' | 'fade-in' | 'slide-left' | 'slide-right' | 'scale'
  delay?: number
  as?: 'section' | 'div' | 'article' | 'li' | 'ol' | 'ul'
  id?: string
}

let registered = false
function ensureRegistered() {
  if (registered) return
  if (typeof window === 'undefined') return
  gsap.registerPlugin(useGSAP, ScrollTrigger)
  registered = true
}

function prefersReducedMotion(): boolean {
  if (typeof window === 'undefined') return false
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches
}

const FROM_VARS: Record<NonNullable<AnimatedSectionProps['animation']>, gsap.TweenVars> = {
  'fade-up': {
    y: 36,
    opacity: 0,
  },
  'fade-in': {
    opacity: 0,
  },
  'slide-left': {
    x: -80,
    opacity: 0,
  },
  'slide-right': {
    x: 80,
    opacity: 0,
  },
  scale: {
    scale: 0.92,
    opacity: 0,
  },
}

export function AnimatedSection({
  children,
  className,
  animation = 'fade-up',
  delay = 0,
  as: Tag = 'div',
  id,
}: Readonly<AnimatedSectionProps>) {
  const previewMode = usePreviewMode()
  const ref = useRef<HTMLElement | null>(null)
  ensureRegistered()

  useGSAP(
    () => {
      if (previewMode || prefersReducedMotion() || !ref.current) return
      gsap.from(ref.current, {
        ...FROM_VARS[animation],
        duration: 1,
        ease: 'power3.out',
        delay: delay / 1000,
        scrollTrigger: {
          trigger: ref.current,
          start: 'top 92%',
          toggleActions: 'play none none none',
        },
      })
    },
    {
      scope: ref,
    }
  )

  if (previewMode) {
    return (
      <Tag id={id} className={className}>
        {children}
      </Tag>
    )
  }

  return (
    <Tag ref={ref as React.Ref<never>} id={id} className={cn(className)}>
      {children}
    </Tag>
  )
}

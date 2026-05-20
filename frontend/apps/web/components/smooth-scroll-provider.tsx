'use client'

import { usePathname } from 'next/navigation'
import { type ReactNode, useEffect } from 'react'

const EXCLUDE_PREFIXES = [
  '/cms-admin',
  '/api',
]

export function SmoothScrollProvider({ children }: { children: ReactNode }) {
  const pathname = usePathname() ?? '/'
  const isExcluded = EXCLUDE_PREFIXES.some((p) => pathname.startsWith(p))

  useEffect(() => {
    if (isExcluded) return
    if (typeof window === 'undefined') return
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return

    let cleanup: (() => void) | undefined
    let cancelled = false

    void (async () => {
      const [{ default: Lenis }, { default: gsap }, { ScrollTrigger }] = await Promise.all([
        import('lenis'),
        import('gsap'),
        import('gsap/ScrollTrigger'),
      ])
      if (cancelled) return

      gsap.registerPlugin(ScrollTrigger)

      const lenis = new Lenis({
        lerp: 0.1,
        smoothWheel: true,
      })
      const onScroll = () => ScrollTrigger.update()
      lenis.on('scroll', onScroll)

      const tick = (time: number) => {
        lenis.raf(time * 1000)
      }
      gsap.ticker.add(tick)
      gsap.ticker.lagSmoothing(0)

      cleanup = () => {
        lenis.off('scroll', onScroll)
        gsap.ticker.remove(tick)
        lenis.destroy()
      }
    })()

    return () => {
      cancelled = true
      cleanup?.()
    }
  }, [
    isExcluded,
  ])

  return <>{children}</>
}

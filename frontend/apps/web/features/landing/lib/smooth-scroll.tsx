'use client'

import gsap from 'gsap'
import { ScrollTrigger } from 'gsap/ScrollTrigger'
import Lenis from 'lenis'
import { useEffect } from 'react'

export function SmoothScroll() {
  useEffect(() => {
    if (typeof window === 'undefined') return
    const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches
    // Smooth scroll (Lenis) is desktop-only. On touch devices the RAF-driven
    // scroll sync runs every frame and thrashes the mobile compositor; combined
    // with the GPU layers from the blobs and scroll reveals it exhausts layer
    // memory and paints the viewport black (iOS Safari / Android Chrome). Native
    // scroll still drives ScrollTrigger, so the reveals keep working on mobile.
    const coarsePointer = window.matchMedia('(pointer: coarse)').matches
    if (reduced || coarsePointer) return

    gsap.registerPlugin(ScrollTrigger)

    const lenis = new Lenis({
      duration: 1.15,
      easing: (t) => Math.min(1, 1.001 - 2 ** (-10 * t)),
      smoothWheel: true,
      touchMultiplier: 1.4,
    })

    lenis.on('scroll', ScrollTrigger.update)

    const tick = (time: number) => {
      lenis.raf(time * 1000)
    }
    gsap.ticker.add(tick)
    gsap.ticker.lagSmoothing(0)

    return () => {
      gsap.ticker.remove(tick)
      lenis.destroy()
    }
  }, [])

  return null
}

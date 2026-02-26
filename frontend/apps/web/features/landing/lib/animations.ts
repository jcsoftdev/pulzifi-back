'use client'

import { useCallback, useEffect, useRef, useState, type RefObject } from 'react'

export function useInView<T extends HTMLElement = HTMLDivElement>(
  options?: IntersectionObserverInit
): [RefObject<T | null>, boolean] {
  const ref = useRef<T | null>(null)
  const [isInView, setIsInView] = useState(false)

  useEffect(() => {
    const el = ref.current
    if (!el) return

    const observer = new IntersectionObserver(
      (entries) => {
        const entry = entries[0]
        if (entry?.isIntersecting) {
          setIsInView(true)
          observer.unobserve(el)
        }
      },
      { threshold: 0.15, ...options }
    )

    observer.observe(el)
    return () => observer.disconnect()
  }, [options])

  return [ref, isInView]
}

export function useCountUp(end: number, duration = 2000, startWhenVisible = false) {
  const [count, setCount] = useState(end)
  const [started, setStarted] = useState(!startWhenVisible)

  useEffect(() => {
    if (!started) return
    let startTime: number | null = null
    let frame: number

    setCount(0)

    const animate = (timestamp: number) => {
      if (!startTime) startTime = timestamp
      const progress = Math.min((timestamp - startTime) / duration, 1)
      const eased = 1 - (1 - progress) ** 3
      setCount(Math.floor(eased * end))
      if (progress < 1) {
        frame = requestAnimationFrame(animate)
      }
    }

    frame = requestAnimationFrame(animate)
    return () => cancelAnimationFrame(frame)
  }, [end, duration, started])

  const start = useCallback(() => setStarted(true), [])

  return { count, start }
}

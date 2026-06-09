'use client'

import { useGSAP } from '@gsap/react'
import gsap from 'gsap'
import { ScrollTrigger } from 'gsap/ScrollTrigger'
import { type RefObject, useRef } from 'react'
import { usePreviewMode } from './preview-mode'

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

/** Animates blobs inside the container with parallax on scroll PLUS organic
 *  idle drift. Parallax uses `y` (px), drift uses `xPercent/yPercent` so the
 *  two transforms compose without conflicting.
 *  Targets: children with `data-pz-blob` attribute. */
export function useParallaxBlob<T extends HTMLElement = HTMLDivElement>(options?: {
  speed?: number
  drift?: boolean
}): RefObject<T | null> {
  const ref = useRef<T | null>(null)
  const previewMode = usePreviewMode()
  ensureRegistered()
  useGSAP(
    () => {
      if (previewMode || prefersReducedMotion() || !ref.current) return
      // Skip on touch devices. Scroll-animating a large filter:blur() blob
      // exhausts the mobile GPU's layer memory and paints the viewport black
      // (iOS Safari / Android Chrome). The blob stays static instead.
      if (typeof window !== 'undefined' && window.matchMedia('(hover: none)').matches) return
      const blobs = ref.current.querySelectorAll<HTMLElement>('[data-pz-blob]')
      if (!blobs.length) return
      const drift = options?.drift ?? true
      blobs.forEach((blob) => {
        const speed = Number(blob.dataset.pzBlobSpeed ?? options?.speed ?? 0.14)
        const distance = Math.round(60 * speed)
        gsap.fromTo(
          blob,
          {
            y: 0,
          },
          {
            y: distance,
            ease: 'none',
            scrollTrigger: {
              trigger: ref.current,
              start: 'top bottom',
              end: 'bottom top',
              scrub: true,
            },
          }
        )
        if (drift) {
          gsap.to(blob, {
            xPercent: 'random(-10, 10)',
            yPercent: 'random(-10, 10)',
            duration: 'random(8, 14)',
            ease: 'sine.inOut',
            repeat: -1,
            yoyo: true,
            repeatRefresh: true,
          })
        }
      })
    },
    {
      scope: ref,
    }
  )
  return ref
}

/** Subtle scroll parallax for a decorative texture overlay. Transform-only
 *  (`yPercent`, GPU-composited) so it never triggers layout or paint. The
 *  overlay must overflow its section box (e.g. `-inset-y-[15%]`) to hide the
 *  travel. Triggers on the overlay's parent section so the scrub spans the
 *  whole section. No-ops on reduced motion / preview mode. */
export function useParallaxPattern<T extends HTMLElement = HTMLDivElement>(options?: {
  range?: number
}): RefObject<T | null> {
  const ref = useRef<T | null>(null)
  const previewMode = usePreviewMode()
  ensureRegistered()
  useGSAP(
    () => {
      if (previewMode || prefersReducedMotion() || !ref.current) return
      const el = ref.current
      const range = options?.range ?? 6
      gsap.fromTo(
        el,
        {
          yPercent: -range,
        },
        {
          yPercent: range,
          ease: 'none',
          scrollTrigger: {
            trigger: el.parentElement ?? el,
            start: 'top bottom',
            end: 'bottom top',
            scrub: true,
          },
        }
      )
    },
    {
      scope: ref,
    }
  )
  return ref
}

/** GSAP-powered stagger reveal for card grids.
 *  Targets children with `data-pz-card` inside the scoped container. */
export function useCardStagger<T extends HTMLElement = HTMLDivElement>(options?: {
  y?: number
  stagger?: number
  start?: string
  scale?: boolean
}): RefObject<T | null> {
  const ref = useRef<T | null>(null)
  const previewMode = usePreviewMode()
  ensureRegistered()
  useGSAP(
    () => {
      if (previewMode || prefersReducedMotion() || !ref.current) return
      const cards = ref.current.querySelectorAll('[data-pz-card]')
      if (!cards.length) return
      gsap.from(cards, {
        opacity: 0,
        y: options?.y ?? 40,
        scale: options?.scale ? 0.96 : 1,
        duration: 0.7,
        ease: 'power3.out',
        stagger: options?.stagger ?? 0.1,
        scrollTrigger: {
          trigger: ref.current,
          start: options?.start ?? 'top 80%',
          toggleActions: 'play none none none',
        },
      })
    },
    {
      scope: ref,
    }
  )
  return ref
}

export function useScrollReveal<T extends HTMLElement = HTMLDivElement>(
  selector = '[data-pz-anim]',
  options?: {
    y?: number
    stagger?: number
    start?: string
  }
): RefObject<T | null> {
  const ref = useRef<T | null>(null)
  const previewMode = usePreviewMode()
  ensureRegistered()
  useGSAP(
    () => {
      if (previewMode || prefersReducedMotion() || !ref.current) return
      const targets = ref.current.querySelectorAll(selector)
      if (!targets.length) return
      gsap.from(targets, {
        opacity: 0,
        y: options?.y ?? 32,
        duration: 0.9,
        ease: 'power3.out',
        stagger: options?.stagger ?? 0.08,
        scrollTrigger: {
          trigger: ref.current,
          start: options?.start ?? 'top 82%',
          toggleActions: 'play none none none',
        },
      })
    },
    {
      scope: ref,
    }
  )
  return ref
}

export function useHeroTimeline<T extends HTMLElement = HTMLDivElement>(): RefObject<T | null> {
  const ref = useRef<T | null>(null)
  const previewMode = usePreviewMode()
  ensureRegistered()
  useGSAP(
    () => {
      if (previewMode || prefersReducedMotion() || !ref.current) return
      const tl = gsap.timeline({
        defaults: {
          ease: 'power3.out',
        },
      })
      tl.from('[data-pz-hero-eyebrow]', {
        opacity: 0,
        y: 12,
        duration: 0.6,
      })
        // Title is animated by useHeadlineReveal — owns its own mask reveal.
        .from(
          '[data-pz-hero-sub]',
          {
            opacity: 0,
            y: 16,
            duration: 0.7,
          },
          0.6
        )
        .from(
          '[data-pz-hero-cta]',
          {
            opacity: 0,
            y: 12,
            duration: 0.6,
          },
          0.85
        )
        .from(
          '[data-pz-hero-trust]',
          {
            opacity: 0,
            y: 8,
            duration: 0.5,
          },
          1.0
        )
      // Card is animated by useClipReveal.
    },
    {
      scope: ref,
    }
  )
  return ref
}

export function useTabCrossFade<T extends HTMLElement = HTMLDivElement>(
  activeIndex: number,
  options?: {
    duration?: number
  }
): RefObject<T | null> {
  const ref = useRef<T | null>(null)
  const previewMode = usePreviewMode()
  ensureRegistered()
  useGSAP(
    () => {
      if (!ref.current) return
      const panels = ref.current.querySelectorAll('[data-pz-tab-panel]')
      if (previewMode || prefersReducedMotion()) {
        panels.forEach((p, i) => {
          ;(p as HTMLElement).style.display = i === activeIndex ? 'block' : 'none'
        })
        return
      }
      panels.forEach((p, i) => {
        const el = p as HTMLElement
        if (i === activeIndex) {
          el.style.display = 'block'
          gsap.fromTo(
            el,
            {
              opacity: 0,
            },
            {
              opacity: 1,
              duration: options?.duration ?? 0.25,
              ease: 'power2.out',
            }
          )
        } else {
          el.style.display = 'none'
        }
      })
    },
    {
      scope: ref,
      dependencies: [
        activeIndex,
      ],
    }
  )
  return ref
}

/** Splits text-node descendants of `root` into word-masks. Element children
 *  (e.g. spans for highlighted words) are preserved — only their text content
 *  is split. Returns the list of inner spans that should be animated. */
function splitWordsInPlace(root: HTMLElement): HTMLElement[] {
  const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT)
  const textNodes: Text[] = []
  let node: Node | null = walker.nextNode()
  while (node) {
    textNodes.push(node as Text)
    node = walker.nextNode()
  }
  const inners: HTMLElement[] = []
  textNodes.forEach((tn) => {
    const text = tn.textContent ?? ''
    if (!text || !text.trim()) return
    const parent = tn.parentNode
    if (!parent) return
    const frag = document.createDocumentFragment()
    text.split(/(\s+)/).forEach((part) => {
      if (!part) return
      if (/^\s+$/.test(part)) {
        frag.appendChild(document.createTextNode(part))
        return
      }
      const mask = document.createElement('span')
      mask.style.display = 'inline-block'
      mask.style.overflow = 'hidden'
      mask.style.verticalAlign = 'baseline'
      mask.style.lineHeight = 'inherit'
      const inner = document.createElement('span')
      inner.style.display = 'inline-block'
      inner.style.willChange = 'transform'
      inner.textContent = part
      mask.appendChild(inner)
      frag.appendChild(mask)
      inners.push(inner)
    })
    parent.replaceChild(frag, tn)
  })
  return inners
}

/** Mask-reveal for headlines. Words rise from yPercent: 110 to 0 with stagger.
 *  After all words land, optionally draws an underline-swoosh stroke marked
 *  with `data-pz-swoosh path`. Cleanup restores the original innerHTML so
 *  Strict Mode / re-renders don't compound mutations. */
export function useHeadlineReveal<T extends HTMLElement = HTMLHeadingElement>(options?: {
  stagger?: number
  duration?: number
  delay?: number
}): RefObject<T | null> {
  const ref = useRef<T | null>(null)
  const previewMode = usePreviewMode()
  ensureRegistered()
  useGSAP(
    () => {
      if (!ref.current) return
      if (previewMode || prefersReducedMotion()) return
      const el = ref.current
      const original = el.innerHTML
      const inners = splitWordsInPlace(el)
      if (!inners.length)
        return () => {
          el.innerHTML = original
        }

      const stagger = options?.stagger ?? 0.05
      const duration = options?.duration ?? 0.9
      const delay = options?.delay ?? 0.1

      gsap.set(inners, {
        yPercent: 110,
      })
      gsap.to(inners, {
        yPercent: 0,
        duration,
        ease: 'power4.out',
        stagger,
        delay,
        // Release the GPU layers once the words have landed. Leaving
        // will-change:transform on every word keeps dozens of promoted layers
        // alive — a needless drain on mobile compositor memory.
        onComplete: () => {
          for (const inner of inners) inner.style.willChange = 'auto'
        },
      })

      const swoosh = el.querySelector<SVGPathElement>('[data-pz-swoosh] path')
      if (swoosh && typeof swoosh.getTotalLength === 'function') {
        const length = swoosh.getTotalLength()
        gsap.set(swoosh, {
          strokeDasharray: length,
          strokeDashoffset: length,
        })
        gsap.to(swoosh, {
          strokeDashoffset: 0,
          duration: 0.65,
          ease: 'power2.out',
          delay: delay + inners.length * stagger + 0.15,
        })
      }

      return () => {
        el.innerHTML = original
      }
    },
    {
      scope: ref,
    }
  )
  return ref
}

/** Clip-path reveal — element uncovers from one edge. Triggered when scrolled
 *  into view. Composes with other transforms (uses clip-path, not transform). */
export function useClipReveal<T extends HTMLElement = HTMLDivElement>(options?: {
  delay?: number
  start?: string
  duration?: number
  direction?: 'top' | 'bottom' | 'left' | 'right'
  triggerOnMount?: boolean
}): RefObject<T | null> {
  const ref = useRef<T | null>(null)
  const previewMode = usePreviewMode()
  ensureRegistered()
  useGSAP(
    () => {
      if (previewMode || prefersReducedMotion() || !ref.current) return
      const dir = options?.direction ?? 'bottom'
      const from =
        dir === 'bottom'
          ? 'inset(0 0 100% 0)'
          : dir === 'top'
            ? 'inset(100% 0 0 0)'
            : dir === 'left'
              ? 'inset(0 100% 0 0)'
              : 'inset(0 0 0 100%)'
      const tween: gsap.TweenVars = {
        clipPath: 'inset(0 0 0 0)',
        duration: options?.duration ?? 1.1,
        ease: 'power4.out',
        delay: options?.delay ?? 0.4,
      }
      if (!options?.triggerOnMount) {
        tween.scrollTrigger = {
          trigger: ref.current,
          start: options?.start ?? 'top 90%',
          toggleActions: 'play none none none',
        }
      }
      gsap.fromTo(
        ref.current,
        {
          clipPath: from,
          willChange: 'clip-path',
        },
        tween
      )
    },
    {
      scope: ref,
    }
  )
  return ref
}

/** Animates a numeric price string from `from` to `to` when the element enters
 *  the viewport. Writes the formatted value into the DOM node's textContent.
 *  Skipped on reduced motion. */
export function usePriceCountUp<T extends HTMLElement = HTMLElement>(
  from: number,
  to: number,
  options?: {
    duration?: number
    prefix?: string
    suffix?: string
    delay?: number
  }
): RefObject<T | null> {
  const ref = useRef<T | null>(null)
  const previewMode = usePreviewMode()
  ensureRegistered()
  useGSAP(
    () => {
      if (previewMode || prefersReducedMotion() || !ref.current) return
      const el = ref.current
      const prefix = options?.prefix ?? ''
      const suffix = options?.suffix ?? ''
      const obj = {
        val: from,
      }
      gsap.to(obj, {
        val: to,
        duration: options?.duration ?? 1.2,
        ease: 'power2.out',
        delay: options?.delay ?? 0,
        onUpdate() {
          el.textContent = `${prefix}${Math.round(obj.val)}${suffix}`
        },
        scrollTrigger: {
          trigger: el,
          start: 'top 85%',
          toggleActions: 'play none none none',
        },
      })
    },
    {
      scope: ref,
      dependencies: [
        from,
        to,
      ],
    }
  )
  return ref
}

/** Magnetic hover — element drifts toward pointer within its bounding rect.
 *  Uses gsap.quickTo for 60fps pointer tracking. Resets on pointerleave. */
export function useMagnetic<T extends HTMLElement = HTMLElement>(options?: {
  strength?: number
  enabled?: boolean
}): RefObject<T | null> {
  const ref = useRef<T | null>(null)
  const previewMode = usePreviewMode()
  ensureRegistered()
  useGSAP(
    () => {
      if (options?.enabled === false) return
      if (previewMode || prefersReducedMotion() || !ref.current) return
      // Skip on touch devices — pointer attraction is meaningless without a mouse.
      if (typeof window !== 'undefined' && window.matchMedia('(hover: none)').matches) return

      const el = ref.current
      const strength = options?.strength ?? 0.35
      const xTo = gsap.quickTo(el, 'x', {
        duration: 0.4,
        ease: 'power3.out',
      })
      const yTo = gsap.quickTo(el, 'y', {
        duration: 0.4,
        ease: 'power3.out',
      })

      const onMove = (e: PointerEvent) => {
        const rect = el.getBoundingClientRect()
        const cx = rect.left + rect.width / 2
        const cy = rect.top + rect.height / 2
        xTo((e.clientX - cx) * strength)
        yTo((e.clientY - cy) * strength)
      }
      const onLeave = () => {
        xTo(0)
        yTo(0)
      }

      el.addEventListener('pointermove', onMove)
      el.addEventListener('pointerleave', onLeave)
      return () => {
        el.removeEventListener('pointermove', onMove)
        el.removeEventListener('pointerleave', onLeave)
      }
    },
    {
      scope: ref,
      dependencies: [
        options?.enabled,
      ],
    }
  )
  return ref
}

/** 3D tilt on pointer move. Element rotates ±max degrees around X/Y axes
 *  based on pointer position within its bounding rect. Resets on leave.
 *  Skipped on touch devices and when reduced motion is requested. */
export function useTilt<T extends HTMLElement = HTMLElement>(options?: {
  max?: number
  perspective?: number
  enabled?: boolean
}): RefObject<T | null> {
  const ref = useRef<T | null>(null)
  const previewMode = usePreviewMode()
  ensureRegistered()
  useGSAP(
    () => {
      if (options?.enabled === false) return
      if (previewMode || prefersReducedMotion() || !ref.current) return
      if (typeof window !== 'undefined' && window.matchMedia('(hover: none)').matches) return

      const el = ref.current
      const max = options?.max ?? 6
      const perspective = options?.perspective ?? 900

      gsap.set(el, {
        transformPerspective: perspective,
      })

      const rxTo = gsap.quickTo(el, 'rotationX', {
        duration: 0.4,
        ease: 'power3.out',
      })
      const ryTo = gsap.quickTo(el, 'rotationY', {
        duration: 0.4,
        ease: 'power3.out',
      })

      const onMove = (e: PointerEvent) => {
        const rect = el.getBoundingClientRect()
        const x = (e.clientX - rect.left) / rect.width
        const y = (e.clientY - rect.top) / rect.height
        rxTo((y - 0.5) * -2 * max)
        ryTo((x - 0.5) * 2 * max)
      }
      const onLeave = () => {
        rxTo(0)
        ryTo(0)
      }

      el.addEventListener('pointermove', onMove)
      el.addEventListener('pointerleave', onLeave)
      return () => {
        el.removeEventListener('pointermove', onMove)
        el.removeEventListener('pointerleave', onLeave)
      }
    },
    {
      scope: ref,
      dependencies: [
        options?.enabled,
      ],
    }
  )
  return ref
}

/** Step cascade — sequentially highlights `[data-pz-step]` cards as the user
 *  scrolls into the section. Optional `[data-pz-step-line]` draws beneath
 *  them; `[data-pz-step-number]` pops with a back-out ease. */
export function useStepCascade<T extends HTMLElement = HTMLElement>(): RefObject<T | null> {
  const ref = useRef<T | null>(null)
  const previewMode = usePreviewMode()
  ensureRegistered()
  useGSAP(
    () => {
      if (previewMode || prefersReducedMotion() || !ref.current) return
      const root = ref.current
      const steps = root.querySelectorAll<HTMLElement>('[data-pz-step]')
      const numbers = root.querySelectorAll<HTMLElement>('[data-pz-step-number]')
      const line = root.querySelector<HTMLElement>('[data-pz-step-line]')
      if (!steps.length) return

      const tl = gsap.timeline({
        scrollTrigger: {
          trigger: root,
          start: 'top 70%',
          toggleActions: 'play none none none',
        },
        defaults: {
          ease: 'power3.out',
        },
      })

      if (line) {
        tl.fromTo(
          line,
          {
            scaleX: 0,
            transformOrigin: 'left center',
          },
          {
            scaleX: 1,
            duration: 1.2,
            ease: 'power2.inOut',
          },
          0
        )
      }
      tl.from(
        steps,
        {
          y: 30,
          opacity: 0,
          duration: 0.7,
          stagger: 0.18,
        },
        0.1
      )
      if (numbers.length) {
        tl.from(
          numbers,
          {
            scale: 0.4,
            opacity: 0,
            duration: 0.5,
            stagger: 0.18,
            ease: 'back.out(2)',
          },
          0.2
        )
      }
    },
    {
      scope: ref,
    }
  )
  return ref
}

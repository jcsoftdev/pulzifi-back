'use client'

import dynamic from 'next/dynamic'

// Loaded client-side only to keep SmoothScroll (~90KB lenis+gsap) off the critical render path.
// `ssr: false` requires a Client Component boundary, so this wrapper exists to be imported
// from Server Components (apex, dynamic [slug], blog) without leaking `dynamic({ ssr: false })`
// into the server module graph.
const SmoothScroll = dynamic(() => import('./smooth-scroll').then((m) => m.SmoothScroll), {
  ssr: false,
})

export function SmoothScrollLazy() {
  return <SmoothScroll />
}

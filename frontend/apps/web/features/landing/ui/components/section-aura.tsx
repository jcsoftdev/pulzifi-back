'use client'

import { useParallaxBlob } from '../../lib/gsap'

type SectionAuraProps = {
  /** Strength of the orbs. 'soft' for light surfaces, 'bold' for dark ones. */
  intensity?: 'soft' | 'bold'
}

// Brand glow orbs blurred behind the content. Both orbs stay in the accent
// (purple) family — accent + accent-muted — so the wash is unmistakably on-brand
// and re-tints with the theme. Motion is scroll parallax only (no idle loop).
export function SectionAura({ intensity = 'soft' }: SectionAuraProps) {
  const ref = useParallaxBlob<HTMLDivElement>({ drift: false })
  const accentPct = intensity === 'bold' ? 26 : 13
  const mutedPct = intensity === 'bold' ? 22 : 10

  return (
    <div ref={ref} aria-hidden className="pointer-events-none absolute inset-0 z-0 overflow-hidden">
      <div
        data-pz-blob
        data-pz-blob-speed="0.12"
        className="pz-blob-soft absolute -top-1/3 right-[-12%] h-[520px] w-[520px] rounded-full"
        style={{
          background: `radial-gradient(circle, color-mix(in srgb, var(--pz-accent) ${accentPct}%, transparent) 0%, transparent 70%)`,
          filter: 'blur(100px)',
        }}
      />
      <div
        data-pz-blob
        data-pz-blob-speed="0.07"
        className="pz-blob-soft absolute bottom-[-30%] left-[-12%] h-[460px] w-[460px] rounded-full"
        style={{
          background: `radial-gradient(circle, color-mix(in srgb, var(--pz-accent-muted) ${mutedPct}%, transparent) 0%, transparent 70%)`,
          filter: 'blur(100px)',
        }}
      />
    </div>
  )
}

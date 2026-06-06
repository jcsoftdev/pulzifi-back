'use client'

import type { CSSProperties } from 'react'
import { useParallaxPattern } from '../../lib/gsap'

type SectionPatternProps = {
  /** Mask SVG that supplies the texture shape (alpha only). */
  url: string
  /** Design-system token painted through the mask. */
  color: string
}

// Radial spotlight: texture is strongest toward the top-center and vignettes out
// to the edges, so it reads as depth rather than a flat tiled wallpaper. Doubles
// as the seam fade — it reaches transparent before the section's edges.
const spotlight = 'radial-gradient(115% 78% at 50% 26%, #000 0%, #000 40%, transparent 82%)'

export function SectionPattern({ url, color }: SectionPatternProps) {
  const ref = useParallaxPattern<HTMLDivElement>()

  const textureStyle: CSSProperties = {
    backgroundColor: color,
    // The SVG masks the shape; the spotlight is a second mask layer, intersected
    // so the painted token color shows only where both are opaque.
    maskImage: `url('${url}'), ${spotlight}`,
    maskRepeat: 'repeat, no-repeat',
    maskComposite: 'intersect',
    WebkitMaskImage: `url('${url}'), ${spotlight}`,
    WebkitMaskRepeat: 'repeat, no-repeat',
    WebkitMaskComposite: 'source-in',
  }

  return (
    <div
      ref={ref}
      aria-hidden
      // Overflows the section box top/bottom so the parallax travel never
      // exposes an edge.
      className="pointer-events-none absolute inset-x-0 -inset-y-[15%] z-0"
      style={textureStyle}
    />
  )
}

import { memo } from 'react'

export const GooeyDefs = memo(function GooeyDefs({
  filterId,
  blur,
}: {
  filterId: string
  blur: number
}) {
  return (
    <defs>
      <filter
        id={filterId}
        x="-20%"
        y="-20%"
        width="140%"
        height="140%"
        colorInterpolationFilters="sRGB"
      >
        <feGaussianBlur in="SourceGraphic" stdDeviation={blur} result="blur" />
        <feColorMatrix
          in="blur"
          mode="matrix"
          values="1 0 0 0 0  0 1 0 0 0  0 0 1 0 0  0 0 0 20 -10"
          result="goo"
        />
        <feComposite in="SourceGraphic" in2="goo" operator="atop" />
      </filter>
    </defs>
  )
})

import type { ReactNode } from 'react'

type UnderlineSwooshProps = {
  children: ReactNode
}

export function UnderlineSwoosh({ children }: UnderlineSwooshProps) {
  return (
    <span className="relative inline-block italic">
      {children}
      {/* biome-ignore lint/a11y/noSvgWithoutTitle: aria-hidden decorative swoosh */}
      <svg
        aria-hidden
        data-pz-swoosh
        className="absolute left-0 w-full"
        style={{
          bottom: '-8px',
          height: '12px',
        }}
        viewBox="0 0 100 12"
        preserveAspectRatio="none"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path
          d="M2 9 C 20 2, 50 1, 80 5 C 90 7, 96 9, 98 8"
          stroke="currentColor"
          strokeWidth="2.5"
          strokeLinecap="round"
          strokeLinejoin="round"
          fill="none"
        />
      </svg>
    </span>
  )
}

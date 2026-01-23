'use client'

import { useEffect, useState } from 'react'

export function CompareBars({
  count = 40,
  variant = 'noisy',
}: {
  count?: number
  variant?: 'noisy' | 'clean'
}) {
  const [bars, setBars] = useState<number[]>(
    Array.from({ length: count }, () => (variant === 'noisy' ? 50 : 40))
  )

  useEffect(() => {
    setBars(
      Array.from({ length: count }, (_, i) =>
        variant === 'noisy'
          ? Math.floor(Math.random() * 100)
          : i % 5 === 0
            ? 0
            : Math.floor(Math.random() * 60 + 20)
      )
    )
  }, [count, variant])

  return (
    <div className="absolute inset-0 flex items-center gap-0.5">
      {bars.map((h, i) => (
        <div
          key={i}
          className={
            variant === 'noisy'
              ? 'w-full bg-foreground opacity-30'
              : `w-full ${i % 5 === 0 ? 'bg-transparent' : 'bg-primary'}`
          }
          style={{ height: `${h}%` }}
        />
      ))}
    </div>
  )
}

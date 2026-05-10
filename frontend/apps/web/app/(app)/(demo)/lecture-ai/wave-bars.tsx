'use client'

import { useEffect, useState } from 'react'

export function WaveBars({
  count = 60,
  noiseRange = [
    21,
    35,
  ],
  intervalMs = 400,
}: {
  count?: number
  noiseRange?: [
    number,
    number,
  ]
  intervalMs?: number
}) {
  const [heights, setHeights] = useState<number[]>(
    Array.from(
      {
        length: count,
      },
      () => 40
    )
  )

  useEffect(() => {
    const id = setInterval(() => {
      setHeights((prev) =>
        prev.map((_, i) =>
          i > noiseRange[0] && i < noiseRange[1] ? 20 : Math.floor(Math.random() * 60 + 20)
        )
      )
    }, intervalMs)
    return () => clearInterval(id)
  }, [
    count,
    noiseRange,
    intervalMs,
  ])

  return (
    <div className="h-8 flex-1 flex items-center gap-0.5">
      {heights.map((h, i) => (
        <div
          key={i}
          className={`w-1 rounded-full transition-all duration-500 ${
            i > noiseRange[0] && i < noiseRange[1] ? 'bg-destructive/30' : 'bg-primary/40'
          }`}
          style={{
            height: `${h}%`,
          }}
        />
      ))}
    </div>
  )
}

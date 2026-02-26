'use client'

import { useEffect } from 'react'
import { useInView, useCountUp } from '../lib/animations'
import { STATS } from '../lib/data'

function StatItem({
  value,
  label,
  inView,
}: Readonly<{ value: string; label: string; inView: boolean }>) {
  const numericMatch = value.match(/^([\d,]+)/)
  const suffix = value.replace(/^[\d,]+/, '')

  if (numericMatch) {
    const numericValue = Number.parseInt(numericMatch[1]!.replace(/,/g, ''), 10)
    return <AnimatedStat value={numericValue} suffix={suffix} label={label} inView={inView} />
  }

  return (
    <div className="flex flex-col gap-1.5 items-start">
      <span className="text-6xl font-normal leading-[72px] tracking-[-3.6px] text-[#131313] max-md:text-4xl max-md:leading-10">
        {value}
      </span>
      <span className="text-base font-medium leading-6 tracking-[-0.96px] text-[#9c9c9c]">
        {label}
      </span>
    </div>
  )
}

function AnimatedStat({
  value,
  suffix,
  label,
  inView,
}: Readonly<{ value: number; suffix: string; label: string; inView: boolean }>) {
  const { count, start } = useCountUp(value, 2000, true)

  useEffect(() => {
    if (inView) start()
  }, [inView, start])

  const formatted = count.toLocaleString('en-US')
  const finalFormatted = value.toLocaleString('en-US')

  return (
    <div className="flex flex-col gap-1.5 items-start">
      <span className="relative inline-block tabular-nums text-6xl font-normal leading-[72px] tracking-[-3.6px] text-[#131313] max-md:text-4xl max-md:leading-10">
        <span className="invisible" aria-hidden="true">{finalFormatted}{suffix}</span>
        <span suppressHydrationWarning className="absolute left-0 top-0">{formatted}{suffix}</span>
      </span>
      <span className="text-base font-medium leading-6 tracking-[-0.96px] text-[#9c9c9c]">
        {label}
      </span>
    </div>
  )
}

export function StatsSection() {
  const [ref, isInView] = useInView()

  return (
    <section
      ref={ref}
      className="mx-auto max-w-[1256px] rounded-3xl bg-white px-6 py-12 md:px-[58px] md:py-[50px]"
    >
      <div className="grid grid-cols-2 gap-8 md:flex md:flex-wrap md:items-center md:justify-between">
        {STATS.map((stat, i) => (
          <div key={stat.label} className="flex items-center gap-8">
            {i > 0 && (
              <div className="hidden h-14 w-px bg-[#e5e5e5] md:block" aria-hidden="true" />
            )}
            <StatItem value={stat.value} label={stat.label} inView={isInView} />
          </div>
        ))}
      </div>
    </section>
  )
}

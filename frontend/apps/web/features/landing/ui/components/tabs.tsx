'use client'

import { cn } from '@workspace/ui/lib/utils'
import { type KeyboardEvent, type ReactNode, useState } from 'react'

type TabItem = {
  label: string
  content: ReactNode
}

type TabsProps = {
  items: TabItem[]
  className?: string
  panelClassName?: string
}

export function Tabs({ items, className, panelClassName }: TabsProps) {
  const [activeIndex, setActiveIndex] = useState(0)

  function onKeyDown(e: KeyboardEvent<HTMLButtonElement>) {
    const last = items.length - 1
    if (e.key === 'ArrowRight') {
      e.preventDefault()
      setActiveIndex((i) => (i === last ? 0 : i + 1))
    } else if (e.key === 'ArrowLeft') {
      e.preventDefault()
      setActiveIndex((i) => (i === 0 ? last : i - 1))
    } else if (e.key === 'Home') {
      e.preventDefault()
      setActiveIndex(0)
    } else if (e.key === 'End') {
      e.preventDefault()
      setActiveIndex(last)
    } else if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault()
      // Already active from focus — nothing extra needed
    }
  }

  return (
    <div className={className}>
      <div role="tablist" aria-label="Tabs" className="flex gap-2">
        {items.map((item, i) => (
          <button
            key={item.label}
            type="button"
            role="tab"
            id={`tab-${i}`}
            aria-selected={activeIndex === i}
            aria-controls={`tabpanel-${i}`}
            tabIndex={activeIndex === i ? 0 : -1}
            onClick={() => setActiveIndex(i)}
            onKeyDown={onKeyDown}
            className={cn(
              'rounded-full px-5 py-2 text-sm font-medium transition-colors',
              'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--pz-accent)] focus-visible:ring-offset-2',
              activeIndex === i
                ? 'bg-[var(--pz-accent)] text-white'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
            )}
          >
            {item.label}
          </button>
        ))}
      </div>

      {items.map((item, i) => (
        <div
          key={item.label}
          role="tabpanel"
          id={`tabpanel-${i}`}
          aria-labelledby={`tab-${i}`}
          data-pz-tab-panel
          hidden={activeIndex !== i}
          className={panelClassName}
        >
          {item.content}
        </div>
      ))}
    </div>
  )
}

type TabPanelProps = {
  id: string
  labelledBy: string
  hidden?: boolean
  children: ReactNode
  className?: string
}

export function TabPanel({ id, labelledBy, hidden = false, children, className }: TabPanelProps) {
  return (
    <div
      role="tabpanel"
      id={id}
      aria-labelledby={labelledBy}
      data-pz-tab-panel
      hidden={hidden}
      className={className}
    >
      {children}
    </div>
  )
}

'use client'

import { useState } from 'react'
import { Plus } from 'lucide-react'
import { cn } from '@workspace/ui/lib/utils'
import { AnimatedSection } from './components/animated-section'
import { Eyebrow } from './components/eyebrow'
import { SectionFrame } from './components/section-frame'

type FaqItem = { question: string; answer: string }

type FaqSectionProps = {
  eyebrow?: string
  headline?: string
  subheadline?: string
  items?: FaqItem[]
}

export function FaqSection({
  eyebrow,
  headline,
  subheadline,
  items,
}: Readonly<FaqSectionProps> = {}) {
  const faqs = items ?? []
  const [openIndex, setOpenIndex] = useState<number | null>(null)

  function toggle(i: number) {
    setOpenIndex((prev) => (prev === i ? null : i))
  }

  function handleKeyDown(e: React.KeyboardEvent, i: number) {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault()
      toggle(i)
    } else if (e.key === 'Escape') {
      setOpenIndex(null)
    }
  }

  return (
    <SectionFrame id="faq" bg="white">
      <AnimatedSection className="mx-auto max-w-2xl text-center">
        {eyebrow && <Eyebrow className="mb-4">{eyebrow}</Eyebrow>}
        <h2 className="font-heading text-4xl font-bold tracking-tight text-[var(--pz-ink)] md:text-5xl">
          {headline ?? 'Questions Answered'}
        </h2>
        {subheadline && (
          <p className="mt-4 text-lg leading-relaxed text-[var(--pz-ink-2)]">{subheadline}</p>
        )}
      </AnimatedSection>

      <AnimatedSection delay={120} className="mx-auto mt-12 max-w-2xl divide-y divide-[var(--pz-card-border)]">
        {faqs.map((item, i) => {
          const isOpen = openIndex === i
          const panelId = `faq-${i}-panel`
          const triggerId = `faq-${i}-trigger`

          return (
            <div key={item.question}>
              <button
                id={triggerId}
                type="button"
                aria-expanded={isOpen}
                aria-controls={panelId}
                onClick={() => toggle(i)}
                onKeyDown={(e) => handleKeyDown(e, i)}
                className="flex w-full items-center justify-between gap-4 py-5 text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--pz-accent)] focus-visible:ring-offset-2"
              >
                <span className="font-heading text-base font-semibold text-[var(--pz-ink)] md:text-lg">
                  {item.question}
                </span>
                <Plus
                  className={cn(
                    'size-5 shrink-0 text-[var(--pz-accent)] transition-transform duration-200',
                    isOpen && 'rotate-45',
                  )}
                  aria-hidden
                />
              </button>

              <div
                id={panelId}
                role="region"
                aria-labelledby={triggerId}
                className="grid transition-[grid-template-rows] duration-300 ease-out"
                style={{ gridTemplateRows: isOpen ? '1fr' : '0fr' }}
              >
                <div className="overflow-hidden">
                  <p className="pb-5 text-base leading-relaxed text-[var(--pz-ink-2)]">
                    {item.answer}
                  </p>
                </div>
              </div>
            </div>
          )
        })}
      </AnimatedSection>
    </SectionFrame>
  )
}

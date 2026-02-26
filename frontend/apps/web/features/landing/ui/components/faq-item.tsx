'use client'

import { useState } from 'react'
import { Plus, X } from 'lucide-react'
import { cn } from '@workspace/ui/lib/utils'

interface FaqItemProps {
  question: string
  answer: string
}

function renderAnswer(answer: string) {
  const paragraphs = answer.split('\n\n')

  return paragraphs.map((block) => {
    const lines = block.split('\n')

    if (lines.length > 2) {
      const firstLine = lines[0] ?? ''
      const restLines = lines.slice(1)
      const looksLikeList = restLines.every((l) => l.trim().length > 0 && !l.trim().endsWith('.'))

      if (looksLikeList) {
        return (
          <div key={block} className="flex flex-col gap-2">
            {firstLine && (
              <p className="text-base leading-6 text-[#111] sm:text-lg sm:leading-7">{firstLine}</p>
            )}
            <ul className="list-disc space-y-1 pl-5 text-base leading-6 text-[#444141] sm:pl-7 sm:text-lg sm:leading-7">
              {restLines.map((item) => (
                <li key={item}>{item}</li>
              ))}
            </ul>
          </div>
        )
      }
    }

    return (
      <p key={block} className="text-base leading-6 text-[#444141] sm:text-lg sm:leading-7">
        {block}
      </p>
    )
  })
}

export function FaqItem({ question, answer }: Readonly<FaqItemProps>) {
  const [isOpen, setIsOpen] = useState(false)

  return (
    <div className="rounded-3xl bg-[#f3f3f3] p-6">
      <button
        type="button"
        className="flex w-full items-center gap-2 text-left"
        onClick={() => setIsOpen(!isOpen)}
        aria-expanded={isOpen}
      >
        <span className="flex-1 text-base font-normal leading-6 text-[#131313] sm:text-xl sm:leading-7">{question}</span>
        <span className="shrink-0">
          {isOpen ? (
            <X className="size-6 text-[#131313]" />
          ) : (
            <Plus className="size-6 text-[#131313]" />
          )}
        </span>
      </button>
      <div
        className={cn(
          'grid transition-all duration-300 ease-in-out',
          isOpen ? 'grid-rows-[1fr] opacity-100 mt-3.5' : 'grid-rows-[0fr] opacity-0'
        )}
      >
        <div className="overflow-hidden">
          <div className="flex flex-col gap-3">{renderAnswer(answer)}</div>
        </div>
      </div>
    </div>
  )
}

'use client'

import { useEffect, useRef } from 'react'
import { TESTIMONIALS } from '../lib/data'
import { AnimatedSection } from './components/animated-section'
import { SectionHeader } from './components/section-header'
import { TestimonialCard } from './components/testimonial-card'

export function TestimonialsSection() {
  const scrollRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const el = scrollRef.current
    if (!el) return

    let animationId: number
    let scrollPos = 0
    const speed = 0.5

    const animate = () => {
      scrollPos += speed
      if (scrollPos >= el.scrollWidth / 2) {
        scrollPos = 0
      }
      el.scrollLeft = scrollPos
      animationId = requestAnimationFrame(animate)
    }

    animationId = requestAnimationFrame(animate)

    const pause = () => cancelAnimationFrame(animationId)
    const resume = () => {
      animationId = requestAnimationFrame(animate)
    }

    el.addEventListener('mouseenter', pause)
    el.addEventListener('mouseleave', resume)

    return () => {
      cancelAnimationFrame(animationId)
      el.removeEventListener('mouseenter', pause)
      el.removeEventListener('mouseleave', resume)
    }
  }, [])

  return (
    <section className="mx-auto max-w-[1256px] overflow-hidden rounded-3xl bg-white px-6 py-12 md:py-16" aria-label="Testimonials carousel">
      <AnimatedSection className="flex flex-col items-center gap-12">
        <SectionHeader
          title={
            <>
              What Our Clients <em className="font-heading">Are Saying</em>
            </>
          }
          subtitle="Our users love how Pulzifi simplifies their decisions and streamlines operations."
        />

        {/* Auto-scrolling testimonials */}
        <div className="relative w-full">
          {/* Edge fades */}
          <div className="pointer-events-none absolute inset-y-0 left-0 z-10 w-[100px] bg-gradient-to-r from-white to-transparent" />
          <div className="pointer-events-none absolute inset-y-0 right-0 z-10 w-[100px] bg-gradient-to-l from-white to-transparent" />

          <div
            ref={scrollRef}
            className="flex gap-6 overflow-hidden"
          >
            {/* Duplicate for infinite scroll effect */}
            {[...TESTIMONIALS, ...TESTIMONIALS].map((testimonial, i) => (
              <TestimonialCard
                key={`${testimonial.author}-${i}`}
                {...testimonial}
              />
            ))}
          </div>
        </div>
      </AnimatedSection>
    </section>
  )
}

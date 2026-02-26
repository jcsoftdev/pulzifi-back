'use client'

import { AnimatedSection } from './components/animated-section'
import { LandingButton } from './components/landing-button'

export function HeroSection() {
  return (
    <section className="relative mx-auto max-w-[1256px] overflow-hidden rounded-3xl bg-white">
      <div className="grid items-center gap-8 px-6 py-10 sm:py-16 md:min-h-[633px] md:grid-cols-2 md:px-[58px] md:py-20">
        {/* Left Content */}
        <AnimatedSection animation="slide-left" className="flex flex-col gap-12">
          <div className="flex flex-col gap-6">
            <span className="inline-flex w-fit items-center rounded-full bg-[rgba(124,58,237,0.1)] px-5 py-2.5 text-sm font-medium tracking-tight text-[#29144c]">
              Real-Time Web Monitoring
            </span>
            <div className="flex flex-col gap-4">
              <h1 className="font-heading text-3xl font-medium leading-[40px] tracking-[-1.5px] text-[#131313] sm:text-4xl sm:leading-[48px] sm:tracking-[-2px] md:text-5xl md:leading-[72px] md:tracking-[-3.6px]">
                Everything you need to{' '}
                <em className="font-heading not-italic" style={{ fontStyle: 'italic' }}>
                  monitor anything
                </em>
              </h1>
              <p className="max-w-[495px] text-base leading-6 text-[#444141]">
                Automate your competitive intelligence. Get instant alerts when sites change, prices
                drop, or content updates.
              </p>
            </div>
          </div>

          {/* CTA Input */}
          <div className="flex w-full max-w-[541px] flex-col gap-2.5 sm:flex-row">
            <div className="flex flex-1 items-center rounded-full border border-[rgba(70,56,174,0.1)] bg-[rgba(130,130,130,0.1)] px-6">
              <input
                type="url"
                placeholder="https://competitor.com/pricing"
                className="w-full bg-transparent py-4 text-base outline-none placeholder:text-[#888]"
                aria-label="URL to monitor"
              />
            </div>
            <LandingButton href="/register" variant="primary" size="lg" withArrow className="w-full justify-center sm:w-auto">
              Start Monitoring Free
            </LandingButton>
          </div>
        </AnimatedSection>

        {/* Right - Dashboard Preview */}
        <AnimatedSection animation="slide-right" delay={200} className="relative hidden md:block">
          <div className="relative">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src="/images/landing/hero-dashboard.png"
              alt="Pulzifi dashboard preview showing competitive intelligence monitoring"
              className="w-full rounded-2xl object-contain"
              loading="eager"
            />
            {/* Floating glow effect */}
            <div className="absolute -inset-4 -z-10 rounded-3xl bg-[#7c3aed]/10 blur-3xl" />
          </div>
        </AnimatedSection>
      </div>
    </section>
  )
}

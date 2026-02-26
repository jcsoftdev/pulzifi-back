import { ArrowRight } from 'lucide-react'
import { HOW_IT_WORKS_STEPS } from '../lib/data'
import { AnimatedSection } from './components/animated-section'
import { LandingButton } from './components/landing-button'
import { SectionHeader } from './components/section-header'

export function HowItWorksSection() {
  return (
    <section
      id="how-it-works"
      className="mx-auto max-w-[1256px] overflow-hidden rounded-3xl bg-[#29144c] px-6 py-12 md:px-[58px] md:py-[50px]"
    >
      <AnimatedSection className="flex flex-col items-center gap-[51px]">
        <SectionHeader
          badge="How It Works:"
          badgeVariant="dark"
          title="Track, analyze, and optimize your business"
          subtitle="Detect changes instantly, understand their impact, and turn insights into smarter business decisions."
          variant="dark"
        />

        <div className="grid w-full gap-5 md:grid-cols-3">
          {HOW_IT_WORKS_STEPS.map((step, i) => (
            <AnimatedSection
              key={step.step}
              animation="fade-up"
              delay={i * 150}
              className="flex flex-col gap-10 rounded-3xl bg-white p-6 transition-shadow duration-300 hover:shadow-xl"
            >
              <span className="font-heading text-4xl font-medium leading-[48px] tracking-[-2px] text-[#131313] sm:text-5xl sm:leading-[56px] sm:tracking-[-2.88px]">
                {step.step}
              </span>
              <div className="flex flex-col gap-4">
                <h3 className="text-2xl font-medium leading-normal tracking-[-1.44px] text-[#111]">
                  {step.title}
                </h3>
                <p className="text-base leading-6 text-[#444141]">{step.description}</p>
              </div>
            </AnimatedSection>
          ))}
        </div>

        <p className="max-w-[657px] text-center text-base leading-6 text-white/80">
          Monitor Any Website Change The Instant It Happens.
        </p>

        <LandingButton href="/register" variant="primary" size="lg" withArrow>
          Get Started
        </LandingButton>
      </AnimatedSection>
    </section>
  )
}

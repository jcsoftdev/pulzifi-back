import { Settings2, Sparkles } from 'lucide-react'
import { AnimatedSection } from './components/animated-section'
import { SectionHeader } from './components/section-header'

const INSIGHT_TYPES = [
  { label: 'Marketing Lents', tagColor: 'bg-[#ebf0ff]' },
  { label: 'Brand and Pricing Strategy', tagColor: 'bg-[#ebfff5]' },
  { label: 'Business Opportunities', tagColor: 'bg-[#fff9f5]' },
]

export function InsightsSection() {
  return (
    <section className="mx-auto flex max-w-[1256px] flex-col gap-6">
      <div className="rounded-3xl bg-white px-6 py-8 md:px-[58px]">
        <AnimatedSection>
          <SectionHeader
            badge="Intelligence"
            title={
              <>
                Transform your insights into
                <br />
                strategic decisions
              </>
            }
            subtitle="Choose guidance tailored to marketing, business strategy, brand positioning, and more. Let AI turn raw data into actionable intelligence."
          />
        </AnimatedSection>
      </div>

      <AnimatedSection
        animation="scale"
        className="relative overflow-hidden rounded-3xl bg-[#2e1a51] px-6 py-10 md:px-[58px] md:py-12"
      >
        {/* Grid background - fades from center */}
        <div
          className="pointer-events-none absolute inset-0"
          style={{
            backgroundImage:
              'linear-gradient(rgba(255,255,255,.15) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,.15) 1px, transparent 1px)',
            backgroundSize: '40px 40px',
            maskImage: 'radial-gradient(ellipse 50% 60% at 50% 50%, black 0%, transparent 100%)',
            WebkitMaskImage: 'radial-gradient(ellipse 50% 60% at 50% 50%, black 0%, transparent 100%)',
          }}
        />

        {/* Stacked insight cards */}
        <div className="relative mx-auto w-full max-w-[520px]">
          {/* Floating "Analyze alerts with..." pill */}
          <div className="relative z-10 mb-[-12px] flex justify-center">
          <div className="inline-flex items-center gap-2 rounded-full bg-white px-4 py-2 shadow-[0_4px_20px_rgba(0,0,0,0.08)]">
            <Sparkles className="size-4 text-[#7c3aed]" />
            <span className="text-sm font-medium text-[#29144c]">Analyze alerts with&hellip;</span>
          </div>
          </div>

          {/* Stacked cards - centered, short to large */}
          <div className="relative" style={{ height: '260px' }}>
            {INSIGHT_TYPES.map((insight, i) => (
              <div
                key={insight.label}
                className="absolute left-1/2 rounded-2xl bg-white px-3 py-2 shadow-[0_4px_20px_rgba(0,0,0,0.06)] sm:px-5 sm:py-3"
                style={{
                  top: `${i * 50}px`,
                  width: `${70 + i * 10}%`,
                  transform: 'translateX(-50%)',
                  zIndex: i + 1,
                }}
              >
                <div className="flex items-center justify-between gap-2">
                  <div className={`rounded-full px-2 py-0.5 sm:px-3 sm:py-1 ${insight.tagColor}`}>
                    <span className="text-xs font-medium text-[#29144c] sm:text-sm">{insight.label}</span>
                  </div>
                  <div className="flex size-7 shrink-0 items-center justify-center rounded-full bg-black/5 sm:size-8">
                    <Settings2 className="size-3.5 text-[#444] sm:size-4" />
                  </div>
                </div>

                {i === INSIGHT_TYPES.length - 1 && (
                  <div className="mt-2 flex flex-col gap-1 sm:mt-3 sm:gap-1.5">
                    <span className="text-xs font-medium text-[#111] sm:text-sm">AI Insight:</span>
                    <p className="text-[11px] leading-relaxed text-[#504170] sm:text-[13px]">
                      &hellip;your competitor updated their homepage messaging from
                      &ldquo;Affordable Listings&rdquo; to &ldquo;Luxury Investment
                      Properties&rdquo; and added new content focused on high-net-worth
                      clients. This suggests a shift toward the premium segment, potentially
                      opening space in the mid-market for your brand to strengthen its
                      positioning&hellip;.
                    </p>
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      </AnimatedSection>
    </section>
  )
}

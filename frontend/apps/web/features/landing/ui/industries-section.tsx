'use client'

import { useState, type ReactNode } from 'react'
import { cn } from '@workspace/ui/lib/utils'
import { AnimatedSection } from './components/animated-section'
import { SectionHeader } from './components/section-header'

function ComplianceVisual() {
  return (
    <div className="flex h-full items-center justify-center p-6">
      <div className="relative rounded-xl bg-white px-5 py-4 shadow-md">
        <p className="mb-1 text-sm font-medium text-[#7c3aed]">New:</p>
        <p className="text-sm leading-relaxed text-[#333]">
          Requirements, Rules, Standard, Policies, Regulations, etc
        </p>
      </div>
    </div>
  )
}

function SaasVisual() {
  return (
    <div className="flex h-full items-center justify-center p-4">
      <div className="w-[180px] overflow-hidden rounded-[24px] border border-gray-200 bg-white shadow-lg">
        <div className="flex items-center justify-between px-3 py-2">
          <span className="text-[10px] font-semibold text-[#111]">Moneta</span>
          <div className="h-2.5 w-4 text-gray-400">
            <svg viewBox="0 0 16 10" fill="none" xmlns="http://www.w3.org/2000/svg">
              <title>Menu</title>
              <path d="M0 0h16v2H0zM0 4h16v2H0zM0 8h16v2H0z" fill="currentColor" />
            </svg>
          </div>
        </div>
        <div className="bg-[#f0fdf4] px-3 py-2">
          <p className="text-[8px] text-[#111]">US Dollar</p>
          <p className="text-[8px] text-gray-500">Total Balance</p>
          <p className="text-lg font-medium tracking-tight text-[#111]">$5,502.45</p>
        </div>
        <div className="px-3 py-2">
          <p className="mb-1 text-[8px] font-semibold text-[#111]">Transaction</p>
          {['Hani Arifin', 'Angelina'].map((name) => (
            <div key={name} className="flex items-center justify-between py-1">
              <span className="text-[8px] text-[#111]">{name}</span>
              <span className="text-[8px] text-[#0eb200]">$239</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

interface Industry {
  name: string
  tagline: string
  description: string
  features: string[]
  image?: string
  visual?: () => ReactNode
}

const INDUSTRIES: Industry[] = [
  {
    name: 'E-commerce',
    tagline: 'Never miss a competitor price drop',
    description:
      'Monitor competitor product pages and get notified the instant prices change. React faster than your rivals and win more sales without constant manual checks.',
    features: ['Price change alerts', 'Product launch detection', 'Promo tracking'],
    image: '/images/landing/industry-ecommerce.png',
  },
  {
    name: 'Compliance',
    tagline: 'Stay ahead of regulatory updates',
    description:
      'Track government portals, legal databases, and regulatory bodies for policy changes. Never be caught off-guard by a terms update or a regulatory shift that affects your business.',
    features: ['Policy change monitoring', 'Regulatory alerts', 'Terms update tracking'],
    visual: ComplianceVisual,
  },
  {
    name: 'Real Estate',
    tagline: 'Gain real estate market intelligence in real time',
    description:
      'Get instant alerts when new properties that match your criteria go live. Monitor real estate agents, brokerage sites, and listing aggregators automatically.',
    features: ['Listing price tracking', 'New property alerts', 'Market positioning insights'],
    image: '/images/landing/industry-realestate.png',
  },
  {
    name: 'Marketing Agencies',
    tagline: 'Stay in control of every client website',
    description:
      "Stay informed when something changes across your clients' sites. Whether it's a broken section, updated copy, removed tracking code, or an unexpected layout change, Pulzifi keeps your team in control.",
    features: ['Client site monitoring', 'Copy change detection', 'Layout shift alerts'],
    image: '/images/landing/industry-marketing.png',
  },
  {
    name: 'PR & Communications',
    tagline: 'Monitor media and search visibility in real time',
    description:
      'Track changes in news sites and search results so you know immediately when your brand is mentioned or when competitors make headlines.',
    features: ['Brand mention tracking', 'Search visibility monitoring', 'News alert detection'],
    image: '/images/landing/industry-pr.png',
  },
  {
    name: 'SaaS & Product Teams',
    tagline: 'Track competitor feature launches',
    description:
      'Know when competitors update pricing, release new features, or change positioning. Detect new feature pages, monitor pricing structure updates, track messaging shifts, and identify roadmap signals.',
    features: ['Feature page monitoring', 'Pricing tier changes', 'GTM strategy shifts'],
    visual: SaasVisual,
  },
]

export function IndustriesSection() {
  const [active, setActive] = useState(0)

  return (
    <section
      id="industries"
      className="mx-auto max-w-[1256px] rounded-3xl bg-white px-6 py-12 md:px-[58px] md:py-16"
    >
      <AnimatedSection className="flex flex-col gap-12">
        <SectionHeader
          title={
            <>
              How other industries{' '}
              <em className="font-heading">are using Pulzifi</em>
            </>
          }
          subtitle="From compliance to competitive intelligence, discover how teams turn website changes into actionable insights."
        />

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 sm:gap-5 lg:grid-cols-3">
          {INDUSTRIES.map((industry, i) => (
            <button
              key={industry.name}
              type="button"
              onClick={() => setActive(i)}
              className={cn(
                'group flex flex-col overflow-hidden rounded-3xl border text-left transition-all duration-300',
                active === i
                  ? 'border-[#7c3aed]/20 shadow-lg ring-2 ring-[#7c3aed]/10'
                  : 'border-black/5 hover:border-black/10 hover:shadow-md'
              )}
            >
              {/* Image area */}
              <div className="relative h-[200px] w-full overflow-hidden bg-[#f3f3f3]">
                {industry.image ? (
                  /* eslint-disable-next-line @next/next/no-img-element */
                  <img
                    src={industry.image}
                    alt={industry.name}
                    className="absolute inset-0 h-full w-full object-contain p-4 transition-transform duration-500 group-hover:scale-105"
                  />
                ) : industry.visual ? (
                  <industry.visual />
                ) : null}
              </div>
              {/* Text area */}
              <div className="flex flex-col gap-2 p-5">
                <span className="text-xs font-medium uppercase tracking-wider text-[#888]">
                  {industry.name}
                </span>
                <h3 className="text-lg font-medium leading-snug text-[#131313]">
                  {industry.tagline}
                </h3>
                <p className="text-sm leading-relaxed text-[#444141] line-clamp-3">
                  {industry.description}
                </p>
              </div>
            </button>
          ))}
        </div>
      </AnimatedSection>
    </section>
  )
}

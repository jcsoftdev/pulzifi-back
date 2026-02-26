import type { Metadata } from 'next'
import { headers } from 'next/headers'
import { redirect } from 'next/navigation'
import { extractTenantFromHostname } from '@workspace/shared-http'

import {
  Navbar,
  HeroSection,
  StatsSection,
  HowItWorksSection,
  FeaturesSection,
  InsightsSection,
  IndustriesSection,
  PricingSection,
  TestimonialsSection,
  FaqSection,
  FooterSection,
} from '@/features/landing'

export const metadata: Metadata = {
  title: 'Pulzifi — AI-Powered Competitive Intelligence & Website Monitoring',
  description:
    'Monitor any website for changes and instantly get AI-powered strategic insights. Track competitor moves, pricing changes, and market shifts — automatically, 24/7.',
  keywords: [
    'competitive intelligence',
    'website monitoring',
    'AI insights',
    'competitor tracking',
    'market intelligence',
    'change detection',
    'competitive analysis',
    'price monitoring',
    'web scraping',
    'business intelligence',
  ],
  authors: [{ name: 'Pulzifi' }],
  creator: 'Pulzifi',
  openGraph: {
    title: 'Pulzifi — AI-Powered Competitive Intelligence & Website Monitoring',
    description:
      'Monitor any website for changes and instantly get AI-powered strategic insights. Track competitor moves, pricing changes, and market shifts — automatically, 24/7.',
    type: 'website',
    siteName: 'Pulzifi',
    images: [
      {
        url: '/images/landing/hero-dashboard.png',
        width: 1200,
        height: 630,
        alt: 'Pulzifi Dashboard — AI-Powered Competitive Intelligence',
      },
    ],
  },
  twitter: {
    title: 'Pulzifi — AI-Powered Competitive Intelligence',
    description:
      'Monitor any website for changes and get AI-powered strategic insights. Know what competitors do before it impacts your business.',
    card: 'summary_large_image',
    images: ['/images/landing/hero-dashboard.png'],
  },
  alternates: {
    canonical: '/',
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      'max-video-preview': -1,
      'max-image-preview': 'large',
      'max-snippet': -1,
    },
  },
}

export default async function HomePage() {
  const headersList = await headers()
  const hostname = headersList.get('host') || ''
  const tenant = extractTenantFromHostname(hostname)

  if (tenant) {
    redirect('/workspaces')
  }

  return (
    <div className="min-h-screen bg-[#f3f3f3]">
      <div className="mx-auto max-w-[1280px] space-y-3 p-3">
        <Navbar />
        <main>
          <div className="space-y-3">
            <HeroSection />
            <StatsSection />
            <HowItWorksSection />
            <FeaturesSection />
            <InsightsSection />
            <IndustriesSection />
            <PricingSection />
            <TestimonialsSection />
            <FaqSection />
          </div>
        </main>
        <FooterSection />
      </div>

      {/* JSON-LD Structured Data for SEO */}
      <script type="application/ld+json">
        {JSON.stringify({
          '@context': 'https://schema.org',
          '@type': 'SoftwareApplication',
          name: 'Pulzifi',
          applicationCategory: 'BusinessApplication',
          operatingSystem: 'Web',
          description:
            'AI-powered competitive intelligence platform that monitors websites for changes and delivers strategic insights.',
          offers: [
            {
              '@type': 'Offer',
              name: 'Starter Plan',
              price: '20',
              priceCurrency: 'USD',
              priceValidUntil: '2027-12-31',
            },
            {
              '@type': 'Offer',
              name: 'Professional Plan',
              price: '62',
              priceCurrency: 'USD',
              priceValidUntil: '2027-12-31',
            },
          ],
          aggregateRating: {
            '@type': 'AggregateRating',
            ratingValue: '5.0',
            ratingCount: '205',
            bestRating: '5',
            worstRating: '1',
          },
        })}
      </script>
    </div>
  )
}

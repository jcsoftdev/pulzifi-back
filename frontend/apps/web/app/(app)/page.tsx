import { AuthApi } from '@workspace/services'
import { extractTenantFromHostname } from '@workspace/shared-http'
import type { Metadata } from 'next'
import { isRedirectError } from 'next/dist/client/components/redirect-error'
import { headers } from 'next/headers'
import { redirect } from 'next/navigation'

import { BlocksRenderer } from '@/features/cms'
import { FooterSection, Navbar } from '@/features/landing'
import { getPayloadClient } from '@/lib/payload'
import { buildThemeStyle } from '@/lib/theme-style'

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
  authors: [
    {
      name: 'Pulzifi',
    },
  ],
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
    images: [
      '/images/landing/hero-dashboard.png',
    ],
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
    try {
      await AuthApi.getCurrentUser()
      redirect('/workspaces')
    } catch (error: unknown) {
      if (isRedirectError(error)) throw error
      // Not authenticated — fall through to show landing page
    }
  }

  let landingBlocks: any[] = []
  let navLinks: { label: string; href: string }[] | undefined
  let navSigninLabel: string | undefined
  let navSigninHref: string | undefined
  let navPrimaryCtaLabel: string | undefined
  let navPrimaryCtaHref: string | undefined
  let navLogoUrl: string | undefined
  let footerGroups: Record<string, { label: string; href: string }[]> | undefined
  let footerTagline: string | undefined
  let footerSocialLinks: { platform: string; href: string }[] | undefined
  let footerLogoUrl: string | undefined
  let themeStyle = ''
  try {
    const payload = await getPayloadClient()
    const [landing, navbar, footer, theme] = await Promise.all([
      payload.findGlobal({ slug: 'landing', depth: 2 }),
      payload.findGlobal({ slug: 'navbar', depth: 1 }),
      payload.findGlobal({ slug: 'footer', depth: 1 }),
      payload.findGlobal({ slug: 'theme', depth: 0 }).catch(() => null),
    ])
    landingBlocks = (landing.blocks as any) ?? []
    const nav = navbar as any
    const rawLinks = nav.links as { label: string; href: string }[] | undefined
    navLinks = rawLinks?.length ? rawLinks : undefined
    navSigninLabel = nav.signinLabel ?? undefined
    navSigninHref = nav.signinHref ?? undefined
    navPrimaryCtaLabel = nav.primaryCtaLabel ?? undefined
    navPrimaryCtaHref = nav.primaryCtaHref ?? undefined
    if (nav.logo && typeof nav.logo === 'object' && nav.logo.url) {
      navLogoUrl = nav.logo.url as string
    }
    const foot = footer as any
    const rawGroups = foot.groups as
      | { heading: string; links: { label: string; href: string }[] }[]
      | undefined
    if (rawGroups?.length) {
      footerGroups = Object.fromEntries(
        rawGroups.map((g) => [g.heading, g.links?.map((l) => ({ label: l.label, href: l.href })) ?? []]),
      )
    }
    footerTagline = foot.tagline ?? undefined
    footerSocialLinks = foot.socialLinks?.length ? foot.socialLinks : undefined
    if (foot.logo && typeof foot.logo === 'object' && foot.logo.url) {
      footerLogoUrl = foot.logo.url as string
    }
    if (theme) {
      themeStyle = buildThemeStyle(theme as unknown as Record<string, string | null | undefined>)
    }
  } catch {
    // DB unavailable — fall through to defaults
  }

  return (
    <div className="min-h-screen bg-[var(--pz-page-bg)]">
      {themeStyle && <style dangerouslySetInnerHTML={{ __html: themeStyle }} />}
      <Navbar
        links={navLinks}
        signinLabel={navSigninLabel}
        signinHref={navSigninHref}
        primaryCtaLabel={navPrimaryCtaLabel}
        primaryCtaHref={navPrimaryCtaHref}
        logoUrl={navLogoUrl}
      />
      <main>
        <BlocksRenderer blocks={landingBlocks} />
      </main>
      <FooterSection
        links={footerGroups}
        tagline={footerTagline}
        socialLinks={footerSocialLinks}
        logoUrl={footerLogoUrl}
      />

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

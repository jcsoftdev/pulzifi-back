import { AuthApi } from '@workspace/services'
import { extractTenantFromHostname } from '@workspace/shared-http'
import type { Metadata } from 'next'
import { isRedirectError } from 'next/dist/client/components/redirect-error'
import { headers } from 'next/headers'
import { redirect } from 'next/navigation'

import {
  FaqSection,
  FooterSection,
  Navbar,
  PricingSection,
} from '@/features/landing'
import { getPayloadClient } from '@/lib/payload'

export const metadata: Metadata = {
  title: 'Pricing — Simple, Transparent Plans',
  description:
    'Choose a plan that fits your business needs and budget. No hidden fees, no surprises, just straightforward pricing for powerful competitive intelligence.',
  openGraph: {
    title: 'Pricing — Pulzifi',
    description:
      'Simple, transparent pricing for AI-powered competitive intelligence. Start with our Starter plan or scale with Professional and Enterprise.',
  },
}

type PlanDoc = {
  id: string | number
  name: string
  price: string
  period?: string | null
  tagline?: string | null
  features?: { text?: string | null; included?: boolean | null }[] | null
  ctaLabel?: string | null
  ctaHref?: string | null
  highlighted?: boolean | null
  popularBadge?: string | null
}

type PricingPageGlobal = {
  header?: {
    eyebrow?: string | null
    headline?: string | null
    headlineHighlight?: string | null
    subheadline?: string | null
  } | null
  plans?: (PlanDoc | string | number)[] | null
  guaranteeNote?: string | null
  faq?: {
    eyebrow?: string | null
    headline?: string | null
    subheadline?: string | null
    items?: { question: string; answer: string }[] | null
  } | null
}

function isPlanDoc(value: PlanDoc | string | number): value is PlanDoc {
  return typeof value === 'object' && value !== null && 'name' in value
}

export default async function PricingPage() {
  const headersList = await headers()
  const hostname = headersList.get('host') || ''
  const tenant = extractTenantFromHostname(hostname)

  if (tenant) {
    try {
      await AuthApi.getCurrentUser()
      redirect('/workspaces')
    } catch (error: unknown) {
      if (isRedirectError(error)) throw error
    }
  }

  let pricingData: PricingPageGlobal = {}
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
    const [pricing, navbar, footer, theme] = await Promise.all([
      payload.findGlobal({ slug: 'pricing-page', depth: 2 }),
      payload.findGlobal({ slug: 'navbar', depth: 1 }),
      payload.findGlobal({ slug: 'footer', depth: 1 }),
      payload.findGlobal({ slug: 'theme', depth: 0 }).catch(() => null),
    ])
    pricingData = pricing as PricingPageGlobal
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
        rawGroups.map((g) => [
          g.heading,
          g.links?.map((l) => ({ label: l.label, href: l.href })) ?? [],
        ]),
      )
    }
    footerTagline = foot.tagline ?? undefined
    footerSocialLinks = foot.socialLinks?.length ? foot.socialLinks : undefined
    if (foot.logo && typeof foot.logo === 'object' && foot.logo.url) {
      footerLogoUrl = foot.logo.url as string
    }
    if (theme) {
      const t = theme as unknown as Record<string, string | null | undefined>
      const map: Array<[string, string | null | undefined]> = [
        ['--pz-page-bg', t.pageBg],
        ['--pz-card-bg', t.cardBg],
        ['--pz-dark-surface', t.darkSurface],
        ['--pz-ink', t.inkPrimary],
        ['--pz-ink-2', t.inkSecondary],
        ['--pz-accent', t.accentPrimary],
        ['--pz-accent-muted', t.accentMuted],
        ['--pz-accent-gold', t.accentGold],
        ['--pz-accent-teal', t.accentTeal],
      ]
      const decls = map
        .filter(([, v]) => typeof v === 'string' && v.trim().length > 0)
        .map(([k, v]) => `${k}: ${v};`)
        .join(' ')
      if (decls) themeStyle = `:root { ${decls} }`
    }
  } catch {
    // DB unavailable — fall through to defaults
  }

  const plans = pricingData.plans
    ?.filter(isPlanDoc)
    .map((p) => ({
      name: p.name,
      price: p.price,
      period: p.period ?? undefined,
      tagline: p.tagline ?? undefined,
      features:
        p.features?.map((f) => ({
          text: f.text ?? '',
          included: f.included ?? true,
        })) ?? undefined,
      ctaLabel: p.ctaLabel ?? undefined,
      ctaHref: p.ctaHref ?? undefined,
      highlighted: p.highlighted ?? false,
      popularBadge: p.popularBadge ?? undefined,
    }))

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
        <PricingSection
          eyebrow={pricingData.header?.eyebrow ?? undefined}
          headline={pricingData.header?.headline ?? undefined}
          headlineHighlight={pricingData.header?.headlineHighlight ?? undefined}
          subheadline={pricingData.header?.subheadline ?? undefined}
          guaranteeNote={pricingData.guaranteeNote ?? undefined}
          plans={plans}
        />
        <FaqSection
          eyebrow={pricingData.faq?.eyebrow ?? undefined}
          headline={pricingData.faq?.headline ?? undefined}
          subheadline={pricingData.faq?.subheadline ?? undefined}
          items={pricingData.faq?.items ?? undefined}
        />
      </main>
      <FooterSection
        links={footerGroups}
        tagline={footerTagline}
        socialLinks={footerSocialLinks}
        logoUrl={footerLogoUrl}
      />

      {/* JSON-LD: Product + Offer array */}
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: JSON.stringify({
            '@context': 'https://schema.org',
            '@type': 'Product',
            name: 'Pulzifi',
            description:
              'AI-powered competitive intelligence platform that monitors websites for changes and delivers strategic insights.',
            brand: { '@type': 'Brand', name: 'Pulzifi' },
            offers: [
              {
                '@type': 'Offer',
                name: 'Starter Plan',
                price: '20',
                priceCurrency: 'USD',
                availability: 'https://schema.org/InStock',
                priceValidUntil: '2027-12-31',
                url: 'https://pulzifi.com/pricing',
              },
              {
                '@type': 'Offer',
                name: 'Professional Plan',
                price: '62',
                priceCurrency: 'USD',
                availability: 'https://schema.org/InStock',
                priceValidUntil: '2027-12-31',
                url: 'https://pulzifi.com/pricing',
              },
              {
                '@type': 'Offer',
                name: 'Enterprise Plan',
                description: 'Custom pricing — contact us',
                availability: 'https://schema.org/InStock',
                url: 'https://pulzifi.com/contact',
              },
            ],
          }),
        }}
      />

      {/* JSON-LD: FAQPage */}
      {(pricingData.faq?.items?.length ?? 0) > 0 && (
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{
            __html: JSON.stringify({
              '@context': 'https://schema.org',
              '@type': 'FAQPage',
              mainEntity: (pricingData.faq?.items ?? []).map((item) => ({
                '@type': 'Question',
                name: item.question,
                acceptedAnswer: {
                  '@type': 'Answer',
                  text: item.answer,
                },
              })),
            }),
          }}
        />
      )}
    </div>
  )
}

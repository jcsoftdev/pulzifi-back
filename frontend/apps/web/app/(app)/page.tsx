import { AuthApi } from '@workspace/services'
import { extractTenantFromHostname } from '@workspace/shared-http'
import type { Metadata } from 'next'
import { isRedirectError } from 'next/dist/client/components/redirect-error'
import { headers } from 'next/headers'
import { redirect } from 'next/navigation'
import { JsonLd } from '@/components/json-ld'
import { buildApexRedirectUrl } from '@/features/auth/application/redirect-authenticated.server'
import { BlocksRenderer } from '@/features/cms'
import { FooterSection, Navbar, SmoothScroll } from '@/features/landing'
import { getPayloadClient } from '@/lib/payload'
import { buildThemeStyle } from '@/lib/theme-style'

export const metadata: Metadata = {
  title: 'Pulzifi — AI-Powered Competitive Intelligence & Website Monitoring',
  description:
    'Track any competitor site for pricing, copy, and visual changes. Pulzifi turns each change into an AI strategic insight, delivered to Slack or email — automatically, 24/7.',
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
      'Track any competitor site for pricing, copy, and visual changes. Pulzifi turns each change into an AI strategic insight, delivered to Slack or email — automatically, 24/7.',
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
      'Monitor competitor sites for text and visual changes. Pulzifi turns each one into an AI insight in Slack — know what rivals do before it impacts your business.',
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
  const hostname = headersList.get('x-forwarded-host') || headersList.get('host') || ''
  const tenant = extractTenantFromHostname(hostname)
  const protocol = (() => {
    const p = headersList.get('x-forwarded-proto')
    return p ? `${p}:` : 'http:'
  })()

  if (tenant) {
    try {
      await AuthApi.getCurrentUser()
      redirect('/workspaces')
    } catch (error: unknown) {
      if (isRedirectError(error)) throw error
      // Anonymous on a tenant subdomain — subdomains require auth, so bounce to
      // the apex marketing/login instead of rendering the landing under a tenant.
      redirect(buildApexRedirectUrl(hostname, protocol, tenant, '/'))
    }
  }

  let blocks: {
    blockType: string
    id?: string
    [key: string]: unknown
  }[] = []
  let navLinks:
    | {
        label: string
        href: string
      }[]
    | undefined
  let navSigninLabel: string | undefined
  let navSigninHref: string | undefined
  let navPrimaryCtaLabel: string | undefined
  let navPrimaryCtaHref: string | undefined
  let navLogoUrl: string | undefined
  let footerGroups:
    | Record<
        string,
        {
          label: string
          href: string
        }[]
      >
    | undefined
  let footerTagline: string | undefined
  let footerSocialLinks:
    | {
        platform: string
        href: string
      }[]
    | undefined
  let footerLogoUrl: string | undefined
  let themeStyle = ''

  try {
    const payload = await getPayloadClient()
    const [pagesResult, navbar, footer, theme] = await Promise.all([
      payload.find({
        collection: 'pages',
        where: {
          slug: {
            equals: 'home',
          },
        },
        depth: 2,
        limit: 1,
      }),
      payload.findGlobal({
        slug: 'navbar',
        depth: 1,
      }),
      payload.findGlobal({
        slug: 'footer',
        depth: 1,
      }),
      payload
        .findGlobal({
          slug: 'theme',
          depth: 0,
        })
        .catch(() => null),
    ])
    blocks =
      (pagesResult.docs[0]?.blocks as {
        blockType: string
        id?: string
        [key: string]: unknown
      }[]) ?? []
    const nav = navbar as unknown as Record<string, unknown>
    const rawLinks = nav.links as
      | {
          label: string
          href: string
        }[]
      | undefined
    navLinks = rawLinks?.length ? rawLinks : undefined
    navSigninLabel = nav.signinLabel as string | undefined
    navSigninHref = nav.signinHref as string | undefined
    navPrimaryCtaLabel = nav.primaryCtaLabel as string | undefined
    navPrimaryCtaHref = nav.primaryCtaHref as string | undefined
    const navLogo = nav.logo as
      | {
          url?: string
        }
      | undefined
    if (navLogo?.url) {
      navLogoUrl = navLogo.url
    }
    const foot = footer as unknown as Record<string, unknown>
    const rawGroups = foot.groups as
      | {
          heading: string
          links: {
            label: string
            href: string
          }[]
        }[]
      | undefined
    if (rawGroups?.length) {
      footerGroups = Object.fromEntries(
        rawGroups.map((g) => [
          g.heading,
          g.links?.map((l: { label: string; href: string }) => ({
            label: l.label,
            href: l.href,
          })) ?? [],
        ])
      )
    }
    footerTagline = foot.tagline as string | undefined
    const rawSocial = foot.socialLinks as
      | {
          platform: string
          href: string
        }[]
      | undefined
    footerSocialLinks = rawSocial?.length ? rawSocial : undefined
    const footLogo = foot.logo as
      | {
          url?: string
        }
      | undefined
    if (footLogo?.url) {
      footerLogoUrl = footLogo.url
    }
    if (theme) {
      themeStyle = buildThemeStyle(theme as unknown as Record<string, string | null | undefined>)
    }
  } catch {
    // DB unavailable — fall through to defaults
  }

  // Extract FAQ items from the CMS blocks for FAQPage JSON-LD
  const faqItems = blocks
    .filter((b) => b.blockType === 'faq')
    .flatMap((b) => {
      const items =
        (
          b as {
            items?: {
              question: string
              answer: string
            }[]
          }
        ).items ?? []
      return items
    })
    .filter((item) => item.question && item.answer)

  return (
    <div className="min-h-screen bg-[var(--pz-page-bg)]">
      {themeStyle && (
        // biome-ignore lint/security/noDangerouslySetInnerHtml: intentional theme CSS injection — controlled server-side input
        <style dangerouslySetInnerHTML={{ __html: themeStyle }} />
      )}
      <SmoothScroll />
      <Navbar
        links={navLinks}
        signinLabel={navSigninLabel}
        signinHref={navSigninHref}
        primaryCtaLabel={navPrimaryCtaLabel}
        primaryCtaHref={navPrimaryCtaHref}
        logoUrl={navLogoUrl}
      />
      <main>
        <BlocksRenderer blocks={blocks} />
      </main>
      <FooterSection
        links={footerGroups}
        tagline={footerTagline}
        socialLinks={footerSocialLinks}
        logoUrl={footerLogoUrl}
      />

      <JsonLd
        data={{
          '@context': 'https://schema.org',
          '@type': 'SoftwareApplication',
          name: 'Pulzifi',
          url: process.env.NEXT_PUBLIC_APP_BASE_URL || 'https://pulzifi.com',
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
        }}
      />
      {faqItems.length > 0 && (
        <JsonLd
          data={{
            '@context': 'https://schema.org',
            '@type': 'FAQPage',
            mainEntity: faqItems.map((item) => ({
              '@type': 'Question',
              name: item.question,
              acceptedAnswer: {
                '@type': 'Answer',
                text: item.answer,
              },
            })),
          }}
        />
      )}
    </div>
  )
}

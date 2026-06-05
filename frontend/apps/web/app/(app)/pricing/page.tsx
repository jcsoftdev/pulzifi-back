import { AuthApi } from '@workspace/services'
import { extractTenantFromHostname } from '@workspace/shared-http'
import type { Metadata } from 'next'
import { isRedirectError } from 'next/dist/client/components/redirect-error'
import { headers } from 'next/headers'
import { redirect } from 'next/navigation'

import { buildApexRedirectUrl } from '@/features/auth/application/redirect-authenticated.server'
import { BlocksRenderer } from '@/features/cms'
import { FooterSection, Navbar, SmoothScroll } from '@/features/landing'
import { getPayloadClient } from '@/lib/payload'
import { buildThemeStyle } from '@/lib/theme-style'

export const metadata: Metadata = {
  title: 'Pricing — Simple, Transparent Plans',
  description:
    'Simple, transparent plans for AI-powered competitive intelligence. Monitor more pages, unlock deeper AI insights, and scale alerts as your team grows. No hidden fees.',
  alternates: {
    canonical: '/pricing',
  },
  openGraph: {
    title: 'Pricing — Pulzifi',
    description:
      'Simple, transparent pricing for AI-powered competitive intelligence. Start with a free trial, then scale pages, AI insights, and alerts as your team grows.',
    type: 'website',
    siteName: 'Pulzifi',
    url: '/pricing',
    images: [
      {
        url: '/opengraph-image',
        width: 1200,
        height: 630,
        alt: 'Pulzifi — Simple, Transparent Pricing',
      },
    ],
  },
  twitter: {
    title: 'Pricing — Pulzifi',
    description:
      'Simple, transparent pricing for AI-powered competitive intelligence. Start with a free trial, then scale pages, AI insights, and alerts as your team grows.',
    card: 'summary_large_image',
    images: [
      '/opengraph-image',
    ],
  },
  robots: {
    index: true,
    follow: true,
  },
}

export default async function PricingPage() {
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
      // Anonymous on a tenant subdomain — bounce to the apex public site.
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
            equals: 'pricing',
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

  return (
    <div className="min-h-screen bg-[var(--pz-page-bg)]">
      {themeStyle && (
        <style
          // biome-ignore lint/security/noDangerouslySetInnerHtml: intentional theme CSS injection — controlled server-side input
          dangerouslySetInnerHTML={{
            __html: themeStyle,
          }}
        />
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
    </div>
  )
}

import { AuthApi } from '@workspace/services'
import { extractTenantFromHostname } from '@workspace/shared-http'
import type { Metadata } from 'next'
import { isRedirectError } from 'next/dist/client/components/redirect-error'
import { headers } from 'next/headers'
import { redirect } from 'next/navigation'

import { BlocksRenderer } from '@/features/cms'
import { FooterSection, Navbar, SmoothScroll } from '@/features/landing'
import { getPayloadClient } from '@/lib/payload'
import { buildThemeStyle } from '@/lib/theme-style'

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

  let blocks: any[] = []
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
    const [pagesResult, navbar, footer, theme] = await Promise.all([
      payload.find({
        collection: 'pages',
        where: { slug: { equals: 'pricing' } },
        depth: 2,
        limit: 1,
      }),
      payload.findGlobal({ slug: 'navbar', depth: 1 }),
      payload.findGlobal({ slug: 'footer', depth: 1 }),
      payload.findGlobal({ slug: 'theme', depth: 0 }).catch(() => null),
    ])
    blocks = (pagesResult.docs[0]?.blocks as any) ?? []
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
          g.links?.map((l: { label: string; href: string }) => ({ label: l.label, href: l.href })) ?? [],
        ]),
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

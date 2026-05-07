'use client'

import { useLivePreview } from '@payloadcms/live-preview-react'
import { useMemo } from 'react'

import { FaqSection, FooterSection, Navbar, PricingSection } from '@/features/landing'
import { PreviewModeProvider } from '@/features/landing/lib/preview-mode'
import { buildThemeStyle } from '@/lib/theme-style'

type PricingHeader = {
  eyebrow?: string | null
  headline?: string | null
  headlineHighlight?: string | null
  subheadline?: string | null
}

type PricingPlan = {
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

type PricingFaq = {
  eyebrow?: string | null
  headline?: string | null
  subheadline?: string | null
  items?: { question: string; answer: string }[] | null
}

type PricingData = {
  header?: PricingHeader | null
  plans?: PricingPlan[] | null
  guaranteeNote?: string | null
  faq?: PricingFaq | null
}

type NavbarData = {
  logo?: { url?: string } | null
  links?: { label: string; href: string }[]
  signinLabel?: string
  signinHref?: string
  primaryCtaLabel?: string
  primaryCtaHref?: string
}

type FooterData = {
  logo?: { url?: string } | null
  tagline?: string
  groups?: { heading: string; links: { label: string; href: string }[] }[]
  socialLinks?: { platform: string; href: string }[]
}

type ThemeData = Record<string, string | null | undefined>

type Props = {
  initialPricing: PricingData
  initialTheme: ThemeData
  initialNavbar: NavbarData
  initialFooter: FooterData
  serverURL: string
}

export function PricingPreviewClient({
  initialPricing,
  initialTheme,
  initialNavbar,
  initialFooter,
  serverURL,
}: Props) {
  const { data: pricing } = useLivePreview<PricingData>({
    initialData: initialPricing,
    serverURL,
    depth: 1,
  })

  const { data: theme } = useLivePreview<ThemeData>({
    initialData: initialTheme,
    serverURL,
    depth: 0,
  })

  const { data: navbar } = useLivePreview<NavbarData>({
    initialData: initialNavbar,
    serverURL,
    depth: 1,
  })

  const { data: footer } = useLivePreview<FooterData>({
    initialData: initialFooter,
    serverURL,
    depth: 1,
  })

  const themeStyle = useMemo(() => buildThemeStyle(theme), [theme])

  const navProps = useMemo(() => {
    const links = navbar?.links?.length ? navbar.links : undefined
    const logoUrl =
      navbar?.logo && typeof navbar.logo === 'object' && navbar.logo.url
        ? navbar.logo.url
        : undefined
    return {
      links,
      signinLabel: navbar?.signinLabel,
      signinHref: navbar?.signinHref,
      primaryCtaLabel: navbar?.primaryCtaLabel,
      primaryCtaHref: navbar?.primaryCtaHref,
      logoUrl,
    }
  }, [navbar])

  const footerProps = useMemo(() => {
    const groups = footer?.groups?.length
      ? Object.fromEntries(
          footer.groups.map((g) => [
            g.heading,
            g.links?.map((l) => ({ label: l.label, href: l.href })) ?? [],
          ]),
        )
      : undefined
    const logoUrl =
      footer?.logo && typeof footer.logo === 'object' && footer.logo.url
        ? footer.logo.url
        : undefined
    return {
      links: groups,
      tagline: footer?.tagline,
      socialLinks: footer?.socialLinks?.length ? footer.socialLinks : undefined,
      logoUrl,
    }
  }, [footer])

  const plans = useMemo(
    () =>
      pricing?.plans?.map((p) => ({
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
      })) ?? undefined,
    [pricing],
  )

  return (
    <PreviewModeProvider value={true}>
      <div className="min-h-screen bg-[var(--pz-page-bg)]">
        {themeStyle && <style dangerouslySetInnerHTML={{ __html: themeStyle }} />}
        <Navbar {...navProps} />
        <main>
          <PricingSection
            eyebrow={pricing?.header?.eyebrow ?? undefined}
            headline={pricing?.header?.headline ?? undefined}
            headlineHighlight={pricing?.header?.headlineHighlight ?? undefined}
            subheadline={pricing?.header?.subheadline ?? undefined}
            guaranteeNote={pricing?.guaranteeNote ?? undefined}
            plans={plans}
          />
          <FaqSection
            eyebrow={pricing?.faq?.eyebrow ?? undefined}
            headline={pricing?.faq?.headline ?? undefined}
            subheadline={pricing?.faq?.subheadline ?? undefined}
            items={pricing?.faq?.items ?? undefined}
          />
        </main>
        <FooterSection {...footerProps} />
      </div>
    </PreviewModeProvider>
  )
}

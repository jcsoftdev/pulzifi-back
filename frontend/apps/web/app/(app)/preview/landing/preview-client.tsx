'use client'

import { useLivePreview } from '@payloadcms/live-preview-react'
import { useMemo } from 'react'

import { BlocksRenderer } from '@/features/cms'
import { FooterSection, Navbar } from '@/features/landing'
import { PreviewModeProvider } from '@/features/landing/lib/preview-mode'
import { buildThemeStyle } from '@/lib/theme-style'

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

type LandingData = {
  blocks?: { blockType: string; id?: string; [key: string]: unknown }[]
}

type ThemeData = Record<string, string | null | undefined>

type Props = {
  initialLanding: LandingData
  initialTheme: ThemeData
  initialNavbar: NavbarData
  initialFooter: FooterData
  serverURL: string
}

export function PreviewClient({
  initialLanding,
  initialTheme,
  initialNavbar,
  initialFooter,
  serverURL,
}: Props) {
  const { data: landing } = useLivePreview<LandingData>({
    initialData: initialLanding,
    serverURL,
    depth: 2,
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

  const blocks = useMemo(() => landing?.blocks ?? [], [landing])

  return (
    <PreviewModeProvider value={true}>
      <div className="min-h-screen bg-[var(--pz-page-bg)]">
        {themeStyle && <style dangerouslySetInnerHTML={{ __html: themeStyle }} />}
        <Navbar {...navProps} />
        <main>
          <BlocksRenderer blocks={blocks} />
        </main>
        <FooterSection {...footerProps} />
      </div>
    </PreviewModeProvider>
  )
}

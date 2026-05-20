'use client'

import { useLivePreview } from '@payloadcms/live-preview-react'
import { useMemo } from 'react'

import { Navbar } from '@/features/landing'
import { PreviewModeProvider } from '@/features/landing/lib/preview-mode'
import { buildThemeStyle } from '@/lib/theme-style'

type NavbarData = {
  logo?: {
    url?: string
  } | null
  links?: {
    label: string
    href: string
  }[]
  signinLabel?: string
  signinHref?: string
  primaryCtaLabel?: string
  primaryCtaHref?: string
}

type ThemeData = Record<string, string | null | undefined>

type Props = {
  initialNavbar: NavbarData
  initialTheme: ThemeData
  serverURL: string
}

export function PreviewClient({ initialNavbar, initialTheme, serverURL }: Props) {
  const { data: navbar } = useLivePreview<NavbarData>({
    initialData: initialNavbar,
    serverURL,
    depth: 1,
  })

  const { data: theme } = useLivePreview<ThemeData>({
    initialData: initialTheme,
    serverURL,
    depth: 0,
  })

  const themeStyle = useMemo(
    () => buildThemeStyle(theme),
    [
      theme,
    ]
  )

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
  }, [
    navbar,
  ])

  return (
    <PreviewModeProvider value={true}>
      <div className="min-h-screen bg-[var(--pz-page-bg)]">
        {themeStyle && (
          // biome-ignore lint/security/noDangerouslySetInnerHtml: intentional theme CSS injection — controlled server-side input
          <style dangerouslySetInnerHTML={{ __html: themeStyle }} />
        )}
        <Navbar {...navProps} />
        <div className="min-h-[60vh] flex items-center justify-center text-[var(--pz-ink-2)] text-sm">
          Navbar preview — scroll to see consolidation
        </div>
      </div>
    </PreviewModeProvider>
  )
}

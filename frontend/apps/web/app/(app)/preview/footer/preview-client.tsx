'use client'

import { useLivePreview } from '@payloadcms/live-preview-react'
import { useMemo } from 'react'

import { FooterSection } from '@/features/landing'
import { PreviewModeProvider } from '@/features/landing/lib/preview-mode'
import { buildThemeStyle } from '@/lib/theme-style'

type FooterData = {
  logo?: {
    url?: string
  } | null
  tagline?: string
  groups?: {
    heading: string
    links: {
      label: string
      href: string
    }[]
  }[]
  socialLinks?: {
    platform: string
    href: string
  }[]
}

type ThemeData = Record<string, string | null | undefined>

type Props = {
  initialFooter: FooterData
  initialTheme: ThemeData
  serverURL: string
}

export function PreviewClient({ initialFooter, initialTheme, serverURL }: Props) {
  const { data: footer } = useLivePreview<FooterData>({
    initialData: initialFooter,
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

  const footerProps = useMemo(() => {
    const groups = footer?.groups?.length
      ? Object.fromEntries(
          footer.groups.map((g) => [
            g.heading,
            g.links?.map((l) => ({
              label: l.label,
              href: l.href,
            })) ?? [],
          ])
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
  }, [
    footer,
  ])

  return (
    <PreviewModeProvider value={true}>
      <div className="min-h-screen bg-[var(--pz-page-bg)] flex flex-col">
        {themeStyle && (
          // biome-ignore lint/security/noDangerouslySetInnerHtml: intentional theme CSS injection — controlled server-side input
          <style dangerouslySetInnerHTML={{ __html: themeStyle }} />
        )}
        <div className="flex-1 flex items-center justify-center text-[var(--pz-ink-2)] text-sm">
          Footer preview
        </div>
        <FooterSection {...footerProps} />
      </div>
    </PreviewModeProvider>
  )
}

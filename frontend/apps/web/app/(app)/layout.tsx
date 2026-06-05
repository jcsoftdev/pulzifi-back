import type { Metadata } from 'next'
import { DM_Sans, DM_Serif_Display, Geist, Geist_Mono, Outfit, Syne } from 'next/font/google'

import '@workspace/ui/globals.css'
import { JsonLd } from '@/components/json-ld'
import { Providers } from '@/components/providers'
import { NotificationProvider } from '@/lib/notification'

const metadataBase = process.env.NEXT_PUBLIC_APP_BASE_URL
  ? new URL(process.env.NEXT_PUBLIC_APP_BASE_URL)
  : new URL('https://pulzifi.com')

export const metadata: Metadata = {
  metadataBase,
  title: {
    template: '%s | Pulzifi',
    default: 'Pulzifi — AI-Powered Competitive Intelligence',
  },
  description:
    'Pulzifi monitors any website for text and visual changes, then turns each one into an AI strategic insight delivered to Slack — so you catch competitor moves first.',
  openGraph: {
    type: 'website',
    siteName: 'Pulzifi',
    locale: 'en_US',
    images: [
      {
        url: '/opengraph-image',
        width: 1200,
        height: 630,
        alt: 'Pulzifi — AI-Powered Competitive Intelligence',
      },
    ],
  },
  twitter: {
    card: 'summary_large_image',
    images: [
      '/opengraph-image',
    ],
  },
}

const fontSans = Geist({
  subsets: [
    'latin',
  ],
  variable: '--font-sans',
})

const fontMono = Geist_Mono({
  subsets: [
    'latin',
  ],
  variable: '--font-mono',
})

// TODO(pulzifi-landing-redesign): --font-family-heading was remapped to var(--font-sans) in globals.css
// so this DM_Serif_Display font is no longer visually active on the landing page. Removal deferred
// until (main) dashboard routes are audited for `font-heading` consumers to avoid regressions.
const fontHeading = DM_Serif_Display({
  weight: '400',
  subsets: [
    'latin',
  ],
  variable: '--font-heading',
})

const fontLogo = Outfit({
  subsets: [
    'latin',
  ],
  variable: '--font-logo',
})

const fontDisplay = Syne({
  subsets: [
    'latin',
  ],
  weight: [
    '400',
    '600',
    '700',
    '800',
  ],
  variable: '--font-display',
  display: 'swap',
})

const fontBody = DM_Sans({
  subsets: [
    'latin',
  ],
  weight: [
    '300',
    '400',
    '500',
    '600',
    '700',
  ],
  variable: '--font-body',
  display: 'swap',
})

const siteUrl = process.env.NEXT_PUBLIC_APP_BASE_URL || 'https://pulzifi.com'

const organizationJsonLd = {
  '@context': 'https://schema.org',
  '@type': 'Organization',
  '@id': `${siteUrl}/#organization`,
  name: 'Pulzifi',
  url: siteUrl,
  logo: {
    '@type': 'ImageObject',
    url: `${siteUrl}/icon.png`,
  },
  description:
    'AI-powered website change monitoring and competitive intelligence platform. Track competitor moves, pricing changes, and market shifts — automatically, 24/7.',
  // Add real social profile URLs here when available
  sameAs: [] as string[],
}

const websiteJsonLd = {
  '@context': 'https://schema.org',
  '@type': 'WebSite',
  '@id': `${siteUrl}/#website`,
  name: 'Pulzifi',
  url: siteUrl,
  publisher: {
    '@id': `${siteUrl}/#organization`,
  },
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body
        className={`${fontSans.variable} ${fontMono.variable} ${fontHeading.variable} ${fontLogo.variable} ${fontDisplay.variable} ${fontBody.variable} font-sans antialiased`}
        suppressHydrationWarning
      >
        <JsonLd data={organizationJsonLd} />
        <JsonLd data={websiteJsonLd} />
        <Providers>{children}</Providers>
        <NotificationProvider />
      </body>
    </html>
  )
}

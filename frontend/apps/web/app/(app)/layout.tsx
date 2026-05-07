import type { Metadata } from 'next'
import { DM_Sans, DM_Serif_Display, Geist, Geist_Mono, Outfit, Syne } from 'next/font/google'

import '@workspace/ui/globals.css'
import { NotificationProvider } from '@/lib/notification'
import { Providers } from '@/components/providers'

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
    'Monitor any website for changes and get AI-powered strategic insights. Track competitor moves automatically.',
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
    images: ['/opengraph-image'],
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
  subsets: ['latin'],
  variable: '--font-heading',
})

const fontLogo = Outfit({
  subsets: ['latin'],
  variable: '--font-logo',
})

const fontDisplay = Syne({
  subsets: ['latin'],
  weight: ['400', '600', '700', '800'],
  variable: '--font-display',
  display: 'swap',
})

const fontBody = DM_Sans({
  subsets: ['latin'],
  weight: ['300', '400', '500', '600', '700'],
  variable: '--font-body',
  display: 'swap',
})

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={`${fontSans.variable} ${fontMono.variable} ${fontHeading.variable} ${fontLogo.variable} ${fontDisplay.variable} ${fontBody.variable} font-sans antialiased`}>
        <Providers>{children}</Providers>
        <NotificationProvider />
      </body>
    </html>
  )
}

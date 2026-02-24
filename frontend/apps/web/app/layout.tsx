import type { Metadata } from 'next'
import { Geist, Geist_Mono } from 'next/font/google'

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

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={`${fontSans.variable} ${fontMono.variable} font-sans antialiased`}>
        <Providers>{children}</Providers>
        <NotificationProvider />
      </body>
    </html>
  )
}

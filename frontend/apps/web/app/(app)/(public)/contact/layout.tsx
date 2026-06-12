import type { Metadata } from 'next'
import type { ReactNode } from 'react'

export const metadata: Metadata = {
  title: 'Contact Us — Pulzifi',
  description:
    'Get in touch with the Pulzifi team. Questions about pricing, enterprise plans, or integrations? We respond within one business day.',
  openGraph: {
    title: 'Contact Pulzifi',
    description:
      'Reach out for enterprise inquiries, support, or partnerships. Our team responds within one business day.',
    type: 'website',
    url: '/contact',
    // Explicit: page-level openGraph REPLACES the layout's wholesale (shallow
    // merge per top-level key), so omitting images here would share without one.
    images: [
      {
        url: '/opengraph-image',
        width: 1200,
        height: 630,
        alt: 'Contact Pulzifi',
      },
    ],
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Contact Pulzifi',
    description: 'Reach out for enterprise inquiries, support, or partnerships.',
    images: [
      '/opengraph-image',
    ],
  },
  alternates: {
    canonical: '/contact',
  },
  robots: {
    index: true,
    follow: true,
  },
}

export default function ContactLayout({ children }: { children: ReactNode }) {
  return <>{children}</>
}

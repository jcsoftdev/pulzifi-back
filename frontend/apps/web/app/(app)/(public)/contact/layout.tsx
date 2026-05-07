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
  },
  twitter: {
    card: 'summary',
    title: 'Contact Pulzifi',
    description: 'Reach out for enterprise inquiries, support, or partnerships.',
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

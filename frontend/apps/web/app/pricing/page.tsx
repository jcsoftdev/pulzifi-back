import { AuthApi } from '@workspace/services'
import { extractTenantFromHostname } from '@workspace/shared-http'
import type { Metadata } from 'next'
import { isRedirectError } from 'next/dist/client/components/redirect-error'
import { headers } from 'next/headers'
import { redirect } from 'next/navigation'

import {
  FaqSection,
  FooterSection,
  Navbar,
  PricingSection,
  TestimonialsSection,
} from '@/features/landing'

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
      // Not authenticated — fall through to show pricing page
    }
  }

  return (
    <div className="min-h-screen bg-[#ededed]">
      <div className="mx-auto max-w-[1280px] space-y-3 p-3">
        <Navbar />
        <main className="space-y-3">
          <PricingSection />
          <FaqSection />
          <TestimonialsSection />
        </main>
        <FooterSection />
      </div>
    </div>
  )
}

import { getPayloadClient } from '@/lib/payload'

import { PricingPreviewClient } from './preview-client'

export const dynamic = 'force-dynamic'

export default async function PreviewPricingPage() {
  const serverURL = process.env.NEXT_PUBLIC_SERVER_URL ?? 'http://localhost:3001'

  let initialPricing: Record<string, unknown> = {}
  let initialTheme: Record<string, string | null | undefined> = {}
  let initialNavbar: Record<string, unknown> = {}
  let initialFooter: Record<string, unknown> = {}

  try {
    const payload = await getPayloadClient()
    const [pricing, theme, navbar, footer] = await Promise.all([
      payload.findGlobal({ slug: 'pricing-page', depth: 1 }).catch(() => ({})),
      payload.findGlobal({ slug: 'theme', depth: 0 }).catch(() => null),
      payload.findGlobal({ slug: 'navbar', depth: 1 }),
      payload.findGlobal({ slug: 'footer', depth: 1 }),
    ])
    initialPricing = pricing as unknown as Record<string, unknown>
    if (theme) initialTheme = theme as unknown as Record<string, string | null | undefined>
    initialNavbar = navbar as unknown as Record<string, unknown>
    initialFooter = footer as unknown as Record<string, unknown>
  } catch {
    // DB unavailable — render with empty initial state
  }

  return (
    <PricingPreviewClient
      initialPricing={initialPricing as never}
      initialTheme={initialTheme}
      initialNavbar={initialNavbar as never}
      initialFooter={initialFooter as never}
      serverURL={serverURL}
    />
  )
}

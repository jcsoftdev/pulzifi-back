import { getPayloadClient } from '@/lib/payload'

import { PreviewClient } from './preview-client'

// This route is embedded in the Payload admin iframe — disable Next.js caching
// so the iframe always gets fresh initial state on load.
export const dynamic = 'force-dynamic'

export default async function PreviewLandingPage() {
  const serverURL = process.env.NEXT_PUBLIC_SERVER_URL ?? 'http://localhost:3001'

  let initialLanding: Record<string, unknown> = { blocks: [] }
  let initialTheme: Record<string, string | null | undefined> = {}
  let initialNavbar: Record<string, unknown> = {}
  let initialFooter: Record<string, unknown> = {}

  try {
    const payload = await getPayloadClient()
    const [landing, theme, navbar, footer] = await Promise.all([
      payload.findGlobal({ slug: 'landing', depth: 2 }),
      payload.findGlobal({ slug: 'theme', depth: 0 }).catch(() => null),
      payload.findGlobal({ slug: 'navbar', depth: 1 }),
      payload.findGlobal({ slug: 'footer', depth: 1 }),
    ])
    initialLanding = landing as unknown as Record<string, unknown>
    if (theme) initialTheme = theme as unknown as Record<string, string | null | undefined>
    initialNavbar = navbar as unknown as Record<string, unknown>
    initialFooter = footer as unknown as Record<string, unknown>
  } catch {
    // DB unavailable — PreviewClient renders with empty initial state
  }

  return (
    <PreviewClient
      initialLanding={initialLanding as never}
      initialTheme={initialTheme}
      initialNavbar={initialNavbar as never}
      initialFooter={initialFooter as never}
      serverURL={serverURL}
    />
  )
}

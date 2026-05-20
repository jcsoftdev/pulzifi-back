import { getPayloadClient } from '@/lib/payload'

import { PreviewClient } from './preview-client'

// This route is embedded in the Payload admin iframe — disable Next.js caching
// so the iframe always gets fresh initial state on load.
export const dynamic = 'force-dynamic'

export default async function PreviewNavbarPage() {
  const serverURL = process.env.NEXT_PUBLIC_SERVER_URL ?? 'http://localhost:3001'

  let initialNavbar: Record<string, unknown> = {}
  let initialTheme: Record<string, string | null | undefined> = {}

  try {
    const payload = await getPayloadClient()
    const [navbar, theme] = await Promise.all([
      payload.findGlobal({
        slug: 'navbar',
        depth: 1,
      }),
      payload
        .findGlobal({
          slug: 'theme',
          depth: 0,
        })
        .catch(() => null),
    ])
    initialNavbar = navbar as unknown as Record<string, unknown>
    if (theme) initialTheme = theme as unknown as Record<string, string | null | undefined>
  } catch {
    // DB unavailable — PreviewClient renders with empty initial state
  }

  return (
    <PreviewClient
      initialNavbar={initialNavbar as never}
      initialTheme={initialTheme}
      serverURL={serverURL}
    />
  )
}

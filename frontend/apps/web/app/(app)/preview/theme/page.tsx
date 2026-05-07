import { getPayloadClient } from '@/lib/payload'

import { ThemePreviewClient } from './preview-client'

export const dynamic = 'force-dynamic'

export default async function PreviewThemePage() {
  const serverURL = process.env.NEXT_PUBLIC_SERVER_URL ?? 'http://localhost:3001'

  let initialTheme: Record<string, string | null | undefined> = {}
  try {
    const payload = await getPayloadClient()
    const theme = await payload.findGlobal({ slug: 'theme', depth: 0 }).catch(() => null)
    if (theme) initialTheme = theme as unknown as Record<string, string | null | undefined>
  } catch {
    // DB unavailable — render with empty initial state
  }

  return <ThemePreviewClient initialTheme={initialTheme} serverURL={serverURL} />
}

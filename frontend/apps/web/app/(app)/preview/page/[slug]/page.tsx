import { getPayloadClient } from '@/lib/payload'

import { PagePreviewClient } from './preview-client'

export const dynamic = 'force-dynamic'

export default async function PreviewPageRoute({
  params,
}: {
  params: Promise<{
    slug: string
  }>
}) {
  const { slug } = await params
  const serverURL = process.env.NEXT_PUBLIC_SERVER_URL ?? 'http://localhost:3001'

  let initialPage: Record<string, unknown> = {
    blocks: [],
  }
  let initialTheme: Record<string, string | null | undefined> = {}
  let initialNavbar: Record<string, unknown> = {}
  let initialFooter: Record<string, unknown> = {}

  try {
    const payload = await getPayloadClient()
    const [pagesResult, theme, navbar, footer] = await Promise.all([
      payload.find({
        collection: 'pages',
        where: {
          slug: {
            equals: slug,
          },
        },
        depth: 2,
        limit: 1,
      }),
      payload
        .findGlobal({
          slug: 'theme',
          depth: 0,
        })
        .catch(() => null),
      payload.findGlobal({
        slug: 'navbar',
        depth: 1,
      }),
      payload.findGlobal({
        slug: 'footer',
        depth: 1,
      }),
    ])
    if (pagesResult.docs[0]) {
      initialPage = pagesResult.docs[0] as unknown as Record<string, unknown>
    }
    if (theme) initialTheme = theme as unknown as Record<string, string | null | undefined>
    initialNavbar = navbar as unknown as Record<string, unknown>
    initialFooter = footer as unknown as Record<string, unknown>
  } catch {
    // DB unavailable — render with empty initial state
  }

  return (
    <PagePreviewClient
      slug={slug}
      initialPage={initialPage}
      initialTheme={initialTheme}
      initialNavbar={initialNavbar as never}
      initialFooter={initialFooter as never}
      serverURL={serverURL}
    />
  )
}

import { getPayloadClient } from '@/lib/payload'

import { BlockPreviewClient } from './preview-client'

export const dynamic = 'force-dynamic'

export default async function PreviewBlockRoute({
  params,
}: {
  params: Promise<{
    id: string
  }>
}) {
  const { id } = await params
  const serverURL = process.env.NEXT_PUBLIC_SERVER_URL ?? 'http://localhost:3001'

  let initialBlock: Record<string, unknown> = {}
  let initialTheme: Record<string, string | null | undefined> = {}

  try {
    const payload = await getPayloadClient()
    const [blockResult, theme] = await Promise.all([
      payload.find({
        collection: 'block-library',
        where: {
          id: {
            equals: Number(id),
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
    ])
    if (blockResult.docs[0]) {
      initialBlock = blockResult.docs[0] as unknown as Record<string, unknown>
    }
    if (theme) initialTheme = theme as unknown as Record<string, string | null | undefined>
  } catch {
    // DB unavailable
  }

  return (
    <BlockPreviewClient
      id={id}
      initialBlock={initialBlock}
      initialTheme={initialTheme}
      serverURL={serverURL}
    />
  )
}

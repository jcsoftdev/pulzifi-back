'use client'

import { useLivePreview } from '@payloadcms/live-preview-react'
import { useMemo } from 'react'

import { BlocksRenderer } from '@/features/cms'
import { PreviewModeProvider } from '@/features/landing/lib/preview-mode'
import { buildThemeStyle } from '@/lib/theme-style'

type BlockLibraryData = {
  block?: {
    blockType: string
    id?: string
    [key: string]: unknown
  }[]
}

type ThemeData = Record<string, string | null | undefined>

type Props = {
  id: string
  initialBlock: BlockLibraryData
  initialTheme: ThemeData
  serverURL: string
}

export function BlockPreviewClient({ initialBlock, initialTheme, serverURL }: Props) {
  const { data: blockEntry } = useLivePreview<BlockLibraryData>({
    initialData: initialBlock,
    serverURL,
    depth: 2,
  })

  const { data: theme } = useLivePreview<ThemeData>({
    initialData: initialTheme,
    serverURL,
    depth: 0,
  })

  const themeStyle = useMemo(
    () => buildThemeStyle(theme),
    [
      theme,
    ]
  )
  const blocks = useMemo(
    () => blockEntry?.block ?? [],
    [
      blockEntry,
    ]
  )

  return (
    <PreviewModeProvider value={true}>
      <div className="min-h-screen bg-[var(--pz-page-bg)]">
        {themeStyle && (
          // biome-ignore lint/security/noDangerouslySetInnerHtml: intentional theme CSS injection — controlled server-side input
          <style dangerouslySetInnerHTML={{ __html: themeStyle }} />
        )}
        <BlocksRenderer blocks={blocks} />
      </div>
    </PreviewModeProvider>
  )
}

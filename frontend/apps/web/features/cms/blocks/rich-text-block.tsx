import { RichText } from '@payloadcms/richtext-lexical/react'
import type { SerializedEditorState } from 'lexical'

type Props = { block: { blockType: 'rich-text'; content: SerializedEditorState } }

export function RichTextBlock({ block }: Props) {
  return (
    <div className="prose mx-auto max-w-3xl px-4 py-8">
      <RichText data={block.content} />
    </div>
  )
}

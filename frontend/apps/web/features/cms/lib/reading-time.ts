// Estimates blog reading time from Payload's Lexical richText content.
// Shared by the `posts.beforeChange` hook (server, on save) so the value is
// persisted on the `readingTime` field and read cheaply on the blog list.

type LexicalNode = {
  text?: string
  children?: LexicalNode[]
}

const WORDS_PER_MINUTE = 200

function countWords(node: LexicalNode | null | undefined): number {
  if (!node || typeof node !== 'object') return 0
  let count = 0
  if (typeof node.text === 'string') {
    const trimmed = node.text.trim()
    if (trimmed.length > 0) count += trimmed.split(/\s+/).length
  }
  if (Array.isArray(node.children)) {
    for (const child of node.children) count += countWords(child)
  }
  return count
}

/** Reading time in whole minutes (minimum 1) for a Lexical editor state. */
export function readingTimeFromContent(content: unknown): number {
  const root = (
    content as
      | {
          root?: LexicalNode
        }
      | null
      | undefined
  )?.root
  const words = countWords(root)
  return Math.max(1, Math.ceil(words / WORDS_PER_MINUTE))
}

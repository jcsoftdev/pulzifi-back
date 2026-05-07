import { FaqSection } from '@/features/landing'

type Props = {
  block: {
    blockType: 'faq'
    eyebrow?: string | null
    headline?: string | null
    subheadline?: string | null
    items?: { question: string; answer: string }[] | null
  }
}

export function FaqBlock({ block }: Props) {
  return (
    <FaqSection
      eyebrow={block.eyebrow ?? undefined}
      headline={block.headline ?? undefined}
      subheadline={block.subheadline ?? undefined}
      items={block.items ?? undefined}
    />
  )
}

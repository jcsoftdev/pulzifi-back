import { AiIntelligenceSection } from '@/features/landing'

type AiTabItem = {
  title: string
  body?: string | null
  image?: string | null
}

type AiTab = {
  label: string
  items: AiTabItem[]
}

type Props = {
  block: {
    blockType: 'ai-intelligence'
    eyebrow?: string | null
    headline?: string | null
    headlineHighlight?: string | null
    subheadline?: string | null
    tabs?: AiTab[] | null
  }
}

export function AiIntelligenceBlock({ block }: Props) {
  return (
    <AiIntelligenceSection
      eyebrow={block.eyebrow ?? undefined}
      headline={block.headline ?? undefined}
      headlineHighlight={block.headlineHighlight ?? undefined}
      subheadline={block.subheadline ?? undefined}
      tabs={
        block.tabs?.map((t) => ({
          label: t.label,
          items: t.items.map((item) => ({
            title: item.title,
            body: item.body ?? undefined,
            image: item.image ?? undefined,
          })),
        })) ?? undefined
      }
    />
  )
}

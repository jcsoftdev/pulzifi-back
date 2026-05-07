import { IndustriesSection } from '@/features/landing'

type Props = {
  block: {
    blockType: 'industries'
    compactMode?: boolean | null
    eyebrow?: string | null
    headline?: string | null
    headlineHighlight?: string | null
    subheadline?: string | null
    items?:
      | {
          icon?: string | null
          title?: string | null
          description?: string | null
          realWin?: string | null
        }[]
      | null
  }
}

export function IndustriesBlock({ block }: Props) {
  return (
    <IndustriesSection
      compactMode={block.compactMode ?? true}
      eyebrow={block.eyebrow ?? undefined}
      headline={block.headline ?? undefined}
      headlineHighlight={block.headlineHighlight ?? undefined}
      subheadline={block.subheadline ?? undefined}
      items={
        block.items
          ?.filter((i) => i.title && i.description)
          .map((i) => ({
            icon: i.icon ?? undefined,
            title: i.title!,
            description: i.description!,
            realWin: i.realWin ?? undefined,
          })) ?? undefined
      }
    />
  )
}

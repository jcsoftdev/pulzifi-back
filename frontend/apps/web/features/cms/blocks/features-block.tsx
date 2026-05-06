import { FeaturesSection } from '@/features/landing'

type Props = {
  block: {
    blockType: 'features'
    cards?: { title: string; description: string; image?: { url?: string | null } | null }[] | null
  }
}

export function FeaturesBlock({ block }: Props) {
  return (
    <FeaturesSection
      cards={(block.cards ?? []).map((c) => ({
        title: c.title,
        description: c.description,
        image: typeof c.image === 'object' ? c.image?.url ?? undefined : undefined,
      }))}
    />
  )
}

import { StatsSection } from '@/features/landing'

type Props = {
  block: {
    blockType: 'stats'
    items?:
      | {
          value: string
          label: string
        }[]
      | null
  }
}

export function StatsBlock({ block }: Props) {
  return <StatsSection items={block.items ?? []} />
}

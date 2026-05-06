import { FaqSection } from '@/features/landing'

type Props = {
  block: {
    blockType: 'faq'
    items?: { question: string; answer: string }[] | null
  }
}

export function FaqBlock({ block }: Props) {
  return <FaqSection items={block.items ?? []} />
}

import { TestimonialsSection } from '@/features/landing'

type Props = {
  block: {
    blockType: 'testimonials'
    items?:
      | {
          quote: string
          author: string
          role?: string | null
          avatar?: {
            url?: string | null
          } | null
        }[]
      | null
  }
}

export function TestimonialsBlock({ block }: Props) {
  return (
    <TestimonialsSection
      items={(block.items ?? []).map((i) => ({
        quote: i.quote,
        author: i.author,
        role: i.role ?? undefined,
        avatar: typeof i.avatar === 'object' ? (i.avatar?.url ?? undefined) : undefined,
      }))}
    />
  )
}

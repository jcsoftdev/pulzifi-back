import { HeroBlock } from '../blocks/hero-block'
import { StatsBlock } from '../blocks/stats-block'
import { HowItWorksBlock } from '../blocks/how-it-works-block'
import { FeaturesBlock } from '../blocks/features-block'
import { InsightsBlock } from '../blocks/insights-block'
import { IndustriesBlock } from '../blocks/industries-block'
import { PricingBlock } from '../blocks/pricing-block'
import { TestimonialsBlock } from '../blocks/testimonials-block'
import { FaqBlock } from '../blocks/faq-block'
import { RichTextBlock } from '../blocks/rich-text-block'
import { CtaBlock } from '../blocks/cta-block'
import { ImageBlock } from '../blocks/image-block'

type Block = { blockType: string; id?: string; [key: string]: unknown }

type Props = { blocks: Block[] }

export function BlocksRenderer({ blocks }: Props) {
  return (
    <div className="space-y-3">
      {blocks.map((block, i) => {
        const key = block.id ?? `${block.blockType}-${i}`
        switch (block.blockType) {
          case 'hero':          return <HeroBlock key={key} block={block as any} />
          case 'stats':         return <StatsBlock key={key} block={block as any} />
          case 'how-it-works':  return <HowItWorksBlock key={key} block={block as any} />
          case 'features':      return <FeaturesBlock key={key} block={block as any} />
          case 'insights':      return <InsightsBlock key={key} block={block as any} />
          case 'industries':    return <IndustriesBlock key={key} block={block as any} />
          case 'pricing':       return <PricingBlock key={key} block={block as any} />
          case 'testimonials':  return <TestimonialsBlock key={key} block={block as any} />
          case 'faq':           return <FaqBlock key={key} block={block as any} />
          case 'rich-text':     return <RichTextBlock key={key} block={block as any} />
          case 'cta':           return <CtaBlock key={key} block={block as any} />
          case 'image':         return <ImageBlock key={key} block={block as any} />
          default:              return null
        }
      })}
    </div>
  )
}

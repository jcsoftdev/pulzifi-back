import { memo } from 'react'

import { AiIntelligenceBlock } from '../blocks/ai-intelligence-block'
import { CtaBlock } from '../blocks/cta-block'
import { FaqBlock } from '../blocks/faq-block'
import { FeaturesBlock } from '../blocks/features-block'
import { HeroBlock } from '../blocks/hero-block'
import { HowItWorksBlock } from '../blocks/how-it-works-block'
import { ImageBlock } from '../blocks/image-block'
import { IndustriesBlock } from '../blocks/industries-block'
import { PricingBlock } from '../blocks/pricing-block'
import { ProblemBlock } from '../blocks/problem-block'
import { RichTextBlock } from '../blocks/rich-text-block'

type Block = { blockType: string; id?: string; [key: string]: unknown }

type Props = { blocks: Block[] }

function renderBlock(block: Block) {
  switch (block.blockType) {
    case 'hero':
      return <HeroBlock block={block as never} />
    case 'problem':
      return <ProblemBlock block={block as never} />
    case 'how-it-works':
      return <HowItWorksBlock block={block as never} />
    case 'features':
      return <FeaturesBlock block={block as never} />
    case 'ai-intelligence':
      return <AiIntelligenceBlock block={block as never} />
    case 'industries':
      return <IndustriesBlock block={block as never} />
    case 'pricing':
      return <PricingBlock block={block as never} />
    case 'faq':
      return <FaqBlock block={block as never} />
    case 'rich-text':
      return <RichTextBlock block={block as never} />
    case 'cta':
      return <CtaBlock block={block as never} />
    case 'image':
      return <ImageBlock block={block as never} />
    // Legacy / removed block types — schemas stay in Payload (avoid migrations) but render nothing
    case 'logos':
    case 'stats':
    case 'testimonials':
    case 'comparison':
    case 'insights':
      return null
    default:
      return null
  }
}

const MemoBlock = memo(
  function MemoBlock({ block }: { block: Block }) {
    return <>{renderBlock(block)}</>
  },
  (prev, next) => {
    // Skip re-render if block object is referentially equal OR JSON-equal
    if (prev.block === next.block) return true
    try {
      return JSON.stringify(prev.block) === JSON.stringify(next.block)
    } catch {
      return false
    }
  },
)

export function BlocksRenderer({ blocks }: Props) {
  return (
    <>
      {blocks.map((block, i) => {
        const key = block.id ?? `${block.blockType}-${i}`
        return <MemoBlock key={key} block={block} />
      })}
    </>
  )
}

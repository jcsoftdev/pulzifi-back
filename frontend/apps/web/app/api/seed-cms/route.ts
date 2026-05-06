import { type NextRequest, NextResponse } from 'next/server'
import { getPayloadClient } from '@/lib/payload'

import {
  NAV_LINKS,
  STATS,
  HOW_IT_WORKS_STEPS,
  FEATURE_CARDS,
  PRICING_PLANS,
  TESTIMONIALS,
  FAQ_ITEMS,
  FOOTER_LINKS,
} from '@/features/landing/lib/data'

export async function POST(req: NextRequest) {
  const secret = req.headers.get('x-seed-secret')
  if (secret !== process.env.SEED_SECRET) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  const payload = await getPayloadClient()

  const landing = await payload.findGlobal({ slug: 'landing', depth: 0 })
  if ((landing as any).blocks && (landing as any).blocks.length > 0) {
    return NextResponse.json({ message: 'Already seeded, skipping.' })
  }

  // Seed navbar
  await payload.updateGlobal({
    slug: 'navbar',
    data: {
      links: NAV_LINKS.map((l) => ({ label: l.label, href: l.href })),
    },
  })

  // Seed footer — FOOTER_LINKS is { [groupName]: readonly { label, href }[] }
  await payload.updateGlobal({
    slug: 'footer',
    data: {
      groups: (
        Object.entries(FOOTER_LINKS) as [
          string,
          ReadonlyArray<{ label: string; href: string }>,
        ][]
      ).map(([heading, links]) => ({
        heading,
        links: links.map((l) => ({ label: l.label, href: l.href })),
      })),
      copyrightText: `© ${new Date().getFullYear()} Pulzifi. All rights reserved.`,
    },
  })

  // Seed landing page blocks
  await payload.updateGlobal({
    slug: 'landing',
    data: {
      blocks: [
        { blockType: 'hero' },
        {
          blockType: 'stats',
          items: STATS.map((s) => ({ value: s.value, label: s.label })),
        },
        {
          blockType: 'how-it-works',
          steps: HOW_IT_WORKS_STEPS.map((s) => ({
            step: s.step,
            title: s.title,
            description: s.description,
          })),
        },
        {
          blockType: 'features',
          cards: FEATURE_CARDS.map((c) => ({
            title: c.title,
            description: c.description,
          })),
        },
        { blockType: 'insights' },
        { blockType: 'industries' },
        {
          blockType: 'pricing',
          plans: PRICING_PLANS.map((p) => ({
            name: p.name,
            price: p.price,
            period: 'period' in p ? p.period : undefined,
            description: p.description,
            features: p.features.map((f) => ({ text: f })),
            ctaLabel: p.cta,
            ctaHref: 'ctaHref' in p ? p.ctaHref : '#',
            highlighted: 'popular' in p ? Boolean(p.popular) : false,
          })),
        },
        {
          blockType: 'testimonials',
          items: TESTIMONIALS.map((t) => ({
            quote: t.quote,
            author: t.author,
            role: t.role,
          })),
        },
        {
          blockType: 'faq',
          items: FAQ_ITEMS.map((f) => ({ question: f.question, answer: f.answer })),
        },
      ],
    },
  })

  return NextResponse.json({ message: 'Seeded successfully.' })
}

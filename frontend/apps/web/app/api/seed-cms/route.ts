import { type NextRequest, NextResponse } from 'next/server'
import { getPayloadClient } from '@/lib/payload'

// Seed data — kept here so the CMS seed is self-contained.
// Once the Payload globals are populated, this route can be deleted.

const NAV_LINKS = [
  { label: 'Home', href: '#' },
  { label: 'Product', href: '#how-it-works' },
  { label: 'How to use', href: '#industries' },
  { label: 'Pricing', href: '#pricing' },
  { label: 'Contact', href: '/contact' },
] as const

const STATS = [
  { value: '25,000+', label: 'Strategy Decisions Created' },
  { value: '150,493', label: 'Monitored Pages' },
  { value: '5.0', label: 'Customer Review' },
  { value: '205+', label: 'Industries' },
] as const

const HOW_IT_WORKS_STEPS = [
  {
    step: 1,
    title: 'Paste any URL',
    description:
      'Enter any public website URL. Pulzifi supports SPAs, JavaScript-rendered pages, and login-gated content.',
  },
  {
    step: 2,
    title: 'Select how to track',
    description: 'Select frequency, type of insights, tags, and workspace.',
  },
  {
    step: 3,
    title: 'Get instant alerts',
    description:
      'Receive notifications via email, Slack, webhook, or SMS the moment a change is detected.',
  },
] as const

const FEATURE_CARDS = [
  {
    title: 'Track Every Word Change',
    description:
      'Instantly see what messaging they adjust, add, or remove to understand shifts in strategy or positioning. Compare text versions and highlight key differences.',
  },
  {
    title: 'Flexible Monitoring Schedules',
    description: 'Check every 5 minutes, once a day or once a month. You set the tempo.',
  },
  {
    title: 'Visual Comparison',
    description:
      'See exactly what changed, revealing hidden tweaks text, branding, or new sections that impact strategy.',
  },
  {
    title: 'Choose How You Get Notified',
    description:
      'Get screenshot alerts of every update, via email, text messages, team channels and more. Stay informed your way.',
  },
] as const

const PRICING_PLANS = [
  {
    name: 'Starter Plan',
    price: '$20',
    period: '/month',
    description: 'Perfect for individual users and business owners.',
    cta: 'Try it Now',
    features: [
      '1 Workspace',
      'Up to 5 single pages',
      'Up to 1 user account',
      'Advanced 4 AI Insights',
      '1 Week Storage',
      'Email and Messages alerts',
    ],
  },
  {
    name: 'Professional Plan',
    price: '$62',
    period: '/month',
    description: 'Perfect for Growing Businesses Ready to Scale Their Operations',
    cta: 'Try it Now',
    popular: true,
    features: [
      'Unlimited Workspaces',
      'Up to 25 single pages',
      'Up to 5 user accounts',
      'Advanced unlimited AI Insights',
      'Email, Messages, Teams, Slack, Telegram alerts',
      '1 month storage',
      'Priority email and chat support',
    ],
  },
  {
    name: 'Enterprise Plan',
    price: 'Custom',
    description: 'Comprehensive and Scalable Solutions for Growing Large Organizations',
    cta: 'Schedule a Call',
    ctaHref: '/contact',
    features: [
      'Unlimited Workspaces',
      'Unlimited user accounts',
      'Unlimited single pages',
      'Advanced unlimited AI Insights',
      'Email, Messages, Teams, Slack, Telegram alerts',
      '3 month storage',
      'Priority email and chat support',
    ],
  },
] as const

const TESTIMONIALS = [
  {
    quote:
      "What I love most is the clarity. Pulzifi doesn't just say something changed, it explains what changed and why it matters for the user experience.",
    author: 'Oscar G.',
    role: 'Manager',
  },
  {
    quote:
      "It's not just another monitoring tool. It connects the dots and translates web changes into business meaning.",
    author: 'Tom W.',
    role: 'Business Intelligence Consultant',
  },
  {
    quote:
      "Pulzifi caught a keyword shift in a competitor's blog before any SEO tool did. That insight helped us reposition our content strategy fast.",
    author: 'Alex D.',
    role: 'SEO Analyst',
  },
  {
    quote:
      'Pulzifi helped us catch subtle homepage and pricing tests from our top competitors, things that traditional trackers completely missed.',
    author: 'Johanna T.',
    role: 'Founder',
  },
  {
    quote:
      'Pulzifi replaced hours of manual checks. Now I know instantly when a competitor changes pricing. I get context, not just alerts.',
    author: 'Julia C.',
    role: 'Product Manager',
  },
] as const

const FAQ_ITEMS = [
  {
    question: 'What is Pulzifi and how does it work?',
    answer: `Pulzifi is an AI-powered website monitoring and market intelligence platform that tracks changes across websites, competitor pages, news sources, and industry portals in real time.

You simply add the pages you want to monitor, choose what to track such as price updates, content changes, new listings, or policy updates, and Pulzifi sends instant alerts with AI insights explaining what changed and why it matters.

Pulzifi helps real estate professionals, marketing agencies, and business teams stay ahead of competitors, trends, and opportunities without checking websites manually.`,
  },
  {
    question: 'What types of changes can Pulzifi detect?',
    answer: `Pulzifi can detect a wide range of website and market changes, including:
Content updates on competitor websites
Real estate price changes or new property listings
Marketing campaign or messaging updates
News releases and press mentions
SEO ranking or keyword changes
Product launches or service updates
Policy or regulation changes related to your industry

With Pulzifi’s AI analysis, you also get business insights like new target audiences, strategy shifts, or emerging trends in your state or industry.`,
  },
  {
    question: 'How is Pulzifi different from other website monitoring tools?',
    answer: `Most website monitoring tools only show that something changed. Pulzifi goes further with AI-driven market intelligence.

Pulzifi explains what changed, why it matters, and what action you can take. It is built for business use cases like real estate trend tracking, marketing agency client monitoring, and competitive intelligence.

Pulzifi also offers:
AI summaries and recommendations
Industry-specific templates for real estate and marketing
White-label options for institutions like IREM training programs
Alerts focused on opportunities, risks, and strategy

Instead of raw alerts, Pulzifi gives strategic insights.`,
  },
  {
    question: 'How secure is my data with Pulzifi?',
    answer: `Pulzifi is built with enterprise-level security and privacy best practices.

Your monitored pages, alerts, and insights are encrypted in transit and at rest. Access is protected with secure authentication, and we never sell or share your data with third parties.

For institutions like real estate education programs or agencies managing client websites, Pulzifi also supports role-based access, audit logs, and secure cloud infrastructure to keep your data safe.`,
  },
] as const

const FOOTER_LINKS = {
  Features: [
    { label: 'Product', href: '#how-it-works' },
    { label: 'Use Cases', href: '#industries' },
    { label: 'Pricing', href: '#pricing' },
  ],
  Support: [
    { label: 'Help', href: '#' },
    { label: 'FAQ', href: '#faq' },
    { label: 'Contact', href: '/contact' },
  ],
  Legal: [
    { label: 'Privacy Policy', href: '#' },
    { label: 'Terms of Services', href: '#' },
  ],
} as const

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

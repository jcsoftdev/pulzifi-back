import type { Payload } from 'payload'

const NAV_LINKS = [
  { label: 'Home', href: '#' },
  { label: 'Product', href: '#how' },
  { label: 'How to use', href: '#usecases' },
  { label: 'Pricing', href: '#pricing' },
  { label: 'Contact', href: '/contact' },
] as const

const HERO = {
  eyebrowBadge: "NEW",
  eyebrowText: "AI-powered competitive intelligence for global teams",
  headline: "Know before your competitors",
  headlineHighlight: "move.",
  subheadline:
    "Pulzifi monitors any website 24/7 and turns raw changes into strategic intelligence — so you react first, win more, and stop losing deals to competitors you didn’t see coming.",
  primaryCtaLabel: "Start Monitoring Free",
  primaryCtaHref: "/register",
  secondaryCtaLabel: "See how it works",
  secondaryCtaHref: "#how",
  trustLine: "Trusted by 2,500+ teams across 40+ countries",
  kpis: [
    { label: "Alerts (24h)", value: "128", delta: "+24% vs yesterday", deltaDirection: "up" },
    { label: "Mentions", value: "43", delta: "+8%", deltaDirection: "up" },
    { label: "Pages Changed", value: "17", delta: "-3%", deltaDirection: "down" },
    { label: "Price Changes", value: "5", delta: "+2", deltaDirection: "up" },
  ],
  dashboardAlerts: [
    {
      tone: "signal",
      icon: "🔴",
      site: "competitor.com — Pricing Change",
      title: "Pro plan dropped from $89 → $69/mo",
      detail: "",
      time: "2m ago",
    },
    {
      tone: "amber",
      icon: "⚡",
      site: "rival-startup.io — New Feature Page",
      title: "”AI Automation Suite” launched",
      detail: "",
      time: "18m ago",
    },
    {
      tone: "teal",
      icon: "📣",
      site: "industry-leader.com — Messaging Shift",
      title: "Homepage CTA changed to “Free Trial”",
      detail: "",
      time: "1h ago",
    },
    {
      tone: "ink",
      icon: "📄",
      site: "regulator.gov — Policy Update",
      title: "New compliance requirements — Q1 2026",
      detail: "",
      time: "3h ago",
    },
  ],
  aiInsightTitle: "AI Insight",
  aiInsightBody:
    "Your competitor cut pricing by 22%. This signals aggressive growth targeting — consider reinforcing your value-add messaging before the weekend traffic spike.",
} as const

const LOGOS = {
  label: 'Trusted by teams at',
  items: [
    { text: 'Shopify Merchants' },
    { text: 'Real Estate Firms' },
    { text: 'Marketing Agencies' },
    { text: 'SaaS Companies' },
    { text: 'Law Firms' },
    { text: 'PE Investors' },
  ],
} as const

const PROBLEM = {
  eyebrow: 'The Problem',
  headline: 'Your competitors are changing. You’re finding out',
  headlineHighlight: 'too late.',
  cards: [
    {
      metric: '6h',
      label: 'Average delay',
      description:
        'The average business learns about a competitor’s pricing change 6 hours after it happens — costing deals that were already in flight.',
    },
    {
      metric: '87%',
      label: 'Miss critical signals',
      description:
        'Of businesses miss critical competitor signals every month because manual monitoring is unreliable, inconsistent, and unsustainable at scale.',
    },
    {
      metric: '$0',
      label: 'Value from raw alerts',
      description:
        'Traditional monitoring tools just tell you something changed. That’s useless. You need to know why it matters and what to do about it — now.',
    },
  ],
} as const

const STATS = [
  { value: '150K+', label: 'Pages monitored daily' },
  { value: '205+', label: 'Industries served' },
  { value: '5★', label: 'Average customer rating' },
  { value: '25K+', label: 'Strategic decisions powered' },
] as const

const HOW_IT_WORKS = {
  eyebrow: 'How It Works',
  headline: 'From URL to',
  headlineHighlight: 'strategic action',
  subheadline:
    'Three steps. Zero complexity. Intelligence that actually moves the needle for your business.',
  steps: [
    {
      step: 1,
      icon: '🔗',
      title: 'Paste any URL',
      description:
        'Add any public page — competitor sites, pricing pages, job boards, regulatory portals, news sources. Pulzifi handles JavaScript-rendered pages, SPAs, and login-gated content. No code. No proxies. Setup in under 60 seconds.',
      mockType: 'url',
      mockText: 'competitor.com/pricing',
    },
    {
      step: 2,
      icon: '⚙️',
      title: 'Set your rules',
      description:
        'Choose monitoring frequency (every 5 minutes to once a month), what to track, and how you want to be notified — email, Slack, SMS, webhook, or Teams. Pulzifi tags changes by importance so your inbox stays clean.',
      mockType: 'insight',
    },
    {
      step: 3,
      icon: '🧠',
      title: 'Get AI intelligence',
      description:
        'Every alert comes with AI-powered analysis: what changed, why it matters for your business, and what action to take. Not raw data. Strategic insight — delivered in plain language the moment it happens.',
      mockType: 'alerts',
    },
  ],
} as const

const FEATURES = {
  eyebrow: 'AI Intelligence',
  headline: 'Other tools say what changed.',
  headlineHighlight: 'We tell you what to do.',
  intro:
    'Every change Pulzifi detects gets analyzed by AI trained for business strategy. You get context, recommended actions, and opportunity scoring — not a screenshot with a timestamp.',
  bullets: [
    {
      title: 'Marketing Lens',
      description:
        'Understand messaging shifts, new USPs, and how their positioning is evolving to counter yours.',
    },
    {
      title: 'Pricing Strategy',
      description:
        'Get instant analysis of competitor price moves with recommended responses for your own pricing.',
    },
    {
      title: 'Business Opportunity Radar',
      description:
        'AI identifies market gaps and openings the moment they appear in competitor activity.',
    },
    {
      title: 'Risk Alerts',
      description:
        'Compliance changes, regulatory updates, and legal risks flagged before they become expensive problems.',
    },
  ],
  demoTitle: 'AI Insight — Live',
  demoBadge: 'Processing',
  demoSite: 'competitor.com/pricing',
  demoChange: '⚡ Pro plan reduced from $89 → $69/month',
  demoAnalysis:
    'This 22% price cut signals a demand generation push — likely targeting your customer base. Their new CTA “Start Free, Upgrade Anytime” combined with this reduction suggests they’re lowering acquisition friction. Recommended action: Reinforce your ROI messaging within 48 hours. Emphasize outcomes over features. A comparison landing page could capture traffic from buyers now reevaluating. Your current pricing may appear higher without additional context.',
  demoActions: [
    { label: 'Share to Slack' },
    { label: 'Mark as priority' },
    { label: 'Take Action →' },
  ],
} as const

const INDUSTRIES = {
  compactMode: true,
  eyebrow: "Use Cases",
  headline: "Built for teams who",
  headlineHighlight: "can’t afford to be second.",
  subheadline:
    "From solo founders to enterprise teams — across 205+ industries worldwide.",
  items: [
    {
      icon: '🛒',
      title: 'E-commerce Brands',
      description:
        'Monitor competitor product pages, pricing, and promotions. React to price drops and inventory changes before your competitors capture the sale.',
      realWin:
        'One merchant recovered $14k/month in revenue by responding to competitor price drops within minutes instead of days.',
    },
    {
      icon: '🏢',
      title: 'Marketing Agencies',
      description:
        'Monitor all client websites automatically. Know instantly when copy changes, tracking breaks, or unexpected updates go live across your entire portfolio.',
      realWin:
        'Agency caught a removed tracking pixel within 6 minutes — saving a $200k ad campaign from running dark.',
    },
    {
      icon: '🏡',
      title: 'Real Estate Teams',
      description:
        'Get instant alerts when properties matching your criteria go live. Monitor agent sites, MLS feeds, and listing aggregators automatically — no more manual refreshing.',
      realWin:
        'Investor submitted an offer 11 minutes after listing — competing offers came in hours later.',
    },
    {
      icon: '⚖️',
      title: 'Compliance & Legal',
      description:
        'Track government portals, regulatory bodies, and legal databases for policy changes. Never be blindsided by a regulatory shift that affects your business or clients.',
      realWin:
        'Law firm identified a compliance change 48 hours before it took effect — briefed clients before competitors knew it happened.',
    },
    {
      icon: '🚀',
      title: 'SaaS & Product Teams',
      description:
        'Know when competitors update pricing pages, launch new features, or pivot positioning. Detect product launches before they’re announced publicly.',
      realWin:
        'Product team detected a competitor’s beta page 3 weeks before official launch — used the time to accelerate their own roadmap.',
    },
    {
      icon: '📣',
      title: 'PR & Communications',
      description:
        'Track brand mentions, search result shifts, and news coverage in real time. Know immediately when your brand or competitors make headlines — and respond first.',
      realWin:
        'PR team responded to a crisis mention within 4 minutes of publication — limiting spread before it went viral.',
    },
  ],
} as const

const AI_INTELLIGENCE = {
  eyebrow: 'AI Intelligence',
  headline: 'Other tools say what changed.',
  headlineHighlight: 'We tell you what to do.',
  subheadline:
    'Every change Pulzifi detects gets analyzed by AI trained for business strategy. You get context, recommended actions, and opportunity scoring — not a screenshot with a timestamp.',
  tabs: [
    {
      label: 'For Marketing',
      items: [
        {
          title: 'Marketing Lens',
          body: 'Understand messaging shifts, new USPs, and how their positioning is evolving to counter yours.',
        },
        {
          title: 'Pricing Strategy',
          body: 'Get instant analysis of competitor price moves with recommended responses for your own pricing.',
        },
        {
          title: 'Business Opportunity Radar',
          body: 'AI identifies market gaps and openings the moment they appear in competitor activity.',
        },
        {
          title: 'Risk Alerts',
          body: 'Compliance changes, regulatory updates, and legal risks flagged before they become expensive problems.',
        },
      ],
    },
    {
      label: 'For Product Teams',
      items: [
        {
          title: 'Feature Launch Detection',
          body: 'Detect competitor product launches before they\'re announced publicly.',
        },
        {
          title: 'Pricing Tier Changes',
          body: 'Track structure updates and packaging shifts in real time.',
        },
        {
          title: 'GTM Strategy Shifts',
          body: 'Spot positioning pivots that signal a roadmap change.',
        },
        {
          title: 'Beta Page Discovery',
          body: 'Find competitor beta pages weeks before official launch.',
        },
      ],
    },
  ],
} as const

const COMPARISON = {
  eyebrow: 'vs Competitors',
  headline: 'Why teams switch to',
  headlineHighlight: 'Pulzifi.',
  columns: [
    { name: 'Pulzifi', isUs: true },
    { name: 'Hexowatch', isUs: false },
    { name: 'Visualping', isUs: false },
  ],
  rows: [
    {
      feature: 'Real-time monitoring',
      cells: [
        { state: 'yes', note: 'Every 5 min' },
        { state: 'yes', note: 'Every 5 min' },
        { state: 'yes', note: 'Every 5 min' },
      ],
    },
    {
      feature: 'AI-powered strategic analysis',
      cells: [
        { state: 'yes', note: 'Full AI context' },
        { state: 'no', note: 'Raw alerts only' },
        { state: 'no', note: 'Raw alerts only' },
      ],
    },
    {
      feature: 'Business opportunity scoring',
      cells: [
        { state: 'yes', note: 'Included' },
        { state: 'no' },
        { state: 'no' },
      ],
    },
    {
      feature: 'Industry-specific insights',
      cells: [
        { state: 'yes', note: '205+ industries' },
        { state: 'no' },
        { state: 'no' },
      ],
    },
    {
      feature: 'Actionable recommendations',
      cells: [
        { state: 'yes', note: 'Per alert' },
        { state: 'no' },
        { state: 'no' },
      ],
    },
    {
      feature: 'Slack / Teams / SMS alerts',
      cells: [
        { state: 'yes' },
        { state: 'yes' },
        { state: 'yes' },
      ],
    },
    {
      feature: 'JavaScript-rendered pages',
      cells: [
        { state: 'yes' },
        { state: 'yes' },
        { state: 'partial', note: 'Limited' },
      ],
    },
    {
      feature: 'Team collaboration workspaces',
      cells: [
        { state: 'yes', note: 'Unlimited' },
        { state: 'partial', note: 'Limited' },
        { state: 'partial', note: 'Limited' },
      ],
    },
    {
      feature: 'White-label options',
      cells: [
        { state: 'yes' },
        { state: 'no' },
        { state: 'no' },
      ],
    },
    {
      feature: 'Starting price',
      cells: [
        { state: 'yes', note: '$20/month' },
        { state: 'partial', note: '~$30/month' },
        { state: 'partial', note: '~$25/month' },
      ],
    },
  ],
} as const

// Plans data — single source of truth. Stored in `plans` collection and
// referenced from both the landing pricing block and the pricing-page global.
// Migrated from https://pulzifi.com/pricing.
const PLANS = [
  {
    name: 'Starter',
    price: '$20',
    period: '/month',
    tagline: 'Perfect for individual users and business owners',
    features: [
      { text: '1 Workspace', included: true },
      { text: 'Up to 5 single pages', included: true },
      { text: 'Up to 1 user account', included: true },
      { text: 'Advanced 4 AI Insights', included: true },
      { text: '1 Week Storage', included: true },
      { text: 'Email and Messages alerts', included: true },
    ],
    ctaLabel: 'Try it Now',
    ctaHref: '/register',
    highlighted: false,
  },
  {
    name: 'Professional',
    price: '$62',
    period: '/month',
    tagline: 'Perfect for Growing Businesses Ready to Scale',
    features: [
      { text: 'Unlimited Workspaces', included: true },
      { text: 'Up to 25 single pages', included: true },
      { text: 'Up to 5 user accounts', included: true },
      { text: 'Advanced unlimited AI Insights', included: true },
      { text: 'Multi-channel alerts (Email, Messages, Teams, Slack, Telegram)', included: true },
      { text: '1 month storage', included: true },
      { text: 'Priority support', included: true },
    ],
    ctaLabel: 'Try it Now',
    ctaHref: '/register',
    highlighted: true,
    popularBadge: 'Most Popular',
  },
  {
    name: 'Enterprise',
    price: 'Custom',
    period: '',
    tagline: 'Comprehensive and Scalable Solutions for Large Organizations',
    features: [
      { text: 'Unlimited Workspaces', included: true },
      { text: 'Unlimited user accounts', included: true },
      { text: 'Unlimited single pages', included: true },
      { text: 'Advanced unlimited AI Insights', included: true },
      { text: 'Multi-channel alerts', included: true },
      { text: '3 month storage', included: true },
      { text: 'Priority support', included: true },
    ],
    ctaLabel: 'Schedule a Call',
    ctaHref: '/contact',
    highlighted: false,
  },
] as const

// Header copy for the landing pricing block (different tone from /pricing).
const LANDING_PRICING_HEADER = {
  eyebrow: 'Pricing',
  headline: 'Transparent pricing.',
  headlineHighlight: 'Zero surprises.',
  subheadline:
    'No contracts. Cancel anytime. Start monitoring in under 60 seconds.',
  guaranteeNote:
    '🔒 No credit card required to start · Cancel anytime · 14-day free trial on all plans',
} as const

// Header + FAQ for the standalone /pricing page. Migrated from prod.
const PRICING_PAGE_HEADER = {
  headline: 'Simple,',
  headlineHighlight: 'Transparent Pricing',
  subheadline:
    'Choose a plan that fits your business needs and budget. No hidden fees, no surprises, just straightforward pricing.',
} as const

const PRICING_PAGE_FAQ = {
  eyebrow: 'FAQ',
  headline: 'Frequently Asked Questions',
  items: [
    {
      question: 'What is Pulzifi and how does it work?',
      answer:
        'Pulzifi is a platform that tracks changes across websites, competitor pages, news sources, and industry portals in real time. Add the pages you care about, choose your tracking preferences, and receive instant alerts with AI-generated insights.',
    },
    {
      question: 'What types of changes can Pulzifi detect?',
      answer:
        'Content updates, real estate price changes, marketing campaign updates, news releases, SEO changes, product launches, and policy changes — plus AI-driven business insights on audience shifts and industry trends.',
    },
    {
      question: 'How is Pulzifi different from other website monitoring tools?',
      answer:
        'Pulzifi goes beyond basic alerts. It provides AI summaries and recommendations, industry-specific templates, white-label options, and strategy-focused insights instead of raw notifications.',
    },
    {
      question: 'How secure is my data with Pulzifi?',
      answer:
        'Enterprise-level security and privacy best practices. Data is encrypted in transit and at rest, authentication is secure, role-based access and audit logs are available, and we never share your data with third parties.',
    },
  ],
} as const

const TESTIMONIALS = {
  eyebrow: 'Loved by Teams',
  headline: 'What customers say after switching.',
  items: [
    {
      quote:
        'Pulzifi caught a keyword shift in a competitor’s blog before any SEO tool did. That insight helped us reposition our content strategy fast — and we ranked #1 within 3 weeks.',
      author: 'Alex D.',
      role: 'SEO Analyst, Growth Agency',
    },
    {
      quote:
        'Pulzifi replaced hours of manual checks. I know instantly when a competitor changes pricing. I get context, not just alerts. It’s become my unfair competitive advantage.',
      author: 'Julia C.',
      role: 'Product Manager, Series B SaaS',
    },
    {
      quote:
        'We caught subtle homepage and pricing tests from our top competitors — things traditional trackers completely missed. Pulzifi operates at a level no other tool does.',
      author: 'Johanna T.',
      role: 'Founder, B2B SaaS',
    },
  ],
} as const

const FAQ = {
  eyebrow: 'FAQ',
  headline: 'Questions answered.',
  subheadline: 'Everything you need to know before you start monitoring.',
  items: [
    {
      question: 'How is Pulzifi different from tools like Hexowatch or Visualping?',
      answer:
        'Traditional monitoring tools tell you something changed. Pulzifi tells you what it means for your business. Every alert comes with AI-powered analysis — the strategic context, recommended actions, and business opportunity scoring. It’s the difference between a smoke alarm and a fire safety system.',
    },
    {
      question: 'Can Pulzifi monitor pages behind login walls or JavaScript apps?',
      answer:
        'Yes. Pulzifi supports JavaScript-rendered SPAs, dynamic content, and select login-gated pages. Our crawler is built to handle modern web infrastructure that simple ping-based tools can’t reach.',
    },
    {
      question: 'How fast are alerts delivered?',
      answer:
        'On the Professional plan, you can set monitoring frequency as often as every 5 minutes. Alerts are delivered via your chosen channel (email, Slack, SMS, Teams, Telegram, or webhook) typically within 2–3 minutes of the change being detected.',
    },
    {
      question: 'Is my data secure?',
      answer:
        'Enterprise-level security. All data is encrypted in transit and at rest. We support role-based access controls, audit logs, and secure cloud infrastructure. We never sell or share your monitored data with third parties — ever.',
    },
    {
      question: 'Do you offer a free trial?',
      answer:
        'Yes — all plans include a 14-day free trial. No credit card required to start. You’ll have full access to the plan’s features so you can test Pulzifi on your real competitors and pages before committing.',
    },
    {
      question: 'What industries does Pulzifi work best for?',
      answer:
        'Pulzifi is active across 205+ industries — from e-commerce and SaaS to real estate, law, finance, PR, and compliance. The AI insight engine is trained to understand business context across all major verticals, so recommendations are relevant to your specific situation.',
    },
  ],
} as const

const FINAL_CTA = {
  eyebrow: 'Get Started',
  headline: 'Stop reacting. Start',
  headlineHighlight: 'winning.',
  subtext:
    'Join 2,500+ businesses who now know what their competitors are doing — before they do it to you.',
  primaryLabel: 'Start Monitoring Free',
  primaryHref: '/register',
  secondaryLabel: 'Talk to our team',
  secondaryHref: '/contact',
  riskNote: 'No credit card required · 14-day free trial · Cancel anytime',
  variant: 'primary' as const,
}

const FOOTER_LINKS = {
  Product: [
    { label: 'How it works', href: '#how' },
    { label: 'Use Cases', href: '#usecases' },
    { label: 'Pricing', href: '#pricing' },
    { label: 'Integrations', href: '#' },
  ],
  Company: [
    { label: 'About', href: '#' },
    { label: 'Contact', href: '/contact' },
    { label: 'Blog', href: '/blog' },
    { label: 'Careers', href: '#' },
  ],
  Legal: [
    { label: 'Privacy Policy', href: '#' },
    { label: 'Terms of Service', href: '#' },
    { label: 'Security', href: '#' },
    { label: 'GDPR', href: '#' },
  ],
} as const

export type SeedResult = { seeded: boolean; reason?: string }

type PlanId = string | number

export async function seedCMSIfEmpty(payload: Payload): Promise<SeedResult> {
  const plansResult = await payload.find({ collection: 'plans', limit: 1 })
  const plansEmpty = plansResult.totalDocs === 0

  const landing = await payload.findGlobal({ slug: 'landing', depth: 0 })
  const landingEmpty = !(landing as { blocks?: unknown[] }).blocks?.length

  if (plansEmpty && landingEmpty) {
    await seedAll(payload)
    return { seeded: true }
  }

  if (plansEmpty) {
    // Existing landing data with old inline pricing plans — migrate to relationship.
    const planIds = await seedPlans(payload)
    await migrateLandingPricingBlock(payload, planIds)
    await seedPricingPage(payload, planIds)
    return { seeded: true, reason: 'migrated-to-plans-collection' }
  }

  return { seeded: false, reason: 'already-seeded' }
}

// Idempotent: returns existing plan IDs if any, otherwise creates from PLANS.
export async function seedPlans(payload: Payload): Promise<PlanId[]> {
  const existing = await payload.find({ collection: 'plans', limit: 100 })
  if (existing.totalDocs > 0) {
    const byName = new Map<string, PlanId>(
      existing.docs.map((d: { id: PlanId; name: string }) => [d.name, d.id]),
    )
    return PLANS.map((p) => byName.get(p.name)).filter(
      (id): id is PlanId => id !== undefined,
    )
  }
  const created: { id: PlanId }[] = []
  for (const plan of PLANS) {
    const doc = await payload.create({
      collection: 'plans',
      data: JSON.parse(JSON.stringify(plan)),
    })
    created.push(doc as { id: PlanId })
  }
  return created.map((d) => d.id)
}

// Replaces inline plans on existing landing pricing block(s) with relationship IDs.
// Other blocks are preserved as-is.
async function migrateLandingPricingBlock(
  payload: Payload,
  planIds: PlanId[],
): Promise<void> {
  const landing = await payload.findGlobal({ slug: 'landing', depth: 0 })
  const blocks = ((landing as { blocks?: unknown[] }).blocks ?? []) as Array<
    Record<string, unknown>
  >
  const migrated = blocks.map((b) => {
    if (b.blockType !== 'pricing') return b
    return {
      blockType: 'pricing',
      eyebrow: b.eyebrow,
      headline: b.headline,
      headlineHighlight: b.headlineHighlight,
      subheadline: b.subheadline,
      guaranteeNote: b.guaranteeNote,
      plans: planIds,
    }
  })
  await payload.updateGlobal({
    slug: 'landing',
    data: { blocks: JSON.parse(JSON.stringify(migrated)) },
  })
}

export async function seedPricingPage(
  payload: Payload,
  planIds: PlanId[],
): Promise<void> {
  await payload.updateGlobal({
    slug: 'pricing-page',
    data: JSON.parse(
      JSON.stringify({
        header: PRICING_PAGE_HEADER,
        plans: planIds,
        faq: PRICING_PAGE_FAQ,
      }),
    ),
  })
}

export async function seedAll(payload: Payload): Promise<void> {
  await payload.updateGlobal({
    slug: 'navbar',
    data: {
      links: NAV_LINKS.map((l) => ({ label: l.label, href: l.href })),
      signinLabel: 'Sign in',
      signinHref: '/login',
      primaryCtaLabel: 'Start Monitoring Free',
      primaryCtaHref: '/register',
    },
  })

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
      tagline: "Built for teams who can't afford to be second.",
      socialLinks: [
        { platform: 'twitter', href: 'https://twitter.com/pulzifi' },
        { platform: 'linkedin', href: 'https://linkedin.com/company/pulzifi' },
      ],
      copyrightText: `© ${new Date().getFullYear()} Pulzifi. All rights reserved.`,
    },
  })

  const planIds = await seedPlans(payload)

  // Slim 8-block landing: hero, problem, how-it-works, ai-intelligence,
  // industries, pricing, faq, cta — logos/stats/testimonials/comparison removed
  const blocks = [
    { blockType: 'hero', ...HERO },
    { blockType: 'problem', ...PROBLEM },
    { blockType: 'how-it-works', ...HOW_IT_WORKS },
    { blockType: 'ai-intelligence', ...AI_INTELLIGENCE },
    { blockType: 'industries', ...INDUSTRIES },
    { blockType: 'pricing', ...LANDING_PRICING_HEADER, plans: planIds },
    { blockType: 'faq', ...FAQ },
    { blockType: 'cta', ...FINAL_CTA },
  ]

  await payload.updateGlobal({
    slug: 'landing',
    data: {
      blocks: JSON.parse(JSON.stringify(blocks)),
    },
  })

  await seedPricingPage(payload, planIds)

  await payload.updateGlobal({
    slug: 'theme',
    data: {
      pageBg: '#f3f3f3',
      pageBgAlt: '#ebebef',
      cardBg: '#ffffff',
      darkSurface: '#29144c',
      inkPrimary: '#131313',
      inkSecondary: '#444141',
      accentPrimary: '#7c3aed',
      accentMuted: '#a78bfa',
      accentGold: '#f59e0b',
      accentTeal: '#14b8a6',
      border: 'rgba(0, 0, 0, 0.08)',
      borderStrong: 'rgba(0, 0, 0, 0.16)',
    },
  })
}

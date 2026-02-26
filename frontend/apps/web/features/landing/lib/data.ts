export const NAV_LINKS = [
  { label: 'Home', href: '#' },
  { label: 'Product', href: '#how-it-works' },
  { label: 'How to use', href: '#industries' },
  { label: 'Pricing', href: '#pricing' },
  { label: 'Contact', href: '#footer' },
] as const

export const STATS = [
  { value: '25,000+', label: 'Strategy Decisions Created' },
  { value: '150,493', label: 'Monitored Pages' },
  { value: '5.0', label: 'Customer Review' },
  { value: '205+', label: 'Industries' },
] as const

export const HOW_IT_WORKS_STEPS = [
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

export const FEATURE_CARDS = [
  {
    title: 'Track Every Word Change',
    description:
      'Instantly see what messaging they adjust, add, or remove to understand shifts in strategy or positioning. Compare text versions and highlight key differences.',
    image: '/images/landing/text-changes.png',
  },
  {
    title: 'Flexible Monitoring Schedules',
    description: 'Check every 5 minutes, once a day or once a month. You set the tempo.',
    image: '/images/landing/monitoring-schedule.png',
  },
  {
    title: 'Visual Comparison',
    description:
      'See exactly what changed, revealing hidden tweaks text, branding, or new sections that impact strategy.',
    image: '/images/landing/visual-comparison.png',
  },
  {
    title: 'Choose How You Get Notified',
    description:
      'Get screenshot alerts of every update, via email, text messages, team channels and more. Stay informed your way.',
    image: '/images/landing/notifications.png',
  },
] as const

export const INSIGHT_CARDS = [
  { label: 'Analyze alerts with...', color: 'bg-white', tagColor: '' },
  { label: 'Marketing Lens', color: 'bg-white', tagColor: 'bg-[#ebf0ff]' },
  { label: 'Brand and Pricing Strategy', color: 'bg-white', tagColor: 'bg-[#ebfff5]' },
  { label: 'Business Opportunities', color: 'bg-white', tagColor: 'bg-[#fff9f5]' },
] as const

export const PRICING_PLANS = [
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

export const TESTIMONIALS = [
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

export const FAQ_ITEMS = [
  {
    question: 'What is Pulzifi and how does it work?',
    answer:
      `Pulzifi is an AI-powered website monitoring and market intelligence platform that tracks changes across websites, competitor pages, news sources, and industry portals in real time.

You simply add the pages you want to monitor, choose what to track such as price updates, content changes, new listings, or policy updates, and Pulzifi sends instant alerts with AI insights explaining what changed and why it matters.

Pulzifi helps real estate professionals, marketing agencies, and business teams stay ahead of competitors, trends, and opportunities without checking websites manually.`,
  },
  {
    question: 'What types of changes can Pulzifi detect?',
    answer:
      `Pulzifi can detect a wide range of website and market changes, including:
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
    answer:
      `Most website monitoring tools only show that something changed. Pulzifi goes further with AI-driven market intelligence.

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
    answer:
      `Pulzifi is built with enterprise-level security and privacy best practices.

Your monitored pages, alerts, and insights are encrypted in transit and at rest. Access is protected with secure authentication, and we never sell or share your data with third parties.

For institutions like real estate education programs or agencies managing client websites, Pulzifi also supports role-based access, audit logs, and secure cloud infrastructure to keep your data safe.`,
  },
] as const

export const FOOTER_LINKS = {
  Features: [
    { label: 'Product', href: '#how-it-works' },
    { label: 'Use Cases', href: '#industries' },
    { label: 'Pricing', href: '#pricing' },
  ],
  Support: [
    { label: 'Help', href: '#' },
    { label: 'FAQ', href: '#faq' },
    { label: 'Contact', href: '#' },
  ],
  Legal: [
    { label: 'Privacy Policy', href: '#' },
    { label: 'Terms of Services', href: '#' },
  ],
} as const

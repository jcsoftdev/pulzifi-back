import { headers } from 'next/headers'
import { redirect } from 'next/navigation'
import { extractTenantFromHostname } from '@workspace/shared-http'
import Link from 'next/link'
import {
  AlertCircle,
  ArrowRight,
  BarChart3,
  Bell,
  BookOpen,
  Brain,
  CheckCircle2,
  ChevronDown,
  Clock,
  Eye,
  FileText,
  Globe,
  History,
  Image,
  LineChart,
  Lock,
  Radar,
  RefreshCw,
  Shield,
  Sparkles,
  Star,
  Target,
  TrendingUp,
  Users,
  Webhook,
  Zap,
} from 'lucide-react'

// ─── Data ────────────────────────────────────────────────────────────────────

const NAV_LINKS = [
  { href: '#features', label: 'Features' },
  { href: '#how-it-works', label: 'How it works' },
  { href: '#insights', label: 'AI Insights' },
  { href: '#use-cases', label: 'Use cases' },
  { href: '#faq', label: 'FAQ' },
]

const STATS = [
  { value: '10M+', label: 'Pages monitored' },
  { value: '<2 min', label: 'Average detection time' },
  { value: '5 lenses', label: 'AI analysis types' },
  { value: '99.9%', label: 'Uptime SLA' },
]

const FEATURES = [
  {
    icon: Eye,
    title: 'Continuous Web Monitoring',
    description:
      'Add any public URL and Pulzifi checks it on your schedule — hourly, daily, or weekly. We capture full-page screenshots, raw HTML, and text hashes to detect every meaningful change, no matter how subtle.',
    tags: ['Playwright-powered', 'Content hashing', 'Visual snapshots'],
    color: 'violet',
  },
  {
    icon: Brain,
    title: 'AI-Powered Strategic Analysis',
    description:
      'Every detected change automatically triggers a five-lens AI analysis: executive overview, marketing signals, market positioning, business opportunities, and talent indicators. Not just "something changed" — but what it means.',
    tags: ['5 analysis types', 'OpenRouter AI', 'Auto-generated'],
    color: 'indigo',
  },
  {
    icon: Bell,
    title: 'Intelligent Alert System',
    description:
      'Configurable alerts sent the moment a change is detected. Route notifications to email, Slack, Teams, Discord, or any custom webhook endpoint. Mark alerts as read, filter by workspace, and never miss a market move.',
    tags: ['Email & Slack', 'Custom webhooks', 'Configurable triggers'],
    color: 'blue',
  },
  {
    icon: History,
    title: 'Full Change History & Diffs',
    description:
      'Every check is stored — screenshots, HTML, and text diffs. Browse a rich timeline of changes per page, compare versions side by side, and understand exactly what shifted in competitor copy, pricing, or positioning.',
    tags: ['Side-by-side diff', 'Screenshot archive', 'Full audit trail'],
    color: 'emerald',
  },
  {
    icon: Users,
    title: 'Collaborative Workspaces',
    description:
      'Organize monitored pages by team, project, or client. Invite members, assign roles, and share insights across your organization. Each workspace is fully isolated with role-based permissions.',
    tags: ['Role-based access', 'Team invites', 'Multi-tenant'],
    color: 'amber',
  },
  {
    icon: FileText,
    title: 'Exportable Intelligence Reports',
    description:
      'Generate polished PDF reports summarizing changes, AI insights, and trends for a given workspace or time range. Share competitive briefs with stakeholders without any manual formatting.',
    tags: ['PDF export', 'Workspace reports', 'Shareable briefs'],
    color: 'rose',
  },
]

const STEPS = [
  {
    step: '01',
    icon: Globe,
    title: 'Add pages to monitor',
    description:
      'Paste any URL into your workspace — competitor homepages, pricing pages, product releases, job boards, regulatory sites, or industry blogs. Tag pages for easy filtering and set your preferred check interval.',
    detail: 'Supports any publicly accessible URL. No browser extension or code required.',
    color: 'violet',
  },
  {
    step: '02',
    icon: Radar,
    title: 'We detect every change',
    description:
      'Our engine visits each page on schedule using a real browser (Playwright), captures a full screenshot, stores the raw HTML, and computes a content hash. Any deviation from the last known state triggers the pipeline.',
    detail: 'SHA-256 content hashing ensures even small copy changes are caught.',
    color: 'indigo',
  },
  {
    step: '03',
    icon: Brain,
    title: 'AI generates strategic insights',
    description:
      'Change detected? Our AI instantly analyzes the content shift across five strategic lenses and writes a structured intelligence report — delivered as alerts to your team and stored in your workspace for reference.',
    detail: 'Analysis runs automatically in the background within seconds of detection.',
    color: 'blue',
  },
  {
    step: '04',
    icon: Zap,
    title: 'Act on intelligence',
    description:
      'Your team receives rich alerts with full context. Explore change diffs, review AI analysis, generate reports for stakeholders, and route insights directly into Slack, CRMs, or any tool via webhooks.',
    detail: 'Close the loop from detection to decision in one platform.',
    color: 'emerald',
  },
]

const INSIGHT_TYPES = [
  {
    title: 'Executive Overview',
    color: 'violet',
    icon: BarChart3,
    description: 'A concise, board-ready summary of what changed and its strategic importance. Cuts through noise to the signal that matters for leadership.',
    example: 'Competitor quietly removed annual pricing tier — likely responding to churn. Their new focus appears to be monthly recurring revenue from SMB.',
  },
  {
    title: 'Marketing & Messaging',
    color: 'blue',
    icon: TrendingUp,
    description: 'Detects shifts in brand voice, taglines, value propositions, and campaign language. Know when competitors are repositioning before your customers do.',
    example: 'Homepage hero shifted from "save time" to "reduce risk" — a pivot toward compliance-driven buyers and regulated industries.',
  },
  {
    title: 'Market Analysis',
    color: 'indigo',
    icon: LineChart,
    description: 'Identifies changes in target segments, competitive differentiation, and market positioning. Surfaces opportunities to outflank competitors in underserved areas.',
    example: 'Three competitors updated their enterprise pages simultaneously — signals a market-wide shift toward larger deal sizes and longer sales cycles.',
  },
  {
    title: 'Business Opportunities',
    color: 'amber',
    icon: Target,
    description: 'Surfaces 3–5 concrete, actionable opportunities derived from the detected change. Goes beyond observation to prescribe what your team should do next.',
    example: 'Opportunity: Competitor dropped free tier — run a win-back campaign targeting their churned users with your free plan as the lead offer.',
  },
  {
    title: 'Talent & Org Signals',
    color: 'emerald',
    icon: Users,
    description: 'Reads career page changes, new job titles, and headcount signals to infer strategic priorities and organizational investments your rivals are making.',
    example: 'Five new VP-level hires in "Revenue Operations" posted this week — indicating a serious push to systematize their enterprise go-to-market.',
  },
]

const TESTIMONIALS = [
  {
    quote:
      'Pulzifi changed how our sales team prepares for calls. We now walk in knowing exactly what changed on a prospect\'s competitor site last week. Win rates are up.',
    name: 'Sarah Chen',
    role: 'VP of Sales',
    company: 'GrowthStack',
    avatar: 'SC',
    color: 'violet',
  },
  {
    quote:
      'I used to spend two hours every Monday manually checking competitor pages. Pulzifi does it automatically and gives me better analysis than I ever could on my own.',
    name: 'Marcus Webb',
    role: 'Senior Product Manager',
    company: 'Orion Software',
    avatar: 'MW',
    color: 'indigo',
  },
  {
    quote:
      'The AI insights are genuinely useful — not generic summaries. When a competitor overhauled their pricing page, Pulzifi surfaced three specific opportunities for us within minutes.',
    name: 'Priya Nair',
    role: 'Head of Strategy',
    company: 'Meridian Analytics',
    avatar: 'PN',
    color: 'blue',
  },
]

const USE_CASES = [
  {
    role: 'Sales & Business Development',
    icon: Target,
    headline: 'Walk into every call with competitive intelligence your rivals don\'t have.',
    description:
      'Monitor competitor pricing, messaging, and product pages for the companies your prospects care about most. Get alerted the moment something changes so your team can craft the perfect counter-narrative — before the meeting.',
    bullets: [
      'Get notified when a competitor changes pricing before your customer notices',
      'Track competitor case studies and social proof to sharpen your pitch',
      'Detect product launches early and prepare objection-handling responses',
      'Monitor win/loss triggers in real time as competitor messaging evolves',
    ],
    color: 'violet',
  },
  {
    role: 'Product Management',
    icon: Sparkles,
    headline: 'Build what the market needs. Informed by what competitors are shipping.',
    description:
      'Stay on top of competitor feature releases, changelog updates, roadmap signals, and UX changes. Feed real market intelligence into your own roadmap decisions rather than relying on sporadic manual checks.',
    bullets: [
      'Track competitor changelog and feature pages for release patterns',
      'Detect UX and onboarding flow changes that signal product-market fit shifts',
      'Monitor job postings for signals about where competitors are investing engineering resources',
      'Generate AI summaries of positioning changes to brief your product team',
    ],
    color: 'indigo',
  },
  {
    role: 'Marketing & Brand',
    icon: TrendingUp,
    headline: 'Know every messaging shift before your audience does.',
    description:
      'Monitor how competitors evolve their brand voice, run campaigns, and adjust their value proposition. Make data-driven decisions on positioning, copy, and channel strategy — not guesswork.',
    bullets: [
      'Detect tagline and hero copy changes across competitor sites',
      'Track campaign pages and landing pages for offer and positioning shifts',
      'Monitor competitor blog and content strategy for topic clusters and trends',
      'Benchmark messaging changes to inform your own content calendar',
    ],
    color: 'blue',
  },
  {
    role: 'Strategy & Intelligence',
    icon: Shield,
    headline: 'Maintain a live picture of market movements that affect your business.',
    description:
      'Build a comprehensive intelligence function across competitors, partners, regulators, and market sources. Pulzifi acts as an always-on analyst that never misses a change — no matter the day or hour.',
    bullets: [
      'Monitor regulatory and compliance sites for policy updates',
      'Track industry analyst sites, news portals, and publication pages',
      'Set up workspace "watchlists" per market segment or geography',
      'Generate executive-ready reports summarizing the week\'s intelligence',
    ],
    color: 'emerald',
  },
]

const FAQS = [
  {
    q: 'What types of websites can Pulzifi monitor?',
    a: 'Any publicly accessible web page — competitor sites, pricing pages, product pages, careers pages, regulatory sources, news publications, partner portals, and more. If a browser can load it, Pulzifi can monitor it.',
  },
  {
    q: 'How does change detection work?',
    a: 'Our engine visits each monitored page using a real browser (Playwright), captures a full screenshot and raw HTML, and computes a SHA-256 hash of the visible text content. A change is flagged when the hash differs from the last recorded state — filtering out cosmetic or dynamic elements like timestamps.',
  },
  {
    q: 'How quickly are changes detected?',
    a: 'Detection speed depends on your configured check interval. With hourly monitoring, changes are typically surfaced within 60 minutes of occurring. AI insight generation runs within seconds of a change being confirmed.',
  },
  {
    q: 'What AI model powers the insights?',
    a: 'Pulzifi uses OpenRouter to access state-of-the-art large language models. The default model is optimized for strategic business analysis. Enterprise plans can configure custom model preferences.',
  },
  {
    q: 'Is my monitored data kept private?',
    a: 'Yes. Each organization operates in a fully isolated tenant environment with schema-level database separation. Your monitored URLs, change history, and AI insights are never shared with other customers or used to train models.',
  },
  {
    q: 'Can I integrate Pulzifi with my existing tools?',
    a: 'Yes. Pulzifi supports outbound webhooks, email notifications, and direct integrations with Slack, Microsoft Teams, and Discord. Any change or insight can trigger a payload to your CRM, workflow automation, or custom system.',
  },
]

// ─── Component ───────────────────────────────────────────────────────────────

export default async function LandingPage() {
  const incomingHeaders = await headers()
  const host = incomingHeaders.get('x-forwarded-host') || incomingHeaders.get('host') || ''
  const tenant = extractTenantFromHostname(host)

  if (tenant) {
    redirect('/dashboard')
  }

  return (
    <div className="min-h-screen bg-[#06060a] text-white overflow-x-hidden">

      {/* Announcement bar */}
      <div className="bg-gradient-to-r from-violet-950/80 via-indigo-950/80 to-violet-950/80 border-b border-violet-500/15 py-2.5 px-4 text-center">
        <p className="text-xs text-white/60">
          <span className="inline-flex items-center gap-1.5 text-violet-300 font-medium">
            <Sparkles className="w-3 h-3" />
            New:
          </span>
          {' '}AI insight reports now include Talent & Org Signals — understand where competitors are investing.
          <Link href="/register" className="ml-3 text-violet-400 hover:text-violet-300 font-medium underline underline-offset-2 transition-colors">
            Try it free →
          </Link>
        </p>
      </div>

      {/* Navigation */}
      <nav className="sticky top-0 z-50 border-b border-white/[0.06] bg-[#06060a]/85 backdrop-blur-2xl">
        <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-violet-500 to-indigo-600 flex items-center justify-center shadow-[0_0_20px_rgba(124,58,237,0.35)]">
              <Radar className="w-4.5 h-4.5 text-white" />
            </div>
            <span className="text-base font-semibold tracking-tight">Pulzifi</span>
            <span className="hidden sm:block text-white/20 text-xs font-normal border border-white/10 rounded-full px-2 py-0.5 ml-1">
              Beta
            </span>
          </div>

          <div className="hidden lg:flex items-center gap-7 text-sm text-white/55">
            {NAV_LINKS.map(({ href, label }) => (
              <a key={href} href={href} className="hover:text-white transition-colors duration-150">
                {label}
              </a>
            ))}
          </div>

          <div className="flex items-center gap-2">
            <Link
              href="/login"
              className="hidden sm:block text-sm text-white/60 hover:text-white transition-colors px-4 py-2"
            >
              Sign in
            </Link>
            <Link
              href="/register"
              className="text-sm bg-white text-[#06060a] hover:bg-white/90 transition-colors px-4 py-2 rounded-lg font-semibold shadow-sm"
            >
              Get started free
            </Link>
          </div>
        </div>
      </nav>

      {/* ── Hero ─────────────────────────────────────────────────────────── */}
      <section className="relative pt-28 pb-20 px-6 overflow-hidden">
        {/* Layered background glows */}
        <div className="absolute inset-0 overflow-hidden pointer-events-none select-none">
          <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[1000px] h-[700px] bg-violet-600/[0.07] rounded-full blur-[140px]" />
          <div className="absolute top-1/3 left-1/5 w-[500px] h-[500px] bg-indigo-600/[0.05] rounded-full blur-[110px]" />
          <div className="absolute top-1/2 right-1/5 w-[400px] h-[400px] bg-blue-600/[0.04] rounded-full blur-[90px]" />
          <div
            className="absolute inset-0 opacity-[0.025]"
            style={{
              backgroundImage: `linear-gradient(rgba(255,255,255,0.6) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.6) 1px, transparent 1px)`,
              backgroundSize: '64px 64px',
            }}
          />
          {/* Radial fade at edges */}
          <div
            className="absolute inset-0"
            style={{
              background: 'radial-gradient(ellipse 80% 60% at 50% 0%, transparent 40%, #06060a 100%)',
            }}
          />
        </div>

        <div className="relative max-w-4xl mx-auto text-center">
          {/* Eyebrow badge */}
          <div className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full border border-violet-500/25 bg-violet-500/[0.08] text-violet-300 text-sm font-medium mb-8 shadow-[inset_0_1px_0_rgba(255,255,255,0.05)]">
            <span className="relative flex h-2 w-2">
              <span className="absolute inline-flex h-full w-full rounded-full bg-violet-400 opacity-75 animate-ping" />
              <span className="relative inline-flex rounded-full h-2 w-2 bg-violet-500" />
            </span>
            AI-Powered Competitive Intelligence Platform
          </div>

          <h1 className="text-5xl sm:text-6xl md:text-[4.5rem] font-bold tracking-[-0.03em] leading-[1.04] mb-6">
            <span className="text-white">Know what competitors do</span>
            <br />
            <span
              className="bg-gradient-to-r from-violet-400 via-indigo-300 to-blue-400 bg-clip-text text-transparent"
              style={{ WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent' }}
            >
              before anyone else does.
            </span>
          </h1>

          <p className="text-lg md:text-xl text-white/50 max-w-2xl mx-auto mb-4 leading-relaxed">
            Pulzifi monitors any website 24/7, detects every change the moment it happens,
            and instantly generates AI-powered strategic intelligence — so your team
            can act, not just observe.
          </p>

          <p className="text-sm text-white/35 mb-10">
            Used by sales, product, marketing, and strategy teams at B2B companies.
          </p>

          <div className="flex flex-col sm:flex-row items-center justify-center gap-3 mb-8">
            <Link
              href="/register"
              className="group flex items-center gap-2.5 bg-violet-600 hover:bg-violet-500 transition-all duration-200 px-8 py-3.5 rounded-xl font-semibold text-white text-sm shadow-[0_0_35px_rgba(124,58,237,0.3)] hover:shadow-[0_0_50px_rgba(124,58,237,0.45)]"
            >
              Start monitoring for free
              <ArrowRight className="w-4 h-4 group-hover:translate-x-0.5 transition-transform" />
            </Link>
            <Link
              href="/login"
              className="flex items-center gap-2 border border-white/[0.12] hover:border-white/20 bg-white/[0.04] hover:bg-white/[0.07] transition-all duration-200 px-8 py-3.5 rounded-xl font-medium text-white/75 text-sm"
            >
              Sign in to dashboard
            </Link>
          </div>

          {/* Trust signals */}
          <div className="flex flex-wrap items-center justify-center gap-6 text-xs text-white/30">
            {[
              { icon: CheckCircle2, text: 'No credit card required' },
              { icon: Lock, text: 'SOC 2-ready infrastructure' },
              { icon: RefreshCw, text: 'Checks run every hour' },
              { icon: Globe, text: 'Monitor any public URL' },
            ].map(({ icon: Icon, text }) => (
              <span key={text} className="flex items-center gap-1.5">
                <Icon className="w-3.5 h-3.5 text-white/25" />
                {text}
              </span>
            ))}
          </div>
        </div>

        {/* ── Rich dashboard mockup ─────────────────────────────────────── */}
        <div className="relative max-w-5xl mx-auto mt-16">
          {/* Bottom fade */}
          <div className="absolute bottom-0 left-0 right-0 h-32 bg-gradient-to-t from-[#06060a] to-transparent z-10 pointer-events-none" />

          <div className="rounded-2xl border border-white/[0.08] bg-[#0d0d14] shadow-[0_40px_100px_rgba(0,0,0,0.7),0_0_0_1px_rgba(255,255,255,0.04)] overflow-hidden">
            {/* Browser chrome */}
            <div className="flex items-center gap-3 px-5 py-3.5 border-b border-white/[0.05] bg-white/[0.015]">
              <div className="flex gap-1.5">
                <div className="w-3 h-3 rounded-full bg-[#ff5f57]" />
                <div className="w-3 h-3 rounded-full bg-[#febc2e]" />
                <div className="w-3 h-3 rounded-full bg-[#28c840]" />
              </div>
              <div className="flex-1 max-w-xs mx-auto h-6 rounded-md bg-white/[0.05] border border-white/[0.06] flex items-center justify-center gap-2 px-3">
                <Lock className="w-3 h-3 text-emerald-500/60" />
                <span className="text-xs text-white/30">acme.pulzifi.com/dashboard</span>
              </div>
              <div className="flex items-center gap-1.5">
                {[1,2,3].map((i) => (
                  <div key={i} className="w-6 h-1.5 rounded-full bg-white/[0.06]" />
                ))}
              </div>
            </div>

            {/* App layout */}
            <div className="flex h-[480px]">
              {/* Sidebar */}
              <div className="w-56 border-r border-white/[0.05] bg-white/[0.01] p-3 flex flex-col gap-0.5 shrink-0">
                <div className="flex items-center gap-2.5 px-3 py-2.5 mb-3">
                  <div className="w-7 h-7 rounded-lg bg-gradient-to-br from-violet-500 to-indigo-600 flex items-center justify-center">
                    <Radar className="w-3.5 h-3.5 text-white" />
                  </div>
                  <span className="text-sm font-semibold">Pulzifi</span>
                </div>
                {[
                  { label: 'Dashboard', active: true, icon: BarChart3 },
                  { label: 'Workspaces', active: false, icon: Globe },
                  { label: 'Alerts', active: false, icon: Bell, badge: '3' },
                  { label: 'Reports', active: false, icon: FileText },
                  { label: 'Team', active: false, icon: Users },
                ].map(({ label, active, icon: Icon, badge }) => (
                  <div
                    key={label}
                    className={`flex items-center justify-between gap-2 px-3 py-2 rounded-lg text-xs font-medium ${
                      active
                        ? 'bg-violet-500/15 text-violet-300'
                        : 'text-white/35 hover:text-white/60'
                    }`}
                  >
                    <span className="flex items-center gap-2">
                      <Icon className="w-3.5 h-3.5" />
                      {label}
                    </span>
                    {badge && (
                      <span className="text-xs bg-amber-500/20 text-amber-300 rounded-full px-1.5 py-0.5">{badge}</span>
                    )}
                  </div>
                ))}
              </div>

              {/* Main content */}
              <div className="flex-1 overflow-hidden p-5 flex flex-col gap-4">
                {/* Header */}
                <div className="flex items-center justify-between">
                  <div>
                    <div className="text-base font-semibold">Dashboard</div>
                    <div className="text-xs text-white/35 mt-0.5">Acme Corp · Competitor Intelligence</div>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="flex items-center gap-1.5 text-xs text-white/35 border border-white/[0.07] rounded-lg px-3 py-1.5">
                      <Clock className="w-3 h-3" />
                      Last 7 days
                      <ChevronDown className="w-3 h-3" />
                    </div>
                    <div className="w-7 h-7 rounded-lg bg-violet-500/15 border border-violet-500/20 flex items-center justify-center">
                      <Bell className="w-3.5 h-3.5 text-violet-400" />
                    </div>
                  </div>
                </div>

                {/* Stat cards */}
                <div className="grid grid-cols-4 gap-3">
                  {[
                    { label: 'Pages Monitored', value: '248', delta: '+12', icon: Globe, color: 'violet' },
                    { label: 'Changes Detected', value: '17', delta: '+5', icon: AlertCircle, color: 'amber' },
                    { label: 'AI Insights', value: '94', delta: '+23', icon: Brain, color: 'indigo' },
                    { label: 'Active Alerts', value: '3', delta: 'new', icon: Bell, color: 'emerald' },
                  ].map(({ label, value, delta, icon: Icon, color }) => (
                    <div key={label} className="rounded-xl p-3 border border-white/[0.07] bg-white/[0.02]">
                      <div className="flex items-center justify-between mb-2.5">
                        <Icon className={`w-3.5 h-3.5 text-${color}-400`} />
                        <span className={`text-[10px] font-medium text-${color}-400 bg-${color}-500/10 rounded-full px-1.5 py-0.5`}>
                          {delta}
                        </span>
                      </div>
                      <div className="text-xl font-bold tracking-tight">{value}</div>
                      <div className="text-[10px] text-white/35 mt-0.5 truncate">{label}</div>
                    </div>
                  ))}
                </div>

                {/* Bottom row */}
                <div className="grid grid-cols-5 gap-3 flex-1 min-h-0">
                  {/* Recent changes feed */}
                  <div className="col-span-2 rounded-xl border border-white/[0.07] bg-white/[0.02] p-3 overflow-hidden">
                    <div className="text-xs font-semibold text-white/60 mb-2.5 flex items-center justify-between">
                      Recent Changes
                      <span className="text-[10px] text-white/25 font-normal">Today</span>
                    </div>
                    <div className="space-y-2">
                      {[
                        { site: 'competitor.io/pricing', type: 'Pricing updated', ago: '2m', severity: 'high' },
                        { site: 'rival.com/features', type: 'New feature page', ago: '18m', severity: 'medium' },
                        { site: 'market-leader.com', type: 'Homepage redesign', ago: '1h', severity: 'high' },
                        { site: 'industry-blog.com', type: 'New article posted', ago: '3h', severity: 'low' },
                      ].map(({ site, type, ago, severity }) => (
                        <div key={site} className="flex items-start gap-2 py-1.5 border-b border-white/[0.04] last:border-0">
                          <div className={`w-1.5 h-1.5 rounded-full mt-1.5 shrink-0 ${
                            severity === 'high' ? 'bg-amber-400' : severity === 'medium' ? 'bg-violet-400' : 'bg-white/20'
                          }`} />
                          <div className="flex-1 min-w-0">
                            <div className="text-[10px] font-medium truncate text-white/70">{site}</div>
                            <div className="text-[10px] text-white/30">{type}</div>
                          </div>
                          <span className="text-[10px] text-white/20 shrink-0">{ago}</span>
                        </div>
                      ))}
                    </div>
                  </div>

                  {/* AI insight highlight */}
                  <div className="col-span-3 rounded-xl border border-violet-500/20 bg-gradient-to-br from-violet-500/[0.06] to-indigo-500/[0.04] p-3.5 flex flex-col">
                    <div className="flex items-center justify-between mb-2.5">
                      <div className="flex items-center gap-1.5">
                        <div className="w-5 h-5 rounded-md bg-violet-500/20 flex items-center justify-center">
                          <Brain className="w-3 h-3 text-violet-400" />
                        </div>
                        <span className="text-xs font-semibold text-violet-300">Latest AI Insight</span>
                      </div>
                      <span className="text-[10px] text-white/25">Market Analysis · 2m ago</span>
                    </div>
                    <p className="text-xs text-white/55 leading-relaxed flex-1">
                      <strong className="text-white/75">competitor.io updated their pricing page</strong> — removed the SMB tier and introduced an
                      Enterprise-only pricing gate. This signals a deliberate pivot upmarket away from
                      self-serve customers.{' '}
                      <span className="text-violet-400">
                        This creates a direct opportunity to capture their churning SMB users with a competitive migration offer.
                      </span>
                    </p>
                    <div className="mt-3 flex items-center justify-between">
                      <div className="flex gap-1.5">
                        {['Pricing shift', 'SMB opportunity', 'Enterprise pivot'].map((tag) => (
                          <span key={tag} className="text-[10px] px-2 py-0.5 rounded-full bg-violet-500/10 text-violet-400 border border-violet-500/15">
                            {tag}
                          </span>
                        ))}
                      </div>
                      <span className="text-[10px] text-violet-400/70 underline cursor-pointer">View all 5 lenses →</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ── Stats strip ──────────────────────────────────────────────────── */}
      <section className="py-14 px-6 border-y border-white/[0.05]">
        <div className="max-w-5xl mx-auto">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-8 text-center">
            {STATS.map(({ value, label }) => (
              <div key={label}>
                <div className="text-3xl md:text-4xl font-bold tracking-tight bg-gradient-to-b from-white to-white/60 bg-clip-text text-transparent">
                  {value}
                </div>
                <div className="text-sm text-white/35 mt-1.5">{label}</div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── Features ─────────────────────────────────────────────────────── */}
      <section id="features" className="py-28 px-6">
        <div className="max-w-6xl mx-auto">
          <div className="text-center mb-16">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-indigo-500/25 bg-indigo-500/[0.07] text-indigo-300 text-xs font-medium mb-5">
              <Zap className="w-3 h-3" />
              Platform Features
            </div>
            <h2 className="text-4xl md:text-5xl font-bold tracking-[-0.02em] mb-5">
              Everything you need to stay
              <br />
              <span className="text-white/35">ahead of the market</span>
            </h2>
            <p className="text-white/45 max-w-xl mx-auto leading-relaxed">
              Pulzifi is a complete competitive intelligence platform — from web monitoring and
              change detection to AI analysis, team collaboration, and automated reporting.
            </p>
          </div>

          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
            {FEATURES.map(({ icon: Icon, title, description, tags, color }) => (
              <div
                key={title}
                className="group rounded-2xl p-6 border border-white/[0.07] bg-white/[0.02] hover:bg-white/[0.04] hover:border-white/[0.12] transition-all duration-300 flex flex-col"
              >
                <div className={`w-11 h-11 rounded-xl bg-${color}-500/10 border border-${color}-500/15 flex items-center justify-center mb-5`}>
                  <Icon className={`w-5 h-5 text-${color}-400`} />
                </div>
                <h3 className="text-base font-semibold mb-2.5">{title}</h3>
                <p className="text-sm text-white/40 leading-relaxed mb-5 flex-1">{description}</p>
                <div className="flex flex-wrap gap-1.5">
                  {tags.map((tag) => (
                    <span
                      key={tag}
                      className={`text-[11px] px-2 py-0.5 rounded-full bg-${color}-500/[0.08] text-${color}-400/80 border border-${color}-500/10 font-medium`}
                    >
                      {tag}
                    </span>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── How it works ─────────────────────────────────────────────────── */}
      <section id="how-it-works" className="py-28 px-6 relative overflow-hidden">
        <div className="absolute inset-0 pointer-events-none">
          <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[700px] h-[700px] bg-indigo-600/[0.05] rounded-full blur-[130px]" />
        </div>

        <div className="relative max-w-5xl mx-auto">
          <div className="text-center mb-16">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-blue-500/25 bg-blue-500/[0.07] text-blue-300 text-xs font-medium mb-5">
              <Target className="w-3 h-3" />
              How It Works
            </div>
            <h2 className="text-4xl md:text-5xl font-bold tracking-[-0.02em] mb-5">
              From zero to intelligence
              <br />
              <span className="text-white/35">in four steps</span>
            </h2>
            <p className="text-white/45 max-w-lg mx-auto leading-relaxed">
              Pulzifi turns any website into a continuous intelligence source without requiring
              any technical setup, browser extensions, or manual checking.
            </p>
          </div>

          <div className="space-y-4">
            {STEPS.map(({ step, icon: Icon, title, description, detail, color }, i) => (
              <div
                key={step}
                className={`grid md:grid-cols-[auto_1fr] gap-6 rounded-2xl p-7 border border-white/[0.07] bg-white/[0.02] items-start ${
                  i % 2 === 0 ? '' : 'md:border-l-0 md:border-r-0 bg-gradient-to-r from-white/[0.02] to-transparent'
                }`}
              >
                <div className={`w-16 h-16 rounded-2xl bg-${color}-500/10 border border-${color}-500/20 flex flex-col items-center justify-center gap-1 shrink-0 relative`}>
                  <Icon className={`w-6 h-6 text-${color}-400`} />
                  <span className={`absolute -top-2 -right-2 text-[10px] font-bold text-${color}-300 bg-[#06060a] px-1.5 py-0.5 rounded-full border border-${color}-500/25 leading-none`}>
                    {step}
                  </span>
                </div>
                <div>
                  <h3 className="text-lg font-semibold mb-2">{title}</h3>
                  <p className="text-sm text-white/45 leading-relaxed mb-3">{description}</p>
                  <div className={`inline-flex items-center gap-1.5 text-xs text-${color}-400/70 bg-${color}-500/[0.06] border border-${color}-500/10 rounded-full px-3 py-1`}>
                    <CheckCircle2 className="w-3 h-3" />
                    {detail}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── AI Insights deep-dive ─────────────────────────────────────────── */}
      <section id="insights" className="py-28 px-6">
        <div className="max-w-6xl mx-auto">
          <div className="text-center mb-16">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-violet-500/25 bg-violet-500/[0.07] text-violet-300 text-xs font-medium mb-5">
              <Brain className="w-3 h-3" />
              AI Intelligence Engine
            </div>
            <h2 className="text-4xl md:text-5xl font-bold tracking-[-0.02em] mb-5">
              Not just alerts. <span className="text-white/35">Strategic intelligence.</span>
            </h2>
            <p className="text-white/45 max-w-2xl mx-auto leading-relaxed">
              When a competitor makes a move, Pulzifi doesn't just tell you something changed —
              it tells you what it means. Powered by advanced AI analysis across five
              strategic lenses, automatically generated within seconds of detection.
            </p>
          </div>

          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4 mb-6">
            {INSIGHT_TYPES.slice(0, 3).map(({ title, color, icon: Icon, description, example }) => (
              <div
                key={title}
                className={`rounded-2xl border border-${color}-500/15 bg-gradient-to-b from-${color}-500/[0.05] to-transparent p-6 flex flex-col gap-3`}
              >
                <div className="flex items-center gap-2">
                  <div className={`w-8 h-8 rounded-lg bg-${color}-500/15 flex items-center justify-center`}>
                    <Icon className={`w-4 h-4 text-${color}-400`} />
                  </div>
                  <span className={`text-sm font-semibold text-${color}-200`}>{title}</span>
                </div>
                <p className="text-sm text-white/45 leading-relaxed">{description}</p>
                <div className="mt-auto p-3 rounded-xl bg-white/[0.03] border border-white/[0.06]">
                  <p className="text-xs text-white/50 leading-relaxed italic">"{example}"</p>
                </div>
              </div>
            ))}
          </div>
          <div className="grid md:grid-cols-2 gap-4">
            {INSIGHT_TYPES.slice(3).map(({ title, color, icon: Icon, description, example }) => (
              <div
                key={title}
                className={`rounded-2xl border border-${color}-500/15 bg-gradient-to-b from-${color}-500/[0.05] to-transparent p-6 flex flex-col gap-3`}
              >
                <div className="flex items-center gap-2">
                  <div className={`w-8 h-8 rounded-lg bg-${color}-500/15 flex items-center justify-center`}>
                    <Icon className={`w-4 h-4 text-${color}-400`} />
                  </div>
                  <span className={`text-sm font-semibold text-${color}-200`}>{title}</span>
                </div>
                <p className="text-sm text-white/45 leading-relaxed">{description}</p>
                <div className="mt-auto p-3 rounded-xl bg-white/[0.03] border border-white/[0.06]">
                  <p className="text-xs text-white/50 leading-relaxed italic">"{example}"</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── Use cases ─────────────────────────────────────────────────────── */}
      <section id="use-cases" className="py-28 px-6 relative overflow-hidden">
        <div className="absolute inset-0 pointer-events-none">
          <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[600px] h-[500px] bg-blue-600/[0.04] rounded-full blur-[110px]" />
        </div>

        <div className="relative max-w-6xl mx-auto">
          <div className="text-center mb-16">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-emerald-500/25 bg-emerald-500/[0.07] text-emerald-300 text-xs font-medium mb-5">
              <BarChart3 className="w-3 h-3" />
              Use Cases
            </div>
            <h2 className="text-4xl md:text-5xl font-bold tracking-[-0.02em] mb-5">
              Built for the teams that
              <br />
              <span className="text-white/35">can't afford to miss a move</span>
            </h2>
            <p className="text-white/45 max-w-xl mx-auto leading-relaxed">
              From sales call prep to quarterly strategy reviews, Pulzifi serves every team
              that relies on market awareness to make better decisions.
            </p>
          </div>

          <div className="grid md:grid-cols-2 gap-5">
            {USE_CASES.map(({ role, icon: Icon, headline, description, bullets, color }) => (
              <div
                key={role}
                className="rounded-2xl p-7 border border-white/[0.07] bg-white/[0.02] hover:bg-white/[0.035] transition-all duration-300"
              >
                <div className={`inline-flex items-center gap-2 text-xs font-semibold text-${color}-300 bg-${color}-500/[0.08] border border-${color}-500/15 rounded-full px-3 py-1 mb-5`}>
                  <Icon className="w-3 h-3" />
                  {role}
                </div>
                <h3 className="text-lg font-semibold mb-3 leading-snug">{headline}</h3>
                <p className="text-sm text-white/40 leading-relaxed mb-5">{description}</p>
                <ul className="space-y-2.5">
                  {bullets.map((bullet) => (
                    <li key={bullet} className="flex items-start gap-2.5 text-sm text-white/50">
                      <CheckCircle2 className={`w-4 h-4 text-${color}-400/70 mt-0.5 shrink-0`} />
                      <span>{bullet}</span>
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── Testimonials ─────────────────────────────────────────────────── */}
      <section className="py-28 px-6">
        <div className="max-w-6xl mx-auto">
          <div className="text-center mb-14">
            <h2 className="text-4xl md:text-5xl font-bold tracking-[-0.02em] mb-4">
              Loved by the teams
              <br />
              <span className="text-white/35">obsessed with winning</span>
            </h2>
            <p className="text-white/40 max-w-md mx-auto">
              Here's what competitive intelligence professionals are saying.
            </p>
          </div>

          <div className="grid md:grid-cols-3 gap-5">
            {TESTIMONIALS.map(({ quote, name, role, company, avatar, color }) => (
              <div
                key={name}
                className="rounded-2xl p-7 border border-white/[0.07] bg-white/[0.02] flex flex-col gap-5"
              >
                <div className="flex gap-0.5">
                  {[1,2,3,4,5].map((s) => (
                    <Star key={s} className="w-4 h-4 fill-amber-400 text-amber-400" />
                  ))}
                </div>
                <p className="text-sm text-white/55 leading-relaxed flex-1">"{quote}"</p>
                <div className="flex items-center gap-3 pt-4 border-t border-white/[0.06]">
                  <div className={`w-10 h-10 rounded-full bg-${color}-500/20 border border-${color}-500/25 flex items-center justify-center text-sm font-bold text-${color}-300`}>
                    {avatar}
                  </div>
                  <div>
                    <div className="text-sm font-semibold">{name}</div>
                    <div className="text-xs text-white/35">{role} · {company}</div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── Trust & Security ─────────────────────────────────────────────── */}
      <section className="py-20 px-6 border-y border-white/[0.05]">
        <div className="max-w-5xl mx-auto">
          <div className="grid md:grid-cols-4 gap-8">
            {[
              {
                icon: Shield,
                title: 'Multi-Tenant Isolation',
                description: 'Every organization gets a dedicated database schema. Your data is never shared, mixed, or accessible to other tenants.',
              },
              {
                icon: Lock,
                title: 'Encrypted at Rest & Transit',
                description: 'All data is encrypted in transit via TLS and encrypted at rest. Access tokens are short-lived and rotated automatically.',
              },
              {
                icon: RefreshCw,
                title: 'Role-Based Access Control',
                description: 'Granular permissions per workspace. Admins, team leads, and members each have appropriately scoped access.',
              },
              {
                icon: History,
                title: 'Complete Audit Trail',
                description: 'Every check, change, insight, and alert is timestamped and logged. Full traceability for compliance requirements.',
              },
            ].map(({ icon: Icon, title, description }) => (
              <div key={title} className="text-center">
                <div className="w-10 h-10 rounded-xl bg-white/[0.04] border border-white/[0.07] flex items-center justify-center mx-auto mb-4">
                  <Icon className="w-5 h-5 text-white/40" />
                </div>
                <h3 className="text-sm font-semibold mb-2">{title}</h3>
                <p className="text-xs text-white/35 leading-relaxed">{description}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── Integrations strip ───────────────────────────────────────────── */}
      <section className="py-20 px-6">
        <div className="max-w-5xl mx-auto text-center">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-white/10 bg-white/[0.03] text-white/40 text-xs font-medium mb-6">
            <Webhook className="w-3 h-3" />
            Integrations
          </div>
          <h2 className="text-3xl font-bold tracking-tight mb-4">
            Fits into the way your team works
          </h2>
          <p className="text-white/40 max-w-lg mx-auto mb-10 text-sm leading-relaxed">
            Pulzifi integrates with the tools your team already uses. Route intelligence
            directly into your workflows so nothing falls through the cracks.
          </p>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {[
              { name: 'Slack', description: 'Post change alerts and insights directly to channels', icon: '⚡' },
              { name: 'Microsoft Teams', description: 'Deliver notifications to your Teams workspace', icon: '💬' },
              { name: 'Discord', description: 'Alert your Discord server on every detected change', icon: '🎮' },
              { name: 'Custom Webhooks', description: 'Trigger any endpoint — CRM, Zapier, Make, or your own system', icon: '🔗' },
            ].map(({ name, description, icon }) => (
              <div key={name} className="rounded-xl p-5 border border-white/[0.07] bg-white/[0.02] text-left">
                <span className="text-2xl mb-3 block">{icon}</span>
                <div className="text-sm font-semibold mb-1">{name}</div>
                <div className="text-xs text-white/35 leading-relaxed">{description}</div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── FAQ ──────────────────────────────────────────────────────────── */}
      <section id="faq" className="py-28 px-6 relative overflow-hidden">
        <div className="absolute inset-0 pointer-events-none">
          <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-[600px] h-[400px] bg-violet-600/[0.04] rounded-full blur-[100px]" />
        </div>

        <div className="relative max-w-3xl mx-auto">
          <div className="text-center mb-14">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-white/10 bg-white/[0.03] text-white/40 text-xs font-medium mb-5">
              <BookOpen className="w-3 h-3" />
              Frequently Asked Questions
            </div>
            <h2 className="text-4xl font-bold tracking-tight mb-4">Got questions?</h2>
            <p className="text-white/40 text-sm">Everything you need to know before getting started.</p>
          </div>

          <div className="space-y-3">
            {FAQS.map(({ q, a }) => (
              <div key={q} className="rounded-2xl border border-white/[0.07] bg-white/[0.02] p-6">
                <h3 className="text-sm font-semibold mb-3 text-white">{q}</h3>
                <p className="text-sm text-white/45 leading-relaxed">{a}</p>
              </div>
            ))}
          </div>

          <div className="text-center mt-10">
            <p className="text-sm text-white/35">
              Still have questions?{' '}
              <Link href="/login" className="text-violet-400 hover:text-violet-300 transition-colors underline underline-offset-2">
                Sign in and explore the platform
              </Link>
              {' '}or{' '}
              <Link href="/register" className="text-violet-400 hover:text-violet-300 transition-colors underline underline-offset-2">
                start for free
              </Link>.
            </p>
          </div>
        </div>
      </section>

      {/* ── Final CTA ────────────────────────────────────────────────────── */}
      <section className="py-28 px-6">
        <div className="max-w-4xl mx-auto">
          <div className="relative rounded-3xl border border-white/[0.08] overflow-hidden">
            {/* Background */}
            <div className="absolute inset-0 bg-gradient-to-br from-violet-950/60 via-indigo-950/40 to-[#06060a]" />
            <div className="absolute inset-0 pointer-events-none">
              <div className="absolute top-0 right-0 w-[400px] h-[400px] bg-violet-600/10 rounded-full blur-[100px]" />
              <div className="absolute bottom-0 left-0 w-[300px] h-[300px] bg-indigo-600/8 rounded-full blur-[80px]" />
            </div>

            <div className="relative px-10 py-20 text-center">
              <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-violet-500 to-indigo-600 flex items-center justify-center mx-auto mb-8 shadow-[0_0_40px_rgba(124,58,237,0.4)]">
                <Radar className="w-8 h-8 text-white" />
              </div>
              <h2 className="text-4xl md:text-5xl font-bold tracking-[-0.025em] mb-5">
                Your competitors are moving.
                <br />
                <span className="text-white/40">Are you watching?</span>
              </h2>
              <p className="text-white/50 max-w-xl mx-auto mb-10 leading-relaxed">
                Join teams that have replaced manual competitive monitoring with Pulzifi.
                Get started free — no credit card, no setup complexity. First change detected in minutes.
              </p>

              <div className="flex flex-col sm:flex-row items-center justify-center gap-3 mb-8">
                <Link
                  href="/register"
                  className="group flex items-center gap-2.5 bg-white text-[#06060a] hover:bg-white/90 transition-all px-8 py-4 rounded-xl font-bold text-sm shadow-[0_0_40px_rgba(255,255,255,0.1)] hover:shadow-[0_0_60px_rgba(255,255,255,0.2)]"
                >
                  Get started free
                  <ArrowRight className="w-4 h-4 group-hover:translate-x-0.5 transition-transform" />
                </Link>
                <Link
                  href="/login"
                  className="text-white/55 hover:text-white transition-colors text-sm font-medium"
                >
                  Already have an account? Sign in →
                </Link>
              </div>

              <div className="flex flex-wrap items-center justify-center gap-6 text-xs text-white/25">
                <span className="flex items-center gap-1.5">
                  <CheckCircle2 className="w-3.5 h-3.5" />
                  Free to start, no card required
                </span>
                <span className="flex items-center gap-1.5">
                  <CheckCircle2 className="w-3.5 h-3.5" />
                  Checks begin within minutes
                </span>
                <span className="flex items-center gap-1.5">
                  <CheckCircle2 className="w-3.5 h-3.5" />
                  Cancel anytime
                </span>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ── Footer ───────────────────────────────────────────────────────── */}
      <footer className="border-t border-white/[0.05] px-6 pt-16 pb-10">
        <div className="max-w-6xl mx-auto">
          {/* Top row */}
          <div className="grid md:grid-cols-4 gap-12 mb-14">
            {/* Brand */}
            <div className="md:col-span-1">
              <div className="flex items-center gap-2.5 mb-5">
                <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-violet-500 to-indigo-600 flex items-center justify-center">
                  <Radar className="w-4.5 h-4.5 text-white" />
                </div>
                <span className="text-base font-semibold">Pulzifi</span>
              </div>
              <p className="text-sm text-white/30 leading-relaxed mb-5">
                AI-powered competitive intelligence platform for B2B teams that refuse
                to be caught off guard by market moves.
              </p>
              <div className="flex gap-3">
                <Link
                  href="/register"
                  className="text-xs bg-violet-600/80 hover:bg-violet-600 transition-colors px-3 py-1.5 rounded-lg text-white font-medium"
                >
                  Get started
                </Link>
              </div>
            </div>

            {/* Product */}
            <div>
              <div className="text-xs font-semibold text-white/50 uppercase tracking-widest mb-5">Product</div>
              <ul className="space-y-3">
                {[
                  { href: '#features', label: 'Features' },
                  { href: '#how-it-works', label: 'How it works' },
                  { href: '#insights', label: 'AI Insights' },
                  { href: '#use-cases', label: 'Use cases' },
                  { href: '/register', label: 'Get started free' },
                ].map(({ href, label }) => (
                  <li key={label}>
                    <a href={href} className="text-sm text-white/35 hover:text-white/70 transition-colors">
                      {label}
                    </a>
                  </li>
                ))}
              </ul>
            </div>

            {/* Capabilities */}
            <div>
              <div className="text-xs font-semibold text-white/50 uppercase tracking-widest mb-5">Capabilities</div>
              <ul className="space-y-3">
                {[
                  'Web monitoring',
                  'Change detection',
                  'AI insight generation',
                  'Screenshot archiving',
                  'Team workspaces',
                  'Alert & notifications',
                  'PDF reports',
                  'Webhook integrations',
                ].map((item) => (
                  <li key={item} className="text-sm text-white/35">{item}</li>
                ))}
              </ul>
            </div>

            {/* Company */}
            <div>
              <div className="text-xs font-semibold text-white/50 uppercase tracking-widest mb-5">Account</div>
              <ul className="space-y-3">
                {[
                  { href: '/login', label: 'Sign in' },
                  { href: '/register', label: 'Create account' },
                  { href: '#faq', label: 'FAQ' },
                ].map(({ href, label }) => (
                  <li key={label}>
                    <a href={href} className="text-sm text-white/35 hover:text-white/70 transition-colors">
                      {label}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          </div>

          {/* Bottom row */}
          <div className="pt-8 border-t border-white/[0.05] flex flex-col md:flex-row items-center justify-between gap-4">
            <p className="text-xs text-white/20">
              © {new Date().getFullYear()} Pulzifi. All rights reserved.
            </p>
            <p className="text-xs text-white/15">
              Competitive intelligence · Change detection · AI analysis
            </p>
          </div>
        </div>
      </footer>
    </div>
  )
}

import { cn } from '@workspace/ui/lib/utils'
import { Sparkles, TrendingDown, TrendingUp } from 'lucide-react'

type AlertTone = 'signal' | 'amber' | 'teal' | 'ink'

export type DashboardAlert = {
  tone?: AlertTone
  icon?: string
  site: string
  title: string
  detail?: string
  time?: string
}

export type KpiItem = {
  label: string
  value: string
  delta?: string
  deltaDirection?: 'up' | 'down'
}

type DashboardMockProps = {
  kpis?: KpiItem[]
  alerts?: DashboardAlert[]
  aiInsightTitle?: string
  aiInsightBody?: string
}

const toneStyles: Record<AlertTone, string> = {
  signal: 'border-[var(--pz-accent)]/20 bg-[var(--pz-accent)]/5',
  amber: 'border-[var(--pz-accent-gold)]/30 bg-[#fff7ed]',
  teal: 'border-[var(--pz-accent-teal)]/30 bg-[#ecfdf5]',
  ink: 'border-black/10 bg-[#f9f9fb]',
}

const SIDEBAR_ITEMS = ['Overview', 'Alerts', 'Monitors', 'AI Insights', 'Reports', 'Settings']

const DEFAULT_KPIS: KpiItem[] = [
  { label: 'Alerts (24h)', value: '128', delta: '+24%', deltaDirection: 'up' },
  { label: 'Mentions', value: '43', delta: '+8%', deltaDirection: 'up' },
  { label: 'Pages Changed', value: '17', delta: '-3%', deltaDirection: 'down' },
  { label: 'Price Changes', value: '5', delta: '+2', deltaDirection: 'up' },
]

function Sidebar() {
  return (
    <div className="flex w-28 shrink-0 flex-col gap-1 border-r border-gray-100 pr-2">
      <div className="mb-2 flex items-center gap-1.5">
        <div className="size-5 rounded bg-[var(--pz-accent)]" aria-hidden />
        <span className="text-[10px] font-bold tracking-wide text-[var(--pz-ink)]">Pulzifi</span>
      </div>
      {SIDEBAR_ITEMS.map((item) => (
        <div
          key={item}
          className={cn(
            'rounded-md px-2 py-1 text-[10px] text-gray-500',
            item === 'Overview' && 'bg-[var(--pz-accent-tint)] font-medium text-[var(--pz-accent)]',
          )}
          aria-hidden
        >
          {item}
        </div>
      ))}
    </div>
  )
}

function KpiTile({ label, value, delta, deltaDirection }: KpiItem) {
  const isUp = deltaDirection === 'up'
  return (
    <div className="rounded-lg border border-gray-100 bg-gray-50 p-2.5">
      <p className="text-[10px] text-gray-500">{label}</p>
      <p className="mt-0.5 text-base font-semibold text-[var(--pz-ink)]">{value}</p>
      {delta && (
        <div
          className={cn(
            'mt-0.5 flex items-center gap-0.5 text-[10px] font-medium',
            isUp ? 'text-emerald-600' : 'text-red-500',
          )}
        >
          {isUp ? (
            <TrendingUp className="size-3" aria-hidden />
          ) : (
            <TrendingDown className="size-3" aria-hidden />
          )}
          {delta}
        </div>
      )}
    </div>
  )
}

function PriceChartDecoration() {
  // Schematic SVG — 3 multi-color lines, purely decorative
  return (
    <div className="rounded-lg border border-gray-100 bg-white p-2.5">
      <p className="mb-1.5 text-[10px] font-medium text-gray-500">Price Positioning</p>
      <svg
        viewBox="0 0 120 40"
        aria-hidden
        className="h-8 w-full"
        preserveAspectRatio="none"
      >
        <polyline
          points="0,30 20,25 40,20 60,22 80,15 100,18 120,10"
          fill="none"
          stroke="var(--pz-accent)"
          strokeWidth="1.5"
        />
        <polyline
          points="0,35 20,32 40,30 60,28 80,26 100,28 120,25"
          fill="none"
          stroke="var(--pz-accent-teal)"
          strokeWidth="1.5"
        />
        <polyline
          points="0,38 20,36 40,35 60,36 80,33 100,35 120,32"
          fill="none"
          stroke="var(--pz-accent-gold)"
          strokeWidth="1.5"
        />
      </svg>
    </div>
  )
}

// Deterministic pseudo-random sequence — must be identical on server and
// client so SSR markup matches hydration. Pattern hand-tuned to look organic.
const HEATMAP_OPACITIES = [
  0.2, 0.4, 0.1, 0.6, 0.8, 0.4, 0.2,
  0.4, 0.6, 0.8, 1.0, 0.6, 0.4, 0.2,
  0.1, 0.2, 0.4, 0.8, 1.0, 0.8, 0.4,
  0.2, 0.1, 0.2, 0.4, 0.6, 0.4, 0.2,
]

function HeatmapDecoration() {
  return (
    <div className="rounded-lg border border-gray-100 bg-white p-2.5">
      <p className="mb-1.5 text-[10px] font-medium text-gray-500">Website Changes</p>
      <div className="grid grid-cols-7 gap-0.5" aria-hidden>
        {HEATMAP_OPACITIES.map((opacity, i) => (
          <div
            key={i}
            className="aspect-square rounded-[2px] bg-[var(--pz-accent)]"
            style={{ opacity }}
          />
        ))}
      </div>
    </div>
  )
}

export function DashboardMock({
  kpis = DEFAULT_KPIS,
  alerts = [],
  aiInsightTitle,
  aiInsightBody,
}: DashboardMockProps) {
  return (
    <div
      data-pz-hero-mock
      className="rounded-2xl border border-gray-200 bg-white p-5 shadow-[0_24px_60px_-12px_rgba(0,0,0,0.08)]"
    >
      <div className="flex gap-3">
        <Sidebar />

        {/* Main content */}
        <div className="min-w-0 flex-1 space-y-3">
          {/* Live indicator */}
          <div className="flex items-center justify-between border-b border-gray-100 pb-2">
            <div className="flex items-center gap-2">
              <span className="relative flex size-2.5 items-center justify-center">
                <span className="absolute size-2.5 animate-ping rounded-full bg-[var(--pz-accent)]/60" />
                <span className="size-2 rounded-full bg-[var(--pz-accent)]" />
              </span>
              <span className="text-xs font-medium text-[var(--pz-ink)]">Live Dashboard</span>
            </div>
            <span className="text-[10px] uppercase tracking-wider text-gray-400">Realtime</span>
          </div>

          {/* KPI row */}
          <div className="grid grid-cols-2 gap-2 lg:grid-cols-4">
            {kpis.slice(0, 4).map((kpi) => (
              <KpiTile key={kpi.label} {...kpi} />
            ))}
          </div>

          {/* Charts row */}
          <div className="grid grid-cols-2 gap-2">
            <PriceChartDecoration />
            <HeatmapDecoration />
          </div>

          {/* Activity feed */}
          {alerts.length > 0 && (
            <div>
              <p className="mb-1.5 text-[10px] font-medium text-gray-500">Recent Alerts</p>
              <ul className="space-y-1.5">
                {alerts.slice(0, 3).map((alert, i) => (
                  <li
                    key={`${alert.site}-${i}`}
                    className={cn(
                      'flex items-start gap-2 rounded-lg border px-2.5 py-1.5',
                      toneStyles[alert.tone ?? 'ink'],
                    )}
                  >
                    {alert.icon && (
                      <span className="text-xs leading-none" aria-hidden>
                        {alert.icon}
                      </span>
                    )}
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-[10px] font-medium text-[var(--pz-ink)]">
                        {alert.title}
                      </p>
                    </div>
                    {alert.time && (
                      <span className="shrink-0 text-[10px] text-gray-400">{alert.time}</span>
                    )}
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* AI Insight callout */}
          {(aiInsightTitle || aiInsightBody) && (
            <div className="rounded-lg border border-[var(--pz-accent)]/15 bg-gradient-to-br from-[#f5f0ff] to-white p-2.5">
              <div className="flex items-center gap-1.5">
                <Sparkles className="size-3 text-[var(--pz-accent)]" aria-hidden />
                <span className="text-[10px] font-semibold uppercase tracking-wider text-[var(--pz-dark-surface)]">
                  {aiInsightTitle ?? 'AI Insight'}
                </span>
              </div>
              {aiInsightBody && (
                <p className="mt-1 text-[11px] leading-relaxed text-[var(--pz-ink-2)]">
                  {aiInsightBody}
                </p>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

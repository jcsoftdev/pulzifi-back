import { AlertCircle, Sparkles } from 'lucide-react'

type MockType = 'url' | 'insight' | 'alerts'

type StepMockProps = {
  mockType: MockType
  mockText?: string
}

function UrlInputMock({ text }: { text?: string }) {
  return (
    <div className="h-[220px] w-full rounded-xl border border-gray-200 bg-gray-50 p-4">
      {/* Browser chrome */}
      <div className="flex items-center gap-1.5 pb-3">
        <span className="size-2.5 rounded-full bg-red-300" />
        <span className="size-2.5 rounded-full bg-yellow-300" />
        <span className="size-2.5 rounded-full bg-green-300" />
      </div>
      {/* URL bar */}
      <div className="flex items-center gap-2 rounded-lg border border-gray-200 bg-white px-3 py-2">
        <div className="size-3 rounded-full bg-gray-300" aria-hidden />
        <span className="flex-1 truncate text-xs text-gray-500">
          {text ?? 'competitor.com/pricing'}
        </span>
        <div className="h-3 w-6 rounded bg-[var(--pz-accent)] opacity-60" aria-hidden />
      </div>
      {/* Simulated page skeleton */}
      <div className="mt-3 space-y-2">
        <div className="h-2.5 w-3/4 rounded bg-gray-200" />
        <div className="h-2.5 w-1/2 rounded bg-gray-200" />
        <div className="mt-3 grid grid-cols-3 gap-2">
          {[1, 2, 3].map((n) => (
            <div key={n} className="h-10 rounded-lg border border-gray-200 bg-white p-1.5">
              <div className="h-1.5 w-2/3 rounded bg-gray-200" />
              <div className="mt-1.5 h-2.5 w-full rounded bg-[var(--pz-accent)] opacity-30" />
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

function InsightCardMock({ text: _text }: { text?: string }) {
  return (
    <div className="h-[220px] w-full rounded-xl border border-gray-200 bg-gray-50 p-4">
      {/* Header */}
      <div className="flex items-center gap-2 pb-2">
        <Sparkles className="size-4 text-[var(--pz-accent)]" aria-hidden />
        <span className="text-xs font-semibold uppercase tracking-wider text-[var(--pz-accent)]">
          AI Insight
        </span>
        <span className="ml-auto rounded-full bg-[var(--pz-accent)] px-2 py-0.5 text-[10px] font-bold text-white">
          NEW
        </span>
      </div>
      {/* Skeleton lines */}
      <div className="mt-2 space-y-2">
        <div className="h-2.5 w-full rounded bg-gray-200" />
        <div className="h-2.5 w-5/6 rounded bg-gray-200" />
        <div className="h-2.5 w-4/5 rounded bg-gray-200" />
      </div>
      {/* Mini card */}
      <div className="mt-4 rounded-lg border border-[var(--pz-accent)]/20 bg-white p-2.5">
        <div className="flex items-start gap-2">
          <div className="mt-0.5 size-4 shrink-0 rounded bg-[var(--pz-accent-tint)]" aria-hidden />
          <div className="flex-1 space-y-1.5">
            <div className="h-2 w-3/4 rounded bg-gray-200" />
            <div className="h-2 w-full rounded bg-gray-200" />
          </div>
        </div>
      </div>
      {/* Action area */}
      <div className="mt-3 flex gap-2">
        <div className="h-6 flex-1 rounded-md bg-[var(--pz-accent)] opacity-50" />
        <div className="h-6 flex-1 rounded-md bg-gray-200" />
      </div>
    </div>
  )
}

function AlertListMock({ text: _text }: { text?: string }) {
  const rows = [
    { icon: '🔴', color: 'border-red-200 bg-red-50' },
    { icon: '⚡', color: 'border-yellow-200 bg-yellow-50' },
    { icon: '📣', color: 'border-teal-200 bg-teal-50' },
  ]
  return (
    <div className="h-[220px] w-full rounded-xl border border-gray-200 bg-gray-50 p-4">
      {/* Header */}
      <div className="flex items-center gap-2 pb-3">
        <AlertCircle className="size-4 text-gray-400" aria-hidden />
        <span className="text-xs font-semibold uppercase tracking-wider text-gray-500">
          Alerts
        </span>
        <span className="ml-auto rounded-full bg-gray-200 px-2 py-0.5 text-[10px] text-gray-500">
          3 new
        </span>
      </div>
      {/* Alert rows */}
      <div className="space-y-2">
        {rows.map((row, i) => (
          <div
            key={i}
            className={`flex items-start gap-2 rounded-lg border px-2.5 py-2 ${row.color}`}
          >
            <span className="text-sm leading-none" aria-hidden>
              {row.icon}
            </span>
            <div className="flex-1 space-y-1">
              <div className="h-2 w-3/4 rounded bg-gray-300 opacity-60" />
              <div className="h-2 w-1/2 rounded bg-gray-300 opacity-40" />
            </div>
            <div className="shrink-0 text-[10px] text-gray-400">now</div>
          </div>
        ))}
      </div>
    </div>
  )
}

export function StepMock({ mockType, mockText }: StepMockProps) {
  switch (mockType) {
    case 'url':
      return <UrlInputMock text={mockText} />
    case 'insight':
      return <InsightCardMock text={mockText} />
    case 'alerts':
      return <AlertListMock text={mockText} />
    default:
      return <UrlInputMock text={mockText} />
  }
}

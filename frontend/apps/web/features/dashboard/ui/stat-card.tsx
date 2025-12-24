'use client'

export interface StatCardProps {
  value: string
  label: string
  max?: string
  tag?: string
  tagColor?: string
}

export function StatCard({ value, label, max, tag, tagColor = 'bg-blue-100 text-blue-700' }: Readonly<StatCardProps>) {
  return (
    <div className="bg-card rounded-lg border border-border p-3 shadow-[0px_2px_4px_0px_rgba(0,0,0,0.02),0px_1px_6px_-1px_rgba(0,0,0,0.02),0px_1px_2px_0px_rgba(0,0,0,0.03)] relative">
      <div className="flex items-baseline gap-0.5 mb-0.5">
        <span className="text-[25px] font-extrabold text-foreground leading-tight">{value}</span>
        {max && <span className="text-[21px] font-semibold text-muted-foreground leading-tight">/{max}</span>}
      </div>
      <div className="flex items-center justify-between">
        <p className="text-[14.6px] text-muted-foreground/50">{label}</p>
      </div>
      {tag && (
        <div className="absolute top-3.5 right-4">
          <span className={`text-[12.5px] px-2 py-0.5 rounded font-normal ${tagColor}`}>
            {tag}
          </span>
        </div>
      )}
    </div>
  )
}

'use client'

import type * as React from 'react'

export interface StatCardProps {
  value: string
  label: string
  max?: string
  tag?: string
  tagColor?: string
  icon?: React.ReactNode
}

export function StatCard({
  value,
  label,
  max,
  tag,
  tagColor = 'bg-blue-100 text-blue-700',
  icon,
}: Readonly<StatCardProps>) {
  return (
    <div className="bg-card rounded-lg border border-border p-3 shadow-sm relative">
      <div className="flex items-baseline gap-0.5 mb-0.5">
        <span className="text-2xl font-extrabold text-foreground leading-tight">{value}</span>
        {max && (
          <span className="text-xl font-semibold text-muted-foreground leading-tight">/{max}</span>
        )}
      </div>
      <div className="flex items-center justify-between">
        <p className="text-sm text-muted-foreground/50">{label}</p>
      </div>
      {tag && (
        <div className="absolute top-3.5 right-4">
          <span className={`text-xs px-2 py-0.5 rounded font-normal ${tagColor}`}>{tag}</span>
        </div>
      )}
      {icon && !tag && (
        <div className="absolute top-3.5 right-4 text-muted-foreground/50">{icon}</div>
      )}
    </div>
  )
}

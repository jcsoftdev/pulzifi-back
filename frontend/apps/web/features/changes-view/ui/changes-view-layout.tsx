'use client'

import { useState } from 'react'
import { useRouter, usePathname, useSearchParams } from 'next/navigation'
import { formatDateTime } from '@workspace/ui'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@workspace/ui/components/atoms/select'
import { cn } from '@workspace/ui/lib/utils'
import type { Check } from '@workspace/services/page-api'

interface ChangesViewLayoutProps {
  checks: Check[]
  activeCheckId: string
  children: React.ReactNode
  activeTab?: string // 'visual' | 'text' | 'insights'
  onTabChange?: (tab: string) => void
}

export function ChangesViewLayout({
  checks,
  activeCheckId,
  children,
  activeTab: controlledActiveTab,
  onTabChange,
}: Readonly<ChangesViewLayoutProps>) {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()
  
  // Local state for tabs if not controlled
  const [internalActiveTab, setInternalActiveTab] = useState('visual')
  const activeTab = controlledActiveTab || internalActiveTab
  const handleTabChange = onTabChange || setInternalActiveTab

  const activeCheck = checks.find((c) => c.id === activeCheckId) || checks[0]

  const handleCheckChange = (checkId: string) => {
    const params = new URLSearchParams(searchParams.toString())
    params.set('checkId', checkId)
    router.push(`${pathname}?${params.toString()}`)
  }

  return (
    <div className="flex flex-col gap-8">
      {/* Header */}
      <div className="flex items-end justify-between">
        <div className="flex flex-col gap-1">
          <span className="text-sm text-muted-foreground">Change detected on:</span>
          <h1 className="text-3xl font-bold text-foreground">
             {activeCheck ? formatDateTime(activeCheck.checkedAt) : 'No check selected'}
          </h1>
        </div>

        <div className="w-64">
           <Select value={activeCheckId} onValueChange={handleCheckChange}>
            <SelectTrigger className="w-full bg-background">
              <SelectValue placeholder="Select a check" />
            </SelectTrigger>
            <SelectContent>
              {checks.map((check) => (
                <SelectItem key={check.id} value={check.id}>
                  {formatDateTime(check.checkedAt)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-border">
        <div className="flex gap-8">
          <button
            onClick={() => handleTabChange('visual')}
            className={cn(
              "pb-3 text-sm font-medium border-b-2 transition-colors",
              activeTab === 'visual' 
                ? "border-primary text-foreground" 
                : "border-transparent text-muted-foreground hover:text-foreground"
            )}
          >
            Visual Pulse
          </button>
          <button
             onClick={() => handleTabChange('text')}
             className={cn(
              "pb-3 text-sm font-medium border-b-2 transition-colors",
              activeTab === 'text' 
                ? "border-primary text-foreground" 
                : "border-transparent text-muted-foreground hover:text-foreground"
            )}
          >
            Text Changes
          </button>
          <button
             onClick={() => handleTabChange('insights')}
             className={cn(
              "pb-3 text-sm font-medium border-b-2 transition-colors",
              activeTab === 'insights' 
                ? "border-primary text-foreground" 
                : "border-transparent text-muted-foreground hover:text-foreground"
            )}
          >
            Intelligent Insights
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="min-h-[500px]">
        {children}
      </div>
    </div>
  )
}

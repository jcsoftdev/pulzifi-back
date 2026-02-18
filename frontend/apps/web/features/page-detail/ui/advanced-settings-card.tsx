'use client'

import { type MonitoringConfig, PageApi } from '@workspace/services/page-api'
import { useRouter } from 'next/navigation'
import { useTransition } from 'react'

interface AdvancedSettingsCardProps {
  pageId: string
  initialConfig: MonitoringConfig | null
}

export function AdvancedSettingsCard({
  pageId,
  initialConfig,
}: Readonly<AdvancedSettingsCardProps>) {
  const router = useRouter()
  const [isPending, startTransition] = useTransition()

  // Defaults if no config
  const scheduleType = initialConfig?.scheduleType || 'all_time'
  const checkFrequency = initialConfig?.checkFrequency || '24h'
  const timezone = initialConfig?.timezone || 'America/Boise'

  const handleUpdate = (updates: Partial<MonitoringConfig>) => {
    startTransition(async () => {
      try {
        await PageApi.updateMonitoringConfig(pageId, updates)
        router.refresh()
      } catch (error) {
        console.error('Failed to update config', error)
      }
    })
  }

  return (
    <div className="flex flex-col gap-6 bg-card border border-border rounded-xl p-6 h-full">
      <h3 className="text-xl font-semibold text-foreground">Advanced settings</h3>

      <div className="flex flex-col gap-4">
        {/* Schedule Type Selection */}
        <div className="flex items-center gap-3 flex-wrap">
          <button
            type="button"
            onClick={() =>
              handleUpdate({
                scheduleType: 'work_days',
              })
            }
            className="flex items-center gap-2 cursor-pointer disabled:opacity-50"
            disabled={isPending}
          >
            <div
              className={`w-4 h-4 rounded-full border ${scheduleType === 'work_days' ? 'border-destructive bg-destructive' : 'border-border bg-muted'}`}
            />
            <span
              className={`text-sm ${scheduleType === 'work_days' ? 'text-foreground font-medium' : 'text-muted-foreground'}`}
            >
              Work days
            </span>
          </button>

          <button
            type="button"
            onClick={() =>
              handleUpdate({
                scheduleType: 'work_hours',
              })
            }
            className="flex items-center gap-2 cursor-pointer disabled:opacity-50"
            disabled={isPending}
          >
            <div
              className={`w-4 h-4 rounded-full border ${scheduleType === 'work_hours' ? 'border-destructive bg-destructive' : 'border-border bg-muted'}`}
            />
            <span
              className={`text-sm ${scheduleType === 'work_hours' ? 'text-foreground font-medium' : 'text-muted-foreground'}`}
            >
              Work days, during work hours
            </span>
          </button>

          <button
            type="button"
            onClick={() =>
              handleUpdate({
                scheduleType: 'all_time',
              })
            }
            className="flex items-center gap-2 cursor-pointer disabled:opacity-50"
            disabled={isPending}
          >
            <div
              className={`w-4 h-4 rounded-full border ${scheduleType === 'all_time' ? 'border-destructive bg-destructive' : 'border-border bg-muted'}`}
            />
            <span
              className={`text-sm ${scheduleType === 'all_time' ? 'text-foreground font-medium' : 'text-muted-foreground'}`}
            >
              Weekdays
            </span>
          </button>
        </div>

        <p className="text-xs text-muted-foreground">
          This selections will run only in your time zone: {timezone}
        </p>

        {/* Frequency Display/Edit */}
        <div className="flex items-center gap-2">
          <div className="px-3 py-1.5 rounded-md border border-border bg-muted/50">
            <span className="text-xs text-foreground">
              {checkFrequency === '24h'
                ? '1 Check a day'
                : checkFrequency === '1h'
                  ? 'Every hour'
                  : checkFrequency === '30m'
                    ? 'Every 30m'
                    : checkFrequency}
            </span>
          </div>
          {/* Mocking calculation for weekly checks */}
          <div className="px-3 py-1.5 rounded-md border border-border bg-muted/50">
            <span className="text-xs text-foreground">
              {checkFrequency === '24h'
                ? '7 checks a week'
                : checkFrequency === '1h'
                  ? '168 checks a week'
                  : 'Multiple checks a week'}
            </span>
          </div>
        </div>

        <p className="text-xs text-muted-foreground">
          {scheduleType === 'all_time'
            ? 'This page will be checked all week all day'
            : scheduleType === 'work_days'
              ? 'This page will be checked only on work days'
              : 'This page will be checked only during work hours'}
        </p>
      </div>
    </div>
  )
}

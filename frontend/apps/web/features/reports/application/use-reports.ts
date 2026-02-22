'use client'

import type { CreateReportDto, Report } from '@workspace/services'
import { ReportApi } from '@workspace/services'
import { useCallback, useEffect, useMemo, useState } from 'react'

export function useReports(pageIds?: string[]) {
  const [reports, setReports] = useState<Report[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  // Stabilize pageIds reference to avoid infinite re-fetching
  const stablePageIds = useMemo(() => pageIds, [JSON.stringify(pageIds)])

  const fetchReports = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const result = await ReportApi.listReports()
      const filtered = stablePageIds
        ? result.data.filter((r) => stablePageIds.includes(r.pageId))
        : result.data
      setReports(filtered)
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to load reports'))
    } finally {
      setLoading(false)
    }
  }, [stablePageIds])

  useEffect(() => {
    fetchReports()
  }, [fetchReports])

  const createReport = useCallback(async (data: CreateReportDto) => {
    const report = await ReportApi.createReport(data)
    setReports((prev) => [report, ...prev])
    return report
  }, [])

  return {
    reports,
    loading,
    error,
    createReport,
    refresh: fetchReports,
  }
}

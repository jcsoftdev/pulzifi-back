import { describe, expect, test } from 'bun:test'
import { transformDashboardStats } from './dashboard-api'

type StatsBackend = Parameters<typeof transformDashboardStats>[0]

describe('transformDashboardStats', () => {
  test('maps counts and nested collections to camelCase', () => {
    const backend: StatsBackend = {
      workspaces_count: 3,
      pages_count: 12,
      today_checks_count: 5,
      changes_per_workspace: [{ workspace_name: 'Marketing', detected_changes: 2 }],
      recent_alerts: [
        {
          checked_at: '2026-01-01T00:00:00Z',
          workspace_name: 'Marketing',
          change_type: 'content_change',
          page_url: 'https://x.com',
        },
      ],
      recent_insights: [
        {
          created_at: '2026-01-02T00:00:00Z',
          workspace_name: 'Marketing',
          page_url: 'https://x.com',
          title: 'Price up',
          content: 'Detail',
        },
      ],
    }

    expect(transformDashboardStats(backend)).toEqual({
      workspacesCount: 3,
      pagesCount: 12,
      todayChecksCount: 5,
      changesPerWorkspace: [{ workspaceName: 'Marketing', detectedChanges: 2 }],
      recentAlerts: [
        {
          checkedAt: '2026-01-01T00:00:00Z',
          workspaceName: 'Marketing',
          changeType: 'content_change',
          pageUrl: 'https://x.com',
        },
      ],
      recentInsights: [
        {
          createdAt: '2026-01-02T00:00:00Z',
          workspaceName: 'Marketing',
          pageUrl: 'https://x.com',
          title: 'Price up',
          content: 'Detail',
        },
      ],
    })
  })

  test('null collections default to empty arrays', () => {
    const backend = {
      workspaces_count: 0,
      pages_count: 0,
      today_checks_count: 0,
      changes_per_workspace: null,
      recent_alerts: null,
      recent_insights: null,
    } as unknown as StatsBackend

    const out = transformDashboardStats(backend)
    expect(out.changesPerWorkspace).toEqual([])
    expect(out.recentAlerts).toEqual([])
    expect(out.recentInsights).toEqual([])
  })
})

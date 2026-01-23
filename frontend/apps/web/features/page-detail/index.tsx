import { WorkspaceBreadcrumbs } from '@/app/(main)/workspaces/[id]/workspace-breadcrumbs'
import { PageDetailService } from './domain/services/page-detail-service'
import { PageInfoWithActions } from './ui/page-info-with-actions'
import { ChecksHistory } from './ui/checks-history'
import { AdvancedSettingsCard } from './ui/advanced-settings-card'
import { GeneralSummaryCard } from './ui/general-summary-card'
import { IntelligentInsightsCard } from './ui/intelligent-insights-card'

export interface PageDetailFeatureProps {
  workspaceId: string
  pageId: string
}

export async function PageDetailFeature({ workspaceId, pageId }: Readonly<PageDetailFeatureProps>) {
  const [page, workspace, config, checks, insights] = await Promise.all([
    PageDetailService.getPage(pageId),
    PageDetailService.getWorkspace(workspaceId),
    PageDetailService.getMonitoringConfig(pageId),
    PageDetailService.listChecks(pageId),
    PageDetailService.listInsights(pageId),
  ])

  return (
    <div className="flex-1 flex flex-col bg-background overflow-auto">
      <WorkspaceBreadcrumbs
        breadcrumbs={[
          {
            label: 'Workspaces',
            href: '/workspaces',
          },
          {
            label: workspace.name,
            href: `/workspaces/${workspaceId}`,
          },
          {
            label: page.name,
            href: `/workspaces/${workspaceId}/pages/${pageId}`,
            isCurrent: true,
          },
        ]}
      />

      <div className="px-8 py-8 space-y-8 max-w-7xl mx-auto w-full">
        <PageInfoWithActions initialPage={page} workspaceId={workspaceId} />

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          <div className="flex flex-col gap-8 lg:col-span-2">
            <ChecksHistory checks={checks} workspaceId={workspaceId} pageId={pageId} />

            <AdvancedSettingsCard pageId={pageId} initialConfig={config} />
          </div>

          <div className="flex flex-col gap-8 lg:col-span-1">
            <GeneralSummaryCard page={page} config={config} />

            <IntelligentInsightsCard insights={insights} />
          </div>
        </div>
      </div>
    </div>
  )
}

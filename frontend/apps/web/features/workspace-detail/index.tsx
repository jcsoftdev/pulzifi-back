import type { Page } from '@/features/page/domain/types'
import type { Workspace } from '@/features/workspace/domain/types'
import { WorkspaceDetailContent } from './ui/workspace-detail-content'

export interface WorkspaceDetailFeatureProps {
  workspace: Workspace
  initialPages?: Page[]
}

export function WorkspaceDetailFeature({
  workspace,
  initialPages = [],
}: Readonly<WorkspaceDetailFeatureProps>) {
  return <WorkspaceDetailContent workspace={workspace} initialPages={initialPages} />
}

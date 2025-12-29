import { WorkspaceFeature } from '@/features/workspace'
import { getWorkspacesServer } from '@/features/workspace/application/services/server'

export default async function WorkspacesPage() {
  const workspaces = await getWorkspacesServer()

  return <WorkspaceFeature initialWorkspaces={workspaces} lastCheckTime="just now" />
}

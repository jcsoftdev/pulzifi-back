import { WorkspaceApi } from '@workspace/services'
import type { Workspace } from '../../domain/types'

export async function getWorkspacesServer(): Promise<Workspace[]> {
  const response = await WorkspaceApi.listWorkspaces()
  return response.workspaces
}

export async function getWorkspaceServer(id: string): Promise<Workspace> {
  return await WorkspaceApi.getWorkspace(id)
}

export async function createWorkspaceServer(data: {
  name: string
  type: 'Personal' | 'Team' | 'Competitor'
}): Promise<Workspace> {
  return await WorkspaceApi.createWorkspace(data)
}

export async function updateWorkspaceServer(
  id: string,
  data: Partial<{
    name: string
    type: 'Personal' | 'Team' | 'Competitor'
    tags: string[]
  }>
): Promise<Workspace> {
  return await WorkspaceApi.updateWorkspace(id, data)
}

export async function deleteWorkspaceServer(id: string): Promise<void> {
  await WorkspaceApi.deleteWorkspace(id)
}

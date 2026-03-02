import { WorkspaceApi } from '@workspace/services'
import { handleServerAuthError } from '@/lib/auth/server-auth'
import type { Workspace } from '../../domain/types'

export async function getWorkspacesServer(): Promise<Workspace[]> {
  try {
    const response = await WorkspaceApi.listWorkspaces()
    return response.workspaces
  } catch (error) {
    return handleServerAuthError(error)
  }
}

export async function getWorkspaceServer(id: string): Promise<Workspace> {
  try {
    return await WorkspaceApi.getWorkspace(id)
  } catch (error) {
    return handleServerAuthError(error)
  }
}

export async function createWorkspaceServer(data: {
  name: string
  type: 'Personal' | 'Team' | 'Competitor'
}): Promise<Workspace> {
  try {
    return await WorkspaceApi.createWorkspace(data)
  } catch (error) {
    return handleServerAuthError(error)
  }
}

export async function updateWorkspaceServer(
  id: string,
  data: Partial<{
    name: string
    type: 'Personal' | 'Team' | 'Competitor'
    tags: string[]
  }>
): Promise<Workspace> {
  try {
    return await WorkspaceApi.updateWorkspace(id, data)
  } catch (error) {
    return handleServerAuthError(error)
  }
}

export async function deleteWorkspaceServer(id: string): Promise<void> {
  try {
    await WorkspaceApi.deleteWorkspace(id)
  } catch (error) {
    return handleServerAuthError(error)
  }
}

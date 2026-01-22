import { getHttpClient } from '@workspace/shared-http'

// Internal: Backend response types (snake_case from Go)
interface WorkspaceBackendDto {
  id: string
  name: string
  type: string
  tags: string[]
  created_by: string
  created_at: string
}

interface ListWorkspacesBackendResponse {
  workspaces: WorkspaceBackendDto[]
}

// Exported: Frontend types (camelCase)
export interface Workspace {
  id: string
  name: string
  type: string
  tags: string[]
  createdBy: string
  createdAt: string
}

export interface ListWorkspacesResponse {
  workspaces: Workspace[]
}

export interface CreateWorkspaceDto {
  name: string
  type: 'Personal' | 'Team' | 'Competitor'
  tags?: string[]
}

export interface ListWorkspacesParams {
  limit?: number
}

// Helper: Transform backend to frontend format
function transformWorkspace(backend: WorkspaceBackendDto): Workspace {
  return {
    id: backend.id,
    name: backend.name,
    type: backend.type,
    tags: backend.tags || [],
    createdBy: backend.created_by,
    createdAt: backend.created_at,
  }
}

export const WorkspaceApi = {
  async listWorkspaces(params?: ListWorkspacesParams): Promise<ListWorkspacesResponse> {
    const http = await getHttpClient()
    const queryParams = params?.limit ? `?limit=${params.limit}` : ''
    const response = await http.get<ListWorkspacesBackendResponse>(
      `/api/v1/workspaces${queryParams}`
    )
    return {
      workspaces: response.workspaces.map(transformWorkspace),
    }
  },

  async getWorkspace(id: string): Promise<Workspace> {
    const http = await getHttpClient()
    const response = await http.get<WorkspaceBackendDto>(`/api/v1/workspaces/${id}`)
    return transformWorkspace(response)
  },

  async createWorkspace(data: CreateWorkspaceDto): Promise<Workspace> {
    const http = await getHttpClient()
    const response = await http.post<WorkspaceBackendDto>('/api/v1/workspaces', data)
    return transformWorkspace(response)
  },

  async updateWorkspace(id: string, data: Partial<CreateWorkspaceDto>): Promise<Workspace> {
    const http = await getHttpClient()
    const response = await http.put<WorkspaceBackendDto>(`/api/v1/workspaces/${id}`, data)
    return transformWorkspace(response)
  },

  async deleteWorkspace(id: string): Promise<void> {
    const http = await getHttpClient()
    await http.delete(`/api/v1/workspaces/${id}`)
  },
}

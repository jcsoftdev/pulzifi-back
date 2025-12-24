
import { getHttpClient } from '@workspace/shared-http'

export interface Workspace {
  id: string
  name: string
  type: 'Personal' | 'Team' | 'Competitor'
  createdAt: string
  updatedAt: string
}

export interface CreateWorkspaceDto {
  name: string
  type: 'Personal' | 'Team' | 'Competitor'
}

export class WorkspaceApi {
  static async getWorkspaces(): Promise<Workspace[]> {
    const http = await getHttpClient()
    return http.get<Workspace[]>('/api/workspaces')
  }

  static async getWorkspace(id: string): Promise<Workspace> {
    const http = await getHttpClient()
    return http.get<Workspace>(`/api/workspaces/${id}`)
  }

  static async createWorkspace(data: CreateWorkspaceDto): Promise<Workspace> {
    const http = await getHttpClient()
    return http.post<Workspace>('/api/workspaces', data)
  }

  static async updateWorkspace(id: string, data: Partial<CreateWorkspaceDto>): Promise<Workspace> {
    const http = await getHttpClient()
    return http.put<Workspace>(`/api/workspaces/${id}`, data)
  }

  static async deleteWorkspace(id: string): Promise<void> {
    const http = await getHttpClient()
    await http.delete(`/api/workspaces/${id}`)
  }
}

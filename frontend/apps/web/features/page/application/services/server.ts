import { PageApi } from '@workspace/services'
import { handleServerAuthError } from '@/lib/auth/server-auth'
import type { CreatePageDto, Page } from '../../domain/types'

export async function getPagesServer(workspaceId: string): Promise<Page[]> {
  try {
    return await PageApi.listPages({
      workspaceId,
    })
  } catch (error) {
    return handleServerAuthError(error)
  }
}

export async function getPageServer(id: string): Promise<Page> {
  try {
    return await PageApi.getPage(id)
  } catch (error) {
    return handleServerAuthError(error)
  }
}

export async function createPageServer(data: CreatePageDto): Promise<Page> {
  try {
    return await PageApi.createPage(data)
  } catch (error) {
    return handleServerAuthError(error)
  }
}

export async function updatePageServer(id: string, data: Partial<CreatePageDto>): Promise<Page> {
  try {
    return await PageApi.updatePage(id, data)
  } catch (error) {
    return handleServerAuthError(error)
  }
}

export async function deletePageServer(id: string): Promise<void> {
  try {
    await PageApi.deletePage(id)
  } catch (error) {
    return handleServerAuthError(error)
  }
}

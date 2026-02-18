import { PageApi } from '@workspace/services'
import type { CreatePageDto, Page } from '../../domain/types'

export async function getPagesServer(workspaceId: string): Promise<Page[]> {
  try {
    return await PageApi.listPages({
      workspaceId,
    })
  } catch (error) {
    const err = error instanceof Error ? error : new Error(String(error))
    console.error('[PageService] Failed to list pages', {
      workspaceId,
      error: err.message,
    })
    throw error
  }
}

export async function getPageServer(id: string): Promise<Page> {
  return await PageApi.getPage(id)
}

export async function createPageServer(data: CreatePageDto): Promise<Page> {
  return await PageApi.createPage(data)
}

export async function updatePageServer(id: string, data: Partial<CreatePageDto>): Promise<Page> {
  return await PageApi.updatePage(id, data)
}

export async function deletePageServer(id: string): Promise<void> {
  await PageApi.deletePage(id)
}

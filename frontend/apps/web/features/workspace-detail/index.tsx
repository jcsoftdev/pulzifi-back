'use client'

import { useState } from 'react'
import { SquarePlus } from 'lucide-react'
import { Button } from '@workspace/ui/components/atoms/button'
import { Input } from '@workspace/ui/components/atoms/input'
import { PagesTable } from '@/features/page/ui/pages-table'
import { AddPageDialog } from '@/features/page/ui/add-page-dialog'
import { createPageServer } from '@/features/page/application/services/server'
import { PageApi } from '@workspace/services/page-api'
import type { Page, CreatePageDto } from '@/features/page/domain/types'

export interface WorkspaceDetailFeatureProps {
  workspaceId: string
  workspaceName: string
  initialPages?: Page[]
}

export function WorkspaceDetailFeature({
  workspaceId,
  workspaceName,
  initialPages = [],
}: Readonly<WorkspaceDetailFeatureProps>) {
  const [pages, setPages] = useState<Page[]>(initialPages)
  const [searchQuery, setSearchQuery] = useState('')
  const [isAddPageOpen, setIsAddPageOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)

  const handleAddPage = async (data: CreatePageDto) => {
    setIsLoading(true)
    setError(null)

    try {
      const newPage = await createPageServer(data)
      setPages((prev) => [...prev, newPage])
      setIsAddPageOpen(false)
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to add page'))
    } finally {
      setIsLoading(false)
    }
  }

  const handleViewChanges = (pageId: string) => {
    console.log('View changes for page:', pageId)
    // NOTE: Navigate to page changes view when route is ready
  }

  const handlePageClick = (pageId: string) => {
    console.log('Page clicked:', pageId)
    // NOTE: Navigate to page detail when route is ready
  }

  const handleCheckFrequencyChange = async (pageId: string, frequency: string) => {
    // Optimistically update UI
    setPages((prev) =>
      prev.map((page) =>
        page.id === pageId ? { ...page, checkFrequency: frequency } : page
      )
    )

    try {
      await PageApi.updateMonitoringConfig(pageId, frequency)
    } catch (err) {
      // Revert on error
      setPages(initialPages)
      console.error('Failed to update check frequency:', err)
    }
  }

  const filteredPages = pages.filter((page) =>
    page.name.toLowerCase().includes(searchQuery.toLowerCase())
  )

  return (
    <div className="flex-1 flex flex-col bg-background">
      {/* Welcome Container */}
      <div className="flex justify-between items-start gap-2 px-24 py-6">
        <div className="flex flex-col gap-2">
          <h1 className="text-2xl font-semibold text-foreground">
            Added pages for {workspaceName}
          </h1>
          <p className="text-base font-normal text-muted-foreground">
            Here are all the pages you've added to this workspace.
          </p>
        </div>
      </div>

      {/* Search and Add Button */}
      <div className="flex justify-between items-center px-24 py-2 gap-4">
        <div className="relative flex-1 max-w-sm">
          <svg
            width="17"
            height="17"
            viewBox="0 0 17 17"
            fill="none"
            className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground"
          >
            <path
              d="M7.79167 13.4583C10.8292 13.4583 13.2917 10.9958 13.2917 7.95833C13.2917 4.92084 10.8292 2.45833 7.79167 2.45833C4.75418 2.45833 2.29167 4.92084 2.29167 7.95833C2.29167 10.9958 4.75418 13.4583 7.79167 13.4583Z"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
            <path
              d="M14.5833 14.75L11.7292 11.8958"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
          <Input
            type="search"
            placeholder="Search workspaces"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>

        <div className="flex items-center gap-4">
          <Button
            variant="default"
            onClick={() => setIsAddPageOpen(true)}
            className="h-10 px-4 gap-2 bg-primary"
          >
            <SquarePlus className="w-4 h-4" />
            Add page
          </Button>
        </div>
      </div>

      {/* Pages Table */}
      <div className="px-24 py-2 pb-6">
        <PagesTable
          pages={filteredPages}
          onViewChanges={handleViewChanges}
          onPageClick={handlePageClick}
          onCheckFrequencyChange={handleCheckFrequencyChange}
        />
      </div>

      {/* Add Page Dialog */}
      <AddPageDialog
        open={isAddPageOpen}
        onOpenChange={setIsAddPageOpen}
        onSubmit={handleAddPage}
        workspaceId={workspaceId}
        isLoading={isLoading}
        error={error}
      />
    </div>
  )
}

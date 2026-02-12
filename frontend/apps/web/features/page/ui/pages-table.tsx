'use client'

import * as React from 'react'
import { useState } from 'react'
import { ChevronDown, Clock, RefreshCcw } from 'lucide-react'
import { Button } from '@workspace/ui/components/atoms/button'
import { Badge } from '@workspace/ui/components/atoms/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@workspace/ui/components/atoms/select'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@workspace/ui/components/molecules/dropdown-menu'
import { cn } from '@workspace/ui/lib/utils'
import type { Page } from '../domain/types'

const CHECK_FREQUENCIES = [
  'Off',
  'Every 1 hour',
  'Every 2 hours',
  'Every 8 hours',
  'Every day',
  'Every 48 hours',
] as const

export interface PagesTableProps {
  pages: Page[]
  onViewChanges?: (pageId: string) => void
  onPageClick?: (pageId: string) => void
  onCheckFrequencyChange?: (pageId: string, frequency: string) => void
  onEdit?: (page: Page) => void
  onDelete?: (page: Page) => void
}

export function PagesTable({
  pages,
  onViewChanges,
  onPageClick,
  onCheckFrequencyChange,
  onEdit,
  onDelete,
}: Readonly<PagesTableProps>) {
  const [selectedPages, setSelectedPages] = useState<Set<string>>(new Set())

  const toggleSelectAll = () => {
    if (selectedPages.size === pages.length) {
      setSelectedPages(new Set())
    } else {
      setSelectedPages(new Set(pages.map((p) => p.id)))
    }
  }

  const toggleSelect = (pageId: string) => {
    const newSelected = new Set(selectedPages)
    if (newSelected.has(pageId)) {
      newSelected.delete(pageId)
    } else {
      newSelected.add(pageId)
    }
    setSelectedPages(newSelected)
  }

  const formatLastChange = (
    lastChangeDetectedAt?: string
  ): {
    text: string
    variant: 'default' | 'success'
  } => {
    if (!lastChangeDetectedAt) {
      return {
        text: 'No changes detected',
        variant: 'default',
      }
    }

    const date = new Date(lastChangeDetectedAt)
    const formatted = date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    })
    return {
      text: formatted,
      variant: 'success',
    }
  }

  return (
    <div className="bg-card border border-border rounded-lg overflow-hidden">
      <div className="overflow-x-auto">
        <div className="min-w-[1000px]">
          {/* Table Header */}
          <div className="flex items-center border-b border-border bg-background">
            {/* Checkbox Column */}
            <div className="flex items-center gap-2.5 px-2 py-2.5 w-8">
          <button
            type="button"
            onClick={toggleSelectAll}
            className={cn(
              'w-4 h-4 border border-border rounded flex items-center justify-center',
              'hover:border-primary transition-colors',
              selectedPages.size === pages.length && 'bg-primary border-primary'
            )}
            aria-label={selectedPages.size === pages.length ? 'Deselect all' : 'Select all'}
          >
            {selectedPages.size === pages.length && (
              <svg width="10" height="8" viewBox="0 0 10 8" fill="none">
                <title>Selected</title>
                <path
                  d="M1 4L3.5 6.5L9 1"
                  stroke="white"
                  strokeWidth="1.5"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
            )}
          </button>
        </div>

        {/* Page Name */}
        <div className="flex items-center gap-2.5 px-2 py-2.5 flex-[0_0_200px]">
          <span className="text-sm font-medium text-foreground/88">Page name</span>
          <ChevronDown className="w-4 h-4 text-foreground/88" />
        </div>

        {/* Tag */}
        <div className="flex items-center gap-2.5 px-2 py-2.5 flex-[0_0_150px]">
          <span className="text-sm font-medium text-foreground/88">Tag</span>
          <ChevronDown className="w-4 h-4 text-foreground/88" />
        </div>

        {/* Check Frequency */}
        <div className="flex items-center gap-2.5 px-2 py-2.5 flex-[0_0_180px]">
          <Clock className="w-4 h-4 text-foreground/88" />
          <span className="text-sm font-medium text-foreground/88">Check Frequency</span>
        </div>

        {/* Thumbnail */}
        <div className="flex items-center justify-center gap-2.5 px-2 py-2.5 flex-[0_0_125px]">
          <span className="text-sm font-medium text-foreground/88">Thumbnail</span>
        </div>

        {/* Last Change */}
        <div className="flex items-center gap-2.5 px-2 py-2.5 flex-[0_0_163px]">
          <RefreshCcw className="w-4 h-4 text-foreground/88" />
          <span className="text-sm font-medium text-foreground/88">Last change</span>
          <ChevronDown className="w-4 h-4 text-foreground/88" />
        </div>

        {/* Detected Changes */}
        <div className="flex items-center gap-2.5 px-2 py-2.5 flex-[0_0_150px]">
          <span className="text-sm font-medium text-foreground/88">Detected Changes</span>
          <ChevronDown className="w-4 h-4 text-foreground/88" />
        </div>

        {/* Actions */}
        <div className="flex items-center px-2 py-2.5 flex-[0_0_60px]" />
      </div>

      {/* Table Body */}
      <div className="divide-y divide-border">
        {pages.length === 0 ? (
          <div className="px-6 py-12 text-center">
            <p className="text-sm text-muted-foreground">No pages found</p>
          </div>
        ) : (
          pages.map((page) => {
            const isSelected = selectedPages.has(page.id)
            const { text: lastChangeText, variant: lastChangeVariant } = formatLastChange(
              page.lastChangeDetectedAt
            )
            const firstTag = page.tags && page.tags.length > 0 ? page.tags[0] : undefined

            return (
              <div key={page.id} className="flex items-center hover:bg-muted/50 transition-colors">
                {/* Checkbox */}
                <div className="flex items-center px-2 py-2 w-8">
                  <button
                    type="button"
                    onClick={() => toggleSelect(page.id)}
                    className={cn(
                      'w-4 h-4 border border-border rounded flex items-center justify-center',
                      'hover:border-primary transition-colors',
                      isSelected && 'bg-primary border-primary'
                    )}
                    aria-label={isSelected ? `Deselect ${page.name}` : `Select ${page.name}`}
                  >
                    {isSelected && (
                      <svg width="10" height="8" viewBox="0 0 10 8" fill="none">
                        <title>Selected</title>
                        <path
                          d="M1 4L3.5 6.5L9 1"
                          stroke="white"
                          strokeWidth="1.5"
                          strokeLinecap="round"
                          strokeLinejoin="round"
                        />
                      </svg>
                    )}
                  </button>
                </div>

                {/* Page Name */}
                <div className="flex items-center px-2 py-2 flex-[0_0_200px]">
                  <button
                    type="button"
                    onClick={() => onPageClick?.(page.id)}
                    className="text-sm font-normal text-foreground hover:underline text-left truncate"
                  >
                    {page.name}
                  </button>
                </div>

                {/* Tag */}
                <div className="flex items-center px-2 py-2 flex-[0_0_150px]">
                  {firstTag && (
                    <Badge
                      variant="outline"
                      className="px-2 py-0.5 text-xs font-medium text-foreground border-border"
                    >
                      {firstTag}
                    </Badge>
                  )}
                </div>

                {/* Check Frequency */}
                <div className="flex items-center px-2 py-2 flex-[0_0_180px]">
                  <Select
                    value={page.checkFrequency}
                    onValueChange={(value) => onCheckFrequencyChange?.(page.id, value)}
                  >
                    <SelectTrigger className="h-7 text-sm border-none shadow-none hover:bg-muted/50 focus:ring-0">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {CHECK_FREQUENCIES.map((freq) => (
                        <SelectItem key={freq} value={freq}>
                          {freq}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                {/* Thumbnail */}
                <div className="flex items-center justify-center px-2 py-2 flex-[0_0_125px]">
                  {page.thumbnailUrl ? (
                    <img
                      src={page.thumbnailUrl}
                      alt={`${page.name} thumbnail`}
                      className="w-14 h-9 object-cover rounded border border-border"
                    />
                  ) : (
                    <div className="w-14 h-9 bg-muted rounded border border-border" />
                  )}
                </div>

                {/* Last Change */}
                <div className="flex items-center px-2 py-2 flex-[0_0_163px]">
                  <span
                    className={cn(
                      'text-sm font-medium',
                      lastChangeVariant === 'success' ? 'text-foreground' : 'text-muted-foreground'
                    )}
                  >
                    {lastChangeText}
                  </span>
                </div>

                {/* Detected Changes */}
                <div className="flex items-center justify-center px-2 py-2 flex-[0_0_150px]">
                  {page.detectedChanges > 0 ? (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => onViewChanges?.(page.id)}
                      className="px-2.5 py-1.5 h-auto text-xs font-medium gap-1 bg-background"
                    >
                      <RefreshCcw className="w-3.5 h-3.5" />
                      View Changes
                      <span className="ml-1 w-2 h-2 bg-destructive rounded-full" />
                    </Button>
                  ) : (
                    <Button
                      variant="outline"
                      size="sm"
                      className="px-2.5 py-1.5 h-auto text-xs font-medium gap-1 bg-muted/50"
                      disabled
                    >
                      <RefreshCcw className="w-3.5 h-3.5" />
                      View Changes
                    </Button>
                  )}
                </div>

                {/* Actions */}
                <div className="flex items-center justify-center px-2 py-2 flex-[0_0_60px]">
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <button
                        type="button"
                        className="p-1 hover:bg-muted rounded transition-colors"
                        aria-label="More actions"
                      >
                        <svg width="21" height="21" viewBox="0 0 21 21" fill="none">
                          <title>More actions</title>
                          <path
                            d="M10.5 11.375C10.9832 11.375 11.375 10.9832 11.375 10.5C11.375 10.0168 10.9832 9.625 10.5 9.625C10.0168 9.625 9.625 10.0168 9.625 10.5C9.625 10.9832 10.0168 11.375 10.5 11.375Z"
                            stroke="currentColor"
                            strokeWidth="1.75"
                            strokeLinecap="round"
                            strokeLinejoin="round"
                          />
                          <path
                            d="M10.5 5.25C10.9832 5.25 11.375 4.85825 11.375 4.375C11.375 3.89175 10.9832 3.5 10.5 3.5C10.0168 3.5 9.625 3.89175 9.625 4.375C9.625 4.85825 10.0168 5.25 10.5 5.25Z"
                            stroke="currentColor"
                            strokeWidth="1.75"
                            strokeLinecap="round"
                            strokeLinejoin="round"
                          />
                          <path
                            d="M10.5 17.5C10.9832 17.5 11.375 17.1082 11.375 16.625C11.375 16.1418 10.9832 15.75 10.5 15.75C10.0168 15.75 9.625 16.1418 9.625 16.625C9.625 17.1082 10.0168 17.5 10.5 17.5Z"
                            stroke="currentColor"
                            strokeWidth="1.75"
                            strokeLinecap="round"
                            strokeLinejoin="round"
                          />
                        </svg>
                      </button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem onClick={() => onEdit?.(page)}>Edit Page</DropdownMenuItem>
                      <DropdownMenuItem
                        onClick={() => onDelete?.(page)}
                        className="text-destructive focus:text-destructive"
                      >
                        Delete Page
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </div>
            )
          })
        )}
      </div>
        </div>
      </div>

      {/* Footer */}
      <div className="flex flex-col md:flex-row items-center justify-between gap-4 px-4 md:px-8 lg:px-24 py-3 border-t border-border bg-background">
        <div className="text-sm font-normal text-muted-foreground w-full md:w-auto text-center md:text-left">
          {selectedPages.size} of {pages.length} row(s) selected.
        </div>
        <div className="flex flex-wrap justify-center items-center gap-4 md:gap-8 w-full md:w-auto">
          {/* Rows per page */}
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium text-foreground">Rows per page</span>
            <select className="px-3 py-2 h-9 text-sm border border-border rounded bg-background">
              <option>10</option>
              <option>20</option>
              <option>50</option>
            </select>
          </div>

          {/* Page info */}
          <div className="text-sm font-medium text-foreground">Page 1 of 1</div>

          {/* Pagination buttons */}
          <div className="flex items-center gap-2">
            <Button variant="outline" size="icon-sm" disabled className="h-8 w-8">
              <svg width="17" height="17" viewBox="0 0 17 17" fill="none">
                <title>Previous page</title>
                <path
                  d="M10.625 12.75L6.375 8.5L10.625 4.25"
                  stroke="currentColor"
                  strokeWidth="1.5"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
            </Button>
            <Button variant="outline" size="icon-sm" disabled className="h-8 w-8">
              <svg width="17" height="17" viewBox="0 0 17 17" fill="none">
                <title>Next page</title>
                <path
                  d="M6.375 4.25L10.625 8.5L6.375 12.75"
                  stroke="currentColor"
                  strokeWidth="1.5"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}

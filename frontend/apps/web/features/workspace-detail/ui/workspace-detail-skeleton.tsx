import { Button } from '@workspace/ui/components/atoms/button'
import { ChevronDown, Clock, FileText, RefreshCcw, Settings, SquarePlus, Trash2 } from 'lucide-react'

export function WorkspaceDetailSkeleton() {
  return (
    <div className="flex-1 flex flex-col bg-background">
      {/* Header */}
      <div className="flex flex-col md:flex-row justify-between items-start gap-4 px-4 md:px-8 lg:px-24 py-6">
        <div className="flex flex-col gap-2">
          <div className="flex flex-wrap items-center gap-3">
            <h1 className="text-2xl font-semibold text-foreground">
              Added pages for{' '}
              <span className="inline-block h-6 w-32 bg-muted rounded animate-pulse align-middle" />
            </h1>
          </div>
          <p className="text-base font-normal text-muted-foreground">
            Here are all the pages you&apos;ve added to this workspace.
          </p>
        </div>

        <div className="flex items-center gap-2 w-full md:w-auto">
          <Button variant="outline" disabled className="gap-2 flex-1 md:flex-none">
            <FileText className="w-4 h-4" />
            Reports
          </Button>
          <Button variant="outline" disabled className="gap-2 flex-1 md:flex-none">
            <Settings className="w-4 h-4" />
            Edit Workspace
          </Button>
          <Button variant="destructive" disabled size="icon" className="h-9 w-9 shrink-0">
            <Trash2 className="w-4 h-4" />
          </Button>
        </div>
      </div>

      {/* Search + Add button */}
      <div className="flex flex-col md:flex-row justify-between items-stretch md:items-center px-4 md:px-8 lg:px-24 py-2 gap-4">
        <div className="relative flex-1 w-full md:max-w-sm">
          <svg
            width="17"
            height="17"
            viewBox="0 0 17 17"
            fill="none"
            className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground"
          >
            <title>Search</title>
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
          <div className="h-9 w-full rounded-md border border-input bg-background pl-10 pr-3 flex items-center">
            <span className="text-sm text-muted-foreground">Search pages</span>
          </div>
        </div>

        <Button variant="default" disabled className="h-9 px-4 gap-2 bg-primary w-full md:w-auto">
          <SquarePlus className="w-4 h-4" />
          Add page
        </Button>
      </div>

      {/* Table — real header, skeleton rows */}
      <div className="px-4 md:px-8 lg:px-24 py-2 pb-6">
        <div className="bg-card border border-border rounded-lg overflow-hidden">
          <div className="overflow-x-auto">
            <div className="min-w-[1000px]">
              {/* Real table header */}
              <div className="flex items-center border-b border-border bg-background">
                <div className="flex items-center gap-2.5 px-2 py-2.5 w-8">
                  <div className="w-4 h-4 border border-border rounded" />
                </div>
                <div className="flex items-center gap-2.5 px-2 py-2.5 flex-[0_0_200px]">
                  <span className="text-sm font-medium text-foreground/88">Page name</span>
                  <ChevronDown className="w-4 h-4 text-foreground/88" />
                </div>
                <div className="flex items-center gap-2.5 px-2 py-2.5 flex-[0_0_150px]">
                  <span className="text-sm font-medium text-foreground/88">Tag</span>
                  <ChevronDown className="w-4 h-4 text-foreground/88" />
                </div>
                <div className="flex items-center gap-2.5 px-2 py-2.5 flex-[0_0_180px]">
                  <Clock className="w-4 h-4 text-foreground/88" />
                  <span className="text-sm font-medium text-foreground/88">Check Frequency</span>
                </div>
                <div className="flex items-center justify-center gap-2.5 px-2 py-2.5 flex-[0_0_125px]">
                  <span className="text-sm font-medium text-foreground/88">Thumbnail</span>
                </div>
                <div className="flex items-center gap-2.5 px-2 py-2.5 flex-[0_0_163px]">
                  <RefreshCcw className="w-4 h-4 text-foreground/88" />
                  <span className="text-sm font-medium text-foreground/88">Last change</span>
                  <ChevronDown className="w-4 h-4 text-foreground/88" />
                </div>
                <div className="flex items-center gap-2.5 px-2 py-2.5 flex-[0_0_150px]">
                  <span className="text-sm font-medium text-foreground/88">Detected Changes</span>
                  <ChevronDown className="w-4 h-4 text-foreground/88" />
                </div>
                <div className="flex items-center px-2 py-2.5 flex-[0_0_60px]" />
              </div>

              {/* Skeleton rows */}
              <div className="divide-y divide-border">
                {['a', 'b', 'c', 'd', 'e'].map((id) => (
                  <div key={id} className="flex items-center">
                    <div className="flex items-center px-2 py-2 w-8">
                      <div className="w-4 h-4 border border-border rounded" />
                    </div>
                    <div className="flex items-center px-2 py-2 flex-[0_0_200px]">
                      <div className="h-4 w-28 bg-muted rounded animate-pulse" />
                    </div>
                    <div className="flex items-center px-2 py-2 flex-[0_0_150px]">
                      <div className="h-5 w-16 bg-muted rounded animate-pulse" />
                    </div>
                    <div className="flex items-center px-2 py-2 flex-[0_0_180px]">
                      <div className="h-4 w-20 bg-muted rounded animate-pulse" />
                    </div>
                    <div className="flex items-center justify-center px-2 py-2 flex-[0_0_125px]">
                      <div className="w-14 h-9 bg-muted rounded border border-border animate-pulse" />
                    </div>
                    <div className="flex items-center px-2 py-2 flex-[0_0_163px]">
                      <div className="h-4 w-20 bg-muted rounded animate-pulse" />
                    </div>
                    <div className="flex items-center px-2 py-2 flex-[0_0_150px]">
                      <div className="h-5 w-8 bg-muted rounded animate-pulse" />
                    </div>
                    <div className="flex items-center justify-center px-2 py-2 flex-[0_0_60px]">
                      <div className="h-5 w-5 bg-muted rounded animate-pulse" />
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Real table footer */}
          <div className="flex flex-col md:flex-row items-center justify-between gap-4 px-4 md:px-8 lg:px-24 py-3 border-t border-border bg-background">
            <div className="text-sm font-normal text-muted-foreground w-full md:w-auto text-center md:text-left">
              0 of 0 row(s) selected.
            </div>
            <div className="flex flex-wrap justify-center items-center gap-4 md:gap-8 w-full md:w-auto">
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium text-foreground">Rows per page</span>
                <div className="px-3 py-2 h-9 text-sm border border-border rounded bg-background">10</div>
              </div>
              <div className="text-sm font-medium text-foreground">Page 1 of 1</div>
              <div className="flex items-center gap-2">
                <Button variant="outline" size="icon-sm" disabled className="h-8 w-8">
                  <svg width="17" height="17" viewBox="0 0 17 17" fill="none">
                    <title>Previous page</title>
                    <path d="M10.625 12.75L6.375 8.5L10.625 4.25" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
                  </svg>
                </Button>
                <Button variant="outline" size="icon-sm" disabled className="h-8 w-8">
                  <svg width="17" height="17" viewBox="0 0 17 17" fill="none">
                    <title>Next page</title>
                    <path d="M6.375 4.25L10.625 8.5L6.375 12.75" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
                  </svg>
                </Button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

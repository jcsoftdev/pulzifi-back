import { Suspense } from 'react'
import { ArrowRight, Search, Settings2, SquarePlus } from 'lucide-react'
import { Button } from '@workspace/ui/components/atoms/button'
import { Card } from '@workspace/ui/components/atoms/card'
import { WorkspaceFeature } from '@/features/workspace'
import { getWorkspacesServer } from '@/features/workspace/application/services/server'

async function WorkspacesLoader() {
  const workspaces = await getWorkspacesServer()
  return <WorkspaceFeature initialWorkspaces={workspaces} lastCheckTime="just now" />
}

function WorkspaceCardSkeleton() {
  return (
    <Card className="w-64 flex flex-col gap-5 shadow-sm rounded-lg py-0">
      <div className="flex flex-col justify-center self-stretch gap-1 p-4 pt-3">
        {/* Header */}
        <div className="flex justify-between items-center self-stretch gap-1">
          <div className="h-3 w-24 bg-muted rounded animate-pulse" />
          <div className="h-5 w-5 bg-muted rounded-sm animate-pulse" />
        </div>
        {/* Title and Badge */}
        <div className="flex flex-col gap-1 py-4">
          <div className="h-6 w-32 bg-muted rounded animate-pulse" />
          <div className="flex gap-1 mt-1">
            <div className="h-5 w-16 bg-muted rounded animate-pulse" />
          </div>
        </div>
      </div>
      {/* Footer */}
      <div className="flex justify-between items-center self-stretch gap-1 p-3 py-2 border-t border-black/6">
        <div className="flex items-center gap-1">
          <div className="flex justify-center items-center gap-1 p-1 w-6 h-6">
            <svg width="13" height="13" viewBox="0 0 13 13" fill="none" className="text-foreground" aria-hidden="true">
              <rect width="5" height="5" rx="1" fill="currentColor" />
              <rect x="7" width="5" height="5" rx="1" fill="currentColor" />
              <rect y="7" width="5" height="5" rx="1" fill="currentColor" />
            </svg>
          </div>
          <p className="text-base font-normal text-black/65">0 pages</p>
        </div>
        <Button variant="ghost" size="icon-sm" disabled className="h-6 w-6">
          <ArrowRight className="h-4 w-4" />
        </Button>
      </div>
    </Card>
  )
}

function WorkspacesListSkeleton() {
  return (
    <div className="flex flex-col gap-1">
      {/* Welcome — static */}
      <div className="flex flex-col md:flex-row justify-between items-start self-stretch gap-4 p-8 px-4 md:px-8 lg:px-24">
        <div className="flex flex-col gap-2">
          <h1 className="text-3xl md:text-5xl font-semibold text-foreground">All Workspaces</h1>
          <p className="text-sm md:text-base font-normal text-black/65">
            Your space to see how you&apos;re doing, what you&apos;ve achieved, and what&apos;s next.
          </p>
        </div>
        <Button variant="ghost" disabled className="flex items-center gap-2">
          <Settings2 className="w-4 h-4" />
          Settings
        </Button>
      </div>

      {/* Search + Create — static */}
      <div className="flex flex-col md:flex-row justify-between items-center self-stretch gap-4 px-4 md:px-8 lg:px-24 py-8">
        <div className="flex justify-stretch items-stretch gap-2.5 px-3 w-96 h-8 rounded-md border border-border bg-background">
          <div className="flex items-center gap-2 py-1 flex-1 h-8">
            <Search className="w-4 h-4 text-foreground" />
            <span className="text-sm font-normal text-muted-foreground/50">Search workspaces</span>
          </div>
        </div>
        <Button variant="default" disabled className="h-9 px-4 gap-2 bg-primary w-full md:w-auto">
          <SquarePlus className="w-4 h-4" />
          Create workplace
        </Button>
      </div>

      {/* Last check — static */}
      <div className="flex justify-between self-stretch gap-1 px-4 md:px-8 lg:px-24 py-8 pb-1">
        <div />
        <div className="flex justify-center items-center gap-1 p-3">
          <p className="text-sm font-normal text-center text-muted-foreground">
            Last check: just now
          </p>
        </div>
      </div>

      {/* Workspace cards — data-dependent, matches real card structure */}
      <div className="flex items-center self-stretch gap-6 px-4 md:px-8 lg:px-24 py-6">
        <div className="flex flex-wrap gap-6">
          {['a', 'b', 'c'].map((i) => (
            <WorkspaceCardSkeleton key={i} />
          ))}
        </div>
      </div>
    </div>
  )
}

export default function WorkspacesPage() {
  return (
    <Suspense fallback={<WorkspacesListSkeleton />}>
      <WorkspacesLoader />
    </Suspense>
  )
}

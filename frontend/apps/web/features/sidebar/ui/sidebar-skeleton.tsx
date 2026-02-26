import { Button } from '@workspace/ui/components/atoms/button'
import { House, Settings, Shapes, Users } from 'lucide-react'
import Link from 'next/link'

export function SidebarSkeleton() {
  return (
    <aside className="w-60 h-screen bg-sidebar border-r border-border flex flex-col p-1">
      {/* Logo — static */}
      <div className="py-2.5 px-3">
        <Button asChild variant="ghost" className="px-1.5 py-1.5 h-auto font-extrabold">
          <Link href="/">
            <span className="text-2xl text-foreground tracking-tight leading-tight">Pulzifi</span>
          </Link>
        </Button>
      </div>

      {/* Organization Selector — data-dependent */}
      <div className="px-2 py-1">
        <div className="h-9 bg-muted rounded-md animate-pulse" />
      </div>

      {/* Divider — static */}
      <div className="h-2 border-t border-border mx-3" />

      {/* Navigation */}
      <div className="flex-1 overflow-y-auto py-2 px-2">
        {/* Home — static */}
        <div className="flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium text-muted-foreground">
          <House className="h-4 w-4" />
          Home
        </div>

        {/* Workspaces list — data-dependent */}
        <div className="mt-1 space-y-1">
          <div className="flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium text-muted-foreground">
            Workspaces
          </div>
          {['a', 'b', 'c'].map((i) => (
            <div key={`workspace-${i}`} className="h-7 bg-muted rounded-md animate-pulse ml-6" />
          ))}
        </div>

        {/* Team — static */}
        <div className="flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium text-muted-foreground mt-1">
          <Users className="h-4 w-4" />
          Team
        </div>
      </div>

      {/* Bottom Section */}
      <div className="p-2 space-y-1">
        {/* Resources — static */}
        <div className="flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium text-muted-foreground">
          <Shapes className="h-4 w-4" />
          Resources
        </div>
        {/* Settings — static */}
        <div className="flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium text-muted-foreground">
          <Settings className="h-4 w-4" />
          Settings
        </div>
        {/* Profile footer — data-dependent */}
        <div className="h-10 bg-muted rounded-md animate-pulse mt-1" />
      </div>
    </aside>
  )
}

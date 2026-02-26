import { Button } from '@workspace/ui/components/atoms/button'
import { Pencil, Trash2 } from 'lucide-react'

const INSIGHT_TYPES = [
  { id: 'marketing', label: 'Marketing Lens' },
  { id: 'market_analysis', label: 'Market Analysis' },
  { id: 'business_opportunities', label: 'Business Opportunities' },
  { id: 'job_recommendation', label: 'Job recommendation' },
] as const

const ALERT_CONDITIONS = [
  { id: 'any_changes', label: 'Any changes' },
  { id: 'new_article', label: 'A new article is published on the site' },
  { id: 'new_comment', label: 'New comment added' },
  { id: 'main_nav_changes', label: "Site's main navigation menu changes" },
] as const

export function PageDetailSkeleton() {
  return (
    <div className="flex-1 flex flex-col bg-background overflow-auto">
      <div className="px-4 md:px-8 py-6 md:py-8 space-y-6 md:space-y-8 max-w-7xl mx-auto w-full">
        {/* PageInfoCard */}
        <div className="flex flex-col gap-6 bg-card p-6 rounded-xl shadow-sm border border-border">
          <div className="flex flex-col md:flex-row justify-between gap-6">
            <div className="flex gap-4">
              <div className="flex flex-col gap-2">
                <div className="flex items-center gap-2">
                  {/* Page name — data */}
                  <div className="h-7 w-48 bg-muted rounded animate-pulse" />
                  <Button type="button" variant="ghost" size="icon-sm" disabled className="text-muted-foreground">
                    <Pencil className="h-4 w-4" />
                  </Button>
                </div>
                {/* URL — data */}
                <div className="h-4 w-72 bg-muted rounded animate-pulse" />
              </div>
            </div>
            <div className="flex items-center gap-3">
              <Button variant="outline" disabled className="h-9 px-4 gap-2 bg-transparent border-border text-foreground">
                <Trash2 className="h-4 w-4" />
                Delete
              </Button>
              <Button type="button" variant="outline" disabled className="px-3 py-2 h-auto gap-2">
                <span className="text-sm">View Changes</span>
                <span className="w-2 h-2 rounded-full bg-foreground" />
              </Button>
            </div>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 md:gap-8">
          <div className="flex flex-col gap-6 md:gap-8 lg:col-span-2">
            {/* ChecksHistory */}
            <div className="flex flex-col gap-6 bg-card border border-border rounded-xl p-6 h-full">
              <div className="flex items-center justify-between">
                <h2 className="text-xl font-semibold text-foreground">Checks history</h2>
              </div>
              <div className="flex flex-col gap-4 max-h-[500px] overflow-y-auto pr-2">
                <div className="flex items-center gap-2">
                  <h3 className="text-sm font-medium text-muted-foreground">Today</h3>
                  <div className="h-px flex-1 bg-border" />
                </div>
                <div className="relative border-l border-border ml-2 space-y-8">
                  {['a', 'b', 'c'].map((i) => (
                    <div key={i} className="relative pl-6">
                      <div className="absolute -left-1.5 top-1.5 h-3 w-3 rounded-full border-2 border-background bg-green-100" />
                      <div className="flex flex-col gap-1 p-2 -ml-2 rounded-md">
                        <div className="h-4 w-20 bg-muted rounded animate-pulse" />
                        <div className="h-4 w-40 bg-muted rounded animate-pulse" />
                        <div className="mt-1 h-4 w-32 bg-muted rounded animate-pulse" />
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>

            {/* AdvancedSettingsCard */}
            <div className="flex flex-col gap-6 bg-card border border-border rounded-xl p-6 h-full">
              <h3 className="text-xl font-semibold text-foreground">Advanced settings</h3>
              <div className="flex flex-col gap-4">
                <div className="flex items-center gap-3 flex-wrap">
                  <div className="flex items-center gap-2">
                    <div className="w-4 h-4 rounded-full border border-border bg-muted" />
                    <span className="text-sm text-muted-foreground">Work days</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="w-4 h-4 rounded-full border border-border bg-muted" />
                    <span className="text-sm text-muted-foreground">Work days, during work hours</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="w-4 h-4 rounded-full border border-border bg-muted" />
                    <span className="text-sm text-muted-foreground">Weekdays</span>
                  </div>
                </div>
                <div className="h-3 w-64 bg-muted rounded animate-pulse" />
                <div className="flex items-center gap-2">
                  <div className="h-7 w-28 bg-muted rounded-md animate-pulse" />
                  <div className="h-7 w-32 bg-muted rounded-md animate-pulse" />
                </div>
                <div className="h-3 w-56 bg-muted rounded animate-pulse" />
              </div>
            </div>
          </div>

          <div className="flex flex-col gap-6 md:gap-8 lg:col-span-1">
            {/* GeneralSummaryCard */}
            <div className="flex flex-col gap-6 bg-card border border-border rounded-xl p-6 h-full">
              <div className="flex items-center justify-between">
                <h3 className="text-xl font-semibold text-foreground">General Summary</h3>
              </div>
              <div className="flex flex-col gap-4">
                {/* Tag */}
                <div className="flex flex-col gap-2">
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium text-muted-foreground">Tag</span>
                    <Button type="button" variant="ghost" size="icon-sm" disabled className="text-muted-foreground">
                      <Pencil className="w-4 h-4" />
                    </Button>
                  </div>
                  <div className="h-5 w-20 bg-muted rounded animate-pulse" />
                </div>
                {/* Check Frequency */}
                <div className="flex flex-col gap-2">
                  <span className="text-sm font-medium text-muted-foreground">Check Frequency</span>
                  <div className="h-5 w-24 bg-muted rounded animate-pulse" />
                </div>
                {/* Block ads option */}
                <div className="flex flex-col gap-3 mt-2">
                  <div className="flex items-center gap-2">
                    <div className="w-4 h-4 rounded-sm border border-muted-foreground" />
                    <span className="text-sm text-foreground">Block ads and cookie banners</span>
                  </div>
                </div>
              </div>
            </div>

            {/* IntelligentInsightsCard */}
            <div className="flex flex-col gap-5 bg-card border border-primary/25 rounded-xl p-6">
              <h3 className="text-xl font-semibold text-foreground">Intelligent Insights</h3>
              <div className="flex flex-col gap-3">
                {INSIGHT_TYPES.map((type) => (
                  <div key={type.id} className="flex items-center gap-3">
                    <div className="w-4 h-4 rounded-sm border border-muted-foreground flex-shrink-0" />
                    <span className="text-sm text-foreground">{type.label}</span>
                  </div>
                ))}
                <p className="text-sm text-muted-foreground pl-7">New insights coming soon...</p>
              </div>
              <div className="border-t border-border" />
              <div className="flex flex-col gap-3">
                <p className="text-sm text-muted-foreground">Alert me when</p>
                {ALERT_CONDITIONS.map((condition) => (
                  <div key={condition.id} className="flex items-center gap-3">
                    <div className="w-4 h-4 rounded-sm border border-muted-foreground flex-shrink-0" />
                    <span className="text-sm text-foreground">{condition.label}</span>
                  </div>
                ))}
                <textarea
                  rows={3}
                  disabled
                  placeholder="Add your own here"
                  className="mt-1 w-full resize-none rounded-lg border border-border bg-muted/30 px-3 py-2 text-sm placeholder:text-muted-foreground"
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

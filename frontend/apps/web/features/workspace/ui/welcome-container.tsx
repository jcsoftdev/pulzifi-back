import { Settings2 } from 'lucide-react'
import { Button } from '@workspace/ui/components/atoms/button'

export interface WelcomeContainerProps {
  onSettings?: () => void
}

export function WelcomeContainer({ onSettings }: Readonly<WelcomeContainerProps>) {
  return (
    <div className="flex flex-col md:flex-row justify-between items-start self-stretch gap-4 p-8 px-4 md:px-8 lg:px-24">
      <div className="flex flex-col gap-2">
        <h1 className="text-3xl md:text-5xl font-semibold text-foreground">All Workspaces</h1>
        <p className="text-sm md:text-base font-normal text-black/65">
          Your space to see how you're doing, what you've achieved, and what's next.
        </p>
      </div>
      <Button variant="ghost" onClick={onSettings} className="flex items-center gap-2">
        <Settings2 className="w-4 h-4" />
        Settings
      </Button>
    </div>
  )
}

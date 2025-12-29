import { Card } from '@workspace/ui/components/atoms/card'

export interface EmptyStateProps {
  message?: string
}

export function EmptyState({ message = 'No workspaces found' }: Readonly<EmptyStateProps>) {
  return (
    <Card className="w-full p-12 flex flex-col items-center justify-center gap-4">
      <div className="text-center">
        <p className="text-base font-normal text-muted-foreground">{message}</p>
      </div>
    </Card>
  )
}

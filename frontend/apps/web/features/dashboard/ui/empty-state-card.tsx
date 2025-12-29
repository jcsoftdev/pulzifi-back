'use client'

import { Button } from '@workspace/ui/index'
import type * as React from 'react'

export interface EmptyStateCardProps {
  title: string
  description: string
  buttonText: string
  onButtonClick: () => void
  icon?: React.ReactNode
}

export function EmptyStateCard({
  title,
  description,
  buttonText,
  onButtonClick,
  icon,
}: Readonly<EmptyStateCardProps>) {
  return (
    <div className="bg-card rounded-lg border border-border p-8 text-center">
      {icon && <div className="mb-4 flex justify-center">{icon}</div>}
      <h3 className="text-lg font-semibold text-foreground mb-2">{title}</h3>
      <p className="text-sm text-muted-foreground mb-6 max-w-md mx-auto">{description}</p>
      <Button onClick={onButtonClick} variant={'default'}>
        {buttonText}
      </Button>
    </div>
  )
}

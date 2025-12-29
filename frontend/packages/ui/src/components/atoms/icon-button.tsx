import * as React from 'react'
import { cva, type VariantProps } from 'class-variance-authority'

import { cn } from '@workspace/ui/lib/utils'

const iconButtonVariants = cva(
  'inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50',
  {
    variants: {
      variant: {
        default: 'bg-primary text-primary-foreground hover:bg-primary/90',
        ghost: 'hover:bg-accent hover:text-accent-foreground',
        outline: 'border border-input bg-background hover:bg-accent hover:text-accent-foreground',
      },
      size: {
        default: 'h-10 w-10',
        sm: 'h-9 w-9',
        lg: 'h-11 w-11',
        icon: 'h-10 w-10',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  }
)

export interface IconButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof iconButtonVariants> {}

const IconButton = React.forwardRef<HTMLButtonElement, IconButtonProps>(
  ({ className, variant, size, ...props }, ref) => (
    <button
      className={cn(
        iconButtonVariants({
          variant,
          size,
          className,
        })
      )}
      ref={ref}
      {...props}
    />
  )
)
IconButton.displayName = 'IconButton'

export { IconButton, iconButtonVariants }

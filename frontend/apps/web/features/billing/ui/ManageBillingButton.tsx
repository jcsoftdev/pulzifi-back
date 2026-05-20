'use client'

import { Button } from '@workspace/ui/components/atoms/button'
import { useCustomerPortal } from '../application/useCustomerPortal'

/**
 * Button that opens the Stripe Customer Portal.
 * Shows a loading state while redirecting.
 */
export function ManageBillingButton() {
  const { isLoading, error, openPortal } = useCustomerPortal()

  return (
    <div className="flex flex-col items-start gap-1">
      <Button
        type="button"
        variant="outline"
        size="sm"
        disabled={isLoading}
        onClick={openPortal}
      >
        {isLoading ? 'Redirecting...' : 'Manage Billing'}
      </Button>
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  )
}

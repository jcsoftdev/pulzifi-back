/**
 * Billing API Service
 * Consumes backend /api/v1/billing/* endpoints
 * Works dynamically in both server-side and client-side contexts
 */

import { getHttpClient } from '@workspace/shared-http'

// ---- DTOs (mirror backend response shapes) ----

export interface SubscriptionDto {
  org_id: string
  stripe_subscription_id: string
  stripe_customer_id: string
  plan_id: string
  /** Stable code from public.plans.code — use this for plan equality checks, not plan_id */
  plan_code: string
  /** Human-readable plan name from public.plans.name */
  plan_name: string
  /** "monthly" | "yearly" | "" — empty when no Stripe subscription exists */
  billing_cycle?: 'monthly' | 'yearly' | ''
  /** null when the org has no active Stripe subscription (FR7) */
  billing_status: BillingStatusDto | null
  current_period_end: string | null
  /** null when the org has no active Stripe subscription (FR7) */
  payment_status: PaymentStatusDto | null
  updated_at: string
  /** Account credit available — auto-applied to upcoming invoices. 0 when none. */
  credit_balance_cents?: number
  credit_balance_currency?: string
}

export type BillingStatusDto = 'active' | 'past_due' | 'canceled' | 'incomplete' | 'trialing'

export type PaymentStatusDto = 'ok' | 'grace_period' | 'failed'

export interface CheckoutSessionDto {
  checkout_url: string
}

export interface PortalSessionDto {
  portal_url: string
}

// ---- API Object ----

export const BillingApi = {
  /**
   * Create a Stripe Checkout session for the given plan and billing cycle.
   * Returns a checkout_url to redirect the user to.
   */
  async createCheckoutSession(
    planId: string,
    billingCycle: 'monthly' | 'yearly'
  ): Promise<{
    checkoutUrl: string
  }> {
    const http = await getHttpClient()
    const response = await http.post<CheckoutSessionDto>('/api/v1/billing/checkout', {
      plan_id: planId,
      billing_cycle: billingCycle,
    })
    return {
      checkoutUrl: response.checkout_url,
    }
  },

  /**
   * Open the Stripe Customer Portal for managing existing subscriptions.
   * Returns a portal_url to redirect the user to.
   */
  async openCustomerPortal(): Promise<{
    portalUrl: string
  }> {
    const http = await getHttpClient()
    const response = await http.post<PortalSessionDto>('/api/v1/billing/portal', {})
    return {
      portalUrl: response.portal_url,
    }
  },

  /**
   * Get the current subscription for the authenticated organization.
   *
   * The backend always returns HTTP 200 (FR7):
   * - Active subscription  → full SubscriptionDto with billing_status set
   * - No subscription yet  → { stripe_status: null, ... } (billing_status: null)
   * - BillingEnabled=false → 404 (caught and returned as null)
   *
   * Returns null only when billing is completely disabled (404).
   * Callers should check `subscription.billing_status !== null` to distinguish
   * "has active Stripe sub" from "no Stripe sub yet".
   */
  async getSubscription(): Promise<SubscriptionDto | null> {
    try {
      const http = await getHttpClient()
      return await http.get<SubscriptionDto>('/api/v1/billing/subscription')
    } catch (err) {
      // BillingEnabled=false → API returns 404 on all billing routes
      if (
        err instanceof Error &&
        'status' in err &&
        (
          err as {
            status: number
          }
        ).status === 404
      ) {
        return null
      }
      if (err instanceof Error && err.message?.includes('404')) {
        return null
      }
      throw err
    }
  },
}

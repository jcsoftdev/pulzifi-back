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
  /** True when a "free month" gift coupon is applied to the subscription. */
  gift_active?: boolean
  /** Discount the gift applies to the next invoice (cents). */
  gift_amount_off_cents?: number
  /** True when the user scheduled cancellation at the end of the paid period. */
  cancel_at_period_end?: boolean
  /** When the scheduled cancellation takes effect (ISO date); null otherwise. */
  cancel_at?: string | null
  /** True while the org is using a gifted higher-tier plan for free. */
  plan_gift_active?: boolean
  /** Plan code being used for free during the gift (e.g. "pro"). */
  gift_plan_code?: string
  /** Plan code the org returns to when the gift ends (e.g. "starter"). */
  gift_revert_plan_code?: string
  /** When the plan gift ends and reverts (ISO date); null otherwise. */
  gift_ends_at?: string | null
}

export interface CancelSubscriptionDto {
  subscription_id: string
  cancel_at_period_end: boolean
  cancel_at: string | null
  message: string
}

export type BillingStatusDto = 'active' | 'past_due' | 'canceled' | 'incomplete' | 'trialing'

export type PaymentStatusDto = 'ok' | 'grace_period' | 'failed'

export interface CheckoutSessionDto {
  checkout_url: string
}

export interface PortalSessionDto {
  portal_url: string
}

export interface UpdateSubscriptionDto {
  subscription_id: string
  new_price_id: string
  billing_cycle: 'monthly' | 'yearly'
  /** Refund for the unused checks of the current plan (cents). */
  usage_credit_cents: number
  /** Stripe customer balance credit auto-applied to invoices (cents). */
  account_credit_cents: number
  /** Full price of the new plan starting today (cents). */
  new_plan_charge_cents: number
  /** Final amount charged today = max(0, new_plan_charge - usage_credit - account_credit). */
  amount_due_cents: number
  currency: string
  checks_remaining: number
  checks_allowed: number
  preview: boolean
}

/** One plan from GET /api/v1/billing/plans. Null fields mean "unlimited". */
export interface PlanDto {
  id: string
  code: string
  name: string
  description: string
  price_monthly_cents: number
  price_yearly_cents: number
  currency: string
  checks_allowed_monthly: number | null
  storage_period_days: number
  ai_insights_allowed_monthly: number | null
  max_pages: number | null
  max_workspaces: number | null
  is_popular: boolean
}

// ---- API Object ----

export const BillingApi = {
  /**
   * Create a Stripe Checkout session for the given plan and billing cycle.
   * Returns a checkout_url to redirect the user to.
   */
  async createCheckoutSession(
    planId: string,
    billingCycle: 'monthly' | 'yearly',
    promotionCode?: string
  ): Promise<{
    checkoutUrl: string
  }> {
    const http = await getHttpClient()
    const response = await http.post<CheckoutSessionDto>('/api/v1/billing/checkout', {
      plan_id: planId,
      billing_cycle: billingCycle,
      promotion_code: promotionCode ?? '',
    })
    return {
      checkoutUrl: response.checkout_url,
    }
  },

  /**
   * In-place subscription change. Mutates the existing Stripe Subscription
   * without sending the user through Stripe-hosted UI. Used for upgrades and
   * downgrades initiated from the app's own pricing UI.
   *
   * When preview=true, returns the prorated amount that would be charged
   * without actually applying the change.
   */
  async updateSubscription(
    planId: string,
    billingCycle: 'monthly' | 'yearly',
    preview = false
  ): Promise<UpdateSubscriptionDto> {
    const http = await getHttpClient()
    return await http.post<UpdateSubscriptionDto>('/api/v1/billing/subscription', {
      plan_id: planId,
      billing_cycle: billingCycle,
      preview,
    })
  },

  /**
   * Schedule cancellation of the current subscription at the end of the paid
   * period. The org keeps access until then, then drops to no active plan.
   */
  async cancelSubscription(): Promise<CancelSubscriptionDto> {
    const http = await getHttpClient()
    return await http.post<CancelSubscriptionDto>('/api/v1/billing/subscription/cancel', {})
  },

  /**
   * Revert a pending period-end cancellation, restoring normal renewal.
   */
  async resumeSubscription(): Promise<CancelSubscriptionDto> {
    const http = await getHttpClient()
    return await http.post<CancelSubscriptionDto>('/api/v1/billing/subscription/resume', {})
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

  /**
   * Fetch the active plan catalog.
   * Public endpoint — does not require authentication.
   * Prices come from public.plans kept in sync by the Stripe webhook.
   */
  async getPlans(): Promise<PlanDto[]> {
    const http = await getHttpClient()
    const response = await http.get<{ plans: PlanDto[] }>('/api/v1/billing/plans')
    return response.plans ?? []
  },
}

// ---- Admin Coupons (SUPER_ADMIN) ----

export interface CouponDto {
  id: string
  code: string
  amount_off_cents: number
  currency: string
  active: boolean
  max_redemptions: number
  times_redeemed: number
  expires_at: number
  plan_code: string
  billing_cycle: string
  apply_url: string
}

export interface CreateCouponInput {
  planCode: string
  billingCycle: 'monthly' | 'yearly'
  code?: string
  maxRedemptions?: number
  expiresAt?: number
}

export interface CreateCouponDto {
  coupon_id: string
  promotion_code_id: string
  code: string
  amount_off_cents: number
  currency: string
  apply_url: string
}

export const AdminCouponsApi = {
  /** Create a first-month-$1 promo for a plan (SUPER_ADMIN). */
  async create(input: CreateCouponInput): Promise<CreateCouponDto> {
    const http = await getHttpClient()
    return await http.post<CreateCouponDto>('/api/v1/billing/admin/coupons', {
      plan_code: input.planCode,
      billing_cycle: input.billingCycle,
      code: input.code ?? '',
      max_redemptions: input.maxRedemptions ?? 0,
      expires_at: input.expiresAt ?? 0,
    })
  },

  /** List all promotion codes (active + inactive). */
  async list(): Promise<CouponDto[]> {
    const http = await getHttpClient()
    const res = await http.get<{ coupons: CouponDto[] }>('/api/v1/billing/admin/coupons')
    return res.coupons ?? []
  },

  /** Deactivate a promotion code. */
  async revoke(promotionCodeId: string): Promise<void> {
    const http = await getHttpClient()
    await http.delete(`/api/v1/billing/admin/coupons/${promotionCodeId}`)
  },
}

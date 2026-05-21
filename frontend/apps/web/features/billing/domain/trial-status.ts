/**
 * Billing Feature — Trial Status Domain
 * Mirrors the backend GET /api/v1/usage/trial-status response shape.
 */
export interface TrialStatus {
  is_trial: boolean
  is_expired: boolean
  needs_upgrade: boolean
  converted: boolean
  days_remaining: number
  /** ISO-8601 timestamp; undefined when the org is not on a trial. */
  trial_ends_at?: string
}

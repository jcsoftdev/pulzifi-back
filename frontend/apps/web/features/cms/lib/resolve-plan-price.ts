import type { Pool } from 'pg'

/**
 * Formats a Stripe unit_amount (cents) into a display price string.
 * NULL amount → "Custom" (e.g. Enterprise has no Stripe price).
 * usd renders as "$20" / "$20.50"; other currencies prefix the ISO code.
 */
export function formatCents(amountCents: number | null | undefined, currency: string): string {
  if (amountCents == null) return 'Custom'
  const major = amountCents / 100
  const hasFraction = amountCents % 100 !== 0
  const num = hasFraction ? major.toFixed(2) : String(major)
  if (currency.toLowerCase() === 'usd') return `$${num}`
  return `${currency.toUpperCase()} ${num}`
}

export type ResolvedPlanPrice = {
  price: string
  priceAnnual: string | null
}

/**
 * Reads the cached Stripe amounts from public.plans for a plan code and returns
 * formatted display strings. Reads are read-only; this never writes public.
 * On any failure or missing row, falls back to "Custom" / null so the CMS read
 * never breaks.
 */
export async function resolvePlanPrice(
  pool: Pool,
  planCode: string | null | undefined
): Promise<ResolvedPlanPrice> {
  if (!planCode)
    return {
      price: 'Custom',
      priceAnnual: null,
    }
  try {
    const { rows } = await pool.query(
      `SELECT amount_cents_monthly, amount_cents_yearly, currency
       FROM public.plans WHERE code = $1 AND is_active = TRUE LIMIT 1`,
      [
        planCode,
      ]
    )
    if (rows.length === 0)
      return {
        price: 'Custom',
        priceAnnual: null,
      }
    const r = rows[0]
    const currency: string = r.currency ?? 'usd'
    const monthly = r.amount_cents_monthly as number | null
    const yearly = r.amount_cents_yearly as number | null
    return {
      price: formatCents(monthly, currency),
      priceAnnual: yearly == null ? null : formatCents(yearly, currency),
    }
  } catch (err) {
    console.warn('[cms] resolvePlanPrice failed; falling back to "Custom"', err)
    return {
      price: 'Custom',
      priceAnnual: null,
    }
  }
}

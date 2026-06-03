'use client'

import { AdminCouponsApi, type CouponDto } from '@workspace/services'
import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Input,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@workspace/ui/components/atoms'
import { Copy } from 'lucide-react'
import { useCallback, useEffect, useId, useState } from 'react'

async function copyToClipboard(text: string): Promise<boolean> {
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(text)
      return true
    }
  } catch {
    // fall through to the legacy path
  }
  try {
    const ta = document.createElement('textarea')
    ta.value = text
    ta.style.position = 'fixed'
    ta.style.opacity = '0'
    document.body.appendChild(ta)
    ta.focus()
    ta.select()
    const ok = document.execCommand('copy')
    document.body.removeChild(ta)
    return ok
  } catch {
    return false
  }
}

export function CouponManagement() {
  const [coupons, setCoupons] = useState<CouponDto[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [planCode, setPlanCode] = useState('pro')
  const [cycle, setCycle] = useState('monthly')
  const [code, setCode] = useState('')
  const [maxRedemptions, setMaxRedemptions] = useState('')
  const [creating, setCreating] = useState(false)
  const [copiedId, setCopiedId] = useState<string | null>(null)
  const planId = useId()
  const cycleId = useId()
  const codeId = useId()
  const maxRedemptionsId = useId()

  const handleCopy = async (id: string, url: string) => {
    const ok = await copyToClipboard(url)
    if (ok) {
      setCopiedId(id)
      setTimeout(() => setCopiedId((cur) => (cur === id ? null : cur)), 2000)
    } else {
      setError(`Could not copy link. Copy it manually: ${url}`)
    }
  }

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      setCoupons(await AdminCouponsApi.list())
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load coupons')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  const handleCreate = async () => {
    setCreating(true)
    setError(null)
    try {
      await AdminCouponsApi.create({
        planCode,
        billingCycle: cycle as 'monthly' | 'yearly',
        code: code.trim() || undefined,
        maxRedemptions: maxRedemptions ? Number(maxRedemptions) : undefined,
      })
      setCode('')
      setMaxRedemptions('')
      await load()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to create coupon')
    } finally {
      setCreating(false)
    }
  }

  const handleRevoke = async (id: string) => {
    setError(null)
    try {
      await AdminCouponsApi.revoke(id)
      await load()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to revoke coupon')
    }
  }

  return (
    <div className="w-full">
      <Card>
        <CardHeader>
          <CardTitle>Discount coupons</CardTitle>
          <CardDescription>
            Create a "first month $1" promo for a plan. Customers redeem the code at checkout or via the apply link.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-6">
          {error && (
            <div className="text-sm text-destructive bg-destructive/10 border border-destructive/20 rounded-md px-4 py-3">
              {error}
            </div>
          )}

          <div className="grid grid-cols-1 md:grid-cols-5 gap-4 items-end">
            <div>
              <label htmlFor={planId} className="block text-sm font-medium text-foreground mb-2">
                Plan
              </label>
              <Select value={planCode} onValueChange={setPlanCode} disabled={creating}>
                <SelectTrigger id={planId}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="starter">starter</SelectItem>
                  <SelectItem value="pro">pro</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div>
              <label htmlFor={cycleId} className="block text-sm font-medium text-foreground mb-2">
                Cycle
              </label>
              <Select value={cycle} onValueChange={setCycle} disabled={creating}>
                <SelectTrigger id={cycleId}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="monthly">monthly</SelectItem>
                  <SelectItem value="yearly">yearly</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div>
              <label htmlFor={codeId} className="block text-sm font-medium text-foreground mb-2">
                Code (optional)
              </label>
              <Input
                id={codeId}
                placeholder="TRY1PRO"
                value={code}
                onChange={(e) => setCode(e.target.value)}
                disabled={creating}
                className="text-sm"
              />
            </div>

            <div>
              <label
                htmlFor={maxRedemptionsId}
                className="block text-sm font-medium text-foreground mb-2"
              >
                Max redemptions
              </label>
              <Input
                id={maxRedemptionsId}
                placeholder="∞"
                value={maxRedemptions}
                onChange={(e) => setMaxRedemptions(e.target.value)}
                disabled={creating}
                type="number"
                min={0}
                className="text-sm"
              />
            </div>

            <Button onClick={handleCreate} disabled={creating}>
              {creating ? 'Creating…' : 'Create coupon'}
            </Button>
          </div>

          <div className="overflow-x-auto">
            {loading ? (
              <div className="p-6 text-sm text-muted-foreground text-center">Loading…</div>
            ) : coupons.length === 0 ? (
              <div className="p-6 text-sm text-muted-foreground text-center">No coupons yet.</div>
            ) : (
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border">
                    <th className="text-left py-3 px-4 font-medium">Code</th>
                    <th className="text-left py-3 px-4 font-medium">Plan</th>
                    <th className="text-left py-3 px-4 font-medium">Off</th>
                    <th className="text-left py-3 px-4 font-medium">Redeemed</th>
                    <th className="text-left py-3 px-4 font-medium">Status</th>
                    <th className="text-right py-3 px-4 font-medium">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {coupons.map((c) => (
                    <tr key={c.id} className="border-b border-border hover:bg-muted/50">
                      <td className="py-3 px-4">{c.code}</td>
                      <td className="py-3 px-4">
                        {c.plan_code} / {c.billing_cycle}
                      </td>
                      <td className="py-3 px-4">
                        {(c.amount_off_cents / 100).toLocaleString('en-US', {
                          style: 'currency',
                          currency: (c.currency || 'usd').toUpperCase(),
                        })}
                      </td>
                      <td className="py-3 px-4 text-muted-foreground">
                        {c.times_redeemed}
                        {c.max_redemptions > 0 ? ` / ${c.max_redemptions}` : ''}
                      </td>
                      <td className="py-3 px-4">
                        {c.active ? (
                          <span className="text-emerald-600 dark:text-emerald-400">Active</span>
                        ) : (
                          <span className="text-muted-foreground">Revoked</span>
                        )}
                      </td>
                      <td className="py-3 px-4 text-right">
                        <div className="flex items-center justify-end gap-2">
                          <button
                            type="button"
                            className="text-primary hover:text-primary/80 inline-flex items-center gap-1 text-xs"
                            onClick={() => handleCopy(c.id, c.apply_url)}
                          >
                            <Copy className="w-3 h-3" />
                            {copiedId === c.id ? 'Copied!' : 'Copy'}
                          </button>
                          {c.active && (
                            <button
                              type="button"
                              className="text-destructive hover:text-destructive/80 text-xs"
                              onClick={() => handleRevoke(c.id)}
                            >
                              Revoke
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

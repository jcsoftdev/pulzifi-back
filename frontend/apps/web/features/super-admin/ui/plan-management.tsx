'use client'

import { type AdminOrganizationPlan, type AdminPlan, SuperAdminApi } from '@workspace/services'
import {
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@workspace/ui/components/atoms'
import { Loader2 } from 'lucide-react'
import { useCallback, useEffect, useState, useTransition } from 'react'

export function PlanManagement() {
  const [isPending, startTransition] = useTransition()
  const [plans, setPlans] = useState<AdminPlan[]>([])
  const [organizations, setOrganizations] = useState<AdminOrganizationPlan[]>([])
  const [error, setError] = useState<string | null>(null)

  const loadData = useCallback(async () => {
    try {
      setError(null)
      const [plansData, orgsData] = await Promise.all([
        SuperAdminApi.listPlans(),
        SuperAdminApi.listOrganizations(),
      ])
      setPlans(plansData)
      setOrganizations(orgsData)
    } catch {
      setError('You need SUPER_ADMIN role to manage organization plans.')
    }
  }, [])

  useEffect(() => {
    loadData()
  }, [
    loadData,
  ])

  const handlePlanChange = (organizationId: string, planCode: string) => {
    startTransition(async () => {
      try {
        await SuperAdminApi.assignPlan(organizationId, planCode)
        await loadData()
      } catch {
        setError('Failed to update plan for this organization.')
      }
    })
  }

  if (error) {
    return (
      <div className="flex-1 p-8 max-w-7xl mx-auto w-full">
        <Card>
          <CardHeader>
            <CardTitle>Super Admin Access</CardTitle>
            <CardDescription>{error}</CardDescription>
          </CardHeader>
        </Card>
      </div>
    )
  }

  return (
    <div className="flex-1 p-8 max-w-7xl mx-auto w-full">
      <div className="flex flex-col gap-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-semibold text-foreground">Company Plans</h1>
            <p className="text-sm text-muted-foreground">
              Super admin panel to change organization plan and usage limits.
            </p>
          </div>
          {isPending && <Loader2 className="w-5 h-5 animate-spin text-muted-foreground" />}
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Available Plans</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-wrap gap-2">
            {plans.map((plan) => (
              <Badge key={plan.id} variant="outline">
                {plan.name} ({plan.checks_allowed_monthly}/month)
              </Badge>
            ))}
          </CardContent>
        </Card>

        <div className="grid grid-cols-1 gap-4">
          {organizations.map((org) => (
            <Card key={org.id}>
              <CardHeader>
                <CardTitle className="text-base">{org.name}</CardTitle>
                <CardDescription>{org.subdomain}</CardDescription>
              </CardHeader>
              <CardContent className="flex flex-col md:flex-row md:items-center gap-4">
                <Badge variant="outline">
                  Current: {org.plan_name || org.plan_code || 'No plan'}
                </Badge>
                <Badge variant="outline">Limit: {org.checks_allowed_monthly}/month</Badge>
                <div className="w-full md:w-72">
                  <Select
                    value={org.plan_code || undefined}
                    onValueChange={(value) => handlePlanChange(org.id, value)}
                    disabled={isPending}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select a plan" />
                    </SelectTrigger>
                    <SelectContent>
                      {plans.map((plan) => (
                        <SelectItem key={plan.id} value={plan.code}>
                          {plan.name} ({plan.checks_allowed_monthly}/month)
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>

        <Button variant="outline" onClick={loadData} disabled={isPending}>
          Refresh
        </Button>
      </div>
    </div>
  )
}

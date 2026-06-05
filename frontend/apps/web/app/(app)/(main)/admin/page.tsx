'use client'

import { Tabs, TabsContent, TabsList, TabsTrigger } from '@workspace/ui/components/atoms'
import { useRouter, useSearchParams } from 'next/navigation'
import { CouponManagement } from '@/features/super-admin/ui/coupon-management'
import { PlanManagement } from '@/features/super-admin/ui/plan-management'
import { UserManagement } from '@/features/super-admin/ui/user-management'

export default function AdminPage() {
  const router = useRouter()
  const params = useSearchParams()
  const tab = params.get('tab') ?? 'plans'

  const setTab = (value: string) => {
    const next = new URLSearchParams(params.toString())
    next.set('tab', value)
    if (value !== 'users') next.delete('org')
    router.replace(`/admin?${next.toString()}`)
  }

  return (
    <div className="flex flex-col flex-1 p-8 max-w-7xl mx-auto w-full">
      <Tabs value={tab} onValueChange={setTab} className="w-full">
        <TabsList className="sticky top-0 z-10">
          <TabsTrigger value="plans">Plans</TabsTrigger>
          <TabsTrigger value="coupons">Coupons</TabsTrigger>
          <TabsTrigger value="users">Users</TabsTrigger>
        </TabsList>
        <TabsContent value="plans">
          <PlanManagement />
        </TabsContent>
        <TabsContent value="coupons">
          <CouponManagement />
        </TabsContent>
        <TabsContent value="users">
          <UserManagement initialOrgId={params.get('org') ?? ''} />
        </TabsContent>
      </Tabs>
    </div>
  )
}

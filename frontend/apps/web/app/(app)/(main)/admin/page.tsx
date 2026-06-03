'use client'

import { CouponManagement } from '@/features/super-admin/ui/coupon-management'
import { PlanManagement } from '@/features/super-admin/ui/plan-management'
import { UserManagement } from '@/features/super-admin/ui/user-management'

export default function AdminPage() {
  return (
    <div className="flex flex-col">
      <PlanManagement />
      <CouponManagement />
      <UserManagement />
    </div>
  )
}

'use client'

import { PlanManagement } from '@/features/super-admin/ui/plan-management'
import { UserApprovals } from '@/features/super-admin/ui/user-approvals'
import { useCallback, useState } from 'react'

export default function AdminPage() {
  const [planRefreshKey, setPlanRefreshKey] = useState(0)

  const handleUserApproved = useCallback(() => {
    setPlanRefreshKey((prev) => prev + 1)
  }, [])

  return (
    <div className="flex flex-col">
      <UserApprovals onUserApproved={handleUserApproved} />
      <PlanManagement key={planRefreshKey} />
    </div>
  )
}

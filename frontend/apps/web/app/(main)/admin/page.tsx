import { PlanManagement } from '@/features/super-admin/ui/plan-management'
import { UserApprovals } from '@/features/super-admin/ui/user-approvals'

export default function AdminPage() {
  return (
    <div className="flex flex-col">
      <UserApprovals />
      <PlanManagement />
    </div>
  )
}

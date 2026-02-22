'use client'

import { type PendingUser, SuperAdminApi } from '@workspace/services'
import { notification } from '@/lib/notification'
import {
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@workspace/ui/components/atoms'
import { Loader2 } from 'lucide-react'
import { useCallback, useEffect, useState, useTransition } from 'react'

interface UserApprovalsProps {
  onUserApproved?: () => void
}

export function UserApprovals({ onUserApproved }: UserApprovalsProps) {
  const [isPending, startTransition] = useTransition()
  const [users, setUsers] = useState<PendingUser[]>([])
  const [error, setError] = useState<string | null>(null)

  const loadData = useCallback(async () => {
    try {
      setError(null)
      const data = await SuperAdminApi.listPendingUsers()
      setUsers(data)
    } catch {
      setError('You need SUPER_ADMIN role to manage user approvals.')
    }
  }, [])

  useEffect(() => {
    loadData()
  }, [
    loadData,
  ])

  const handleApprove = (requestId: string) => {
    startTransition(async () => {
      try {
        await SuperAdminApi.approveUser(requestId)
        await loadData()
        onUserApproved?.()
        notification.success({ title: 'User approved' })
      } catch {
        setError('Failed to approve user.')
        notification.error({ title: 'Failed to approve user', description: 'Please try again.' })
      }
    })
  }

  const handleReject = (requestId: string) => {
    startTransition(async () => {
      try {
        await SuperAdminApi.rejectUser(requestId)
        await loadData()
        notification.success({ title: 'User rejected' })
      } catch {
        setError('Failed to reject user.')
        notification.error({ title: 'Failed to reject user', description: 'Please try again.' })
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
            <h1 className="text-2xl font-semibold text-foreground">User Approvals</h1>
            <p className="text-sm text-muted-foreground">
              Review and approve pending user registrations.
            </p>
          </div>
          {isPending && <Loader2 className="w-5 h-5 animate-spin text-muted-foreground" />}
        </div>

        {users.length === 0 ? (
          <Card>
            <CardHeader>
              <CardDescription>No pending registrations</CardDescription>
            </CardHeader>
          </Card>
        ) : (
          <div className="grid grid-cols-1 gap-4">
            {users.map((user) => (
              <Card key={user.request_id}>
                <CardHeader>
                  <CardTitle className="text-base">
                    {user.first_name} {user.last_name}
                  </CardTitle>
                  <CardDescription>{user.email}</CardDescription>
                </CardHeader>
                <CardContent className="flex flex-col md:flex-row md:items-center gap-4">
                  <Badge variant="outline">{user.organization_name}</Badge>
                  <Badge variant="outline">{user.organization_subdomain}</Badge>
                  <span className="text-xs text-muted-foreground">
                    {new Date(user.created_at).toLocaleDateString()}
                  </span>
                  <div className="flex gap-2 md:ml-auto">
                    <Button
                      variant="default"
                      size="sm"
                      onClick={() => handleApprove(user.request_id)}
                      disabled={isPending}
                    >
                      Approve
                    </Button>
                    <Button
                      variant="destructive"
                      size="sm"
                      onClick={() => handleReject(user.request_id)}
                      disabled={isPending}
                    >
                      Reject
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}

        <Button variant="outline" onClick={loadData} disabled={isPending}>
          Refresh
        </Button>
      </div>
    </div>
  )
}

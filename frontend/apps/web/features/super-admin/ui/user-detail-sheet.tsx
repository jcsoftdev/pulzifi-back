'use client'

import { type AdminMembership, type AdminUserDetail, SuperAdminApi } from '@workspace/services'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  Badge,
  Button,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@workspace/ui/components/atoms'
import { Loader2, Lock } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'
import { notification } from '@/lib/notification'

interface Props {
  userId: string
  open: boolean
  onClose: () => void
  onChanged: () => void
}

export function UserDetailSheet({ userId, open, onClose, onChanged }: Props) {
  const [detail, setDetail] = useState<AdminUserDetail | null>(null)
  const [loading, setLoading] = useState(false)
  const [busy, setBusy] = useState(false)
  const [confirmSuspend, setConfirmSuspend] = useState(false)
  const [confirmRemove, setConfirmRemove] = useState<string | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      setDetail(await SuperAdminApi.getUser(userId))
    } catch {
      notification.error({ title: 'Failed to load user' })
    } finally {
      setLoading(false)
    }
  }, [userId])

  useEffect(() => {
    if (open) load()
  }, [open, load])

  const toggleStatus = async () => {
    if (!detail) return
    const next = detail.status === 'suspended' ? 'approved' : 'suspended'
    setBusy(true)
    try {
      await SuperAdminApi.setUserStatus(userId, next)
      notification.success({ title: next === 'suspended' ? 'User suspended' : 'User activated' })
      await load()
      onChanged()
    } catch (err) {
      notification.error({
        title: 'Action failed',
        description: err instanceof Error ? err.message : undefined,
      })
    } finally {
      setBusy(false)
      setConfirmSuspend(false)
    }
  }

  const promoteUser = async () => {
    setBusy(true)
    try {
      await SuperAdminApi.promoteUser(userId)
      notification.success({ title: 'Promoted to Super Admin' })
      await load()
      onChanged()
    } catch (err) {
      notification.error({
        title: 'Promotion failed',
        description: err instanceof Error ? err.message : undefined,
      })
    } finally {
      setBusy(false)
    }
  }

  const changeRole = async (orgId: string, role: string) => {
    setBusy(true)
    try {
      await SuperAdminApi.setMembershipRole(userId, orgId, role as 'OWNER' | 'ADMIN' | 'MEMBER')
      notification.success({ title: 'Role updated' })
      await load()
      onChanged()
    } catch (err) {
      notification.error({
        title: 'Role change failed',
        description: err instanceof Error ? err.message : undefined,
      })
    } finally {
      setBusy(false)
    }
  }

  const removeMembership = async (orgId: string) => {
    setBusy(true)
    try {
      await SuperAdminApi.removeMembership(userId, orgId)
      notification.success({ title: 'Removed from organization' })
      await load()
      onChanged()
    } catch (err) {
      notification.error({
        title: 'Remove failed',
        description: err instanceof Error ? err.message : undefined,
      })
    } finally {
      setBusy(false)
      setConfirmRemove(null)
    }
  }

  return (
    <Sheet open={open} onOpenChange={(o) => !o && onClose()}>
      <SheetContent className="w-full sm:max-w-lg overflow-y-auto">
        <SheetHeader>
          <SheetTitle>User detail</SheetTitle>
          <SheetDescription>Review and manage this user.</SheetDescription>
        </SheetHeader>

        {loading || !detail ? (
          <div className="flex justify-center py-12">
            <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
          </div>
        ) : (
          <div className="flex flex-col gap-6 py-4">
            {/* Profile */}
            <div className="space-y-1">
              <p className="text-lg font-semibold">
                {detail.firstName} {detail.lastName}
              </p>
              <p className="text-sm text-muted-foreground">{detail.email}</p>
              <div className="flex items-center gap-2 pt-2">
                <Badge variant={detail.status === 'suspended' ? 'destructive' : 'secondary'}>
                  {detail.status}
                </Badge>
                {detail.emailVerified && <Badge variant="outline">Email verified</Badge>}
                {detail.isSuperAdmin && <Badge>Super Admin</Badge>}
              </div>
            </div>

            {/* Actions */}
            <div className="flex flex-col gap-2">
              <Button
                variant={detail.status === 'suspended' ? 'default' : 'destructive'}
                disabled={busy}
                onClick={() =>
                  detail.status === 'suspended' ? toggleStatus() : setConfirmSuspend(true)
                }
              >
                {detail.status === 'suspended' ? 'Activate account' : 'Suspend account'}
              </Button>

              {!detail.isSuperAdmin && (
                <Button variant="outline" disabled={busy} onClick={promoteUser}>
                  Promote to Super Admin
                </Button>
              )}
            </div>

            {/* Memberships */}
            <div className="space-y-3">
              <p className="text-sm font-medium">
                Organizations ({detail.memberships.length})
              </p>
              {detail.memberships.length === 0 ? (
                <p className="text-sm text-muted-foreground">No organization memberships.</p>
              ) : (
                detail.memberships.map((m: AdminMembership) => (
                  <div
                    key={m.orgId}
                    className="flex items-center gap-2 border border-border rounded-md p-3"
                  >
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium truncate">{m.orgName}</p>
                      <p className="text-xs text-muted-foreground">{m.subdomain}</p>
                    </div>
                    {m.isOwner ? (
                      <Badge variant="outline" className="gap-1">
                        <Lock className="w-3 h-3" /> Owner
                      </Badge>
                    ) : (
                      <>
                        <Select
                          value={m.role}
                          disabled={busy}
                          onValueChange={(v) => changeRole(m.orgId, v)}
                        >
                          <SelectTrigger className="w-[120px]">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="ADMIN">Admin</SelectItem>
                            <SelectItem value="MEMBER">Member</SelectItem>
                          </SelectContent>
                        </Select>
                        <Button
                          variant="destructive"
                          size="sm"
                          disabled={busy}
                          onClick={() => setConfirmRemove(m.orgId)}
                        >
                          Remove
                        </Button>
                      </>
                    )}
                  </div>
                ))
              )}
            </div>
          </div>
        )}
      </SheetContent>

      {/* Confirm suspend */}
      <AlertDialog open={confirmSuspend} onOpenChange={setConfirmSuspend}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Suspend this user?</AlertDialogTitle>
            <AlertDialogDescription>
              They will be blocked from logging in until reactivated.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={toggleStatus}>Suspend</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Confirm remove membership */}
      <AlertDialog
        open={!!confirmRemove}
        onOpenChange={(o) => !o && setConfirmRemove(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove from organization?</AlertDialogTitle>
            <AlertDialogDescription>
              The user will lose access to this organization.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => confirmRemove && removeMembership(confirmRemove)}
            >
              Remove
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </Sheet>
  )
}

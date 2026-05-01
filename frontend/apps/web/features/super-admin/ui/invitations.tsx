'use client'

import {
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@workspace/ui/components/atoms'
import { AlertTriangle, Loader2, Mail, Plus, RefreshCw, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { notification } from '@/lib/notification'
import { useInvitations } from '../application/use-invitations'
import type { InvitationStatus } from '../domain/types'
import { InvitePlatformDialog } from './invite-platform-dialog'

function formatDate(value: string): string {
  try {
    return new Date(value).toLocaleString()
  } catch {
    return value
  }
}

function StatusBadge({ status }: { status: InvitationStatus }) {
  switch (status) {
    case 'pending':
      return <Badge variant="outline">Pending</Badge>
    case 'accepted':
      return <Badge variant="default">Accepted</Badge>
    case 'revoked':
      return <Badge variant="secondary">Revoked</Badge>
    case 'expired':
      return <Badge variant="secondary">Expired</Badge>
    default:
      return <Badge variant="outline">{status}</Badge>
  }
}

function DeliveryBadge({
  emailSentAt,
  emailError,
}: {
  emailSentAt: string | null
  emailError: string | null
}) {
  if (emailError) {
    return (
      <Badge variant="destructive" title={emailError}>
        <AlertTriangle className="w-3 h-3" />
        Failed
      </Badge>
    )
  }
  if (emailSentAt) {
    return (
      <Badge variant="outline">
        <Mail className="w-3 h-3" />
        Sent
      </Badge>
    )
  }
  return <Badge variant="outline">Queued</Badge>
}

export function Invitations() {
  const { invitations, loading, error, create, revoke, resend, refresh } = useInvitations()
  const [inviteOpen, setInviteOpen] = useState(false)
  const [actionError, setActionError] = useState<string | null>(null)
  const [busyId, setBusyId] = useState<string | null>(null)
  const [createError, setCreateError] = useState<string | null>(null)
  const [creating, setCreating] = useState(false)

  const handleCreate = async (email: string) => {
    setCreateError(null)
    setCreating(true)
    try {
      const res = await create(email)
      setInviteOpen(false)
      if (res.emailDelivery === 'failed') {
        notification.success({
          title: 'Invitation created',
          description: 'But the email delivery failed. You can resend from the list.',
        })
      } else {
        notification.success({
          title: 'Invitation sent',
          description: `An invite was emailed to ${res.email}.`,
        })
      }
    } catch (err: unknown) {
      const axiosErr = err as {
        response?: {
          status?: number
          data?: {
            error?: string
          }
        }
      }
      const code = axiosErr?.response?.data?.error
      let message = 'Failed to create invitation.'
      if (code === 'cannot_invite_email') {
        message = 'This email is already registered or already has a pending invitation.'
      } else if (code === 'daily_cap_exceeded') {
        message = 'Daily invitation limit reached. Try again tomorrow.'
      } else if (axiosErr?.response?.status === 403) {
        message = 'You do not have permission to create invitations.'
      }
      setCreateError(message)
      notification.error({
        title: 'Failed to create invitation',
        description: message,
      })
    } finally {
      setCreating(false)
    }
  }

  const handleResend = async (id: string, email: string) => {
    setActionError(null)
    setBusyId(id)
    try {
      const res = await resend(id)
      if (res.emailDelivery === 'failed') {
        notification.error({
          title: 'Resend attempted',
          description: 'A new token was issued, but the email failed to send.',
        })
      } else {
        notification.success({
          title: 'Invitation resent',
          description: `A fresh email was sent to ${email}.`,
        })
      }
    } catch {
      const message = 'Failed to resend invitation.'
      setActionError(message)
      notification.error({
        title: message,
        description: 'Please try again.',
      })
    } finally {
      setBusyId(null)
    }
  }

  const handleRevoke = async (id: string) => {
    setActionError(null)
    setBusyId(id)
    try {
      await revoke(id)
      notification.success({
        title: 'Invitation revoked',
      })
    } catch {
      const message = 'Failed to revoke invitation.'
      setActionError(message)
      notification.error({
        title: message,
        description: 'Please try again.',
      })
    } finally {
      setBusyId(null)
    }
  }

  if (error) {
    return (
      <div className="flex-1 p-8 max-w-7xl mx-auto w-full">
        <Card>
          <CardHeader>
            <CardTitle>Invitations</CardTitle>
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
            <h1 className="text-2xl font-semibold text-foreground">Invitations</h1>
            <p className="text-sm text-muted-foreground">
              Invite people directly to the platform. They sign up with their own organization.
            </p>
          </div>
          <div className="flex items-center gap-3">
            {loading && <Loader2 className="w-5 h-5 animate-spin text-muted-foreground" />}
            <Button
              onClick={() => {
                setCreateError(null)
                setInviteOpen(true)
              }}
            >
              <Plus className="w-4 h-4 mr-2" />
              Invite to platform
            </Button>
          </div>
        </div>

        {actionError && (
          <div className="text-sm text-destructive bg-destructive/10 border border-destructive/20 rounded-md px-4 py-3">
            {actionError}
          </div>
        )}

        {invitations.length === 0 && !loading ? (
          <Card>
            <CardHeader>
              <CardDescription>No invitations yet</CardDescription>
            </CardHeader>
          </Card>
        ) : (
          <Card>
            <CardContent className="p-0">
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b text-left text-xs uppercase tracking-wide text-muted-foreground">
                      <th className="px-4 py-3 font-medium">Email</th>
                      <th className="px-4 py-3 font-medium">Status</th>
                      <th className="px-4 py-3 font-medium">Delivery</th>
                      <th className="px-4 py-3 font-medium">Expires</th>
                      <th className="px-4 py-3 font-medium text-right">Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {invitations.map((inv) => {
                      const isPending = inv.status === 'pending'
                      const canResend = isPending
                      const canRevoke = isPending
                      const rowBusy = busyId === inv.id
                      return (
                        <tr key={inv.id} className="border-b last:border-b-0">
                          <td className="px-4 py-3 align-middle">
                            <div className="font-medium text-foreground">{inv.email}</div>
                            {inv.resentCount > 0 && (
                              <div className="text-xs text-muted-foreground">
                                Resent {inv.resentCount} time
                                {inv.resentCount === 1 ? '' : 's'}
                              </div>
                            )}
                          </td>
                          <td className="px-4 py-3 align-middle">
                            <StatusBadge status={inv.status} />
                          </td>
                          <td className="px-4 py-3 align-middle">
                            <DeliveryBadge
                              emailSentAt={inv.emailSentAt}
                              emailError={inv.emailError}
                            />
                          </td>
                          <td className="px-4 py-3 align-middle text-muted-foreground">
                            {formatDate(inv.expiresAt)}
                          </td>
                          <td className="px-4 py-3 align-middle">
                            <div className="flex justify-end gap-2">
                              {canResend && (
                                <Button
                                  variant="outline"
                                  size="sm"
                                  disabled={rowBusy}
                                  onClick={() => handleResend(inv.id, inv.email)}
                                >
                                  <RefreshCw className="w-4 h-4 mr-1" />
                                  Resend
                                </Button>
                              )}
                              {canRevoke && (
                                <Button
                                  variant="destructive"
                                  size="sm"
                                  disabled={rowBusy}
                                  onClick={() => handleRevoke(inv.id)}
                                >
                                  <Trash2 className="w-4 h-4 mr-1" />
                                  Revoke
                                </Button>
                              )}
                            </div>
                          </td>
                        </tr>
                      )
                    })}
                  </tbody>
                </table>
              </div>
            </CardContent>
          </Card>
        )}

        <Button variant="outline" onClick={refresh} disabled={loading}>
          Refresh
        </Button>
      </div>

      <InvitePlatformDialog
        open={inviteOpen}
        onOpenChange={(o) => {
          setInviteOpen(o)
          if (!o) setCreateError(null)
        }}
        onSubmit={handleCreate}
        isLoading={creating}
        error={createError}
      />
    </div>
  )
}

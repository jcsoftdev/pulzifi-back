'use client'

import { Plus } from 'lucide-react'
import { useState } from 'react'
import { notification } from '@/lib/notification'
import { useTeam } from './application/use-team'
import type { TeamMember } from './domain/types'
import { DeleteMemberDialog } from './ui/delete-member-dialog'
import { EditMemberDialog } from './ui/edit-member-dialog'
import { InviteMemberDialog } from './ui/invite-member-dialog'
import { MemberCard } from './ui/member-card'

interface TeamFeatureProps {
  currentUserId?: string
}

export function TeamFeature({ currentUserId }: Readonly<TeamFeatureProps>) {
  const { members, loading, inviteMember, updateMember, removeMember } = useTeam()

  const [inviteOpen, setInviteOpen] = useState(false)
  const [editingMember, setEditingMember] = useState<TeamMember | null>(null)
  const [deletingMember, setDeletingMember] = useState<TeamMember | null>(null)
  const [actionError, setActionError] = useState<Error | null>(null)
  const [actionLoading, setActionLoading] = useState(false)

  const handleInvite = async (email: string, role: string) => {
    setActionError(null)
    setActionLoading(true)
    try {
      await inviteMember(email, role)
      setInviteOpen(false)
      notification.success({ title: 'Invitation sent', description: `An invite has been sent to ${email}.` })
    } catch (err) {
      setActionError(err instanceof Error ? err : new Error('Failed to invite member'))
      notification.error({ title: 'Failed to send invitation', description: err instanceof Error ? err.message : 'Please try again.' })
    } finally {
      setActionLoading(false)
    }
  }

  const handleEdit = async (memberId: string, role: string) => {
    setActionError(null)
    setActionLoading(true)
    try {
      await updateMember(memberId, role)
      setEditingMember(null)
      notification.success({ title: 'Member role updated' })
    } catch (err) {
      setActionError(err instanceof Error ? err : new Error('Failed to update member'))
      notification.error({ title: 'Failed to update member', description: err instanceof Error ? err.message : 'Please try again.' })
    } finally {
      setActionLoading(false)
    }
  }

  const handleDelete = async () => {
    if (!deletingMember) return
    setActionError(null)
    setActionLoading(true)
    try {
      await removeMember(deletingMember.id)
      setDeletingMember(null)
      notification.success({ title: 'Member removed' })
    } catch (err) {
      setActionError(err instanceof Error ? err : new Error('Failed to remove member'))
      notification.error({ title: 'Failed to remove member', description: err instanceof Error ? err.message : 'Please try again.' })
    } finally {
      setActionLoading(false)
    }
  }

  return (
    <div className="px-4 md:px-8 lg:px-24 py-8">
      {/* Page header */}
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-foreground">Team</h1>
        <p className="text-sm text-muted-foreground mt-1">Edit your team members</p>
      </div>

      {/* Members grid */}
      <div className="flex flex-wrap gap-8 items-start">
        {/* Invite button */}
        <button
          type="button"
          onClick={() => {
            setActionError(null)
            setInviteOpen(true)
          }}
          className="flex flex-col items-center gap-2 w-20 group"
        >
          <div className="w-16 h-16 rounded-full border-2 border-dashed border-border flex items-center justify-center group-hover:border-foreground transition-colors">
            <Plus className="w-6 h-6 text-muted-foreground group-hover:text-foreground transition-colors" />
          </div>
          <span className="text-xs text-muted-foreground group-hover:text-foreground transition-colors">
            Invite member
          </span>
        </button>

        {/* Member cards */}
        {loading ? (
          <>
            {[1, 2, 3].map((n) => (
              <div key={n} className="flex flex-col items-center gap-2 w-20">
                <div className="w-16 h-16 rounded-full bg-muted animate-pulse" />
                <div className="h-3 w-14 bg-muted animate-pulse rounded" />
              </div>
            ))}
          </>
        ) : (
          members.map((member) => (
            <MemberCard
              key={member.id}
              member={member}
              isCurrentUser={member.userId === currentUserId}
              onEdit={(m) => {
                setActionError(null)
                setEditingMember(m)
              }}
              onDelete={(m) => {
                setActionError(null)
                setDeletingMember(m)
              }}
            />
          ))
        )}
      </div>

      {/* Dialogs */}
      <InviteMemberDialog
        open={inviteOpen}
        onOpenChange={setInviteOpen}
        onSubmit={handleInvite}
        isLoading={actionLoading}
        error={actionError}
      />

      <EditMemberDialog
        open={!!editingMember}
        onOpenChange={(open) => {
          if (!open) setEditingMember(null)
        }}
        member={editingMember}
        onSubmit={handleEdit}
        isLoading={actionLoading}
        error={actionError}
      />

      <DeleteMemberDialog
        open={!!deletingMember}
        onOpenChange={(open) => {
          if (!open) setDeletingMember(null)
        }}
        member={deletingMember}
        onConfirm={handleDelete}
        isLoading={actionLoading}
      />
    </div>
  )
}

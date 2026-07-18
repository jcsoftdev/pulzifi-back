'use client'

import { notix } from '@workspace/notix'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@workspace/ui/components/atoms/alert-dialog'
import { Button } from '@workspace/ui/components/atoms/button'
import { SquarePlus } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useSocialProfiles } from '../application/use-social-profiles'
import { useSocialQuota } from '../application/use-social-quota'
import { useSocialWorkspaceChanges } from '../application/use-social-workspace-changes'
import type { CreateSocialProfileDto, SocialProfile } from '../domain/types'
import { AddSocialProfileDialog } from './add-social-profile-dialog'
import { SocialChangesSection } from './changes/social-changes-section'
import { SocialProfileGrid } from './social-profile-grid'
import { SocialQuotaBadge } from './social-quota-badge'

interface SocialTabContentProps {
  workspaceId: string
}

export function SocialTabContent({ workspaceId }: Readonly<SocialTabContentProps>) {
  const { profiles, isLoading: profilesLoading, fetchProfiles, createProfile, deleteProfile } =
    useSocialProfiles(workspaceId)
  const { quota, fetchQuota } = useSocialQuota(workspaceId)
  const { changes, isLoading: changesLoading } = useSocialWorkspaceChanges(workspaceId)
  const [isAddOpen, setIsAddOpen] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [profileToDelete, setProfileToDelete] = useState<SocialProfile | null>(null)
  const [isDeleting, setIsDeleting] = useState(false)

  useEffect(() => {
    fetchProfiles()
    fetchQuota()
  }, [fetchProfiles, fetchQuota])

  const handleAddProfile = async (data: CreateSocialProfileDto) => {
    setIsSubmitting(true)
    try {
      await createProfile(data)
      setIsAddOpen(false)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleConfirmDelete = async () => {
    if (!profileToDelete) return
    setIsDeleting(true)
    try {
      await deleteProfile(profileToDelete.id)
      await fetchQuota() // deleting a profile frees a monitored-profile slot
      notix.success({
        title: 'Profile removed',
        description: `@${profileToDelete.handle} is no longer being tracked.`,
      })
      setProfileToDelete(null)
    } catch {
      notix.error({
        title: 'Could not remove profile',
        description: 'Something went wrong. Please try again.',
      })
    } finally {
      setIsDeleting(false)
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between gap-4">
        <SocialQuotaBadge quota={quota} />
        <Button
          variant="default"
          size="sm"
          className="gap-2"
          onClick={() => setIsAddOpen(true)}
        >
          <SquarePlus className="w-4 h-4" />
          Add Profile
        </Button>
      </div>

      {profilesLoading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-24 bg-muted animate-pulse rounded-lg" />
          ))}
        </div>
      ) : (
        <SocialProfileGrid
          profiles={profiles}
          workspaceId={workspaceId}
          onDelete={setProfileToDelete}
        />
      )}

      <SocialChangesSection changes={changes} profiles={profiles} isLoading={changesLoading} />

      <AddSocialProfileDialog
        open={isAddOpen}
        onOpenChange={setIsAddOpen}
        onSubmit={handleAddProfile}
        workspaceId={workspaceId}
        isLoading={isSubmitting}
      />

      <AlertDialog
        open={profileToDelete !== null}
        onOpenChange={(open) => !open && setProfileToDelete(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove social profile</AlertDialogTitle>
            <AlertDialogDescription>
              {profileToDelete
                ? `Are you sure you want to remove @${profileToDelete.handle}? This deletes the profile and all its snapshots and changes. This action cannot be undone.`
                : ''}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeleting}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={(e) => {
                e.preventDefault()
                handleConfirmDelete()
              }}
              disabled={isDeleting}
              className="bg-destructive text-white hover:bg-destructive/90"
            >
              {isDeleting ? 'Removing…' : 'Remove profile'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}

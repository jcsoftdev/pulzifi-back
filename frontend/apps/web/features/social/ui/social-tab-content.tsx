'use client'

import { Button } from '@workspace/ui/components/atoms/button'
import { SquarePlus } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useSocialProfiles } from '../application/use-social-profiles'
import { useSocialQuota } from '../application/use-social-quota'
import { useSocialWorkspaceChanges } from '../application/use-social-workspace-changes'
import type { CreateSocialProfileDto } from '../domain/types'
import { AddSocialProfileDialog } from './add-social-profile-dialog'
import { SocialChangesSection } from './changes/social-changes-section'
import { SocialProfileGrid } from './social-profile-grid'
import { SocialQuotaBadge } from './social-quota-badge'

interface SocialTabContentProps {
  workspaceId: string
}

export function SocialTabContent({ workspaceId }: Readonly<SocialTabContentProps>) {
  const { profiles, isLoading: profilesLoading, fetchProfiles, createProfile } =
    useSocialProfiles(workspaceId)
  const { quota, fetchQuota } = useSocialQuota(workspaceId)
  const { changes, isLoading: changesLoading } = useSocialWorkspaceChanges(workspaceId)
  const [isAddOpen, setIsAddOpen] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)

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
        <SocialProfileGrid profiles={profiles} workspaceId={workspaceId} />
      )}

      <SocialChangesSection changes={changes} profiles={profiles} isLoading={changesLoading} />

      <AddSocialProfileDialog
        open={isAddOpen}
        onOpenChange={setIsAddOpen}
        onSubmit={handleAddProfile}
        workspaceId={workspaceId}
        isLoading={isSubmitting}
      />
    </div>
  )
}

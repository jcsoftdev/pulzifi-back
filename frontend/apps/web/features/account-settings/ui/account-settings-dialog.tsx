'use client'

import { Avatar, AvatarFallback, AvatarImage } from '@workspace/ui/components/atoms/avatar'
import { Button } from '@workspace/ui/components/atoms/button'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@workspace/ui/components/atoms/dialog'
import { Input } from '@workspace/ui/components/atoms/input'
import { Label } from '@workspace/ui/components/atoms/label'
import { InfoIcon, LockIcon, MailIcon, UserIcon } from 'lucide-react'
import { useEffect, useId, useState } from 'react'
import type { User } from '@workspace/services'
import { notification } from '../../../lib/notification'
import { useAccountSettings } from '../application/use-account-settings'

interface AccountSettingsDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  user: User
}

export function AccountSettingsDialog({
  open,
  onOpenChange,
  user,
}: Readonly<AccountSettingsDialogProps>) {
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')

  const uid = useId()
  const { isSubmitting, profileError, passwordError, updateProfile, changePassword } =
    useAccountSettings()

  useEffect(() => {
    if (open) {
      const parts = user.name.split(' ')
      setFirstName(parts[0] ?? '')
      setLastName(parts.slice(1).join(' ') ?? '')
      setCurrentPassword('')
      setNewPassword('')
      setConfirmPassword('')
    }
  }, [open, user.name])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    const hasProfileChanges =
      firstName !== (user.name.split(' ')[0] ?? '') ||
      lastName !== (user.name.split(' ').slice(1).join(' ') ?? '')
    const hasPasswordChanges = currentPassword || newPassword || confirmPassword

    try {
      if (hasProfileChanges) {
        await updateProfile({ firstName, lastName })
      }
      if (hasPasswordChanges) {
        await changePassword({ currentPassword, newPassword, confirmPassword })
      }

      notification.success({
        title: 'Settings saved',
        description: 'Your account has been updated.',
      })
      onOpenChange(false)
    } catch {
      // errors are surfaced via profileError / passwordError
    }
  }

  const busy = isSubmitting

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[512px] p-0 gap-0 overflow-hidden">
        {/* Header */}
        <DialogHeader className="px-6 py-5 border-b border-border">
          <DialogTitle className="text-xl font-bold text-foreground">Account Settings</DialogTitle>
          <p className="text-sm text-muted-foreground mt-0.5">
            Manage your profile and security preferences.
          </p>
        </DialogHeader>

        <form onSubmit={handleSubmit}>
          <div className="overflow-y-auto max-h-[600px]">
            {/* Profile photo */}
            <div className="flex flex-col items-center pt-6 pb-4">
              <div className="relative">
                <div className="border-4 border-white rounded-full shadow-md">
                  <Avatar className="w-20 h-20 rounded-full">
                    {user.avatar && <AvatarImage src={user.avatar} alt={user.name} />}
                    <AvatarFallback className="text-xl rounded-full bg-violet-100 text-violet-700">
                      {(firstName[0] ?? user.name[0] ?? '').toUpperCase()}
                    </AvatarFallback>
                  </Avatar>
                </div>
              </div>
              <div className="mt-3 text-center">
                <p className="text-base font-semibold text-foreground">{user.name}</p>
                <p className="text-xs text-muted-foreground">{user.email}</p>
              </div>
            </div>

            {/* Tabs */}
            <div className="px-6 border-b border-border">
              <div className="inline-flex items-center pb-2.5 border-b-2 border-violet-600">
                <span className="text-sm font-medium text-violet-600">Profile Details</span>
              </div>
            </div>

            {/* Form body */}
            <div className="px-6 py-6 flex flex-col gap-6">
              {/* First & Last name */}
              <div className="flex gap-4">
                <div className="flex-1 flex flex-col gap-1.5">
                  <Label htmlFor={`${uid}-firstName`} className="text-xs font-semibold text-gray-700">
                    First Name
                  </Label>
                  <div className="relative">
                    <UserIcon className="absolute left-3 top-1/2 -translate-y-1/2 h-3 w-3 text-muted-foreground" />
                    <Input
                      id={`${uid}-firstName`}
                      value={firstName}
                      onChange={(e) => setFirstName(e.target.value)}
                      className="pl-8"
                      placeholder="First name"
                      disabled={busy}
                    />
                  </div>
                </div>
                <div className="flex-1 flex flex-col gap-1.5">
                  <Label htmlFor={`${uid}-lastName`} className="text-xs font-semibold text-gray-700">
                    Last Name
                  </Label>
                  <div className="relative">
                    <UserIcon className="absolute left-3 top-1/2 -translate-y-1/2 h-3 w-3 text-muted-foreground" />
                    <Input
                      id={`${uid}-lastName`}
                      value={lastName}
                      onChange={(e) => setLastName(e.target.value)}
                      className="pl-8"
                      placeholder="Last name"
                      disabled={busy}
                    />
                  </div>
                </div>
              </div>

              {profileError && (
                <p className="text-xs text-destructive">{profileError}</p>
              )}

              {/* Email */}
              <div className="flex flex-col gap-1.5">
                <Label className="text-xs font-semibold text-gray-700">Email Address</Label>
                <div className="relative">
                  <MailIcon className="absolute left-3 top-1/2 -translate-y-1/2 h-3 w-3 text-muted-foreground" />
                  <Input
                    value={user.email}
                    className="pl-8 pr-24 bg-muted text-muted-foreground"
                    disabled
                    readOnly
                  />
                  <div className="absolute right-3 top-1/2 -translate-y-1/2">
                    <span className="text-[10px] text-muted-foreground bg-muted border border-border rounded px-1.5 py-0.5">
                      Managed
                    </span>
                  </div>
                </div>
              </div>

              {/* Change password */}
              <div className="border-t border-gray-100 pt-4 flex flex-col gap-4">
                <div className="flex items-center gap-2">
                  <LockIcon className="h-3 w-3 text-foreground" />
                  <span className="text-sm font-bold text-foreground">Change Password</span>
                </div>

                <div className="flex flex-col gap-1.5">
                  <Label htmlFor={`${uid}-currentPassword`} className="text-xs font-semibold text-gray-700">
                    Current Password
                  </Label>
                  <Input
                    id={`${uid}-currentPassword`}
                    type="password"
                    value={currentPassword}
                    onChange={(e) => setCurrentPassword(e.target.value)}
                    placeholder="Enter current password"
                    disabled={busy}
                  />
                </div>

                <div className="flex gap-4">
                  <div className="flex-1 flex flex-col gap-1.5">
                    <Label htmlFor={`${uid}-newPassword`} className="text-xs font-semibold text-gray-700">
                      New Password
                    </Label>
                    <Input
                      id={`${uid}-newPassword`}
                      type="password"
                      value={newPassword}
                      onChange={(e) => setNewPassword(e.target.value)}
                      placeholder="New password"
                      disabled={busy}
                    />
                  </div>
                  <div className="flex-1 flex flex-col gap-1.5">
                    <Label
                      htmlFor={`${uid}-confirmPassword`}
                      className="text-xs font-semibold text-gray-700"
                    >
                      Confirm Password
                    </Label>
                    <Input
                      id={`${uid}-confirmPassword`}
                      type="password"
                      value={confirmPassword}
                      onChange={(e) => setConfirmPassword(e.target.value)}
                      placeholder="Confirm password"
                      disabled={busy}
                    />
                  </div>
                </div>

                {passwordError && (
                  <p className="text-xs text-destructive">{passwordError}</p>
                )}

                <div className="flex gap-2 items-start bg-blue-50 rounded-md px-3 py-2">
                  <InfoIcon className="h-3 w-3 text-blue-700 mt-0.5 shrink-0" />
                  <p className="text-xs text-blue-700 leading-snug">
                    Password must be at least 8 characters long and include a number and a special
                    character.
                  </p>
                </div>
              </div>
            </div>
          </div>

          {/* Footer */}
          <DialogFooter className="px-6 py-4 bg-muted/50 border-t border-border">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={busy}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={busy} className="bg-violet-600 hover:bg-violet-700">
              {busy ? 'Saving...' : 'Save Changes'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

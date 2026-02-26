'use client'

import { AuthApi } from '@workspace/services'
import { useState } from 'react'

export interface ProfileFormData {
  firstName: string
  lastName: string
}

export interface PasswordFormData {
  currentPassword: string
  newPassword: string
  confirmPassword: string
}

export function useAccountSettings() {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [profileError, setProfileError] = useState<string | null>(null)
  const [passwordError, setPasswordError] = useState<string | null>(null)

  const updateProfile = async (data: ProfileFormData) => {
    setProfileError(null)
    setIsSubmitting(true)
    try {
      await AuthApi.updateProfile({ firstName: data.firstName, lastName: data.lastName })
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to update profile'
      setProfileError(message)
      throw err
    } finally {
      setIsSubmitting(false)
    }
  }

  const changePassword = async (data: PasswordFormData) => {
    setPasswordError(null)
    if (data.newPassword !== data.confirmPassword) {
      setPasswordError('Passwords do not match')
      throw new Error('Passwords do not match')
    }
    setIsSubmitting(true)
    try {
      await AuthApi.changePassword({
        currentPassword: data.currentPassword,
        newPassword: data.newPassword,
      })
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to update password'
      setPasswordError(message)
      throw err
    } finally {
      setIsSubmitting(false)
    }
  }

  return {
    isSubmitting,
    profileError,
    passwordError,
    updateProfile,
    changePassword,
  }
}

'use client'

import { useId, useState } from 'react'
import { Button } from '@workspace/ui/components/atoms/button'
import { Input } from '@workspace/ui/components/atoms/input'
import { Label } from '@workspace/ui/components/atoms/label'
import type { RegisterData } from '../domain/types'

export interface RegisterFormProps {
  onSubmit: (data: RegisterData) => Promise<void>
  isLoading?: boolean
  error?: string
}

export function RegisterForm({ onSubmit, isLoading = false, error }: Readonly<RegisterFormProps>) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const firstNameId = useId()
  const lastNameId = useId()
  const emailId = useId()
  const passwordId = useId()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    await onSubmit({
      email,
      password,
      firstName,
      lastName,
    })
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4 w-full">
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor={firstNameId}>First Name</Label>
          <Input
            id={firstNameId}
            type="text"
            value={firstName}
            onChange={(e) => setFirstName(e.target.value)}
            required
            placeholder="John"
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor={lastNameId}>Last Name</Label>
          <Input
            id={lastNameId}
            type="text"
            value={lastName}
            onChange={(e) => setLastName(e.target.value)}
            required
            placeholder="Doe"
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor={emailId}>Email</Label>
        <Input
          id={emailId}
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
          placeholder="you@example.com"
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor={passwordId}>Password</Label>
        <Input
          id={passwordId}
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
          minLength={8}
          placeholder="••••••••"
        />
        <p className="text-xs text-muted-foreground mt-1">Minimum 8 characters</p>
      </div>

      {error && (
        <div className="text-destructive text-sm bg-destructive/10 p-3 rounded-md border border-destructive/20">
          {error}
        </div>
      )}

      <Button type="submit" disabled={isLoading} className="w-full">
        {isLoading ? 'Creating account...' : 'Create account'}
      </Button>
    </form>
  )
}

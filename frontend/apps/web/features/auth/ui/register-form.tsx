'use client'

import { Button } from '@workspace/ui/components/atoms/button'
import { Input } from '@workspace/ui/components/atoms/input'
import { Label } from '@workspace/ui/components/atoms/label'
import { CheckCircle, Loader2, XCircle } from 'lucide-react'
import { useId, useState } from 'react'
import type { SubdomainStatus } from '../application/use-register'
import type { RegisterData } from '../domain/types'

export interface RegisterFormProps {
  onSubmit: (data: RegisterData) => Promise<void>
  isLoading?: boolean
  error?: string
  onSubdomainChange?: (subdomain: string) => void
  subdomainStatus?: SubdomainStatus
  subdomainMessage?: string
}

export function RegisterForm({
  onSubmit,
  isLoading = false,
  error,
  onSubdomainChange,
  subdomainStatus = 'idle',
  subdomainMessage,
}: Readonly<RegisterFormProps>) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const [organizationName, setOrganizationName] = useState('')
  const [organizationSubdomain, setOrganizationSubdomain] = useState('')
  const firstNameId = useId()
  const lastNameId = useId()
  const emailId = useId()
  const passwordId = useId()
  const orgNameId = useId()
  const subdomainId = useId()

  const handleSubdomainChange = (value: string) => {
    const normalized = value.toLowerCase()
    setOrganizationSubdomain(normalized)
    onSubdomainChange?.(normalized)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    await onSubmit({
      email,
      password,
      firstName,
      lastName,
      organizationName,
      organizationSubdomain,
    })
  }

  const isSubdomainUnavailable = subdomainStatus === 'unavailable'

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
        <Label htmlFor={orgNameId}>Organization Name</Label>
        <Input
          id={orgNameId}
          type="text"
          value={organizationName}
          onChange={(e) => setOrganizationName(e.target.value)}
          required
          placeholder="Acme Inc."
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor={subdomainId}>Subdomain</Label>
        <div className="relative">
          <Input
            id={subdomainId}
            type="text"
            value={organizationSubdomain}
            onChange={(e) => handleSubdomainChange(e.target.value)}
            required
            placeholder="your-company"
            className={
              isSubdomainUnavailable
                ? 'border-destructive pr-9 focus-visible:ring-destructive'
                : subdomainStatus === 'available'
                  ? 'border-green-500 pr-9 focus-visible:ring-green-500'
                  : 'pr-9'
            }
          />
          <div className="absolute right-3 top-1/2 -translate-y-1/2">
            {subdomainStatus === 'checking' && (
              <Loader2 className="w-4 h-4 animate-spin text-muted-foreground" />
            )}
            {subdomainStatus === 'available' && (
              <CheckCircle className="w-4 h-4 text-green-500" />
            )}
            {subdomainStatus === 'unavailable' && (
              <XCircle className="w-4 h-4 text-destructive" />
            )}
          </div>
        </div>
        {subdomainStatus === 'unavailable' && subdomainMessage ? (
          <p className="text-xs text-destructive mt-1">{subdomainMessage}</p>
        ) : subdomainStatus === 'available' ? (
          <p className="text-xs text-green-500 mt-1">Subdomain is available</p>
        ) : (
          <p className="text-xs text-muted-foreground mt-1">Lowercase letters, numbers, and hyphens only</p>
        )}
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

      <Button
        type="submit"
        disabled={isLoading || isSubdomainUnavailable || subdomainStatus === 'checking'}
        className="w-full"
      >
        {isLoading ? 'Creating account...' : 'Create account'}
      </Button>
    </form>
  )
}

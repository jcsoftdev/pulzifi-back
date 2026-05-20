'use client'

import { Button } from '@workspace/ui/components/atoms/button'
import { cn } from '@workspace/ui/lib/utils'
import { CheckCircle, Eye, EyeOff, Loader2, XCircle } from 'lucide-react'
import Link from 'next/link'
import { useId, useState } from 'react'
import type { SubdomainStatus } from '../application/use-register'
import type { RegisterData } from '../domain/types'
import { AuthLabel, ErrorBanner, FieldMessage } from './form-atoms'

function GoogleIcon() {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <path
        d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z"
        fill="#4285F4"
      />
      <path
        d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
        fill="#34A853"
      />
      <path
        d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18A10.96 10.96 0 0 0 1 12c0 1.77.42 3.45 1.18 4.93l3.66-2.84z"
        fill="#FBBC05"
      />
      <path
        d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
        fill="#EA4335"
      />
    </svg>
  )
}

const baseInput =
  'h-10 w-full rounded-xl border border-[var(--pz-ink)]/10 bg-[var(--pz-page-bg,#f9f9f9)] px-4 text-sm text-[var(--pz-ink)] outline-none transition-[border-color,box-shadow] placeholder:text-[var(--pz-ink)]/30 focus:border-[var(--pz-accent)]/40 focus:bg-white focus:ring-2 focus:ring-[var(--pz-accent)]/15'

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
  const [confirmPassword, setConfirmPassword] = useState('')
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const [organizationName, setOrganizationName] = useState('')
  const [organizationSubdomain, setOrganizationSubdomain] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  const [passwordError, setPasswordError] = useState<string>()

  const firstNameId = useId()
  const lastNameId = useId()
  const emailId = useId()
  const passwordId = useId()
  const confirmPasswordId = useId()
  const orgNameId = useId()
  const subdomainId = useId()
  const errorId = useId()
  const subdomainHintId = useId()
  const passwordErrorId = useId()

  const handleSubdomainChange = (value: string) => {
    const normalized = value.toLowerCase()
    setOrganizationSubdomain(normalized)
    onSubdomainChange?.(normalized)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setPasswordError(undefined)
    if (password !== confirmPassword) {
      setPasswordError('Passwords do not match')
      return
    }
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

  const subdomainHintVariant = isSubdomainUnavailable
    ? 'error'
    : subdomainStatus === 'available'
      ? 'success'
      : 'hint'

  const subdomainHintText = isSubdomainUnavailable
    ? (subdomainMessage ?? 'Subdomain unavailable')
    : subdomainStatus === 'available'
      ? 'Available'
      : 'Lowercase, numbers, hyphens'

  return (
    <div className="flex flex-col gap-4">
      {/* Google OAuth */}
      <a
        href="/api/v1/auth/oauth/google"
        aria-label="Continue with Google"
        className={cn(
          'flex h-10 items-center justify-center gap-3 rounded-xl border border-[var(--pz-ink)]/10 bg-white text-sm font-medium text-[var(--pz-ink)] shadow-sm transition-[box-shadow,background-color] hover:bg-gray-50 hover:shadow-md',
          isLoading && 'pointer-events-none opacity-50'
        )}
      >
        <GoogleIcon />
        Continue with Google
      </a>

      {/* Divider */}
      <div className="flex items-center gap-3" aria-hidden="true">
        <div className="h-px flex-1 bg-[var(--pz-ink)]/8" />
        <span className="text-xs text-[var(--pz-ink)]/35">or sign up with email</span>
        <div className="h-px flex-1 bg-[var(--pz-ink)]/8" />
      </div>

      {/* Form */}
      <form onSubmit={handleSubmit} className="flex flex-col gap-3" aria-label="Create an account">
        {/* First / Last name */}
        <div className="grid grid-cols-2 gap-3">
          <div className="flex flex-col gap-1.5">
            <AuthLabel htmlFor={firstNameId}>First name</AuthLabel>
            <input
              id={firstNameId}
              type="text"
              value={firstName}
              onChange={(e) => setFirstName(e.target.value)}
              required
              aria-required="true"
              placeholder="John"
              className={baseInput}
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <AuthLabel htmlFor={lastNameId}>Last name</AuthLabel>
            <input
              id={lastNameId}
              type="text"
              value={lastName}
              onChange={(e) => setLastName(e.target.value)}
              required
              aria-required="true"
              placeholder="Smith"
              className={baseInput}
            />
          </div>
        </div>

        {/* Email */}
        <div className="flex flex-col gap-1.5">
          <AuthLabel htmlFor={emailId}>Email address</AuthLabel>
          <input
            id={emailId}
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            aria-required="true"
            placeholder="you@company.com"
            className={baseInput}
          />
        </div>

        {/* Org name + subdomain */}
        <div className="grid grid-cols-2 gap-3">
          <div className="flex flex-col gap-1.5">
            <AuthLabel htmlFor={orgNameId}>Organization</AuthLabel>
            <input
              id={orgNameId}
              type="text"
              value={organizationName}
              onChange={(e) => setOrganizationName(e.target.value)}
              required
              aria-required="true"
              placeholder="Acme Inc."
              className={baseInput}
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <AuthLabel htmlFor={subdomainId}>Subdomain</AuthLabel>
            <div className="relative">
              <input
                id={subdomainId}
                type="text"
                value={organizationSubdomain}
                onChange={(e) => handleSubdomainChange(e.target.value)}
                required
                aria-required="true"
                aria-invalid={isSubdomainUnavailable}
                aria-describedby={subdomainHintId}
                placeholder="your-company"
                className={cn(
                  baseInput,
                  'pr-10',
                  isSubdomainUnavailable &&
                    'border-destructive/50 focus:border-destructive/50 focus:ring-destructive/15',
                  subdomainStatus === 'available' &&
                    'border-green-500/50 focus:border-green-500/50 focus:ring-green-500/15'
                )}
              />
              <div className="absolute right-3 top-1/2 -translate-y-1/2" aria-hidden="true">
                {subdomainStatus === 'checking' && (
                  <Loader2 className="size-4 animate-spin text-[var(--pz-ink)]/30" />
                )}
                {subdomainStatus === 'available' && (
                  <CheckCircle className="size-4 text-green-500" />
                )}
                {subdomainStatus === 'unavailable' && (
                  <XCircle className="size-4 text-destructive" />
                )}
              </div>
            </div>
            <FieldMessage id={subdomainHintId} variant={subdomainHintVariant}>
              {subdomainStatus === 'checking' ? (
                <span>
                  <span className="sr-only">Checking subdomain availability</span>
                  <span aria-hidden="true">Lowercase, numbers, hyphens</span>
                </span>
              ) : (
                subdomainHintText
              )}
            </FieldMessage>
          </div>
        </div>

        {/* Password */}
        <div className="grid grid-cols-2 gap-3">
          <div className="flex flex-col gap-1.5">
            <AuthLabel htmlFor={passwordId}>Password</AuthLabel>
            <div className="relative">
              <input
                id={passwordId}
                type={showPassword ? 'text' : 'password'}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                aria-required="true"
                minLength={8}
                placeholder="Min. 8 characters"
                className={cn(baseInput, 'pr-10')}
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-[var(--pz-ink)]/40 transition-colors hover:text-[var(--pz-ink)]"
                aria-label={showPassword ? 'Hide password' : 'Show password'}
              >
                {showPassword ? (
                  <EyeOff className="size-4" aria-hidden="true" />
                ) : (
                  <Eye className="size-4" aria-hidden="true" />
                )}
              </button>
            </div>
          </div>

          <div className="flex flex-col gap-1.5">
            <AuthLabel htmlFor={confirmPasswordId}>Confirm password</AuthLabel>
            <div className="relative">
              <input
                id={confirmPasswordId}
                type={showConfirmPassword ? 'text' : 'password'}
                value={confirmPassword}
                onChange={(e) => {
                  setConfirmPassword(e.target.value)
                  setPasswordError(undefined)
                }}
                required
                aria-required="true"
                aria-invalid={!!passwordError}
                aria-describedby={passwordError ? passwordErrorId : undefined}
                placeholder="Repeat password"
                className={cn(
                  baseInput,
                  'pr-10',
                  passwordError &&
                    'border-destructive/50 focus:border-destructive/50 focus:ring-destructive/15'
                )}
              />
              <button
                type="button"
                onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-[var(--pz-ink)]/40 transition-colors hover:text-[var(--pz-ink)]"
                aria-label={showConfirmPassword ? 'Hide confirm password' : 'Show confirm password'}
              >
                {showConfirmPassword ? (
                  <EyeOff className="size-4" aria-hidden="true" />
                ) : (
                  <Eye className="size-4" aria-hidden="true" />
                )}
              </button>
            </div>
            {passwordError && (
              <FieldMessage id={passwordErrorId} variant="error">
                {passwordError}
              </FieldMessage>
            )}
          </div>
        </div>

        {/* General error */}
        {error && <ErrorBanner id={errorId} message={error} />}

        {/* Submit */}
        <Button
          type="submit"
          disabled={isLoading || isSubdomainUnavailable || subdomainStatus === 'checking'}
          className="mt-0.5 h-10 w-full rounded-xl bg-[var(--pz-accent)] text-sm font-semibold shadow-[var(--pz-shadow-accent)] transition-[opacity,box-shadow,transform] hover:opacity-90 hover:shadow-[var(--pz-shadow-accent-lg)] hover:scale-[1.01] active:scale-[0.99]"
        >
          {isLoading ? 'Creating account...' : 'Create free account'}
        </Button>

        <p className="text-center text-sm text-[var(--pz-ink-2)]">
          Already have an account?{' '}
          <Link href="/login" className="font-semibold text-[var(--pz-accent)] hover:underline">
            Sign in
          </Link>
        </p>
      </form>
    </div>
  )
}

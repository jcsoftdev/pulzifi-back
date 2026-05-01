'use client'

import { AuthApi } from '@workspace/services'
import { Button } from '@workspace/ui/components/atoms/button'
import { cn } from '@workspace/ui/lib/utils'
import { CheckCircle, Eye, EyeOff, Loader2, XCircle } from 'lucide-react'
import { useId, useRef, useState } from 'react'
import { env } from '@/lib/env'

interface AcceptPlatformInviteFormProps {
  token: string
  email: string
  invitedByName: string
}

type SubdomainStatus = 'idle' | 'checking' | 'available' | 'unavailable'

const inputClass =
  'h-14 w-full rounded-full bg-[#f3f3f3] px-6 text-base font-medium leading-6 tracking-[-0.128px] text-[#111] outline-none placeholder:text-[#111]/40 focus:ring-2 focus:ring-primary/30 disabled:opacity-60'

function getHostInfo() {
  const hostname = globalThis.location.hostname
  const isLocalhost =
    hostname === 'localhost' || hostname === '127.0.0.1' || hostname.endsWith('.localhost')

  const appBaseUrl = env.NEXT_PUBLIC_APP_BASE_URL
  const base = isLocalhost && appBaseUrl ? new URL(appBaseUrl) : null

  let protocol: string
  if (base) {
    protocol = base.protocol
  } else {
    protocol = globalThis.location.protocol
  }

  let port: string | undefined
  if (base) {
    port = base.port
  } else if (isLocalhost) {
    port = globalThis.location.port || '3000'
  } else {
    port = globalThis.location.port
  }

  const appDomain = env.NEXT_PUBLIC_APP_DOMAIN
  let baseDomain = appDomain === 'localhost' && !isLocalhost ? undefined : appDomain
  if (!baseDomain) {
    if (base) {
      baseDomain = base.hostname.split('.').slice(-2).join('.')
    } else if (isLocalhost) {
      baseDomain = 'localhost'
    } else {
      baseDomain = hostname.split('.').slice(-2).join('.')
    }
  }

  return {
    hostname,
    isLocalhost,
    protocol,
    port,
    baseDomain,
  }
}

function buildTenantCallbackUrl(
  protocol: string,
  targetHost: string,
  port: string | undefined,
  nonce: string | null | undefined,
  redirectTo = '/'
) {
  const portSuffix = port ? `:${port}` : ''
  const url = new URL(`${protocol}//${targetHost}${portSuffix}/api/auth/callback`)
  if (nonce) url.searchParams.set('nonce', nonce)
  url.searchParams.set('redirectTo', redirectTo)
  return url
}

function buildBaseSessionUrl(
  protocol: string,
  baseDomain: string,
  port: string | undefined,
  nonce: string | null | undefined,
  tenant: string,
  returnTo: string
) {
  const portSuffix = port ? `:${port}` : ''
  const url = new URL(`${protocol}//${baseDomain}${portSuffix}/api/auth/set-base-session`)
  if (nonce) url.searchParams.set('nonce', nonce)
  url.searchParams.set('tenant', tenant)
  url.searchParams.set('returnTo', returnTo)
  return url
}

export function AcceptPlatformInviteForm({
  token,
  email,
  invitedByName,
}: Readonly<AcceptPlatformInviteFormProps>) {
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [organizationName, setOrganizationName] = useState('')
  const [organizationSubdomain, setOrganizationSubdomain] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  const [passwordError, setPasswordError] = useState<string>()
  const [error, setError] = useState<string>()
  const [isLoading, setIsLoading] = useState(false)
  const [subdomainStatus, setSubdomainStatus] = useState<SubdomainStatus>('idle')
  const [subdomainMessage, setSubdomainMessage] = useState<string>()

  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const firstNameId = useId()
  const lastNameId = useId()
  const emailId = useId()
  const passwordId = useId()
  const confirmPasswordId = useId()
  const orgNameId = useId()
  const subdomainId = useId()

  const handleSubdomainChange = (value: string) => {
    const normalized = value.toLowerCase()
    setOrganizationSubdomain(normalized)
    if (debounceRef.current) clearTimeout(debounceRef.current)
    if (!normalized) {
      setSubdomainStatus('idle')
      setSubdomainMessage(undefined)
      return
    }
    setSubdomainStatus('checking')
    debounceRef.current = setTimeout(async () => {
      try {
        const result = await AuthApi.checkSubdomain(normalized)
        setSubdomainStatus(result.available ? 'available' : 'unavailable')
        setSubdomainMessage(result.message)
      } catch {
        setSubdomainStatus('idle')
        setSubdomainMessage(undefined)
      }
    }, 500)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(undefined)
    setPasswordError(undefined)

    if (password.length < 8) {
      setPasswordError('Password must be at least 8 characters.')
      return
    }
    if (password !== confirmPassword) {
      setPasswordError('Passwords do not match.')
      return
    }

    setIsLoading(true)
    try {
      const result = await AuthApi.acceptInvitation(token, {
        firstName,
        lastName,
        password,
        organizationName,
        organizationSubdomain,
      })

      if ('error' in result && result.error === 'session_issue_failed') {
        // Org/user provisioned but session failed: send to login on the new tenant.
        globalThis.location.href = result.loginUrl
        return
      }

      // Success: redirect into tenant via base session + tenant callback (mirror login flow).
      const success = result as {
        nonce: string
        tenant: string
        expiresIn: number
      }
      const tenant = success.tenant
      const nonce = success.nonce
      const { hostname, protocol, port, baseDomain } = getHostInfo()
      const targetHost = `${tenant}.${baseDomain}`
      const tenantCallbackUrl = buildTenantCallbackUrl(protocol, targetHost, port, nonce, '/')

      const isOnSubdomain = hostname !== baseDomain
      if (isOnSubdomain && nonce) {
        const baseSessionUrl = buildBaseSessionUrl(
          protocol,
          baseDomain,
          port,
          nonce,
          tenant,
          tenantCallbackUrl.toString()
        )
        globalThis.location.href = baseSessionUrl.toString()
      } else {
        globalThis.location.href = tenantCallbackUrl.toString()
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
      let message = 'Failed to accept invitation. Please try again.'
      if (code === 'invitation_not_found') {
        message = 'This invitation is no longer valid.'
      } else if (code === 'cannot_invite_email') {
        message = 'This email cannot be used. It may already be registered.'
      } else if (code === 'subdomain_taken') {
        message = 'That subdomain is already taken. Please pick another.'
      } else if (code === 'invalid_subdomain') {
        message = 'Subdomain must be lowercase letters, numbers, and hyphens.'
      } else if (code === 'schema_provisioning_failed') {
        message = 'We could not provision your workspace. Please contact support.'
      }
      setError(message)
    } finally {
      setIsLoading(false)
    }
  }

  const isSubdomainUnavailable = subdomainStatus === 'unavailable'

  return (
    <form onSubmit={handleSubmit} className="flex w-full flex-col gap-[30px]">
      <p className="text-base text-[#111]/70">
        Invited by <span className="font-medium text-[#111]">{invitedByName}</span>. Set up your
        account and create your organization below.
      </p>

      <div className="flex flex-col gap-3.5">
        {/* First Name / Last Name */}
        <div className="flex gap-3.5">
          <div className="flex flex-1 flex-col gap-3">
            <label
              htmlFor={firstNameId}
              className="text-base font-medium leading-6 tracking-[-0.128px] text-[#111]"
            >
              First Name
            </label>
            <input
              id={firstNameId}
              type="text"
              value={firstName}
              onChange={(e) => setFirstName(e.target.value)}
              required
              placeholder="John"
              className={inputClass}
            />
          </div>
          <div className="flex flex-1 flex-col gap-3">
            <label
              htmlFor={lastNameId}
              className="text-base font-medium leading-6 tracking-[-0.128px] text-[#111]"
            >
              Last Name
            </label>
            <input
              id={lastNameId}
              type="text"
              value={lastName}
              onChange={(e) => setLastName(e.target.value)}
              required
              placeholder="Smith"
              className={inputClass}
            />
          </div>
        </div>

        {/* Locked email */}
        <div className="flex flex-col gap-3">
          <label
            htmlFor={emailId}
            className="text-base font-medium leading-6 tracking-[-0.128px] text-[#111]"
          >
            Email address
          </label>
          <input
            id={emailId}
            type="email"
            value={email}
            readOnly
            disabled
            className={cn(inputClass, 'cursor-not-allowed')}
          />
        </div>

        {/* Organization Name */}
        <div className="flex flex-col gap-3">
          <label
            htmlFor={orgNameId}
            className="text-base font-medium leading-6 tracking-[-0.128px] text-[#111]"
          >
            Organization Name
          </label>
          <input
            id={orgNameId}
            type="text"
            value={organizationName}
            onChange={(e) => setOrganizationName(e.target.value)}
            required
            placeholder="Acme Inc."
            className={inputClass}
          />
        </div>

        {/* Subdomain */}
        <div className="flex flex-col gap-3">
          <label
            htmlFor={subdomainId}
            className="text-base font-medium leading-6 tracking-[-0.128px] text-[#111]"
          >
            Subdomain
          </label>
          <div className="relative">
            <input
              id={subdomainId}
              type="text"
              value={organizationSubdomain}
              onChange={(e) => handleSubdomainChange(e.target.value)}
              required
              placeholder="your-company"
              className={cn(
                inputClass,
                'pr-14',
                isSubdomainUnavailable && 'ring-2 ring-destructive/30',
                subdomainStatus === 'available' && 'ring-2 ring-green-500/30'
              )}
            />
            <div className="absolute right-5 top-1/2 -translate-y-1/2">
              {subdomainStatus === 'checking' && (
                <Loader2 className="size-5 animate-spin text-[#111]/40" />
              )}
              {subdomainStatus === 'available' && <CheckCircle className="size-5 text-green-500" />}
              {subdomainStatus === 'unavailable' && <XCircle className="size-5 text-destructive" />}
            </div>
          </div>
          {subdomainStatus === 'unavailable' && subdomainMessage ? (
            <p className="text-xs text-destructive">{subdomainMessage}</p>
          ) : subdomainStatus === 'available' ? (
            <p className="text-xs text-green-500">Subdomain is available</p>
          ) : (
            <p className="text-xs text-[#111]/40">Lowercase letters, numbers, and hyphens only</p>
          )}
        </div>

        {/* Password */}
        <div className="flex flex-col gap-3">
          <label
            htmlFor={passwordId}
            className="text-base font-medium leading-6 tracking-[-0.128px] text-[#111]"
          >
            Password
          </label>
          <div className="relative">
            <input
              id={passwordId}
              type={showPassword ? 'text' : 'password'}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={8}
              placeholder="Your password"
              className={cn(inputClass, 'pr-14')}
            />
            <button
              type="button"
              onClick={() => setShowPassword(!showPassword)}
              className="absolute right-5 top-1/2 -translate-y-1/2 text-[#111]/60 transition-colors hover:text-[#111]"
              aria-label={showPassword ? 'Hide password' : 'Show password'}
            >
              {showPassword ? <EyeOff className="size-6" /> : <Eye className="size-6" />}
            </button>
          </div>
        </div>

        {/* Confirm Password */}
        <div className="flex flex-col gap-3">
          <label
            htmlFor={confirmPasswordId}
            className="text-base font-medium leading-6 tracking-[-0.128px] text-[#111]"
          >
            Confirm password
          </label>
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
              placeholder="Your password"
              className={cn(inputClass, 'pr-14', passwordError && 'ring-2 ring-destructive/30')}
            />
            <button
              type="button"
              onClick={() => setShowConfirmPassword(!showConfirmPassword)}
              className="absolute right-5 top-1/2 -translate-y-1/2 text-[#111]/60 transition-colors hover:text-[#111]"
              aria-label={showConfirmPassword ? 'Hide password' : 'Show password'}
            >
              {showConfirmPassword ? <EyeOff className="size-6" /> : <Eye className="size-6" />}
            </button>
          </div>
          {passwordError && <p className="text-xs text-destructive">{passwordError}</p>}
        </div>
      </div>

      {error && (
        <div className="rounded-xl border border-destructive/20 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}

      <Button
        type="submit"
        disabled={isLoading || isSubdomainUnavailable || subdomainStatus === 'checking'}
        className="h-14 w-full rounded-full bg-[#29144c] text-base font-medium tracking-[-0.128px] hover:bg-[#3d1d6e]"
      >
        {isLoading ? 'Creating account...' : 'Accept invitation'}
      </Button>
    </form>
  )
}

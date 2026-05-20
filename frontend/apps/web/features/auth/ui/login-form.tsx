'use client'

import { Button } from '@workspace/ui/components/atoms/button'
import { cn } from '@workspace/ui/lib/utils'
import { Eye, EyeOff } from 'lucide-react'
import Link from 'next/link'
import { useId, useState } from 'react'
import type { LoginCredentials } from '../domain/types'
import { AuthLabel, ErrorBanner } from './form-atoms'

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
  'h-11 w-full rounded-xl border border-[var(--pz-ink)]/10 bg-white px-4 text-sm text-[var(--pz-ink)] outline-none transition-[border-color,box-shadow] placeholder:text-[var(--pz-ink)]/35 focus:border-[var(--pz-accent)]/40 focus:ring-2 focus:ring-[var(--pz-accent)]/15'

export interface LoginFormProps {
  onSubmit: (credentials: LoginCredentials) => Promise<void>
  isLoading?: boolean
  error?: string
}

export function LoginForm({ onSubmit, isLoading = false, error }: Readonly<LoginFormProps>) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const emailId = useId()
  const passwordId = useId()
  const errorId = useId()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      await onSubmit({
        email,
        password,
      })
    } catch (err) {
      console.error('[LoginForm] onSubmit error:', err)
    }
  }

  return (
    <div className="flex flex-col gap-5">
      {/* Google OAuth */}
      <a
        href="/api/v1/auth/oauth/google"
        aria-label="Continue with Google"
        className={cn(
          'flex h-11 items-center justify-center gap-3 rounded-xl border border-[var(--pz-ink)]/10 bg-white text-sm font-medium text-[var(--pz-ink)] transition-colors hover:bg-gray-50',
          isLoading && 'pointer-events-none opacity-50'
        )}
      >
        <GoogleIcon />
        Continue with Google
      </a>

      {/* Divider */}
      <div className="flex items-center gap-3" aria-hidden="true">
        <div className="h-px flex-1 bg-[var(--pz-ink)]/8" />
        <span className="text-xs text-[var(--pz-ink)]/35">or sign in with email</span>
        <div className="h-px flex-1 bg-[var(--pz-ink)]/8" />
      </div>

      {/* Email + password form */}
      <form onSubmit={handleSubmit} className="flex flex-col gap-4" aria-label="Sign in with email">
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
            aria-invalid={!!error}
            aria-describedby={error ? errorId : undefined}
            placeholder="you@company.com"
            className={baseInput}
          />
        </div>

        {/* Password */}
        <div className="flex flex-col gap-1.5">
          <div className="flex items-center justify-between">
            <AuthLabel htmlFor={passwordId}>Password</AuthLabel>
            <Link
              href="/forgot-password"
              className="text-xs font-medium text-[var(--pz-accent)] hover:underline"
            >
              Forgot password?
            </Link>
          </div>
          <div className="relative">
            <input
              id={passwordId}
              type={showPassword ? 'text' : 'password'}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              aria-required="true"
              aria-invalid={!!error}
              aria-describedby={error ? errorId : undefined}
              placeholder="Your password"
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

        {/* Error */}
        {error && <ErrorBanner id={errorId} message={error} />}

        {/* Submit */}
        <Button
          type="submit"
          disabled={isLoading}
          className="mt-1 h-11 w-full rounded-xl bg-[var(--pz-accent)] text-sm font-semibold shadow-[var(--pz-shadow-accent)] transition-[opacity,box-shadow,transform] hover:opacity-90 hover:shadow-[var(--pz-shadow-accent-lg)] hover:scale-[1.01] active:scale-[0.99]"
        >
          {isLoading ? 'Signing in...' : 'Sign in'}
        </Button>

        <p className="text-center text-sm text-[var(--pz-ink-2)]">
          Don&apos;t have an account?{' '}
          <Link href="/register" className="font-semibold text-[var(--pz-accent)] hover:underline">
            Sign up free
          </Link>
        </p>
      </form>
    </div>
  )
}

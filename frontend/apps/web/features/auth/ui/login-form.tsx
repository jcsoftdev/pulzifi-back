'use client'

import { Button } from '@workspace/ui/components/atoms/button'
import { cn } from '@workspace/ui/lib/utils'
import { Eye, EyeOff } from 'lucide-react'
import Link from 'next/link'
import { useId, useState } from 'react'
import type { LoginCredentials } from '../domain/types'

function GoogleIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" aria-label="Google" role="img">
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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      await onSubmit({
        email,
        password,
      })
    } catch (error) {
      console.error('[LoginForm] onSubmit error:', error)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="flex w-full flex-col gap-[30px]">
      {/* Input fields */}
      <div className="flex flex-col gap-3.5">
        {/* Email field */}
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
            onChange={(e) => setEmail(e.target.value)}
            required
            placeholder="Your email address"
            className="h-14 w-full rounded-full bg-[#f3f3f3] px-6 text-base font-medium leading-6 tracking-[-0.128px] text-[#111] outline-none placeholder:text-[#111]/40 focus:ring-2 focus:ring-primary/30"
          />
        </div>

        {/* Password field */}
        <div className="flex flex-col gap-3">
          <div className="flex items-center justify-between">
            <label
              htmlFor={passwordId}
              className="text-base font-medium leading-6 tracking-[-0.128px] text-[#111]"
            >
              Password
            </label>
            <Link
              href="/forgot-password"
              className="text-sm font-medium leading-6 tracking-[-0.112px] text-[#2e47ba] hover:underline"
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
              placeholder="Your password"
              className="h-14 w-full rounded-full bg-[#f3f3f3] px-6 pr-14 text-base font-medium leading-6 tracking-[-0.128px] text-[#111] outline-none placeholder:text-[#111]/40 focus:ring-2 focus:ring-primary/30"
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
      </div>

      {/* Error message */}
      {error && (
        <div className="rounded-xl bg-destructive/10 px-4 py-3 text-sm text-destructive border border-destructive/20">
          {error}
        </div>
      )}

      {/* Divider */}
      <div className="flex items-center gap-3.5">
        <div className="h-px flex-1 bg-[#111]/10" />
        <span className="text-base font-medium leading-6 tracking-[-0.128px] text-[#111]/40">
          Or
        </span>
        <div className="h-px flex-1 bg-[#111]/10" />
      </div>

      {/* Google sign in + Submit */}
      <div className="flex flex-col gap-3.5">
        {/* Google OAuth */}
        <a
          href="/api/v1/auth/oauth/google"
          className={cn(
            'flex h-[52px] items-center justify-center rounded-full border border-[#111]/10 bg-white transition-colors hover:bg-gray-50',
            isLoading && 'pointer-events-none opacity-50'
          )}
        >
          <GoogleIcon />
        </a>

        {/* Sign In button */}
        <Button
          type="submit"
          disabled={isLoading}
          className="h-14 w-full rounded-full bg-[#29144c] text-base font-medium tracking-[-0.128px] hover:bg-[#3d1d6e]"
        >
          {isLoading ? 'Signing in...' : 'Sign In'}
        </Button>

        {/* Sign up link */}
        <p className="text-base leading-6 text-[#111]">
          Don&apos;t have account?{' '}
          <Link href="/register" className="font-bold text-[#29144c] hover:underline">
            Sign Up
          </Link>
        </p>
      </div>
    </form>
  )
}

'use client'

import { AuthApi } from '@workspace/services'
import { useCallback, useRef, useState } from 'react'
import { performTenantHandoff } from './tenant-session-handoff'
import type { RegisterData } from '../domain/types'

export type SubdomainStatus = 'idle' | 'checking' | 'available' | 'unavailable'

export function useRegister() {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string>()
  const [subdomainStatus, setSubdomainStatus] = useState<SubdomainStatus>('idle')
  const [subdomainMessage, setSubdomainMessage] = useState<string>()
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const checkSubdomain = useCallback((subdomain: string) => {
    if (debounceRef.current) clearTimeout(debounceRef.current)

    if (!subdomain) {
      setSubdomainStatus('idle')
      setSubdomainMessage(undefined)
      return
    }

    setSubdomainStatus('checking')
    debounceRef.current = setTimeout(async () => {
      try {
        const result = await AuthApi.checkSubdomain(subdomain)
        setSubdomainStatus(result.available ? 'available' : 'unavailable')
        setSubdomainMessage(result.message)
      } catch {
        setSubdomainStatus('idle')
        setSubdomainMessage(undefined)
      }
    }, 500)
  }, [])

  const register = async (data: RegisterData) => {
    setIsLoading(true)
    setError(undefined)
    try {
      await AuthApi.register(data)
      const loginResponse = await AuthApi.login({ email: data.email, password: data.password })
      performTenantHandoff(loginResponse, '/onboarding', () => {
        window.location.assign('/onboarding')
      })
    } catch (err: unknown) {
      const apiError = err as {
        response?: {
          data?: {
            error?: string
          }
        }
        message?: string
      }
      setError(
        apiError?.response?.data?.error ||
          apiError?.message ||
          'Registration failed. Please try again.'
      )
    } finally {
      setIsLoading(false)
    }
  }

  return {
    register,
    isLoading,
    error,
    checkSubdomain,
    subdomainStatus,
    subdomainMessage,
  }
}

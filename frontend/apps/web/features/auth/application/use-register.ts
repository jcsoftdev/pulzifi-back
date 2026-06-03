'use client'

import { AuthApi } from '@workspace/services'
import { useCallback, useRef, useState } from 'react'
import type { RegisterData } from '../domain/types'

export type SubdomainStatus = 'idle' | 'checking' | 'available' | 'unavailable'

function buildTenantOnboardingUrl(subdomain: string): string {
  const baseUrl = process.env.NEXT_PUBLIC_APP_BASE_URL
  if (baseUrl) {
    const base = new URL(baseUrl)
    const portSuffix = base.port ? `:${base.port}` : ''
    return `${base.protocol}//${subdomain}.${base.hostname}${portSuffix}/onboarding`
  }

  const appDomain = process.env.NEXT_PUBLIC_APP_DOMAIN
  if (appDomain) {
    return `https://${subdomain}.${appDomain}/onboarding`
  }

  const { protocol, hostname, port } = window.location
  const hostWithoutPort = hostname
  let baseDomain: string
  if (hostWithoutPort === 'localhost' || hostWithoutPort === '127.0.0.1') {
    baseDomain = 'localhost'
  } else {
    const parts = hostWithoutPort.split('.')
    baseDomain = parts.length >= 2 ? parts.slice(-2).join('.') : hostWithoutPort
  }
  const portSuffix = port ? `:${port}` : ''
  return `${protocol}//${subdomain}.${baseDomain}${portSuffix}/onboarding`
}

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
      const result = await AuthApi.register(data)
      window.location.assign(buildTenantOnboardingUrl(result.organizationSubdomain))
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

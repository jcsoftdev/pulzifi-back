'use client'

import { AuthApi } from '@workspace/services'
import { useCallback, useRef, useState } from 'react'
import type { OnboardingFormValues } from '../domain/types'

export type SubdomainStatus = 'idle' | 'checking' | 'available' | 'unavailable'

function buildTenantUrl(subdomain: string): string {
  // NEXT_PUBLIC_APP_BASE_URL overrides — e.g. https://localhost:3000
  const baseUrl = process.env.NEXT_PUBLIC_APP_BASE_URL
  if (baseUrl) {
    const base = new URL(baseUrl)
    const portSuffix = base.port ? `:${base.port}` : ''
    return `${base.protocol}//${subdomain}.${base.hostname}${portSuffix}/`
  }

  // NEXT_PUBLIC_APP_DOMAIN for production — e.g. pulzifi.com
  const appDomain = process.env.NEXT_PUBLIC_APP_DOMAIN
  if (appDomain) {
    return `https://${subdomain}.${appDomain}/`
  }

  // Dev fallback: derive from window.location
  const { protocol, hostname, port } = window.location
  const hostWithoutPort = hostname
  let baseDomain: string
  if (hostWithoutPort === 'localhost' || hostWithoutPort === '127.0.0.1') {
    baseDomain = 'localhost'
  } else {
    const parts = hostWithoutPort.split('.')
    baseDomain = parts.length >= 2 ? parts.slice(-2).join('.') : hostWithoutPort
  }
  const isLocal = baseDomain === 'localhost' || baseDomain === '127.0.0.1'
  const portSuffix = isLocal && port ? `:${port}` : ''
  return `${protocol}//${subdomain}.${baseDomain}${portSuffix}/`
}

export function useOnboarding() {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | undefined>()
  const [subdomainStatus, setSubdomainStatus] = useState<SubdomainStatus>('idle')
  const [subdomainMessage, setSubdomainMessage] = useState<string | undefined>()
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

  const submit = async (values: OnboardingFormValues) => {
    setIsLoading(true)
    setError(undefined)
    try {
      const result = await AuthApi.submitOnboarding({
        org_name: values.org_name,
        subdomain: values.subdomain,
      })
      window.location.assign(buildTenantUrl(result.subdomain))
    } catch (err: unknown) {
      const apiError = err as {
        response?: { data?: { error?: string }; status?: number }
        message?: string
      }
      const status = apiError?.response?.status
      if (status === 409) {
        // Race condition: user already has an org — refetch and redirect
        try {
          const user = await AuthApi.getCurrentUser()
          if (user.tenant) {
            window.location.assign(buildTenantUrl(user.tenant))
            return
          }
        } catch {
          // Ignore secondary error — fall through to generic message
        }
        setError('Your organization has already been created.')
      } else {
        setError(
          apiError?.response?.data?.error ||
            apiError?.message ||
            'Failed to create organization. Please try again.'
        )
      }
    } finally {
      setIsLoading(false)
    }
  }

  return {
    submit,
    isLoading,
    error,
    checkSubdomain,
    subdomainStatus,
    subdomainMessage,
  }
}

'use client'

import { AuthApi } from '@workspace/services'
import { ArrowRight } from 'lucide-react'
import Link from 'next/link'
import { useRouter, useSearchParams } from 'next/navigation'
import { useEffect, useState } from 'react'
import { LoginForm } from '@/features/auth/ui/login-form'
import { FooterSection } from '@/features/landing/ui/footer-section'
import { Navbar } from '@/features/landing/ui/navbar'
import { env } from '@/lib/env'

function CTASection() {
  return (
    <section className="relative mx-auto max-w-[1256px] overflow-hidden rounded-3xl bg-[#29144c] px-6 py-12 md:px-[58px] md:py-[49px]">
      {/* Grid background */}
      <div className="pointer-events-none absolute inset-0 opacity-10">
        <div
          className="size-full"
          style={{
            backgroundImage:
              'linear-gradient(rgba(255,255,255,0.3) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.3) 1px, transparent 1px)',
            backgroundSize: '60px 60px',
          }}
        />
      </div>

      <div className="relative z-10 flex flex-col gap-10 lg:flex-row lg:items-center lg:justify-between">
        {/* Left content */}
        <div className="flex max-w-[575px] flex-col gap-[30px]">
          <div className="flex flex-col gap-[30px]">
            <h2 className="font-heading text-3xl font-medium leading-[40px] tracking-[-1.5px] text-white sm:text-4xl sm:leading-[48px] md:text-[48px] md:leading-[56px] md:tracking-[-2.88px]">
              Ready to Transform your strategic decisions, fast!
            </h2>
            <p className="max-w-[486px] text-base leading-6 text-white/80">
              Choose guidance tailored to marketing, business strategy, brand positioning, and more.
              Let AI turn raw data into actionable intelligence.
            </p>
          </div>
          <Link
            href="/register"
            className="inline-flex h-14 w-fit items-center gap-2 rounded-full bg-white px-8 text-base font-medium text-[#7c3aed] transition-all hover:scale-[1.02] active:scale-[0.98]"
          >
            Try It Now
            <ArrowRight className="size-5" />
          </Link>
        </div>

        {/* Right floating cards */}
        <div className="hidden lg:block relative h-[400px] w-[540px] shrink-0">
          <FloatingCard
            className="right-[50px] top-0 w-[214px]"
            tagColor=""
            label="Analyze alerts with..."
            tagText=""
            labelColor="text-[#7c3aed]"
          />
          <FloatingCard
            className="right-[67px] top-[53px] w-[466px]"
            tagColor="bg-[#ebf0ff]"
            tagText="Marketing Lens"
          />
          <FloatingCard
            className="right-[83px] top-[107px] w-[503px]"
            tagColor="bg-[#ebfff5]"
            tagText="Brand and Pricing Strategy"
          />
          <FloatingCard
            className="right-[95px] top-[160px] w-[541px]"
            tagColor="bg-[#fff9f5]"
            tagText="Business Opportunities"
            showInsight
          />
        </div>
      </div>
    </section>
  )
}

function FloatingCard({
  className,
  tagColor,
  tagText,
  label,
  labelColor,
  showInsight,
}: {
  className?: string
  tagColor: string
  tagText: string
  label?: string
  labelColor?: string
  showInsight?: boolean
}) {
  return (
    <div
      className={`absolute rounded-[28px] bg-white p-4 shadow-[0px_-18px_28px_0px_rgba(0,0,0,0.02)] ${className}`}
    >
      <div className="flex items-start justify-between">
        {tagText ? (
          <span
            className={`rounded-full ${tagColor} px-4 py-1.5 text-sm font-medium text-[#29144c]`}
          >
            {tagText}
          </span>
        ) : label ? (
          <span className={`text-sm font-bold ${labelColor}`}>{label}</span>
        ) : null}
      </div>
      {showInsight && (
        <div className="mt-3">
          <p className="text-xs text-[#111]">AI Insight:</p>
          <p className="mt-1 text-xs leading-relaxed text-[#504170]">
            ...your competitor updated their homepage messaging from &quot;Affordable Listings&quot;
            to &quot;Luxury Investment Properties&quot; and added new content focused on
            high-net-worth clients...
          </p>
        </div>
      )}
    </div>
  )
}

export default function LoginPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string>()
  const [infoBanner, setInfoBanner] = useState<string>()

  useEffect(() => {
    const errorParam = searchParams.get('error')
    if (errorParam === 'SessionExpired') {
      setError('Your session has expired. Please sign in again.')
    } else if (errorParam === 'PendingApproval') {
      setInfoBanner(
        'Your account is pending approval by an administrator. Please check back later.'
      )
    }

    if (searchParams.get('registered') === 'true') {
      setInfoBanner('Registration successful! Please wait for admin approval before logging in.')
    }
  }, [
    searchParams,
  ])

  const getHostInfo = () => {
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
      base,
    }
  }

  const buildTenantCallbackUrl = (
    protocol: string,
    targetHost: string,
    port?: string,
    nonce?: string | null,
    redirectTo = '/'
  ) => {
    const portSuffix = port ? `:${port}` : ''
    const tenantCallbackUrl = new URL(`${protocol}//${targetHost}${portSuffix}/api/auth/callback`)
    if (nonce) {
      tenantCallbackUrl.searchParams.set('nonce', nonce)
    }
    tenantCallbackUrl.searchParams.set('redirectTo', redirectTo)
    return tenantCallbackUrl
  }

  const buildBaseSessionUrl = (
    protocol: string,
    baseDomain: string,
    port?: string,
    nonce?: string | null,
    tenant?: string,
    returnTo?: string
  ) => {
    const portSuffix = port ? `:${port}` : ''
    const baseSessionUrl = new URL(
      `${protocol}//${baseDomain}${portSuffix}/api/auth/set-base-session`
    )
    if (nonce) baseSessionUrl.searchParams.set('nonce', nonce)
    if (tenant) baseSessionUrl.searchParams.set('tenant', tenant)
    if (returnTo) baseSessionUrl.searchParams.set('returnTo', returnTo)
    return baseSessionUrl
  }

  const handleLogin = async (credentials: { email: string; password: string }) => {
    setIsLoading(true)
    setError(undefined)

    try {
      const loginResponse = await AuthApi.login(credentials)
      const tenant = loginResponse.tenant

      if (!tenant) {
        router.push('/')
        router.refresh()
        return
      }

      const { hostname, protocol, port, baseDomain } = getHostInfo()
      const targetHost = `${tenant}.${baseDomain}`
      const redirectTo = searchParams.get('callbackUrl') || '/'

      const tenantCallbackUrl = buildTenantCallbackUrl(
        protocol,
        targetHost,
        port,
        loginResponse.nonce,
        redirectTo
      )

      const isOnSubdomain = hostname !== baseDomain
      if (isOnSubdomain && loginResponse.nonce) {
        const baseSessionUrl = buildBaseSessionUrl(
          protocol,
          baseDomain,
          port,
          loginResponse.nonce,
          tenant,
          tenantCallbackUrl.toString()
        )
        globalThis.location.href = baseSessionUrl.toString()
      } else {
        globalThis.location.href = tenantCallbackUrl.toString()
      }
    } catch (err: unknown) {
      const axiosError = err as {
        response?: {
          status?: number
          data?: {
            error?: string
            code?: string
          }
        }
      }

      if (axiosError?.response?.status === 403) {
        const code = axiosError.response.data?.code
        if (code === 'USER_REJECTED') {
          setError('Your account has been rejected. Please contact support.')
        } else {
          setError('Your account is pending approval by an administrator.')
        }
      } else {
        setError('Invalid email or password')
      }
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-[#f3f3f3]">
      <div className="mx-auto max-w-[1280px] space-y-3 p-3">
        <Navbar />

        {/* Sign-in form section */}
        <section className="relative mx-auto max-w-[1256px] overflow-hidden rounded-3xl bg-white px-6 py-12 md:px-0 md:py-[50px]">
          <div className="mx-auto flex w-full max-w-[583px] flex-col gap-[60px]">
            <h1 className="font-heading text-[48px] font-medium italic leading-[56px] tracking-[-1px] text-[#111] max-md:text-4xl max-md:leading-[44px]">
              Welcome back!
            </h1>

            {infoBanner && (
              <div className="rounded-xl border border-blue-200 bg-blue-50 p-3 text-sm text-blue-700 dark:border-blue-800 dark:bg-blue-950 dark:text-blue-300">
                {infoBanner}
              </div>
            )}

            <LoginForm onSubmit={handleLogin} isLoading={isLoading} error={error} />
          </div>
        </section>

        <CTASection />
        <FooterSection />
      </div>
    </div>
  )
}

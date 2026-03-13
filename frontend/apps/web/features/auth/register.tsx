'use client'

import Link from 'next/link'
import { FooterSection } from '@/features/landing/ui/footer-section'
import { Navbar } from '@/features/landing/ui/navbar'
import { useRegister } from './application/use-register'
import { RegisterForm } from './ui/register-form'

export function RegisterFeature() {
  const {
    register,
    isLoading,
    error,
    submitted,
    checkSubdomain,
    subdomainStatus,
    subdomainMessage,
  } = useRegister()

  if (submitted) {
    return (
      <div className="min-h-screen bg-[#f3f3f3]">
        <div className="mx-auto max-w-[1280px] space-y-3 p-3">
          <Navbar />
          <section className="mx-auto max-w-[1256px] overflow-hidden rounded-3xl bg-white px-6 py-12 md:px-0 md:py-[50px]">
            <div className="mx-auto flex w-full max-w-[583px] flex-col items-center gap-6 text-center">
              <h1 className="font-heading text-[48px] font-medium italic leading-[56px] tracking-[-1px] text-[#111] max-md:text-4xl max-md:leading-[44px]">
                Registration submitted!
              </h1>
              <p className="text-base leading-6 text-[#111]/60">
                Your account is pending approval by an administrator. You will be able to log in
                once your account has been approved.
              </p>
              <Link
                href="/login"
                className="inline-flex h-14 items-center rounded-full bg-[#29144c] px-8 text-base font-medium text-white transition-colors hover:bg-[#3d1d6e]"
              >
                Back to login
              </Link>
            </div>
          </section>
          <FooterSection />
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-[#f3f3f3]">
      <div className="mx-auto max-w-[1280px] space-y-3 p-3">
        <Navbar />

        {/* Sign-up form section */}
        <section className="mx-auto max-w-[1256px] overflow-hidden rounded-3xl bg-white px-6 py-12 md:px-0 md:py-[50px]">
          <div className="mx-auto flex w-full max-w-[583px] flex-col gap-[60px]">
            <h1 className="font-heading text-[48px] font-medium italic leading-[56px] tracking-[-1px] text-[#111] max-md:text-4xl max-md:leading-[44px]">
              Let&apos;s create your account
            </h1>

            <RegisterForm
              onSubmit={register}
              isLoading={isLoading}
              error={error}
              onSubdomainChange={checkSubdomain}
              subdomainStatus={subdomainStatus}
              subdomainMessage={subdomainMessage}
            />
          </div>
        </section>

        <FooterSection />
      </div>
    </div>
  )
}

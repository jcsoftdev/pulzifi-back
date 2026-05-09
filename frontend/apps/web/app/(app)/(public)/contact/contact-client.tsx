'use client'

import { cn } from '@workspace/ui/lib/utils'
import { CheckCircle, Linkedin, Mail, MapPin, Twitter, Youtube } from 'lucide-react'
import { useId, useState } from 'react'
import { Eyebrow } from '@/features/landing/ui/components/eyebrow'
import { Highlight } from '@/features/landing/ui/components/highlight'
import { FooterSection } from '@/features/landing/ui/footer-section'
import { Navbar } from '@/features/landing/ui/navbar'

type NavLink = { label: string; href: string }
type SocialLink = { platform: string; href: string }

export type ContactPageCmsProps = {
  headline: string
  subheadline: string
  email: string
  address: string
  navLinks?: NavLink[]
  navSigninLabel?: string
  navSigninHref?: string
  navPrimaryCtaLabel?: string
  navPrimaryCtaHref?: string
  navLogoUrl?: string
  footerGroups?: Record<string, { label: string; href: string }[]>
  footerTagline?: string
  footerSocialLinks?: SocialLink[]
  footerLogoUrl?: string
  themeStyle?: string
}

const inputClass =
  'h-14 w-full rounded-xl border border-[var(--pz-card-border)] bg-[var(--pz-page-bg-alt)] px-5 text-base text-[var(--pz-ink)] outline-none placeholder:text-[var(--pz-ink-2)]/50 focus:border-[var(--pz-accent)]/40 focus:ring-2 focus:ring-[var(--pz-accent)]/20 transition-colors'

export function ContactClient({
  headline,
  subheadline,
  email,
  address,
  navLinks,
  navSigninLabel,
  navSigninHref,
  navPrimaryCtaLabel,
  navPrimaryCtaHref,
  navLogoUrl,
  footerGroups,
  footerTagline,
  footerSocialLinks,
  footerLogoUrl,
  themeStyle,
}: ContactPageCmsProps) {
  const [name, setName] = useState('')
  const [userEmail, setUserEmail] = useState('')
  const [message, setMessage] = useState('')
  const [submitted, setSubmitted] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const nameId = useId()
  const emailId = useId()
  const messageId = useId()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)
    const subject = encodeURIComponent(`Contact inquiry from ${name}`)
    const body = encodeURIComponent(`Name: ${name}\nEmail: ${userEmail}\n\n${message}`)
    window.open(`mailto:${email}?subject=${subject}&body=${body}`, '_self')
    setSubmitted(true)
    setIsLoading(false)
  }

  return (
    <>
      {themeStyle && <style dangerouslySetInnerHTML={{ __html: `:root{${themeStyle}}` }} />}
      <div className="flex min-h-screen flex-col bg-[var(--pz-page-bg)]">
        <div className="mx-auto w-full max-w-[1280px] px-3 pt-3">
          <Navbar
            links={navLinks}
            signinLabel={navSigninLabel}
            signinHref={navSigninHref}
            primaryCtaLabel={navPrimaryCtaLabel}
            primaryCtaHref={navPrimaryCtaHref}
            logoUrl={navLogoUrl}
          />
        </div>

        <div className="h-28 shrink-0" />
        <main className="mx-auto flex w-full max-w-[1280px] flex-1 flex-col justify-center px-4 py-16 lg:px-8">
          <div className="grid grid-cols-1 gap-10 lg:grid-cols-5 lg:items-center lg:gap-16">

            {/* Left */}
            <div className="flex flex-col gap-10 lg:col-span-2">
              <div>
                <Eyebrow className="mb-4">Contact</Eyebrow>
                <h1 className="font-heading text-4xl font-bold leading-[1.1] tracking-tight text-[var(--pz-ink)] md:text-5xl">
                  {headline.includes('touch') ? (
                    <>Get in <Highlight tone="accent">touch</Highlight></>
                  ) : headline}
                </h1>
                <p className="mt-4 text-lg leading-relaxed text-[var(--pz-ink-2)]">{subheadline}</p>
              </div>

              <div className="flex flex-col gap-6">
                <div className="flex items-start gap-4">
                  <div className="flex size-11 shrink-0 items-center justify-center rounded-xl bg-[var(--pz-accent)]/10">
                    <Mail className="size-5 text-[var(--pz-accent)]" />
                  </div>
                  <div>
                    <p className="text-sm font-medium text-[var(--pz-ink-2)]">Email</p>
                    <a href={`mailto:${email}`} className="text-base font-semibold text-[var(--pz-ink)] transition-colors hover:text-[var(--pz-accent)]">{email}</a>
                  </div>
                </div>
                <div className="flex items-start gap-4">
                  <div className="flex size-11 shrink-0 items-center justify-center rounded-xl bg-[var(--pz-accent)]/10">
                    <MapPin className="size-5 text-[var(--pz-accent)]" />
                  </div>
                  <div>
                    <p className="text-sm font-medium text-[var(--pz-ink-2)]">Address</p>
                    <span className="text-base font-semibold text-[var(--pz-ink)]">{address}</span>
                  </div>
                </div>
              </div>

              <div>
                <p className="mb-3 text-sm font-medium text-[var(--pz-ink-2)]">Follow us</p>
                <div className="flex gap-3">
                  {[
                    { href: 'https://x.com/pulzifi', Icon: Twitter, label: 'X (Twitter)' },
                    { href: 'https://linkedin.com/company/pulzifi', Icon: Linkedin, label: 'LinkedIn' },
                    { href: 'https://youtube.com/@pulzifi', Icon: Youtube, label: 'YouTube' },
                  ].map(({ href, Icon, label }) => (
                    <a key={label} href={href} target="_blank" rel="noopener noreferrer" aria-label={label}
                      className="flex size-10 items-center justify-center rounded-xl border border-[var(--pz-card-border)] bg-white text-[var(--pz-ink-2)] transition-all hover:border-[var(--pz-accent)]/30 hover:text-[var(--pz-accent)] hover:shadow-sm">
                      <Icon className="size-4" />
                    </a>
                  ))}
                </div>
              </div>
            </div>

            {/* Right: form */}
            <div className="rounded-2xl border border-[var(--pz-card-border)] bg-white p-8 shadow-[var(--pz-card-shadow-rest)] lg:col-span-3 md:p-10">
              {submitted ? (
                <div className="flex flex-col items-center justify-center gap-4 py-12 text-center">
                  <CheckCircle className="size-12 text-emerald-500" />
                  <h3 className="text-xl font-bold text-[var(--pz-ink)]">Message sent!</h3>
                  <p className="text-sm text-[var(--pz-ink-2)]">Thank you for reaching out. We'll get back to you within one business day.</p>
                  <button type="button" onClick={() => { setSubmitted(false); setName(''); setUserEmail(''); setMessage('') }}
                    className="text-sm font-medium text-[var(--pz-accent)] hover:underline">Send another message</button>
                </div>
              ) : (
                <form onSubmit={handleSubmit} className="flex flex-col gap-6">
                  <div className="flex flex-col gap-4 sm:flex-row">
                    <div className="flex flex-1 flex-col gap-2">
                      <label htmlFor={nameId} className="text-sm font-medium text-[var(--pz-ink)]">Your name</label>
                      <input id={nameId} type="text" value={name} onChange={(e) => setName(e.target.value)} required placeholder="Jane Smith" className={inputClass} />
                    </div>
                    <div className="flex flex-1 flex-col gap-2">
                      <label htmlFor={emailId} className="text-sm font-medium text-[var(--pz-ink)]">Email address</label>
                      <input id={emailId} type="email" value={userEmail} onChange={(e) => setUserEmail(e.target.value)} required placeholder="jane@company.com" className={inputClass} />
                    </div>
                  </div>
                  <div className="flex flex-col gap-2">
                    <label htmlFor={messageId} className="text-sm font-medium text-[var(--pz-ink)]">Message</label>
                    <textarea id={messageId} value={message} onChange={(e) => setMessage(e.target.value)} required placeholder="How can we help you?"
                      className={cn('min-h-[160px] resize-none rounded-xl border border-[var(--pz-card-border)] bg-[var(--pz-page-bg-alt)] px-5 py-4 text-base text-[var(--pz-ink)] outline-none placeholder:text-[var(--pz-ink-2)]/50 focus:border-[var(--pz-accent)]/40 focus:ring-2 focus:ring-[var(--pz-accent)]/20 transition-colors')} />
                  </div>
                  <button type="submit" disabled={isLoading}
                    className="h-12 w-full rounded-xl bg-[var(--pz-accent)] text-sm font-semibold text-white shadow-[var(--pz-shadow-accent)] transition-all hover:opacity-90 hover:scale-[1.01] active:scale-[0.99] disabled:opacity-50">
                    {isLoading ? 'Sending...' : 'Send Message'}
                  </button>
                </form>
              )}
            </div>
          </div>
        </main>

        <FooterSection
          links={footerGroups}
          tagline={footerTagline}
          socialLinks={footerSocialLinks}
          logoUrl={footerLogoUrl}
        />
      </div>
    </>
  )
}

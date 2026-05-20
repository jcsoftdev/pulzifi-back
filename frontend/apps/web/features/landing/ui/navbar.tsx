'use client'

import { cn } from '@workspace/ui/lib/utils'
import { Menu, X } from 'lucide-react'
import Image from 'next/image'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { LandingButton } from './components/landing-button'

type NavLink = {
  label: string
  href: string
}
type NavbarProps = {
  links?: ReadonlyArray<NavLink>
  signinLabel?: string
  signinHref?: string
  primaryCtaLabel?: string
  primaryCtaHref?: string
  logoUrl?: string
}

const DEFAULT_NAV_LINKS: ReadonlyArray<NavLink> = [
  {
    label: 'Home',
    href: '#',
  },
  {
    label: 'Product',
    href: '#how-it-works',
  },
  {
    label: 'How to use',
    href: '#industries',
  },
  {
    label: 'Pricing',
    href: '#pricing',
  },
  {
    label: 'Contact',
    href: '/contact',
  },
]

const NAV_LINK_PATHS: Record<string, string> = {
  Pricing: '/pricing',
}

function resolveHref(href: string, isHome: boolean) {
  if (!isHome && href.startsWith('#')) return `/${href}`
  return href
}

function isLinkActive(label: string, pathname: string) {
  if (NAV_LINK_PATHS[label]) return pathname === NAV_LINK_PATHS[label]
  return false
}

function MobileMenu({
  open,
  onClose,
  isHome,
  links,
  signinLabel,
  signinHref,
  primaryCtaLabel,
  primaryCtaHref,
  logoUrl,
}: {
  open: boolean
  onClose: () => void
  isHome: boolean
  links: ReadonlyArray<NavLink>
  signinLabel?: string
  signinHref?: string
  primaryCtaLabel?: string
  primaryCtaHref?: string
  logoUrl?: string
}) {
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted) return null

  return createPortal(
    <div className="md:hidden">
      {/* Backdrop */}
      <div
        className={cn(
          'fixed inset-0 z-[90] bg-black/40 transition-opacity duration-300',
          open ? 'opacity-100' : 'pointer-events-none opacity-0'
        )}
        onClick={onClose}
        aria-hidden="true"
      />

      {/* Dropdown panel */}
      <div
        className={cn(
          'fixed left-3 right-3 top-3 z-[100] rounded-[20px] bg-white px-6 py-4 shadow-xl transition-all duration-300 ease-in-out',
          open ? 'translate-y-0 opacity-100' : '-translate-y-2 pointer-events-none opacity-0'
        )}
      >
        {/* Header */}
        <div className="flex items-center justify-between">
          <Link
            href="/"
            className="font-logo text-xl font-extrabold tracking-[0.2px] text-[var(--pz-dark-surface)] focus-visible:ring-2 focus-visible:ring-[var(--pz-accent)] focus-visible:ring-offset-2 focus-visible:outline-none"
            onClick={onClose}
          >
            {logoUrl ? (
              <Image src={logoUrl} alt="Pulzifi" width={96} height={28} className="h-7 w-auto" />
            ) : (
              'Pulzifi'
            )}
          </Link>
          <button
            type="button"
            className="flex items-center justify-center rounded-lg p-2 text-[var(--pz-ink)] focus-visible:ring-2 focus-visible:ring-[var(--pz-accent)] focus-visible:outline-none"
            onClick={onClose}
            aria-label="Close menu"
          >
            <X className="size-6" />
          </button>
        </div>

        {/* Links */}
        <div className="flex flex-col gap-1 pt-4">
          {links.map((link) => (
            <Link
              key={link.label}
              href={NAV_LINK_PATHS[link.label] || resolveHref(link.href, isHome)}
              className="rounded-xl px-4 py-3 text-base font-medium text-[var(--pz-ink)] transition-colors hover:bg-gray-50 focus-visible:ring-2 focus-visible:ring-[var(--pz-accent)] focus-visible:outline-none"
              onClick={onClose}
            >
              {link.label}
            </Link>
          ))}
        </div>

        {/* Buttons */}
        <div className="mt-3 flex gap-2 border-t border-gray-100 pt-4">
          <LandingButton
            href={signinHref ?? '/login'}
            variant="outline"
            className="flex-1 justify-center"
          >
            {signinLabel ?? 'Sign in'}
          </LandingButton>
          <LandingButton
            href={primaryCtaHref ?? '/register'}
            variant="primary"
            className="flex-1 justify-center"
          >
            {primaryCtaLabel ?? 'Sign up'}
          </LandingButton>
        </div>
      </div>
    </div>,
    document.body
  )
}

export function Navbar({
  links = DEFAULT_NAV_LINKS,
  signinLabel,
  signinHref,
  primaryCtaLabel,
  primaryCtaHref,
  logoUrl,
}: NavbarProps = {}) {
  const [mobileOpen, setMobileOpen] = useState(false)
  const [scrolled, setScrolled] = useState(false)
  const pathname = usePathname()
  const isHome = pathname === '/'

  useEffect(() => {
    let ticking = false
    const onScroll = () => {
      if (ticking) return
      ticking = true
      requestAnimationFrame(() => {
        setScrolled(window.scrollY > 40)
        ticking = false
      })
    }
    window.addEventListener('scroll', onScroll, {
      passive: true,
    })
    onScroll()
    return () => window.removeEventListener('scroll', onScroll)
  }, [])

  return (
    <>
      <div className="fixed inset-x-0 top-0 z-50 px-5 py-4">
        <nav
          aria-label="Main"
          data-scrolled={scrolled}
          className={cn(
            'navbar-outer flex items-center justify-between rounded-full px-3 py-1.5',
            'transition-all duration-[450ms] ease-[cubic-bezier(0.22,1,0.36,1)]',
            'mx-auto',
            scrolled
              ? 'max-w-[880px] border border-black/5 bg-white/70 backdrop-blur-xl supports-[not_(backdrop-filter:blur(1px))]:bg-white'
              : 'max-w-[1200px] border border-transparent bg-transparent'
          )}
        >
          {/* Logo group */}
          <Link
            href="/"
            className="shrink-0 font-logo text-xl font-extrabold tracking-[0.2px] text-[var(--pz-dark-surface)] focus-visible:rounded focus-visible:ring-2 focus-visible:ring-[var(--pz-accent)] focus-visible:ring-offset-2 focus-visible:outline-none"
          >
            {logoUrl ? (
              <Image src={logoUrl} alt="Pulzifi" width={96} height={28} className="h-7 w-auto" />
            ) : (
              'Pulzifi'
            )}
          </Link>

          {/* Nav links pill — desktop only */}
          <div
            className={cn(
              'hidden rounded-full px-4 py-1.5 transition-colors duration-[450ms] ease-[cubic-bezier(0.22,1,0.36,1)] md:flex md:items-center md:gap-1',
              scrolled ? 'bg-transparent' : 'bg-gray-100'
            )}
          >
            {links.map((link) => {
              const href = NAV_LINK_PATHS[link.label] || resolveHref(link.href, isHome)
              const active = isLinkActive(link.label, pathname)
              return (
                <Link
                  key={link.label}
                  href={href}
                  className={cn(
                    'rounded-full px-3 py-1.5 text-sm font-medium leading-5 tracking-tight text-[var(--pz-ink)] transition-colors hover:bg-black/5 focus-visible:ring-2 focus-visible:ring-[var(--pz-accent)] focus-visible:ring-offset-1 focus-visible:outline-none',
                    active && 'text-[var(--pz-accent)]'
                  )}
                >
                  {link.label}
                </Link>
              )
            })}
          </div>

          {/* CTAs pill — desktop only */}
          <div
            className={cn(
              'hidden items-center gap-2 rounded-full px-3 py-1 transition-colors duration-[450ms] ease-[cubic-bezier(0.22,1,0.36,1)] md:flex',
              scrolled ? 'bg-transparent' : 'bg-gray-100'
            )}
          >
            {(signinLabel || signinHref) && (
              <Link
                href={signinHref ?? '/login'}
                className="rounded-full px-3 py-1.5 text-sm font-medium text-[var(--pz-ink-2)] transition-colors hover:text-[var(--pz-ink)] focus-visible:ring-2 focus-visible:ring-[var(--pz-accent)] focus-visible:ring-offset-1 focus-visible:outline-none"
              >
                {signinLabel ?? 'Sign in'}
              </Link>
            )}
            {!signinLabel && !signinHref && (
              <Link
                href="/login"
                className="rounded-full px-3 py-1.5 text-sm font-medium text-[var(--pz-ink-2)] transition-colors hover:text-[var(--pz-ink)] focus-visible:ring-2 focus-visible:ring-[var(--pz-accent)] focus-visible:ring-offset-1 focus-visible:outline-none"
              >
                Sign in
              </Link>
            )}
            <LandingButton href={primaryCtaHref ?? '/register'} variant="primary" withArrow>
              {primaryCtaLabel ?? 'Sign up'}
            </LandingButton>
          </div>

          {/* Mobile toggle */}
          <button
            type="button"
            className="flex items-center justify-center rounded-full p-2 text-[var(--pz-ink)] transition-colors hover:bg-black/5 focus-visible:ring-2 focus-visible:ring-[var(--pz-accent)] focus-visible:outline-none md:hidden"
            onClick={() => setMobileOpen(true)}
            aria-label="Open menu"
            aria-expanded={mobileOpen}
          >
            <Menu className="size-6" />
          </button>
        </nav>
      </div>

      <MobileMenu
        open={mobileOpen}
        onClose={() => setMobileOpen(false)}
        isHome={isHome}
        links={links}
        signinLabel={signinLabel}
        signinHref={signinHref}
        primaryCtaLabel={primaryCtaLabel}
        primaryCtaHref={primaryCtaHref}
        logoUrl={logoUrl}
      />
    </>
  )
}

'use client'

import { useState, useEffect } from 'react'
import { createPortal } from 'react-dom'
import Link from 'next/link'
import { Menu, X } from 'lucide-react'
import { cn } from '@workspace/ui/lib/utils'
import { NAV_LINKS } from '../lib/data'
import { LandingButton } from './components/landing-button'

function MobileMenu({ open, onClose }: { open: boolean; onClose: () => void }) {
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
            className="font-logo text-xl font-extrabold tracking-[0.2px] text-[#2d1359]"
            onClick={onClose}
          >
            Pulzifi
          </Link>
          <button
            type="button"
            className="flex items-center justify-center rounded-lg p-2 text-[#121217]"
            onClick={onClose}
            aria-label="Close menu"
          >
            <X className="size-6" />
          </button>
        </div>

        {/* Links */}
        <div className="flex flex-col gap-1 pt-4">
          {NAV_LINKS.map((link) => (
            <Link
              key={link.label}
              href={link.href}
              className="rounded-xl px-4 py-3 text-base font-medium text-[#121217] transition-colors hover:bg-gray-50"
              onClick={onClose}
            >
              {link.label}
            </Link>
          ))}
        </div>

        {/* Buttons */}
        <div className="mt-3 flex gap-2 border-t border-gray-100 pt-4">
          <LandingButton href="/login" variant="outline" className="flex-1 justify-center">
            Sign in
          </LandingButton>
          <LandingButton href="/register" variant="primary" className="flex-1 justify-center">
            Sign up
          </LandingButton>
        </div>
      </div>
    </div>,
    document.body
  )
}

export function Navbar() {
  const [mobileOpen, setMobileOpen] = useState(false)

  return (
    <>
      <nav className="sticky top-3 z-50 mx-auto max-w-[1256px] rounded-[20px] bg-white/90 px-6 py-4 backdrop-blur-md shadow-sm md:px-[58px]">
        <div className="flex items-center justify-between">
          <Link
            href="/"
            className="font-logo text-xl font-extrabold tracking-[0.2px] text-[#2d1359]"
          >
            Pulzifi
          </Link>

          {/* Desktop Nav */}
          <div className="hidden items-center gap-1 md:flex">
            {NAV_LINKS.map((link) => (
              <Link
                key={link.label}
                href={link.href}
                className="rounded-lg px-3 py-2 text-sm font-medium leading-5 tracking-tight text-[#121217] transition-colors hover:bg-gray-100"
              >
                {link.label}
              </Link>
            ))}
          </div>

          <div className="hidden items-center gap-2.5 md:flex">
            <LandingButton href="/login" variant="outline">
              Sign in
            </LandingButton>
            <LandingButton href="/register" variant="primary">
              Sign up
            </LandingButton>
          </div>

          {/* Mobile toggle */}
          <button
            type="button"
            className="flex items-center justify-center rounded-lg p-2 text-[#121217] md:hidden"
            onClick={() => setMobileOpen(true)}
            aria-label="Open menu"
          >
            <Menu className="size-6" />
          </button>
        </div>
      </nav>

      <MobileMenu open={mobileOpen} onClose={() => setMobileOpen(false)} />
    </>
  )
}

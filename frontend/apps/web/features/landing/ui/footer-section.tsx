import { cn } from '@workspace/ui/lib/utils'
import type { LucideIcon } from 'lucide-react'
import { Github, Globe, Linkedin, Twitter, Youtube } from 'lucide-react'
import Image from 'next/image'
import Link from 'next/link'

type FooterLink = {
  label: string
  href: string
}
type FooterLinks = Record<string, ReadonlyArray<FooterLink>>
type SocialPlatform = 'twitter' | 'linkedin' | 'github' | 'youtube' | 'other'
type SocialLink = {
  platform: string
  href: string
}

type FooterSectionProps = {
  links?: FooterLinks
  copyrightText?: string
  tagline?: string
  socialLinks?: SocialLink[]
  logoUrl?: string
}

const SOCIAL_ICONS: Record<SocialPlatform, LucideIcon> = {
  twitter: Twitter,
  linkedin: Linkedin,
  github: Github,
  youtube: Youtube,
  other: Globe,
}

function resolveSocialIcon(platform: string): LucideIcon {
  const key = platform.toLowerCase() as SocialPlatform
  return SOCIAL_ICONS[key] ?? Globe
}

const LEGAL_LINKS: FooterLink[] = [
  {
    label: 'Privacy Policy',
    href: '#',
  },
  {
    label: 'Terms of Service',
    href: '#',
  },
]

const DEFAULT_FOOTER_LINKS: FooterLinks = {
  Features: [
    {
      label: 'Product',
      href: '#how-it-works',
    },
    {
      label: 'Use Cases',
      href: '#industries',
    },
    {
      label: 'Pricing',
      href: '#pricing',
    },
  ],
  Support: [
    {
      label: 'Help',
      href: '#',
    },
    {
      label: 'FAQ',
      href: '#faq',
    },
    {
      label: 'Contact',
      href: '/contact',
    },
  ],
  Legal: [
    {
      label: 'Privacy Policy',
      href: '#',
    },
    {
      label: 'Terms of Services',
      href: '#',
    },
  ],
}

export function FooterSection({
  links = DEFAULT_FOOTER_LINKS,
  copyrightText,
  tagline,
  socialLinks,
  logoUrl,
}: FooterSectionProps = {}) {
  const groups = Object.entries(links)
  const year = new Date().getFullYear()
  const copyright = copyrightText ?? `© ${year} Pulzifi. All rights reserved.`

  return (
    // biome-ignore lint/correctness/useUniqueElementIds: static nav anchor ID
    <footer id="footer" className="bg-white py-16 lg:py-20">
      <div className="mx-auto max-w-[1200px] px-6 lg:px-8">
        {/* 4-col grid */}
        <div
          className={cn(
            'grid grid-cols-1 gap-8 md:gap-12',
            groups.length > 0 ? 'md:grid-cols-[2fr_1fr_1fr_1fr]' : 'md:grid-cols-1'
          )}
        >
          {/* Col 1 — Brand */}
          <div className="flex flex-col gap-4">
            <div>
              {logoUrl ? (
                <Image
                  src={logoUrl}
                  alt="Pulzifi"
                  width={120}
                  height={32}
                  className="h-8 w-auto object-contain"
                />
              ) : (
                <span className="font-logo text-2xl font-extrabold tracking-tight text-[var(--pz-ink)]">
                  Pulzifi
                </span>
              )}
            </div>

            {tagline && (
              <p className="max-w-[240px] text-sm leading-relaxed text-[var(--pz-ink-2)]">
                {tagline}
              </p>
            )}

            {socialLinks && socialLinks.length > 0 && (
              <div className="mt-2 flex items-center gap-3">
                {socialLinks.map((link) => {
                  const Icon = resolveSocialIcon(link.platform)
                  return (
                    <a
                      key={link.platform}
                      href={link.href}
                      aria-label={link.platform}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-[var(--pz-ink-2)] transition-colors hover:text-[var(--pz-accent)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--pz-accent)] focus-visible:ring-offset-2"
                    >
                      <Icon className="size-5" />
                    </a>
                  )
                })}
              </div>
            )}
          </div>

          {/* Cols 2-4 — Link groups */}
          {groups.map(([heading, groupLinks]) => (
            <div key={heading} className="flex flex-col gap-4">
              <h3 className="text-xs font-semibold uppercase tracking-[0.12em] text-[var(--pz-ink)]">
                {heading}
              </h3>
              <ul className="flex flex-col gap-2.5">
                {groupLinks.map((link) => (
                  <li key={link.label}>
                    <Link
                      href={link.href}
                      className="text-sm text-[var(--pz-ink-2)] transition-colors hover:text-[var(--pz-accent)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--pz-accent)] focus-visible:ring-offset-2"
                    >
                      {link.label}
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        {/* Bottom row */}
        <div className="mt-12 flex flex-col justify-between gap-4 border-t border-[var(--pz-card-border)] pt-6 md:flex-row md:items-center">
          <p className="text-sm text-[var(--pz-ink-2)]">{copyright}</p>
          <div className="flex flex-wrap gap-4">
            {LEGAL_LINKS.map((link) => (
              <Link
                key={link.label}
                href={link.href}
                className="text-sm text-[var(--pz-ink-2)] transition-colors hover:text-[var(--pz-accent)]"
              >
                {link.label}
              </Link>
            ))}
          </div>
        </div>
      </div>
    </footer>
  )
}

import { getPayloadClient } from '@/lib/payload'
import { buildThemeStyle } from '@/lib/theme-style'

type NavLink = {
  label: string
  href: string
}
type FooterGroups = Record<
  string,
  {
    label: string
    href: string
  }[]
>
type SocialLink = {
  platform: string
  href: string
}

export type NavbarChromeProps = {
  links?: NavLink[]
  signinLabel?: string
  signinHref?: string
  primaryCtaLabel?: string
  primaryCtaHref?: string
  logoUrl?: string
}

export type FooterChromeProps = {
  links?: FooterGroups
  tagline?: string
  socialLinks?: SocialLink[]
  logoUrl?: string
}

export type CmsChrome = {
  themeStyle: string
  navbar: NavbarChromeProps
  footer: FooterChromeProps
}

/**
 * Fetches the CMS-managed navbar, footer, and theme globals and maps them into
 * ready-to-spread props for `<Navbar>` and `<FooterSection>`, plus the injectable
 * theme `<style>` string. Chrome is best-effort: any failure falls back to the
 * component defaults so a page still renders.
 *
 * Used by every landing-style page (apex, dynamic `[slug]`, blog) so they all
 * share identical header/footer/theming.
 */
export async function fetchCmsChrome(): Promise<CmsChrome> {
  let navbar: Record<string, unknown> | null = null
  let footer: Record<string, unknown> | null = null
  let theme: Record<string, unknown> | null = null

  try {
    const payload = await getPayloadClient()
    const [nav, foot, th] = await Promise.all([
      payload
        .findGlobal({
          slug: 'navbar',
          depth: 1,
        })
        .catch(() => null),
      payload
        .findGlobal({
          slug: 'footer',
          depth: 1,
        })
        .catch(() => null),
      payload
        .findGlobal({
          slug: 'theme',
          depth: 0,
        })
        .catch(() => null),
    ])
    navbar = nav as unknown as Record<string, unknown> | null
    footer = foot as unknown as Record<string, unknown> | null
    theme = th as unknown as Record<string, unknown> | null
  } catch {
    // Chrome is non-critical — fall through to defaults.
  }

  const themeStyle = theme
    ? buildThemeStyle(theme as unknown as Record<string, string | null | undefined>)
    : ''

  const navLinks = (navbar?.links as NavLink[] | undefined)?.length
    ? (navbar?.links as NavLink[])
    : undefined

  let footerGroups: FooterGroups | undefined
  if ((footer?.groups as unknown[] | undefined)?.length) {
    footerGroups = Object.fromEntries(
      (
        footer?.groups as {
          heading: string
          links: NavLink[]
        }[]
      ).map((g) => [
        g.heading,
        g.links?.map((l) => ({
          label: l.label,
          href: l.href,
        })) ?? [],
      ])
    )
  }

  const socialLinks = (footer?.socialLinks as SocialLink[] | undefined)?.length
    ? (footer?.socialLinks as SocialLink[])
    : undefined

  return {
    themeStyle,
    navbar: {
      links: navLinks,
      signinLabel: navbar?.signinLabel as string | undefined,
      signinHref: navbar?.signinHref as string | undefined,
      primaryCtaLabel: navbar?.primaryCtaLabel as string | undefined,
      primaryCtaHref: navbar?.primaryCtaHref as string | undefined,
      logoUrl: (
        navbar?.logo as
          | {
              url?: string
            }
          | undefined
      )?.url,
    },
    footer: {
      links: footerGroups,
      tagline: footer?.tagline as string | undefined,
      socialLinks,
      logoUrl: (
        footer?.logo as
          | {
              url?: string
            }
          | undefined
      )?.url,
    },
  }
}

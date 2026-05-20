import { getPayloadClient } from '@/lib/payload'
import { buildThemeStyle } from '@/lib/theme-style'
import { ContactClient } from './contact-client'

const FALLBACK = {
  headline: 'Get in touch',
  subheadline: "We're here to help. Our team responds within one business day.",
  email: 'support@pulzifi.com',
  address: 'Boise, ID',
}

export default async function ContactPage() {
  const props = {
    ...FALLBACK,
  } as Parameters<typeof ContactClient>[0]

  try {
    const payload = await getPayloadClient()
    const [contactPage, navbar, footer, theme] = await Promise.all([
      payload
        .findGlobal({
          slug: 'contact-page',
          depth: 0,
        })
        .catch(() => null),
      payload.findGlobal({
        slug: 'navbar',
        depth: 1,
      }),
      payload.findGlobal({
        slug: 'footer',
        depth: 1,
      }),
      payload
        .findGlobal({
          slug: 'theme',
          depth: 0,
        })
        .catch(() => null),
    ])

    const cp = contactPage as unknown as Record<string, unknown>
    if (cp) {
      props.headline = (cp.headline as string) ?? FALLBACK.headline
      props.subheadline = (cp.subheadline as string) ?? FALLBACK.subheadline
      props.email = (cp.email as string) ?? FALLBACK.email
      props.address = (cp.address as string) ?? FALLBACK.address
    }

    const nav = navbar as unknown as Record<string, unknown>
    props.navLinks = (nav.links as typeof props.navLinks)?.length
      ? (nav.links as typeof props.navLinks)
      : undefined
    props.navSigninLabel = nav.signinLabel as string | undefined
    props.navSigninHref = nav.signinHref as string | undefined
    props.navPrimaryCtaLabel = nav.primaryCtaLabel as string | undefined
    props.navPrimaryCtaHref = nav.primaryCtaHref as string | undefined
    const navLogo = nav.logo as Record<string, unknown> | undefined
    if (navLogo?.url) props.navLogoUrl = navLogo.url as string

    const foot = footer as unknown as Record<string, unknown>
    const footGroups = foot.groups as
      | {
          heading: string
          links: {
            label: string
            href: string
          }[]
        }[]
      | undefined
    if (footGroups?.length) {
      props.footerGroups = Object.fromEntries(
        footGroups.map((g) => [
          g.heading,
          g.links?.map((l) => ({
            label: l.label,
            href: l.href,
          })) ?? [],
        ])
      )
    }
    props.footerTagline = foot.tagline as string | undefined
    props.footerSocialLinks = (foot.socialLinks as typeof props.footerSocialLinks)?.length
      ? (foot.socialLinks as typeof props.footerSocialLinks)
      : undefined
    const footLogo = foot.logo as Record<string, unknown> | undefined
    if (footLogo?.url) props.footerLogoUrl = footLogo.url as string

    if (theme)
      props.themeStyle = buildThemeStyle(
        theme as unknown as Record<string, string | null | undefined>
      )
  } catch {
    // DB unavailable — use fallback
  }

  return <ContactClient {...props} />
}

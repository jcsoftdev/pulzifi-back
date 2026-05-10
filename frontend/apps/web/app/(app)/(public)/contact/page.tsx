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
  let props = { ...FALLBACK } as Parameters<typeof ContactClient>[0]

  try {
    const payload = await getPayloadClient()
    const [contactPage, navbar, footer, theme] = await Promise.all([
      payload.findGlobal({ slug: 'contact-page', depth: 0 }).catch(() => null),
      payload.findGlobal({ slug: 'navbar', depth: 1 }),
      payload.findGlobal({ slug: 'footer', depth: 1 }),
      payload.findGlobal({ slug: 'theme', depth: 0 }).catch(() => null),
    ])

    const cp = contactPage as any
    if (cp) {
      props.headline = cp.headline ?? FALLBACK.headline
      props.subheadline = cp.subheadline ?? FALLBACK.subheadline
      props.email = cp.email ?? FALLBACK.email
      props.address = cp.address ?? FALLBACK.address
    }

    const nav = navbar as any
    props.navLinks = nav.links?.length ? nav.links : undefined
    props.navSigninLabel = nav.signinLabel ?? undefined
    props.navSigninHref = nav.signinHref ?? undefined
    props.navPrimaryCtaLabel = nav.primaryCtaLabel ?? undefined
    props.navPrimaryCtaHref = nav.primaryCtaHref ?? undefined
    if (nav.logo?.url) props.navLogoUrl = nav.logo.url

    const foot = footer as any
    if (foot.groups?.length) {
      props.footerGroups = Object.fromEntries(
        foot.groups.map((g: any) => [g.heading, g.links?.map((l: any) => ({ label: l.label, href: l.href })) ?? []])
      )
    }
    props.footerTagline = foot.tagline ?? undefined
    props.footerSocialLinks = foot.socialLinks?.length ? foot.socialLinks : undefined
    if (foot.logo?.url) props.footerLogoUrl = foot.logo.url

    if (theme) props.themeStyle = buildThemeStyle(theme as unknown as Record<string, string | null | undefined>)
  } catch {
    // DB unavailable — use fallback
  }

  return <ContactClient {...props} />
}

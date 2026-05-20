import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { BlocksRenderer } from '@/features/cms'
import { FooterSection, Navbar, SmoothScroll } from '@/features/landing'
import { getPayloadClient } from '@/lib/payload'
import { buildThemeStyle } from '@/lib/theme-style'

export const dynamic = 'force-dynamic'

type Props = {
  params: Promise<{
    slug: string
  }>
}

export async function generateStaticParams() {
  if (process.env.NEXT_PHASE === 'phase-production-build') return []
  try {
    const payload = await getPayloadClient()
    const pages = await payload.find({
      collection: 'pages',
      where: {
        _status: {
          equals: 'published',
        },
      },
      select: {
        slug: true,
      },
      limit: 1000,
    })
    return pages.docs.map((page: { slug?: string | null }) => ({
      slug: page.slug,
    }))
  } catch {
    return []
  }
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params
  try {
    const payload = await getPayloadClient()
    const result = await payload.find({
      collection: 'pages',
      where: {
        slug: {
          equals: slug,
        },
        _status: {
          equals: 'published',
        },
      },
      limit: 1,
    })
    const page = result.docs[0] as
      | {
          title?: string
          meta?: {
            title?: string
            description?: string
          }
        }
      | undefined
    if (!page) return {}
    return {
      title: page.meta?.title ?? page.title,
      description: page.meta?.description ?? undefined,
    }
  } catch {
    return {}
  }
}

export default async function DynamicPage({ params }: Props) {
  const { slug } = await params
  let page: Record<string, unknown> | null = null
  let navbar: Record<string, unknown> | null = null
  let footer: Record<string, unknown> | null = null
  let theme: Record<string, unknown> | null = null
  try {
    const payload = await getPayloadClient()
    const [result, nav, foot, th] = await Promise.all([
      payload.find({
        collection: 'pages',
        where: {
          slug: {
            equals: slug,
          },
          _status: {
            equals: 'published',
          },
        },
        limit: 1,
        depth: 2,
      }),
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
    page = (result.docs[0] as unknown as Record<string, unknown> | undefined) ?? null
    navbar = nav as unknown as Record<string, unknown> | null
    footer = foot as unknown as Record<string, unknown> | null
    theme = th as unknown as Record<string, unknown> | null
  } catch {
    notFound()
  }
  if (!page) notFound()

  const nav = navbar as Record<string, unknown> | null
  const foot = footer as Record<string, unknown> | null
  const themeStyle = theme
    ? buildThemeStyle(theme as unknown as Record<string, string | null | undefined>)
    : ''

  const navLinks = (
    nav?.links as
      | {
          label: string
          href: string
        }[]
      | undefined
  )?.length
    ? (nav?.links as {
        label: string
        href: string
      }[])
    : undefined
  let footerGroups:
    | Record<
        string,
        {
          label: string
          href: string
        }[]
      >
    | undefined
  if ((foot?.groups as unknown[] | undefined)?.length) {
    footerGroups = Object.fromEntries(
      (
        // biome-ignore lint/style/noNonNullAssertion: narrowed by length check above
        foot!.groups as {
          heading: string
          links: {
            label: string
            href: string
          }[]
        }[]
      ).map((g) => [
        g.heading,
        g.links?.map((l: { label: string; href: string }) => ({
          label: l.label,
          href: l.href,
        })) ?? [],
      ])
    )
  }

  return (
    <div className="flex min-h-screen flex-col bg-[var(--pz-page-bg)]">
      {themeStyle && (
        // biome-ignore lint/security/noDangerouslySetInnerHtml: intentional theme CSS injection — controlled server-side input
        <style dangerouslySetInnerHTML={{ __html: themeStyle }} />
      )}
      <SmoothScroll />
      <Navbar
        links={navLinks}
        signinLabel={nav?.signinLabel as string | undefined}
        signinHref={nav?.signinHref as string | undefined}
        primaryCtaLabel={nav?.primaryCtaLabel as string | undefined}
        primaryCtaHref={nav?.primaryCtaHref as string | undefined}
        logoUrl={(nav?.logo as { url?: string } | undefined)?.url}
      />
      <main className="flex flex-1 flex-col">
        <BlocksRenderer blocks={(page.blocks as { blockType: string; id?: string; [key: string]: unknown }[]) ?? []} />
      </main>
      <FooterSection
        links={footerGroups}
        tagline={foot?.tagline as string | undefined}
        socialLinks={(foot?.socialLinks as { platform: string; href: string }[] | undefined)?.length ? (foot?.socialLinks as { platform: string; href: string }[]) : undefined}
        logoUrl={(foot?.logo as { url?: string } | undefined)?.url}
      />
    </div>
  )
}

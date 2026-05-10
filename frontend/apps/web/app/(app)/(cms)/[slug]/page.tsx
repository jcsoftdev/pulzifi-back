import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { BlocksRenderer } from '@/features/cms'
import { FooterSection, Navbar, SmoothScroll } from '@/features/landing'
import { getPayloadClient } from '@/lib/payload'
import { buildThemeStyle } from '@/lib/theme-style'

export const dynamic = 'force-dynamic'

type Props = {
  params: Promise<{ slug: string }>
}

export async function generateStaticParams() {
  if (process.env.NEXT_PHASE === 'phase-production-build') return []
  try {
    const payload = await getPayloadClient()
    const pages = await payload.find({
      collection: 'pages',
      where: { _status: { equals: 'published' } },
      select: { slug: true },
      limit: 1000,
    })
    return pages.docs.map((page: any) => ({ slug: page.slug }))
  } catch {
    return []
  }
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params
  const payload = await getPayloadClient()
  const result = await payload.find({
    collection: 'pages',
    where: { slug: { equals: slug }, _status: { equals: 'published' } },
    limit: 1,
  })
  const page = result.docs[0] as any
  if (!page) return {}
  return {
    title: page.meta?.title ?? page.title,
    description: page.meta?.description ?? undefined,
  }
}

export default async function DynamicPage({ params }: Props) {
  const { slug } = await params
  const payload = await getPayloadClient()
  const [result, navbar, footer, theme] = await Promise.all([
    payload.find({
      collection: 'pages',
      where: { slug: { equals: slug }, _status: { equals: 'published' } },
      limit: 1,
      depth: 2,
    }),
    payload.findGlobal({ slug: 'navbar', depth: 1 }).catch(() => null),
    payload.findGlobal({ slug: 'footer', depth: 1 }).catch(() => null),
    payload.findGlobal({ slug: 'theme', depth: 0 }).catch(() => null),
  ])
  const page = result.docs[0] as any
  if (!page) notFound()

  const nav = navbar as any
  const foot = footer as any
  const themeStyle = theme
    ? buildThemeStyle(theme as unknown as Record<string, string | null | undefined>)
    : ''

  const navLinks = (nav?.links as { label: string; href: string }[] | undefined)?.length
    ? (nav.links as { label: string; href: string }[])
    : undefined
  let footerGroups: Record<string, { label: string; href: string }[]> | undefined
  if (foot?.groups?.length) {
    footerGroups = Object.fromEntries(
      (foot.groups as { heading: string; links: { label: string; href: string }[] }[]).map((g) => [
        g.heading,
        g.links?.map((l: { label: string; href: string }) => ({ label: l.label, href: l.href })) ?? [],
      ]),
    )
  }

  return (
    <div className="flex min-h-screen flex-col bg-[var(--pz-page-bg)]">
      {themeStyle && <style dangerouslySetInnerHTML={{ __html: themeStyle }} />}
      <SmoothScroll />
      <Navbar
        links={navLinks}
        signinLabel={nav?.signinLabel ?? undefined}
        signinHref={nav?.signinHref ?? undefined}
        primaryCtaLabel={nav?.primaryCtaLabel ?? undefined}
        primaryCtaHref={nav?.primaryCtaHref ?? undefined}
        logoUrl={nav?.logo?.url ?? undefined}
      />
      <main className="flex flex-1 flex-col">
        <BlocksRenderer blocks={page.blocks ?? []} />
      </main>
      <FooterSection
        links={footerGroups}
        tagline={foot?.tagline ?? undefined}
        socialLinks={foot?.socialLinks?.length ? foot.socialLinks : undefined}
        logoUrl={foot?.logo?.url ?? undefined}
      />
    </div>
  )
}

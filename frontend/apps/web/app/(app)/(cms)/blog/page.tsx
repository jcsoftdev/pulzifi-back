import type { Metadata } from 'next'
import { JsonLd } from '@/components/json-ld'
import type { BlogPostItem } from '@/features/cms'
import { BlogList } from '@/features/cms'
import { FooterSection, Navbar, SmoothScrollLazy } from '@/features/landing'
import { fetchCmsChrome } from '@/lib/cms-chrome'
import { getPayloadClient } from '@/lib/payload'

type PostDoc = Record<string, unknown>

function toBlogPostItem(post: PostDoc): BlogPostItem {
  const heroImage =
    typeof post.heroImage === 'object' && post.heroImage !== null
      ? (post.heroImage as Record<string, unknown>)
      : null
  return {
    id: post.id as string,
    title: post.title as string,
    slug: post.slug as string,
    excerpt: post.excerpt as string | undefined,
    heroImageUrl: heroImage?.url as string | undefined,
    publishedAt: post.publishedAt as string | undefined,
    category: post.category as string | undefined,
    author: post.author as string | undefined,
    readingTime: post.readingTime as number | undefined,
  }
}

// Dynamic on purpose: this route has no params, so ISR would prerender it AT
// BUILD TIME, where Payload/Postgres is unreachable — the empty "No posts yet"
// shell would then be cached for the whole revalidate window after every
// deploy. Detail pages keep ISR because they generate on demand (at runtime,
// with the DB available). The Payload local query here is a few ms.
export const dynamic = 'force-dynamic'

const BLOG_TITLE = 'Website Monitoring Blog — Insights & Guides | Pulzifi'
const BLOG_DESCRIPTION =
  'Expert guides on website change monitoring, competitive intelligence, and AI-powered tracking. Stay ahead with Pulzifi product updates and industry insights.'

export const metadata: Metadata = {
  title: BLOG_TITLE,
  description: BLOG_DESCRIPTION,
  alternates: {
    canonical: '/blog',
  },
  openGraph: {
    type: 'website',
    title: BLOG_TITLE,
    description: BLOG_DESCRIPTION,
    url: '/blog',
  },
  twitter: {
    card: 'summary_large_image',
    title: BLOG_TITLE,
    description: BLOG_DESCRIPTION,
  },
  robots: {
    index: true,
    follow: true,
  },
}

export default async function BlogIndexPage() {
  const siteUrl = process.env.NEXT_PUBLIC_APP_BASE_URL || 'https://pulzifi.com'

  // Payload is unavailable at build time (getPayloadClient throws). Fall back to
  // an empty list so the static shell prerenders; ISR (revalidate) then fills in
  // real posts on the first runtime revalidation.
  let posts = { docs: [] as PostDoc[] }
  let chrome = {
    themeStyle: '',
    navbar: {},
    footer: {},
  } as Awaited<ReturnType<typeof fetchCmsChrome>>
  try {
    const payload = await getPayloadClient()
    const [postsResult, cmsChrome] = await Promise.all([
      payload.find({
        collection: 'posts',
        where: {
          _status: {
            equals: 'published',
          },
        },
        sort: '-publishedAt',
        limit: 50,
        depth: 1,
      }),
      fetchCmsChrome(),
    ])
    posts = postsResult as unknown as { docs: PostDoc[] }
    chrome = cmsChrome
  } catch {
    // build-time fallback — keep defaults
  }

  const blogJsonLd: Record<string, unknown> = {
    '@context': 'https://schema.org',
    '@type': 'Blog',
    name: BLOG_TITLE,
    description: BLOG_DESCRIPTION,
    url: `${siteUrl}/blog`,
  }

  const itemListJsonLd: Record<string, unknown> = {
    '@context': 'https://schema.org',
    '@type': 'ItemList',
    name: BLOG_TITLE,
    itemListElement: (posts.docs as unknown as PostDoc[]).map((post, index) => ({
      '@type': 'ListItem',
      position: index + 1,
      url: `${siteUrl}/blog/${post.slug as string}`,
      name: post.title as string,
    })),
  }

  return (
    <div className="min-h-screen bg-[var(--pz-page-bg)]">
      {chrome.themeStyle && (
        <style
          // biome-ignore lint/security/noDangerouslySetInnerHtml: intentional theme CSS injection — controlled server-side input
          dangerouslySetInnerHTML={{
            __html: chrome.themeStyle,
          }}
        />
      )}
      <SmoothScrollLazy />
      <Navbar {...chrome.navbar} />
      <main>
        <div className="mx-auto max-w-[1280px] px-4 pb-16 pt-28 sm:px-6">
          <header className="mb-12 max-w-2xl">
            <p className="mb-3 text-xs font-semibold uppercase tracking-[0.2em] text-[var(--pz-accent)]">
              Insights & Guides
            </p>
            <h1 className="pz-display mb-4 text-5xl leading-tight text-[var(--pz-ink)]">Blog</h1>
            <p className="text-lg text-[var(--pz-ink-2)]">
              Insights, product updates, and guides from the Pulzifi team.
            </p>
          </header>
          <BlogList posts={(posts.docs as unknown as PostDoc[]).map(toBlogPostItem)} />
        </div>
      </main>
      <FooterSection {...chrome.footer} />
      <JsonLd data={blogJsonLd} />
      <JsonLd data={itemListJsonLd} />
    </div>
  )
}

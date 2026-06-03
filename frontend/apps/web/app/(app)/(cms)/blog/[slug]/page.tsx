import { RichText } from '@payloadcms/richtext-lexical/react'
import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { FooterSection, Navbar } from '@/features/landing'
import { JsonLd } from '@/components/json-ld'
import { getPayloadClient } from '@/lib/payload'

type PostDoc = Record<string, unknown>

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
    const posts = await payload.find({
      collection: 'posts',
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
    return (posts.docs as unknown as PostDoc[]).map((post) => ({
      slug: post.slug,
    }))
  } catch {
    return []
  }
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params
  const siteUrl = process.env.NEXT_PUBLIC_APP_BASE_URL || 'https://pulzifi.com'
  try {
    const payload = await getPayloadClient()
    const result = await payload.find({
      collection: 'posts',
      where: {
        slug: { equals: slug },
        _status: { equals: 'published' },
      },
      limit: 1,
      depth: 1,
    })
    const post = result.docs[0] as unknown as PostDoc
    if (!post) return {}

    const meta =
      typeof post.meta === 'object' && post.meta !== null
        ? (post.meta as Record<string, unknown>)
        : null

    const title = (meta?.title ?? post.title) as string | undefined
    const description = (meta?.description ?? post.excerpt) as string | undefined
    const heroImageObj =
      typeof post.heroImage === 'object' && post.heroImage !== null
        ? (post.heroImage as Record<string, unknown>)
        : null
    const heroImageUrl = heroImageObj?.url as string | undefined
    const publishedAt = post.publishedAt as string | undefined
    const author = post.author as string | undefined

    return {
      title,
      description,
      alternates: {
        canonical: `${siteUrl}/blog/${slug}`,
      },
      openGraph: {
        type: 'article',
        title: title ?? undefined,
        description: description ?? undefined,
        url: `${siteUrl}/blog/${slug}`,
        siteName: 'Pulzifi',
        ...(publishedAt ? { publishedTime: publishedAt } : {}),
        ...(author ? { authors: [author] } : {}),
        ...(heroImageUrl
          ? {
              images: [
                {
                  url: heroImageUrl,
                  width: 1200,
                  height: 630,
                  alt: title ?? 'Blog post hero image',
                },
              ],
            }
          : {}),
      },
      twitter: {
        card: 'summary_large_image',
        title: title ?? undefined,
        description: description ?? undefined,
        ...(heroImageUrl ? { images: [heroImageUrl] } : {}),
      },
    }
  } catch {
    return {}
  }
}

export default async function BlogArticlePage({ params }: Props) {
  const { slug } = await params
  const siteUrl = process.env.NEXT_PUBLIC_APP_BASE_URL || 'https://pulzifi.com'

  let post: PostDoc | null = null
  try {
    const payload = await getPayloadClient()
    const result = await payload.find({
      collection: 'posts',
      where: {
        slug: { equals: slug },
        _status: { equals: 'published' },
      },
      limit: 1,
      depth: 2,
    })
    post = result.docs[0] as unknown as PostDoc
  } catch {
    notFound()
  }
  if (!post) notFound()

  const heroImageObj =
    typeof post.heroImage === 'object' && post.heroImage !== null
      ? (post.heroImage as Record<string, unknown>)
      : null
  const heroImageUrl = heroImageObj?.url as string | null | undefined

  const postUrl = `${siteUrl}/blog/${slug}`
  const publishedAt = post.publishedAt as string | undefined
  const updatedAt = (post.updatedAt ?? post.publishedAt) as string | undefined
  const author = post.author as string | undefined
  const title = post.title as string

  const blogPostingJsonLd: Record<string, unknown> = {
    '@context': 'https://schema.org',
    '@type': 'BlogPosting',
    headline: title,
    description: post.excerpt as string | undefined,
    url: postUrl,
    ...(publishedAt ? { datePublished: publishedAt } : {}),
    ...(updatedAt ? { dateModified: updatedAt } : {}),
    ...(author
      ? {
          author: {
            '@type': 'Person',
            name: author,
          },
        }
      : {}),
    publisher: {
      '@type': 'Organization',
      '@id': `${siteUrl}/#organization`,
      name: 'Pulzifi',
      logo: {
        '@type': 'ImageObject',
        url: `${siteUrl}/icon.png`,
      },
    },
    ...(heroImageUrl
      ? {
          image: {
            '@type': 'ImageObject',
            url: heroImageUrl,
          },
        }
      : {}),
    mainEntityOfPage: {
      '@type': 'WebPage',
      '@id': postUrl,
    },
  }

  const breadcrumbJsonLd = {
    '@context': 'https://schema.org',
    '@type': 'BreadcrumbList',
    itemListElement: [
      {
        '@type': 'ListItem',
        position: 1,
        name: 'Home',
        item: siteUrl,
      },
      {
        '@type': 'ListItem',
        position: 2,
        name: 'Blog',
        item: `${siteUrl}/blog`,
      },
      {
        '@type': 'ListItem',
        position: 3,
        name: title,
        item: postUrl,
      },
    ],
  }

  return (
    <div className="min-h-screen bg-[var(--pz-page-bg)]">
      <div className="mx-auto max-w-[1280px] space-y-3 p-3">
        <Navbar />
        <main>
          <article className="rounded-3xl bg-white px-6 py-12 sm:px-12 lg:px-20">
            {/* Hero image */}
            {heroImageUrl && (
              // biome-ignore lint/performance/noImgElement: external CMS image, dimensions unknown at build time
              <img
                src={heroImageUrl}
                alt={title}
                className="mb-10 h-64 w-full rounded-2xl object-cover sm:h-96"
              />
            )}

            {/* Article header */}
            <div className="mx-auto max-w-2xl">
              {(post.category as string | undefined) && (
                <span className="mb-4 inline-block rounded-full bg-[var(--pz-accent-tint)] px-3 py-1 text-xs font-medium text-[var(--pz-accent)]">
                  {post.category as string}
                </span>
              )}
              <h1 className="pz-display mb-4 text-3xl text-[var(--pz-ink)] sm:text-4xl lg:text-5xl">
                {title}
              </h1>
              <div className="mb-10 flex flex-wrap items-center gap-3 text-sm text-[var(--pz-ink-2)]">
                {author && <span className="font-medium">{author}</span>}
                {author && publishedAt && (
                  <span aria-hidden="true" className="opacity-40">
                    ·
                  </span>
                )}
                {publishedAt && (
                  <time dateTime={publishedAt}>
                    {new Date(publishedAt).toLocaleDateString('en-US', {
                      year: 'numeric',
                      month: 'long',
                      day: 'numeric',
                    })}
                  </time>
                )}
              </div>
            </div>

            {/* Article body */}
            {(post.content as object | undefined) && (
              <div className="prose prose-lg mx-auto max-w-2xl">
                {/* biome-ignore lint/suspicious/noExplicitAny: RichText requires SerializedEditorState — content comes from Payload CMS */}
                <RichText data={post.content as any} />
              </div>
            )}
          </article>
        </main>
        <FooterSection />
      </div>

      <JsonLd data={blogPostingJsonLd} />
      <JsonLd data={breadcrumbJsonLd} />
    </div>
  )
}

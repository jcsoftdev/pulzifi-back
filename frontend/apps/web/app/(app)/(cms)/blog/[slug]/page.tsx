import { RichText } from '@payloadcms/richtext-lexical/react'
import type { Metadata } from 'next'
import Image from 'next/image'
import Link from 'next/link'
import { notFound } from 'next/navigation'
import { JsonLd } from '@/components/json-ld'
import { categoryStyle, formatBlogDate } from '@/features/cms'
import { FooterSection, Navbar, SmoothScrollLazy } from '@/features/landing'
import { fetchCmsChrome } from '@/lib/cms-chrome'
import { getPayloadClient } from '@/lib/payload'

type PostDoc = Record<string, unknown>

export const revalidate = 3600

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
        slug: {
          equals: slug,
        },
        _status: {
          equals: 'published',
        },
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
    const metaImageObj =
      typeof meta?.image === 'object' && meta.image !== null
        ? (meta.image as Record<string, unknown>)
        : null
    // Share image prefers the dedicated SEO/OG image (meta.image), falling back
    // to the article hero so posts without a custom OG image still get one.
    const heroImageUrl = (metaImageObj?.url ?? heroImageObj?.url) as string | undefined
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
        ...(publishedAt
          ? {
              publishedTime: publishedAt,
            }
          : {}),
        ...(author
          ? {
              authors: [
                author,
              ],
            }
          : {}),
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
        site: '@pulzifi',
        creator: '@pulzifi',
        title: title ?? undefined,
        description: description ?? undefined,
        ...(heroImageUrl
          ? {
              images: [
                heroImageUrl,
              ],
            }
          : {}),
      },
      robots: {
        index: true,
        follow: true,
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
  // fetchCmsChrome never throws (best-effort), so it's safe inside this try.
  let chrome = {
    themeStyle: '',
    navbar: {},
    footer: {},
  } as Awaited<ReturnType<typeof fetchCmsChrome>>
  try {
    const payload = await getPayloadClient()
    const [result, cmsChrome] = await Promise.all([
      payload.find({
        collection: 'posts',
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
      fetchCmsChrome(),
    ])
    post = result.docs[0] as unknown as PostDoc
    chrome = cmsChrome
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
  const category = post.category as string | undefined
  const readingTime = post.readingTime as number | undefined
  const publishedDate = formatBlogDate(publishedAt)
  const pillStyle = categoryStyle(category)

  const blogPostingJsonLd: Record<string, unknown> = {
    '@context': 'https://schema.org',
    '@type': 'BlogPosting',
    headline: title,
    description: post.excerpt as string | undefined,
    url: postUrl,
    ...(publishedAt
      ? {
          datePublished: publishedAt,
        }
      : {}),
    ...(updatedAt
      ? {
          dateModified: updatedAt,
        }
      : {}),
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
      <div className="mx-auto max-w-[1280px] px-4 pb-16 pt-28 sm:px-6">
        <main>
          <article className="overflow-hidden rounded-3xl bg-[var(--pz-card-bg)] shadow-[var(--pz-card-shadow-rest)] ring-1 ring-[var(--pz-card-border)]">
            {/* Article header */}
            <div className="px-6 pt-12 sm:px-12 lg:px-20">
              <div className="mx-auto max-w-2xl">
                <Link
                  href="/blog"
                  className="mb-6 inline-flex items-center gap-1 text-sm font-medium text-[var(--pz-ink-2)] transition-colors hover:text-[var(--pz-accent)]"
                >
                  <span aria-hidden="true">←</span> Back to blog
                </Link>
                {category && (
                  <span
                    className={`mb-4 inline-block rounded-full px-3 py-1 text-xs font-medium ${pillStyle.pill}`}
                  >
                    {category}
                  </span>
                )}
                <h1 className="pz-display mb-4 text-3xl leading-tight text-[var(--pz-ink)] sm:text-4xl lg:text-5xl">
                  {title}
                </h1>
                <div className="flex flex-wrap items-center gap-x-2 gap-y-1 text-sm text-[var(--pz-ink-2)]">
                  {author && <span className="font-medium">{author}</span>}
                  {author && readingTime ? <span aria-hidden="true">·</span> : null}
                  {readingTime ? <span>{readingTime} min read</span> : null}
                  {publishedDate && (readingTime || author) ? (
                    <span aria-hidden="true">·</span>
                  ) : null}
                  {publishedDate && <time dateTime={publishedAt}>{publishedDate}</time>}
                </div>
              </div>
            </div>

            {/* Hero image */}
            {heroImageUrl && (
              <div className="px-6 pt-10 sm:px-12 lg:px-20">
                <div className="relative mx-auto h-64 w-full max-w-3xl sm:h-96">
                  <Image
                    src={heroImageUrl}
                    alt={title ?? 'Blog post hero image'}
                    fill
                    priority
                    sizes="(max-width: 768px) 100vw, 768px"
                    className="rounded-2xl object-cover"
                  />
                </div>
              </div>
            )}

            {/* Article body */}
            {(post.content as object | undefined) && (
              <div className="px-6 py-12 sm:px-12 lg:px-20">
                <div className="prose prose-lg mx-auto max-w-2xl">
                  {/* biome-ignore lint/suspicious/noExplicitAny: RichText requires SerializedEditorState — content comes from Payload CMS */}
                  <RichText data={post.content as any} />
                </div>
              </div>
            )}
          </article>
        </main>
      </div>
      <FooterSection {...chrome.footer} />

      <JsonLd data={blogPostingJsonLd} />
      <JsonLd data={breadcrumbJsonLd} />
    </div>
  )
}

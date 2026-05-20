import { RichText } from '@payloadcms/richtext-lexical/react'
import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { FooterSection, Navbar } from '@/features/landing'
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
    })
    const post = result.docs[0] as unknown as PostDoc
    if (!post) return {}
    const meta =
      typeof post.meta === 'object' && post.meta !== null
        ? (post.meta as Record<string, unknown>)
        : null
    return {
      title: (meta?.title ?? post.title) as string | undefined,
      description: (meta?.description ?? post.excerpt) as string | undefined,
    }
  } catch {
    return {}
  }
}

export default async function BlogArticlePage({ params }: Props) {
  const { slug } = await params
  let post: PostDoc | null = null
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

  return (
    <div className="min-h-screen bg-[#f3f3f3]">
      <div className="mx-auto max-w-[1280px] space-y-3 p-3">
        <Navbar />
        <main>
          <article className="rounded-3xl bg-white px-8 py-12">
            {heroImageUrl && (
              // biome-ignore lint/performance/noImgElement: external CMS image, dimensions unknown at build time
              <img
                src={heroImageUrl}
                alt={post.title as string}
                className="mb-8 h-64 w-full rounded-2xl object-cover sm:h-96"
              />
            )}
            {(post.category as string | undefined) && (
              <span className="mb-3 inline-block rounded-full bg-purple-100 px-3 py-1 text-xs font-medium text-purple-700">
                {post.category as string}
              </span>
            )}
            <h1 className="mb-4 text-4xl font-bold text-gray-900">{post.title as string}</h1>
            <div className="mb-8 flex items-center gap-4 text-sm text-gray-500">
              {(post.author as string | undefined) && <span>{post.author as string}</span>}
              {(post.publishedAt as string | undefined) && (
                <span>{new Date(post.publishedAt as string).toLocaleDateString()}</span>
              )}
            </div>
            {(post.content as object | undefined) && (
              <div className="prose max-w-none">
                {/* biome-ignore lint/suspicious/noExplicitAny: RichText requires SerializedEditorState — content comes from Payload CMS */}
                <RichText data={post.content as any} />
              </div>
            )}
          </article>
        </main>
        <FooterSection />
      </div>
    </div>
  )
}

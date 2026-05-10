import { RichText } from '@payloadcms/richtext-lexical/react'
import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { FooterSection, Navbar } from '@/features/landing'
import { getPayloadClient } from '@/lib/payload'

export const revalidate = 60

type Props = {
  params: Promise<{ slug: string }>
}

export async function generateStaticParams() {
  try {
    const payload = await getPayloadClient()
    const posts = await payload.find({
      collection: 'posts',
      where: { _status: { equals: 'published' } },
      select: { slug: true },
      limit: 1000,
    })
    return posts.docs.map((post: any) => ({ slug: post.slug }))
  } catch {
    return []
  }
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params
  const payload = await getPayloadClient()
  const result = await payload.find({
    collection: 'posts',
    where: { slug: { equals: slug }, _status: { equals: 'published' } },
    limit: 1,
  })
  const post = result.docs[0] as any
  if (!post) return {}
  return {
    title: post.meta?.title ?? post.title,
    description: post.meta?.description ?? post.excerpt ?? undefined,
  }
}

export default async function BlogArticlePage({ params }: Props) {
  const { slug } = await params
  const payload = await getPayloadClient()
  const result = await payload.find({
    collection: 'posts',
    where: { slug: { equals: slug }, _status: { equals: 'published' } },
    limit: 1,
    depth: 2,
  })
  const post = result.docs[0] as any
  if (!post) notFound()

  const heroImageUrl =
    typeof post.heroImage === 'object' ? (post.heroImage?.url ?? null) : null

  return (
    <div className="min-h-screen bg-[#f3f3f3]">
      <div className="mx-auto max-w-[1280px] space-y-3 p-3">
        <Navbar />
        <main>
          <article className="rounded-3xl bg-white px-8 py-12">
            {heroImageUrl && (
              <img
                src={heroImageUrl}
                alt={post.title}
                className="mb-8 h-64 w-full rounded-2xl object-cover sm:h-96"
              />
            )}
            {post.category && (
              <span className="mb-3 inline-block rounded-full bg-purple-100 px-3 py-1 text-xs font-medium text-purple-700">
                {post.category}
              </span>
            )}
            <h1 className="mb-4 text-4xl font-bold text-gray-900">{post.title}</h1>
            <div className="mb-8 flex items-center gap-4 text-sm text-gray-500">
              {post.author && <span>{post.author}</span>}
              {post.publishedAt && (
                <span>{new Date(post.publishedAt).toLocaleDateString()}</span>
              )}
            </div>
            {post.content && (
              <div className="prose max-w-none">
                <RichText data={post.content} />
              </div>
            )}
          </article>
        </main>
        <FooterSection />
      </div>
    </div>
  )
}

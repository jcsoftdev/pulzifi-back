import type { Metadata } from 'next'
import { BlogCard } from '@/features/cms'
import { FooterSection, Navbar } from '@/features/landing'
import { getPayloadClient } from '@/lib/payload'

type PostDoc = Record<string, unknown>

export const dynamic = 'force-dynamic'

export const metadata: Metadata = {
  title: 'Blog — Pulzifi',
  description: 'Insights, product updates, and guides from the Pulzifi team.',
}

export default async function BlogIndexPage() {
  const payload = await getPayloadClient()
  const posts = await payload.find({
    collection: 'posts',
    where: {
      _status: {
        equals: 'published',
      },
    },
    sort: '-publishedAt',
    limit: 50,
    depth: 1,
  })

  return (
    <div className="min-h-screen bg-[#f3f3f3]">
      <div className="mx-auto max-w-[1280px] space-y-3 p-3">
        <Navbar />
        <main>
          <div className="rounded-3xl bg-white px-8 py-12">
            <h1 className="mb-8 text-4xl font-bold text-gray-900">Blog</h1>
            <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
              {(posts.docs as unknown as PostDoc[]).map((post) => (
                <BlogCard
                  key={post.id as string}
                  title={post.title as string}
                  slug={post.slug as string}
                  excerpt={post.excerpt as string | undefined}
                  heroImageUrl={
                    typeof post.heroImage === 'object' && post.heroImage !== null
                      ? ((post.heroImage as Record<string, unknown>).url as string | undefined)
                      : undefined
                  }
                  publishedAt={post.publishedAt as string | undefined}
                  category={post.category as string | undefined}
                />
              ))}
            </div>
            {posts.docs.length === 0 && <p className="text-gray-400">No posts published yet.</p>}
          </div>
        </main>
        <FooterSection />
      </div>
    </div>
  )
}

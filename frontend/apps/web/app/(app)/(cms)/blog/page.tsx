import type { Metadata } from 'next'
import { BlogCard } from '@/features/cms'
import { FooterSection, Navbar } from '@/features/landing'
import { getPayloadClient } from '@/lib/payload'

export const dynamic = 'force-dynamic'

export const metadata: Metadata = {
  title: 'Blog — Pulzifi',
  description: 'Insights, product updates, and guides from the Pulzifi team.',
}

export default async function BlogIndexPage() {
  const payload = await getPayloadClient()
  const posts = await payload.find({
    collection: 'posts',
    where: { _status: { equals: 'published' } },
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
              {posts.docs.map((post: any) => (
                <BlogCard
                  key={post.id}
                  title={post.title}
                  slug={post.slug}
                  excerpt={post.excerpt ?? undefined}
                  heroImageUrl={
                    typeof post.heroImage === 'object'
                      ? (post.heroImage?.url ?? undefined)
                      : undefined
                  }
                  publishedAt={post.publishedAt ?? undefined}
                  category={post.category ?? undefined}
                />
              ))}
            </div>
            {posts.docs.length === 0 && (
              <p className="text-gray-400">No posts published yet.</p>
            )}
          </div>
        </main>
        <FooterSection />
      </div>
    </div>
  )
}

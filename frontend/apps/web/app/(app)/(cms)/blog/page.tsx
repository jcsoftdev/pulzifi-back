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
    <div className="min-h-screen bg-[var(--pz-page-bg)]">
      <div className="mx-auto max-w-[1280px] space-y-3 p-3">
        <Navbar />
        <main>
          <div className="rounded-3xl bg-[var(--pz-card-bg)] px-8 py-12">
            <h1 className="pz-display mb-2 text-4xl text-[var(--pz-ink)]">Blog</h1>
            <p className="mb-10 text-[var(--pz-ink-2)]">
              Insights, product updates, and guides from the Pulzifi team.
            </p>
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
            {posts.docs.length === 0 && (
              <p className="text-[var(--pz-ink-2)] opacity-50">No posts published yet.</p>
            )}
          </div>
        </main>
        <FooterSection />
      </div>
    </div>
  )
}

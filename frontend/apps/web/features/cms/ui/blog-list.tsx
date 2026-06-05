'use client'

import { useMemo, useState } from 'react'
import type { BlogPostItem } from '../lib/blog-post'
import { BlogCard } from './blog-card'
import { BlogFeaturedCard } from './blog-featured-card'

type Props = {
  posts: BlogPostItem[]
}

const ALL = 'All'

export function BlogList({ posts }: Props) {
  const [active, setActive] = useState<string>(ALL)

  // Category tabs: All + only the categories actually present, in post order.
  const categories = useMemo(() => {
    const seen: string[] = []
    for (const p of posts) {
      if (p.category && !seen.includes(p.category)) seen.push(p.category)
    }
    return [
      ALL,
      ...seen,
    ]
  }, [
    posts,
  ])

  const filtered = useMemo(
    () => (active === ALL ? posts : posts.filter((p) => p.category === active)),
    [
      posts,
      active,
    ]
  )

  const [featured, ...rest] = filtered

  if (posts.length === 0) {
    return <p className="text-[var(--pz-ink-2)] opacity-50">No posts published yet.</p>
  }

  return (
    <div className="space-y-10">
      {categories.length > 2 && (
        <div className="flex flex-wrap gap-2">
          {categories.map((cat) => {
            const isActive = cat === active
            return (
              <button
                key={cat}
                type="button"
                onClick={() => setActive(cat)}
                className={`rounded-full px-4 py-1.5 text-sm font-medium transition-colors ${
                  isActive
                    ? 'bg-[var(--pz-accent)] text-white'
                    : 'bg-[var(--pz-card-bg)] text-[var(--pz-ink-2)] ring-1 ring-[var(--pz-card-border)] hover:ring-[var(--pz-accent-line)]'
                }`}
              >
                {cat}
              </button>
            )
          })}
        </div>
      )}

      {filtered.length === 0 ? (
        <p className="text-[var(--pz-ink-2)] opacity-50">No posts in this category yet.</p>
      ) : (
        <div className="space-y-10">
          {featured && <BlogFeaturedCard post={featured} />}
          {rest.length > 0 && (
            <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
              {rest.map((post) => (
                <BlogCard key={post.id} post={post} />
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

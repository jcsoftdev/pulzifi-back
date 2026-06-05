import Link from 'next/link'
import type { BlogPostItem } from '../lib/blog-post'
import { formatBlogDate } from '../lib/blog-post'
import { categoryStyle } from '../lib/category-style'
import { BlogCardMedia } from './blog-card-media'

type Props = {
  post: BlogPostItem
}

export function BlogCard({ post }: Props) {
  const { title, slug, excerpt, heroImageUrl, publishedAt, category, author, readingTime } = post
  const style = categoryStyle(category)
  const date = formatBlogDate(publishedAt)

  return (
    <Link
      href={`/blog/${slug}`}
      className="group flex flex-col overflow-hidden rounded-2xl bg-[var(--pz-card-bg)] shadow-[var(--pz-card-shadow-rest)] ring-1 ring-[var(--pz-card-border)] transition-all duration-200 hover:-translate-y-0.5 hover:shadow-[var(--pz-card-shadow-hover)] hover:ring-[var(--pz-accent-line)]"
    >
      <BlogCardMedia
        title={title}
        heroImageUrl={heroImageUrl}
        category={category}
        className="h-44 w-full"
        sizes="(max-width: 640px) 100vw, (max-width: 1024px) 50vw, 33vw"
      />
      <div className="flex flex-1 flex-col p-6">
        {category && (
          <span
            className={`mb-3 inline-block self-start rounded-full px-3 py-1 text-xs font-medium ${style.pill}`}
          >
            {category}
          </span>
        )}
        <h2 className="mb-2 text-xl font-bold leading-snug text-[var(--pz-ink)] transition-colors group-hover:text-[var(--pz-accent)]">
          {title}
        </h2>
        {excerpt && (
          <p className="line-clamp-3 flex-1 text-sm leading-relaxed text-[var(--pz-ink-2)]">
            {excerpt}
          </p>
        )}
        <div className="mt-4 flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-[var(--pz-ink-2)] opacity-70">
          {author && <span className="font-medium">{author}</span>}
          {author && readingTime ? <span aria-hidden="true">·</span> : null}
          {readingTime ? <span>{readingTime} min read</span> : null}
          {date && (readingTime || author) ? <span aria-hidden="true">·</span> : null}
          {date && (
            <time dateTime={publishedAt} className="opacity-90">
              {date}
            </time>
          )}
        </div>
      </div>
    </Link>
  )
}

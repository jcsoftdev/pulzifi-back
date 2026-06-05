import Link from 'next/link'
import type { BlogPostItem } from '../lib/blog-post'
import { formatBlogDate } from '../lib/blog-post'
import { categoryStyle } from '../lib/category-style'
import { BlogCardMedia } from './blog-card-media'

type Props = {
  post: BlogPostItem
}

/** Wide hero card for the newest post in the active filter. */
export function BlogFeaturedCard({ post }: Props) {
  const { title, slug, excerpt, heroImageUrl, publishedAt, category, author, readingTime } = post
  const style = categoryStyle(category)
  const date = formatBlogDate(publishedAt)

  return (
    <Link
      href={`/blog/${slug}`}
      className="group grid overflow-hidden rounded-3xl bg-[var(--pz-card-bg)] shadow-[var(--pz-card-shadow-rest)] ring-1 ring-[var(--pz-card-border)] transition-all duration-200 hover:shadow-[var(--pz-card-shadow-hover)] hover:ring-[var(--pz-accent-line)] lg:grid-cols-2"
    >
      <BlogCardMedia
        title={title}
        heroImageUrl={heroImageUrl}
        category={category}
        className="h-56 w-full lg:h-full lg:min-h-[20rem]"
        sizes="(max-width: 1024px) 100vw, 50vw"
      />
      <div className="flex flex-col justify-center gap-4 p-8 lg:p-12">
        <div className="flex flex-wrap items-center gap-3">
          {category && (
            <span
              className={`inline-block rounded-full px-3 py-1 text-xs font-medium ${style.pill}`}
            >
              {category}
            </span>
          )}
          <span className="text-xs font-medium uppercase tracking-[0.2em] text-[var(--pz-accent)]">
            Latest
          </span>
        </div>
        <h2 className="pz-display text-2xl leading-tight text-[var(--pz-ink)] transition-colors group-hover:text-[var(--pz-accent)] sm:text-3xl lg:text-4xl">
          {title}
        </h2>
        {excerpt && (
          <p className="line-clamp-3 text-base leading-relaxed text-[var(--pz-ink-2)]">{excerpt}</p>
        )}
        <div className="flex flex-wrap items-center gap-x-2 gap-y-1 text-sm text-[var(--pz-ink-2)] opacity-70">
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
